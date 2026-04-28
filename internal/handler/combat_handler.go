package handler

import (
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/starfall-warsong/sws/internal/service"
	"github.com/starfall-warsong/sws/pkg/response"
)

type CombatHandler struct {
	svc *service.CombatService
}

func NewCombatHandler(svc *service.CombatService) *CombatHandler {
	return &CombatHandler{svc: svc}
}

func (h *CombatHandler) ScanEnemies(c *gin.Context) {
	type NPCBrief struct {
		ID     int64  `db:"id" json:"id"`
		Name   string `db:"name" json:"name"`
		Type   string `db:"npc_type" json:"npc_type"`
		Tier   int    `db:"tier" json:"tier"`
		Bounty int64  `db:"bounty" json:"bounty"`
	}
	var npcs []NPCBrief
	// Return all available NPCs for now (later: filter by system)
	h.svc.DB().SelectContext(c.Request.Context(), &npcs,
		`SELECT id, name, npc_type, tier, bounty FROM npc_defs ORDER BY tier, id`)
	response.OK(c, gin.H{"enemies": npcs, "count": len(npcs)})
}

func (h *CombatHandler) EngageNPC(c *gin.Context) {
	charID := getCharIDFromContext(c)
	var req service.EngageNPCRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}
	state, err := h.svc.EngageNPC(c.Request.Context(), charID, &req)
	if err != nil {
		handleCombatError(c, err)
		return
	}
	response.OK(c, state)
}

func (h *CombatHandler) NextTick(c *gin.Context) {
	charID := getCharIDFromContext(c)
	state, err := h.svc.NextTick(c.Request.Context(), charID)
	if err != nil {
		handleCombatError(c, err)
		return
	}
	response.OK(c, state)
}

func (h *CombatHandler) GetState(c *gin.Context) {
	charID := getCharIDFromContext(c)
	state, err := h.svc.GetCombatState(c.Request.Context(), charID)
	if err != nil {
		handleCombatError(c, err)
		return
	}
	response.OK(c, state)
}

func (h *CombatHandler) AutoFight(c *gin.Context) {
	charID := getCharIDFromContext(c)
	var allLogs []string
	for i := 0; i < 100; i++ {
		state, err := h.svc.NextTick(c.Request.Context(), charID)
		if err != nil {
			break
		}
		allLogs = append(allLogs, state.Logs...)
		if state.Status != "active" {
			response.OK(c, gin.H{
				"result": state.Status,
				"ticks":  state.Tick,
				"logs":   allLogs,
				"participants": state.Participants,
			})
			return
		}
	}
	response.OK(c, gin.H{"result": "timeout", "message": "战斗未在100Tick内结束"})
}

func (h *CombatHandler) Command(c *gin.Context) {
	charID := getCharIDFromContext(c)
	var cmd service.CombatCommand
	if err := c.ShouldBindJSON(&cmd); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}
	state, err := h.svc.IssueCommand(c.Request.Context(), charID, &cmd)
	if err != nil {
		handleCombatError(c, err)
		return
	}
	response.OK(c, state)
}

func handleCombatError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrNoCombat):
		response.NotFound(c, err.Error())
	case errors.Is(err, service.ErrAlreadyInCombat):
		response.Conflict(c, err.Error())
	case errors.Is(err, service.ErrNPCNotFound):
		response.NotFound(c, err.Error())
	case errors.Is(err, service.ErrCombatFinished):
		response.BadRequest(c, err.Error())
	default:
		response.InternalError(c, "战斗系统错误")
	}
}
