package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/starfall-warsong/sws/internal/service"
	"github.com/starfall-warsong/sws/pkg/response"
)

type CombatSiteHandler struct {
	svc *service.CombatSiteService
}

func NewCombatSiteHandler(svc *service.CombatSiteService) *CombatSiteHandler {
	return &CombatSiteHandler{svc: svc}
}

func (h *CombatSiteHandler) ListSites(c *gin.Context) {
	systemID, _ := strconv.ParseInt(c.Query("system_id"), 10, 64)
	if systemID == 0 {
		charID := getCharIDFromContext(c)
		h.svc.DB().QueryRowContext(c.Request.Context(),
			`SELECT current_system_id FROM characters WHERE id=$1`, charID).Scan(&systemID)
	}
	showAll := c.Query("show_all") == "true"
	sites, err := h.svc.ListSites(c.Request.Context(), systemID, showAll)
	if err != nil {
		response.InternalError(c, "获取地点列表失败")
		return
	}
	response.OK(c, gin.H{"sites": sites, "count": len(sites), "system_id": systemID})
}

func (h *CombatSiteHandler) Scan(c *gin.Context) {
	charID := getCharIDFromContext(c)
	var systemID int64
	h.svc.DB().QueryRowContext(c.Request.Context(),
		`SELECT current_system_id FROM characters WHERE id=$1`, charID).Scan(&systemID)
	sites, err := h.svc.ScanSystem(c.Request.Context(), systemID)
	if err != nil {
		response.InternalError(c, "扫描失败")
		return
	}
	response.OK(c, gin.H{"sites": sites, "count": len(sites), "system_id": systemID, "message": "扫描完成"})
}

func (h *CombatSiteHandler) Enter(c *gin.Context) {
	charID := getCharIDFromContext(c)
	var req struct {
		SiteID int64 `json:"site_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}
	result, err := h.svc.EnterSite(c.Request.Context(), charID, req.SiteID)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.OK(c, result)
}

func (h *CombatSiteHandler) Tick(c *gin.Context) {
	charID := getCharIDFromContext(c)
	result, err := h.svc.SiteNextTick(c.Request.Context(), charID)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.OK(c, result)
}

func (h *CombatSiteHandler) AutoWave(c *gin.Context) {
	charID := getCharIDFromContext(c)
	result, err := h.svc.SiteAutoFight(c.Request.Context(), charID)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.OK(c, result)
}

func (h *CombatSiteHandler) Leave(c *gin.Context) {
	charID := getCharIDFromContext(c)
	h.svc.LeaveSite(c.Request.Context(), charID)
	response.OK(c, gin.H{"message": "已离开地点"})
}
