package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/starfall-warsong/sws/internal/service"
	"github.com/starfall-warsong/sws/pkg/response"
)

type ShipHandler struct {
	svc *service.ShipService
}

func NewShipHandler(svc *service.ShipService) *ShipHandler {
	return &ShipHandler{svc: svc}
}

func (h *ShipHandler) GetShipDefs(c *gin.Context) {
	raceID, _ := strconv.Atoi(c.Query("race_id"))
	defs, err := h.svc.GetShipDefs(c.Request.Context(), raceID)
	if err != nil {
		response.InternalError(c, "获取舰船定义失败")
		return
	}
	response.OK(c, gin.H{"ships": defs, "count": len(defs)})
}

func (h *ShipHandler) BoardShip(c *gin.Context) {
	charID := getCharIDFromContext(c)
	var req struct {
		ShipDefID int64  `json:"ship_def_id" binding:"required"`
		Name      string `json:"name"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}
	if req.Name == "" {
		req.Name = "我的舰船"
	}
	ship, err := h.svc.BoardShip(c.Request.Context(), charID, req.ShipDefID, req.Name)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.Created(c, ship)
}

func (h *ShipHandler) GetMyShips(c *gin.Context) {
	charID := getCharIDFromContext(c)
	ships, err := h.svc.GetMyShips(c.Request.Context(), charID)
	if err != nil {
		response.InternalError(c, "获取舰船列表失败")
		return
	}
	response.OK(c, gin.H{"ships": ships, "count": len(ships)})
}

func (h *ShipHandler) FitModule(c *gin.Context) {
	charID := getCharIDFromContext(c)
	var req service.FitModuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}
	if err := h.svc.FitModule(c.Request.Context(), charID, &req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.OK(c, gin.H{"message": "模块已装配"})
}

func (h *ShipHandler) GetFitting(c *gin.Context) {
	shipID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的舰船ID")
		return
	}
	fittings, err := h.svc.GetFitting(c.Request.Context(), shipID)
	if err != nil {
		response.InternalError(c, "获取装配失败")
		return
	}
	response.OK(c, gin.H{"fittings": fittings})
}

func (h *ShipHandler) RemoveModule(c *gin.Context) {
	charID := getCharIDFromContext(c)
	var req struct {
		ShipID    int64  `json:"ship_id" binding:"required"`
		SlotType  string `json:"slot_type" binding:"required"`
		SlotIndex int    `json:"slot_index"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}
	if err := h.svc.RemoveModule(c.Request.Context(), charID, req.ShipID, req.SlotType, req.SlotIndex); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.OK(c, gin.H{"message": "模块已卸载"})
}
