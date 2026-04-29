package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/starfall-warsong/sws/internal/service"
	"github.com/starfall-warsong/sws/pkg/response"
)

type WreckHandler struct {
	svc *service.WreckService
}

func NewWreckHandler(svc *service.WreckService) *WreckHandler {
	return &WreckHandler{svc: svc}
}

func (h *WreckHandler) ListWrecks(c *gin.Context) {
	charID := getCharIDFromContext(c)
	var systemID int64
	if sid, err := strconv.ParseInt(c.Query("system_id"), 10, 64); err == nil && sid > 0 {
		systemID = sid
	} else {
		h.svc.DB().QueryRowContext(c.Request.Context(),
			`SELECT current_system_id FROM characters WHERE id=$1`, charID).Scan(&systemID)
	}

	wrecks, err := h.svc.ListWrecks(c.Request.Context(), systemID)
	if err != nil {
		response.InternalError(c, "获取残骸列表失败")
		return
	}
	response.OK(c, gin.H{"wrecks": wrecks, "count": len(wrecks), "system_id": systemID})
}

func (h *WreckHandler) LootWreck(c *gin.Context) {
	charID := getCharIDFromContext(c)
	wreckID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的残骸ID")
		return
	}

	items, err := h.svc.LootWreck(c.Request.Context(), charID, wreckID)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.OK(c, gin.H{"message": "拾取成功", "items": items})
}
