package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/starfall-warsong/sws/internal/service"
	"github.com/starfall-warsong/sws/pkg/response"
)

type WorldHandler struct {
	buildingSvc   *service.BuildingService
	encounterSvc  *service.EncounterService
}

func NewWorldHandler(buildingSvc *service.BuildingService, encounterSvc *service.EncounterService) *WorldHandler {
	return &WorldHandler{buildingSvc: buildingSvc, encounterSvc: encounterSvc}
}

// ========== 建筑 ==========

func (h *WorldHandler) GetBuildingDefs(c *gin.Context) {
	category := c.Query("category")
	defs, err := h.buildingSvc.GetBuildingDefs(c.Request.Context(), category)
	if err != nil {
		response.InternalError(c, "获取建筑定义失败")
		return
	}
	response.OK(c, gin.H{"buildings": defs, "count": len(defs)})
}

func (h *WorldHandler) DeployBuilding(c *gin.Context) {
	charID := getCharIDFromContext(c)
	var req service.DeployRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}

	var charName string
	// simplified: use charID as owner
	building, err := h.buildingSvc.Deploy(c.Request.Context(), "character", charID, charName, &req)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.Created(c, building)
}

func (h *WorldHandler) GetSystemBuildings(c *gin.Context) {
	systemID, err := strconv.ParseInt(c.Param("system_id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的星系ID")
		return
	}
	buildings, err := h.buildingSvc.GetSystemBuildings(c.Request.Context(), systemID)
	if err != nil {
		response.InternalError(c, "获取建筑列表失败")
		return
	}
	response.OK(c, gin.H{"buildings": buildings, "count": len(buildings)})
}

// ========== 奇遇 ==========

func (h *WorldHandler) TryEncounter(c *gin.Context) {
	charID := getCharIDFromContext(c)
	event, err := h.encounterSvc.TryTrigger(c.Request.Context(), charID)
	if err != nil {
		response.OK(c, gin.H{"triggered": false, "message": err.Error()})
		return
	}
	response.OK(c, gin.H{"triggered": true, "event": event})
}

func (h *WorldHandler) MakeChoice(c *gin.Context) {
	charID := getCharIDFromContext(c)
	var req service.MakeChoiceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}
	result, err := h.encounterSvc.MakeChoice(c.Request.Context(), charID, &req)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.OK(c, result)
}
