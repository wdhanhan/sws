package handler

import (
	"errors"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/starfall-warsong/sws/internal/middleware"
	"github.com/starfall-warsong/sws/internal/repository"
	"github.com/starfall-warsong/sws/internal/service"
	"github.com/starfall-warsong/sws/pkg/response"
)

type EconomyHandler struct {
	miningSvc  *service.MiningService
	refineSvc  *service.RefineService
	marketSvc  *service.MarketService
	invRepo    *repository.InventoryRepo
}

func NewEconomyHandler(
	miningSvc *service.MiningService,
	refineSvc *service.RefineService,
	marketSvc *service.MarketService,
	invRepo *repository.InventoryRepo,
) *EconomyHandler {
	return &EconomyHandler{
		miningSvc: miningSvc,
		refineSvc: refineSvc,
		marketSvc: marketSvc,
		invRepo:   invRepo,
	}
}

// ========== 采矿 ==========

func (h *EconomyHandler) StartMining(c *gin.Context) {
	charID := getCharIDFromContext(c)
	var req service.MiningStartRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}
	result, err := h.miningSvc.StartMining(c.Request.Context(), charID, &req)
	if err != nil {
		handleEconomyError(c, err)
		return
	}
	response.OK(c, result)
}

func (h *EconomyHandler) CollectMining(c *gin.Context) {
	charID := getCharIDFromContext(c)
	result, err := h.miningSvc.CollectMining(c.Request.Context(), charID)
	if err != nil {
		handleEconomyError(c, err)
		return
	}
	response.OK(c, result)
}

func (h *EconomyHandler) StopMining(c *gin.Context) {
	charID := getCharIDFromContext(c)
	if err := h.miningSvc.StopMining(c.Request.Context(), charID); err != nil {
		handleEconomyError(c, err)
		return
	}
	response.OK(c, gin.H{"message": "采矿已停止"})
}

// ========== 精炼 ==========

func (h *EconomyHandler) Refine(c *gin.Context) {
	charID := getCharIDFromContext(c)
	var req service.RefineRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}
	result, err := h.refineSvc.Refine(c.Request.Context(), charID, &req)
	if err != nil {
		handleEconomyError(c, err)
		return
	}
	response.OK(c, result)
}

func (h *EconomyHandler) GetRefineRecipes(c *gin.Context) {
	recipes, err := h.refineSvc.GetRecipes(c.Request.Context())
	if err != nil {
		response.InternalError(c, "获取配方失败")
		return
	}
	response.OK(c, gin.H{"recipes": recipes})
}

// ========== 背包 ==========

func (h *EconomyHandler) GetInventory(c *gin.Context) {
	charID := getCharIDFromContext(c)
	var systemID int64
	// Use character's current system
	if sid, err := strconv.ParseInt(c.Query("system_id"), 10, 64); err == nil {
		systemID = sid
	}

	items, err := h.invRepo.ListByOwner(c.Request.Context(), "character", charID, systemID)
	if err != nil {
		response.InternalError(c, "获取背包失败")
		return
	}

	type ItemWithName struct {
		ID        int64  `json:"id"`
		ItemDefID int64  `json:"item_def_id"`
		Name      string `json:"name"`
		Quantity  int64  `json:"quantity"`
	}

	var enriched []ItemWithName
	for _, item := range items {
		def, _ := h.invRepo.GetItemDef(c.Request.Context(), item.ItemDefID)
		name := "未知物品"
		if def != nil {
			name = def.Name
		}
		enriched = append(enriched, ItemWithName{
			ID: item.ID, ItemDefID: item.ItemDefID, Name: name, Quantity: item.Quantity,
		})
	}

	response.OK(c, gin.H{"items": enriched, "count": len(enriched)})
}

func (h *EconomyHandler) NPCShop(c *gin.Context) {
	charID := getCharIDFromContext(c)
	var req struct {
		ItemDefID int64 `json:"item_def_id" binding:"required"`
		Quantity  int64 `json:"quantity" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}

	def, err := h.invRepo.GetItemDef(c.Request.Context(), req.ItemDefID)
	if err != nil || def == nil {
		response.NotFound(c, "物品不存在")
		return
	}

	totalCost := def.BasePrice * req.Quantity * 2 // NPC售价=保底价×2

	// 扣钱
	result, _ := h.invRepo.DB().ExecContext(c.Request.Context(),
		`UPDATE characters SET balance = balance - $1 WHERE id = $2 AND balance >= $1`,
		totalCost, charID)
	rows, _ := result.RowsAffected()
	if rows == 0 {
		response.Forbidden(c, "星币不足")
		return
	}

	var systemID int64
	h.invRepo.DB().QueryRowContext(c.Request.Context(),
		`SELECT current_system_id FROM characters WHERE id = $1`, charID).Scan(&systemID)

	h.invRepo.AddOrUpsertItem(c.Request.Context(), "character", charID, req.ItemDefID, req.Quantity, systemID)

	response.OK(c, gin.H{
		"message":  "购买成功",
		"item":     def.Name,
		"quantity": req.Quantity,
		"cost":     totalCost,
	})
}

// TransferItem 舰船↔空间站转移物品
func (h *EconomyHandler) TransferItem(c *gin.Context) {
	charID := getCharIDFromContext(c)
	var req struct {
		ItemDefID int64  `json:"item_def_id" binding:"required"`
		Quantity  int64  `json:"quantity" binding:"required"`
		Direction string `json:"direction" binding:"required"` // to_station / to_ship
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}

	var systemID int64
	h.invRepo.DB().QueryRowContext(c.Request.Context(),
		`SELECT current_system_id FROM characters WHERE id=$1`, charID).Scan(&systemID)

	if req.Direction == "to_station" {
		// 从舰船移到空间站
		ok, _ := h.invRepo.RemoveItemByType(c.Request.Context(), "character", charID, req.ItemDefID, req.Quantity, systemID, "ship")
		if !ok {
			response.Forbidden(c, "舰船货仓中没有足够的物品")
			return
		}
		h.invRepo.AddItemWithType(c.Request.Context(), "character", charID, req.ItemDefID, req.Quantity, systemID, "station")
		response.OK(c, gin.H{"message": "已转移到空间站仓库"})
	} else {
		ok, _ := h.invRepo.RemoveItemByType(c.Request.Context(), "character", charID, req.ItemDefID, req.Quantity, systemID, "station")
		if !ok {
			response.Forbidden(c, "空间站仓库中没有足够的物品")
			return
		}
		h.invRepo.AddItemWithType(c.Request.Context(), "character", charID, req.ItemDefID, req.Quantity, systemID, "ship")
		response.OK(c, gin.H{"message": "已转移到舰船货仓"})
	}
}

// GetAssets 资产总览（所有位置的物品）
func (h *EconomyHandler) GetAssets(c *gin.Context) {
	charID := getCharIDFromContext(c)

	type Asset struct {
		ItemDefID  int64  `db:"item_def_id" json:"item_def_id"`
		Name       string `json:"name"`
		Quantity   int64  `db:"quantity" json:"quantity"`
		LocationID int64  `db:"location_id" json:"location_id"`
		LocType    string `db:"location_type" json:"location_type"`
	}
	var assets []Asset
	h.invRepo.DB().SelectContext(c.Request.Context(), &assets,
		`SELECT item_def_id, quantity, location_id, location_type FROM inventory
		 WHERE owner_type='character' AND owner_id=$1 ORDER BY location_id, location_type, item_def_id`, charID)

	for i := range assets {
		def, _ := h.invRepo.GetItemDef(c.Request.Context(), assets[i].ItemDefID)
		if def != nil {
			assets[i].Name = def.Name
		}
	}

	// 按位置分组
	grouped := map[string][]Asset{}
	for _, a := range assets {
		key := ""
		if a.LocType == "station" {
			key = "空间站#" + strconv.FormatInt(a.LocationID, 10)
		} else {
			key = "舰船(星系#" + strconv.FormatInt(a.LocationID, 10) + ")"
		}
		grouped[key] = append(grouped[key], a)
	}

	response.OK(c, gin.H{"assets": grouped, "total_items": len(assets)})
}

func (h *EconomyHandler) GetItemDefs(c *gin.Context) {
	category := c.Query("category")
	slotType := c.Query("slot_type")
	items, err := h.invRepo.ListItemDefs(c.Request.Context(), category, slotType)
	if err != nil {
		response.InternalError(c, "获取物品定义失败")
		return
	}
	response.OK(c, gin.H{"items": items, "count": len(items)})
}

// ========== 市场 ==========

func (h *EconomyHandler) CreateSellOrder(c *gin.Context) {
	charID := getCharIDFromContext(c)
	var req service.SellOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}
	order, err := h.marketSvc.CreateSellOrder(c.Request.Context(), charID, &req)
	if err != nil {
		handleEconomyError(c, err)
		return
	}
	response.Created(c, order)
}

func (h *EconomyHandler) CreateBuyOrder(c *gin.Context) {
	charID := getCharIDFromContext(c)
	var req service.BuyOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}
	order, err := h.marketSvc.CreateBuyOrder(c.Request.Context(), charID, &req)
	if err != nil {
		handleEconomyError(c, err)
		return
	}
	response.Created(c, order)
}

func (h *EconomyHandler) BuyFromOrder(c *gin.Context) {
	charID := getCharIDFromContext(c)
	orderID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的订单ID")
		return
	}
	var req struct {
		Quantity int64 `json:"quantity" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}
	if err := h.marketSvc.FulfillSellOrder(c.Request.Context(), charID, orderID, req.Quantity); err != nil {
		handleEconomyError(c, err)
		return
	}
	response.OK(c, gin.H{"message": "购买成功"})
}

func (h *EconomyHandler) SearchOrders(c *gin.Context) {
	itemID, _ := strconv.ParseInt(c.Query("item_id"), 10, 64)
	systemID, _ := strconv.ParseInt(c.Query("system_id"), 10, 64)
	orderType := c.Query("type")
	limit := 50
	if l, err := strconv.Atoi(c.Query("limit")); err == nil && l > 0 {
		limit = l
	}
	orders, err := h.marketSvc.SearchOrders(c.Request.Context(), itemID, systemID, orderType, limit)
	if err != nil {
		response.InternalError(c, "搜索订单失败")
		return
	}
	response.OK(c, gin.H{"orders": orders, "count": len(orders)})
}

func (h *EconomyHandler) CancelOrder(c *gin.Context) {
	charID := getCharIDFromContext(c)
	orderID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的订单ID")
		return
	}
	if err := h.marketSvc.CancelOrder(c.Request.Context(), charID, orderID); err != nil {
		handleEconomyError(c, err)
		return
	}
	response.OK(c, gin.H{"message": "订单已取消"})
}

// ========== helpers ==========

func getCharIDFromContext(c *gin.Context) int64 {
	// For now use account_id as char_id proxy; later add char selection
	// In real impl, client sends active character ID in header
	if cid, err := strconv.ParseInt(c.GetHeader("X-Character-ID"), 10, 64); err == nil {
		return cid
	}
	return middleware.GetAccountID(c)
}

func handleEconomyError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrBeltNotFound), errors.Is(err, service.ErrOrderNotFound):
		response.NotFound(c, err.Error())
	case errors.Is(err, service.ErrBeltDepleted), errors.Is(err, service.ErrAlreadyMining),
		errors.Is(err, service.ErrNotMining), errors.Is(err, service.ErrCannotBuyOwnOrder),
		errors.Is(err, service.ErrOrderNotActive):
		response.BadRequest(c, err.Error())
	case errors.Is(err, service.ErrInsufficientFunds), errors.Is(err, service.ErrInsufficientItems),
		errors.Is(err, service.ErrInsufficientOre), errors.Is(err, service.ErrNoRecipe):
		response.Forbidden(c, err.Error())
	default:
		response.InternalError(c, "操作失败")
	}
}
