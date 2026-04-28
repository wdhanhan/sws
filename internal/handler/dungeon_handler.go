package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/starfall-warsong/sws/internal/service"
	"github.com/starfall-warsong/sws/pkg/response"
)

type DungeonHandler struct {
	svc *service.DungeonService
}

func NewDungeonHandler(svc *service.DungeonService) *DungeonHandler {
	return &DungeonHandler{svc: svc}
}

func (h *DungeonHandler) ListDungeons(c *gin.Context) {
	race, _ := strconv.Atoi(c.Query("race"))
	diff, _ := strconv.Atoi(c.Query("difficulty"))
	dungeons, err := h.svc.ListDungeons(c.Request.Context(), race, diff)
	if err != nil {
		response.InternalError(c, "获取远征列表失败")
		return
	}
	response.OK(c, gin.H{"dungeons": dungeons, "count": len(dungeons)})
}

func (h *DungeonHandler) Enter(c *gin.Context) {
	charID := getCharIDFromContext(c)
	var req struct {
		DungeonDefID int64 `json:"dungeon_def_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}
	info, err := h.svc.Enter(c.Request.Context(), charID, req.DungeonDefID)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.OK(c, info)
}

func (h *DungeonHandler) GetStatus(c *gin.Context) {
	charID := getCharIDFromContext(c)
	info, err := h.svc.GetStatus(c.Request.Context(), charID)
	if err != nil {
		response.NotFound(c, err.Error())
		return
	}
	response.OK(c, info)
}

func (h *DungeonHandler) FightWave(c *gin.Context) {
	charID := getCharIDFromContext(c)
	info, err := h.svc.FightWave(c.Request.Context(), charID)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.OK(c, info)
}

func (h *DungeonHandler) Leave(c *gin.Context) {
	charID := getCharIDFromContext(c)
	if err := h.svc.Leave(c.Request.Context(), charID); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.OK(c, gin.H{"message": "已离开远征"})
}
