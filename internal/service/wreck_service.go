package service

import (
	"context"
	"errors"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/starfall-warsong/sws/internal/repository"
)

var (
	ErrWreckNotFound = errors.New("残骸不存在")
	ErrWreckLooted   = errors.New("残骸已被拾取")
)

type WreckService struct {
	db      *sqlx.DB
	invRepo *repository.InventoryRepo
}

func NewWreckService(db *sqlx.DB, invRepo *repository.InventoryRepo) *WreckService {
	return &WreckService{db: db, invRepo: invRepo}
}

func (s *WreckService) DB() *sqlx.DB { return s.db }

type WreckInfo struct {
	ID        int64     `db:"id" json:"id"`
	SystemID  int64     `db:"system_id" json:"system_id"`
	OwnerName string    `db:"owner_name" json:"owner_name"`
	ShipName  string    `db:"ship_name" json:"ship_name"`
	ShipDefID int64     `db:"ship_def_id" json:"ship_def_id"`
	IsLooted  bool      `db:"is_looted" json:"is_looted"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	ExpiresAt time.Time `db:"expires_at" json:"expires_at"`
	ItemCount int       `json:"item_count"`
	ShipType  string    `json:"ship_type"`
}

type WreckItem struct {
	ItemDefID int64  `db:"item_def_id" json:"item_def_id"`
	Name      string `db:"name" json:"name"`
	Quantity  int    `db:"quantity" json:"quantity"`
}

func (s *WreckService) ListWrecks(ctx context.Context, systemID int64) ([]WreckInfo, error) {
	// Clean up expired wrecks
	s.db.ExecContext(ctx, `DELETE FROM wrecks WHERE expires_at < NOW()`)

	var wrecks []WreckInfo
	err := s.db.SelectContext(ctx, &wrecks,
		`SELECT id, system_id, owner_name, ship_name, COALESCE(ship_def_id,0) as ship_def_id, is_looted, created_at, expires_at
		 FROM wrecks WHERE system_id=$1 AND is_looted=false AND expires_at > NOW()
		 ORDER BY created_at DESC`, systemID)
	if err != nil {
		return nil, err
	}

	for i := range wrecks {
		var cnt int
		s.db.GetContext(ctx, &cnt, `SELECT COUNT(*) FROM wreck_items WHERE wreck_id=$1`, wrecks[i].ID)
		wrecks[i].ItemCount = cnt
		if wrecks[i].ShipDefID > 0 {
			var sn string
			s.db.GetContext(ctx, &sn, `SELECT name FROM ship_defs WHERE id=$1`, wrecks[i].ShipDefID)
			wrecks[i].ShipType = sn
		}
	}

	return wrecks, nil
}

func (s *WreckService) GetWreckItems(ctx context.Context, wreckID int64) ([]WreckItem, error) {
	var items []WreckItem
	err := s.db.SelectContext(ctx, &items,
		`SELECT wi.item_def_id, i.name, wi.quantity
		 FROM wreck_items wi JOIN item_defs i ON i.id = wi.item_def_id
		 WHERE wi.wreck_id=$1`, wreckID)
	return items, err
}

func (s *WreckService) LootWreck(ctx context.Context, charID, wreckID int64) ([]WreckItem, error) {
	var wreck struct {
		ID       int64 `db:"id"`
		SystemID int64 `db:"system_id"`
		IsLooted bool  `db:"is_looted"`
	}
	if err := s.db.GetContext(ctx, &wreck, `SELECT id, system_id, is_looted FROM wrecks WHERE id=$1`, wreckID); err != nil {
		return nil, ErrWreckNotFound
	}
	if wreck.IsLooted {
		return nil, ErrWreckLooted
	}

	items, err := s.GetWreckItems(ctx, wreckID)
	if err != nil {
		return nil, err
	}

	// Transfer items to character inventory
	for _, item := range items {
		s.invRepo.AddOrUpsertItem(ctx, "character", charID, item.ItemDefID, int64(item.Quantity), wreck.SystemID)
	}

	// Mark wreck as looted
	s.db.ExecContext(ctx, `UPDATE wrecks SET is_looted=true WHERE id=$1`, wreckID)

	return items, nil
}
