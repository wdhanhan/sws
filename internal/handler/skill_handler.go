package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/starfall-warsong/sws/internal/service"
	"github.com/starfall-warsong/sws/pkg/response"
)

type SkillHandler struct {
	skillSvc *service.SkillService
	deathSvc *service.DeathService
}

func NewSkillHandler(skillSvc *service.SkillService, deathSvc *service.DeathService) *SkillHandler {
	return &SkillHandler{skillSvc: skillSvc, deathSvc: deathSvc}
}

func (h *SkillHandler) GetSkillDefs(c *gin.Context) {
	category := c.Query("category")
	skills, err := h.skillSvc.GetSkillDefs(c.Request.Context(), category)
	if err != nil {
		response.InternalError(c, "获取技能定义失败")
		return
	}
	response.OK(c, gin.H{"skills": skills, "count": len(skills)})
}

func (h *SkillHandler) GetCharacterSkills(c *gin.Context) {
	charID := getCharIDFromContext(c)
	// First process any completed training
	completed, _ := h.skillSvc.ProcessCompleted(c.Request.Context(), charID)
	skills, err := h.skillSvc.GetCharacterSkills(c.Request.Context(), charID)
	if err != nil {
		response.InternalError(c, "获取角色技能失败")
		return
	}
	response.OK(c, gin.H{"skills": skills, "completed": completed})
}

func (h *SkillHandler) AddToQueue(c *gin.Context) {
	charID := getCharIDFromContext(c)
	var req service.TrainRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}
	item, err := h.skillSvc.AddToQueue(c.Request.Context(), charID, &req)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.OK(c, item)
}

func (h *SkillHandler) GetQueue(c *gin.Context) {
	charID := getCharIDFromContext(c)
	h.skillSvc.ProcessCompleted(c.Request.Context(), charID)
	queue, err := h.skillSvc.GetQueue(c.Request.Context(), charID)
	if err != nil {
		response.InternalError(c, "获取训练队列失败")
		return
	}
	response.OK(c, gin.H{"queue": queue, "count": len(queue)})
}

func (h *SkillHandler) ProcessDeath(c *gin.Context) {
	charID := getCharIDFromContext(c)
	var req struct {
		KilledBy string `json:"killed_by"`
	}
	c.ShouldBindJSON(&req)
	if req.KilledBy == "" {
		req.KilledBy = "unknown"
	}
	result, err := h.deathSvc.ProcessDeath(c.Request.Context(), charID, req.KilledBy)
	if err != nil {
		response.InternalError(c, "死亡处理失败")
		return
	}
	response.OK(c, result)
}
