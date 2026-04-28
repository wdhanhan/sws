package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/starfall-warsong/sws/internal/repository"
)

var (
	ErrInsufficientFunds = errors.New("星币不足")
	ErrInsufficientItems = errors.New("物品数量不足")
	ErrOrderNotFound     = errors.New("订单不存在")
	ErrCannotBuyOwnOrder = errors.New("不能购买自己的订单")
	ErrOrderNotActive    = errors.New("订单不在有效状态")
)

type MarketService struct {
	db      *sqlx.DB
	invRepo *repository.InventoryRepo
}

func NewMarketService(db *sqlx.DB, invRepo *repository.InventoryRepo) *MarketService {
	return &MarketService{db: db, invRepo: invRepo}
}

type SellOrderRequest struct {
	ItemDefID int64 `json:"item_def_id" binding:"required"`
	Quantity  int64 `json:"quantity" binding:"required"`
	Price     int64 `json:"price" binding:"required"` // 单价(分)
}

type BuyOrderRequest struct {
	ItemDefID int64 `json:"item_def_id" binding:"required"`
	Quantity  int64 `json:"quantity" binding:"required"`
	Price     int64 `json:"price" binding:"required"`
}

type OrderInfo struct {
	ID           int64  `db:"id" json:"id"`
	CharacterID  int64  `db:"character_id" json:"character_id"`
	ItemDefID    int64  `db:"item_def_id" json:"item_def_id"`
	ItemName     string `json:"item_name"`
	OrderType    string `db:"order_type" json:"order_type"`
	Price        int64  `db:"price" json:"price"`
	Quantity     int64  `db:"quantity" json:"quantity"`
	QuantityLeft int64  `db:"quantity_filled" json:"quantity_filled"`
	StationID    int64  `db:"station_id" json:"station_id"`
	SystemID     int64  `db:"system_id" json:"system_id"`
	Status       string `db:"status" json:"status"`
}

func (s *MarketService) CreateSellOrder(ctx context.Context, charID int64, req *SellOrderRequest) (*OrderInfo, error) {
	var systemID int64
	s.db.GetContext(ctx, &systemID, `SELECT current_system_id FROM characters WHERE id = $1`, charID)

	// Remove items from inventory
	ok, err := s.invRepo.RemoveItem(ctx, "character", charID, req.ItemDefID, req.Quantity, systemID)
	if err != nil || !ok {
		return nil, ErrInsufficientItems
	}

	// Create sell order
	var order OrderInfo
	err = s.db.QueryRowxContext(ctx,
		`INSERT INTO market_orders (character_id, item_def_id, order_type, price, quantity, station_id, system_id, expires_at)
		 VALUES ($1, $2, 'sell', $3, $4, $5, $5, $6)
		 RETURNING *`,
		charID, req.ItemDefID, req.Price, req.Quantity, systemID, time.Now().Add(30*24*time.Hour),
	).StructScan(&order)
	if err != nil {
		// Rollback: return items
		s.invRepo.AddOrUpsertItem(ctx, "character", charID, req.ItemDefID, req.Quantity, systemID)
		return nil, err
	}

	return &order, nil
}

func (s *MarketService) CreateBuyOrder(ctx context.Context, charID int64, req *BuyOrderRequest) (*OrderInfo, error) {
	totalCost := req.Price * req.Quantity

	// Deduct money
	result, err := s.db.ExecContext(ctx,
		`UPDATE characters SET balance = balance - $1 WHERE id = $2 AND balance >= $1`,
		totalCost, charID)
	if err != nil {
		return nil, err
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return nil, ErrInsufficientFunds
	}

	var systemID int64
	s.db.GetContext(ctx, &systemID, `SELECT current_system_id FROM characters WHERE id = $1`, charID)

	var order OrderInfo
	err = s.db.QueryRowxContext(ctx,
		`INSERT INTO market_orders (character_id, item_def_id, order_type, price, quantity, station_id, system_id, expires_at)
		 VALUES ($1, $2, 'buy', $3, $4, $5, $5, $6)
		 RETURNING *`,
		charID, req.ItemDefID, req.Price, req.Quantity, systemID, time.Now().Add(30*24*time.Hour),
	).StructScan(&order)
	if err != nil {
		// Rollback money
		s.db.ExecContext(ctx, `UPDATE characters SET balance = balance + $1 WHERE id = $2`, totalCost, charID)
		return nil, err
	}

	return &order, nil
}

func (s *MarketService) FulfillSellOrder(ctx context.Context, buyerCharID, orderID int64, quantity int64) error {
	var order OrderInfo
	err := s.db.GetContext(ctx, &order, `SELECT * FROM market_orders WHERE id = $1 AND status = 'active'`, orderID)
	if err != nil {
		return ErrOrderNotFound
	}
	if order.OrderType != "sell" {
		return ErrOrderNotActive
	}
	if order.CharacterID == buyerCharID {
		return ErrCannotBuyOwnOrder
	}

	remaining := order.Quantity - order.QuantityLeft
	if quantity > remaining {
		quantity = remaining
	}

	totalCost := order.Price * quantity

	// Deduct buyer's money
	result, err := s.db.ExecContext(ctx,
		`UPDATE characters SET balance = balance - $1 WHERE id = $2 AND balance >= $1`,
		totalCost, buyerCharID)
	if err != nil {
		return err
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return ErrInsufficientFunds
	}

	// Add items to buyer
	s.invRepo.AddOrUpsertItem(ctx, "character", buyerCharID, order.ItemDefID, quantity, order.SystemID)

	// Pay seller
	s.db.ExecContext(ctx, `UPDATE characters SET balance = balance + $1 WHERE id = $2`,
		totalCost*95/100, order.CharacterID) // 5% market tax

	// Update order
	newFilled := order.QuantityLeft + quantity
	if newFilled >= order.Quantity {
		s.db.ExecContext(ctx, `UPDATE market_orders SET quantity_filled = $1, status = 'filled' WHERE id = $2`, newFilled, orderID)
	} else {
		s.db.ExecContext(ctx, `UPDATE market_orders SET quantity_filled = $1 WHERE id = $2`, newFilled, orderID)
	}

	// Record transaction
	s.db.ExecContext(ctx,
		`INSERT INTO market_transactions (buyer_id, seller_id, item_def_id, quantity, price, total, station_id, system_id)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		buyerCharID, order.CharacterID, order.ItemDefID, quantity, order.Price, totalCost, order.StationID, order.SystemID)

	return nil
}

func (s *MarketService) SearchOrders(ctx context.Context, itemDefID int64, systemID int64, orderType string, limit int) ([]OrderInfo, error) {
	query := `SELECT * FROM market_orders WHERE status = 'active'`
	args := []interface{}{}
	idx := 1

	if itemDefID > 0 {
		query += fmt.Sprintf(` AND item_def_id = $%d`, idx)
		args = append(args, itemDefID)
		idx++
	}
	if systemID > 0 {
		query += fmt.Sprintf(` AND system_id = $%d`, idx)
		args = append(args, systemID)
		idx++
	}
	if orderType != "" {
		query += fmt.Sprintf(` AND order_type = $%d`, idx)
		args = append(args, orderType)
		idx++
	}

	if orderType == "sell" {
		query += ` ORDER BY price ASC`
	} else {
		query += ` ORDER BY price DESC`
	}
	query += fmt.Sprintf(` LIMIT $%d`, idx)
	args = append(args, limit)

	var orders []OrderInfo
	err := s.db.SelectContext(ctx, &orders, query, args...)
	if err != nil {
		return nil, err
	}
	for i := range orders {
		var name string
		s.db.QueryRowContext(ctx, `SELECT name FROM item_defs WHERE id = $1`, orders[i].ItemDefID).Scan(&name)
		orders[i].ItemName = name
	}
	return orders, nil
}

func (s *MarketService) CancelOrder(ctx context.Context, charID, orderID int64) error {
	var order OrderInfo
	err := s.db.GetContext(ctx, &order, `SELECT * FROM market_orders WHERE id = $1 AND character_id = $2 AND status = 'active'`, orderID, charID)
	if err != nil {
		return ErrOrderNotFound
	}

	remaining := order.Quantity - order.QuantityLeft

	if order.OrderType == "sell" {
		// Return unsold items
		var systemID int64
		s.db.GetContext(ctx, &systemID, `SELECT current_system_id FROM characters WHERE id = $1`, charID)
		s.invRepo.AddOrUpsertItem(ctx, "character", charID, order.ItemDefID, remaining, systemID)
	} else {
		// Return unspent money
		refund := order.Price * remaining
		s.db.ExecContext(ctx, `UPDATE characters SET balance = balance + $1 WHERE id = $2`, refund, charID)
	}

	s.db.ExecContext(ctx, `UPDATE market_orders SET status = 'cancelled' WHERE id = $1`, orderID)
	return nil
}
