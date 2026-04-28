package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/starfall-warsong/sws/internal/model"
	"github.com/starfall-warsong/sws/internal/repository"
	"github.com/starfall-warsong/sws/pkg/response"
)

type StarmapHandler struct {
	repo *repository.StarmapRepo
}

func NewStarmapHandler(repo *repository.StarmapRepo) *StarmapHandler {
	return &StarmapHandler{repo: repo}
}

func (h *StarmapHandler) GetSystem(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的星系ID")
		return
	}

	sys, err := h.repo.GetSystem(c.Request.Context(), id)
	if err != nil {
		response.InternalError(c, "查询星系失败")
		return
	}
	if sys == nil {
		response.NotFound(c, "星系不存在")
		return
	}

	planets, _ := h.repo.GetSystemPlanets(c.Request.Context(), id)
	belts, _ := h.repo.GetSystemBelts(c.Request.Context(), id)
	adjacent, _ := h.repo.GetAdjacentSystems(c.Request.Context(), id)

	// 查询本星系空间站
	type StationBrief struct {
		ID   int64  `db:"id" json:"id"`
		Name string `db:"name" json:"name"`
	}
	var stations []StationBrief
	h.repo.DB().SelectContext(c.Request.Context(), &stations,
		`SELECT id, name FROM stations WHERE system_id = $1`, id)

	adjNames := make([]gin.H, len(adjacent))
	for i, a := range adjacent {
		adjNames[i] = gin.H{"id": a.ID, "name": a.Name, "security_level": a.SecurityLevel}
	}

	response.OK(c, gin.H{
		"system":    sys,
		"arm_name":  model.ArmNames[sys.ArmID],
		"star_name": model.StarTypeNames[sys.StarType],
		"zone":      sys.SecurityZone(),
		"planets":   planets,
		"belts":     belts,
		"gates_to":  adjNames,
		"stations":  stations,
	})
}

func (h *StarmapHandler) SearchSystems(c *gin.Context) {
	limit := 50
	if l, err := strconv.Atoi(c.Query("limit")); err == nil && l > 0 && l <= 200 {
		limit = l
	}

	var armID *model.ArmID
	if a, err := strconv.Atoi(c.Query("arm")); err == nil {
		aid := model.ArmID(a)
		armID = &aid
	}

	var secMin, secMax *float64
	if v, err := strconv.ParseFloat(c.Query("sec_min"), 64); err == nil {
		secMin = &v
	}
	if v, err := strconv.ParseFloat(c.Query("sec_max"), 64); err == nil {
		secMax = &v
	}

	systems, err := h.repo.SearchSystems(c.Request.Context(), armID, secMin, secMax, limit)
	if err != nil {
		response.InternalError(c, "搜索星系失败")
		return
	}

	response.OK(c, gin.H{
		"systems": systems,
		"count":   len(systems),
	})
}

func (h *StarmapHandler) GetStats(c *gin.Context) {
	count, _ := h.repo.GetSystemCount(c.Request.Context())
	response.OK(c, gin.H{
		"total_systems": count,
	})
}
