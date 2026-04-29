package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/starfall-warsong/sws/internal/service"
	"github.com/starfall-warsong/sws/pkg/response"
)

type FleetHandler struct {
	svc *service.FleetService
}

func NewFleetHandler(svc *service.FleetService) *FleetHandler {
	return &FleetHandler{svc: svc}
}

func (h *FleetHandler) Create(c *gin.Context) {
	charID := getCharIDFromContext(c)
	fleet, err := h.svc.Create(c.Request.Context(), charID)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.Created(c, fleet)
}

func (h *FleetHandler) Invite(c *gin.Context) {
	charID := getCharIDFromContext(c)
	var req struct {
		CharID int64 `json:"char_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}
	if err := h.svc.Invite(c.Request.Context(), charID, req.CharID); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.OK(c, gin.H{"message": "已发送邀请"})
}

func (h *FleetHandler) Accept(c *gin.Context) {
	charID := getCharIDFromContext(c)
	var req struct {
		FleetID int64 `json:"fleet_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}
	if err := h.svc.Accept(c.Request.Context(), charID, req.FleetID); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.OK(c, gin.H{"message": "已加入舰队"})
}

func (h *FleetHandler) Decline(c *gin.Context) {
	charID := getCharIDFromContext(c)
	var req struct {
		FleetID int64 `json:"fleet_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}
	h.svc.Decline(c.Request.Context(), charID, req.FleetID)
	response.OK(c, gin.H{"message": "已拒绝邀请"})
}

func (h *FleetHandler) Leave(c *gin.Context) {
	charID := getCharIDFromContext(c)
	if err := h.svc.Leave(c.Request.Context(), charID); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.OK(c, gin.H{"message": "已离开舰队"})
}

func (h *FleetHandler) Kick(c *gin.Context) {
	charID := getCharIDFromContext(c)
	var req struct {
		CharID int64 `json:"char_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}
	if err := h.svc.Kick(c.Request.Context(), charID, req.CharID); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.OK(c, gin.H{"message": "已踢出成员"})
}

func (h *FleetHandler) Disband(c *gin.Context) {
	charID := getCharIDFromContext(c)
	if err := h.svc.Disband(c.Request.Context(), charID); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.OK(c, gin.H{"message": "舰队已解散"})
}

func (h *FleetHandler) GetFleet(c *gin.Context) {
	charID := getCharIDFromContext(c)
	fleet, err := h.svc.GetFleet(c.Request.Context(), charID)
	invites, _ := h.svc.GetPendingInvites(c.Request.Context(), charID)
	myChars := h.svc.GetSameAccountChars(c.Request.Context(), charID)
	if err != nil {
		response.OK(c, gin.H{"fleet": nil, "invites": invites, "my_chars": myChars})
		return
	}
	response.OK(c, gin.H{"fleet": fleet, "invites": invites, "my_chars": myChars})
}
