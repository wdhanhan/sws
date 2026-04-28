package handler

import (
	"errors"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/starfall-warsong/sws/internal/middleware"
	"github.com/starfall-warsong/sws/internal/model"
	"github.com/starfall-warsong/sws/internal/service"
	"github.com/starfall-warsong/sws/pkg/response"
)

type CharacterHandler struct {
	svc *service.CharacterService
}

func NewCharacterHandler(svc *service.CharacterService) *CharacterHandler {
	return &CharacterHandler{svc: svc}
}

func (h *CharacterHandler) Create(c *gin.Context) {
	accountID := middleware.GetAccountID(c)

	var req service.CreateCharacterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}

	char, err := h.svc.Create(c.Request.Context(), accountID, &req)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrSlotsFull):
			response.Forbidden(c, err.Error())
		case errors.Is(err, service.ErrNameExists):
			response.Conflict(c, err.Error())
		case errors.Is(err, service.ErrInvalidRace):
			response.BadRequest(c, err.Error())
		case errors.Is(err, service.ErrNameTooShort), errors.Is(err, service.ErrNameTooLong):
			response.BadRequest(c, err.Error())
		default:
			response.InternalError(c, "创建角色失败")
		}
		return
	}

	response.Created(c, char)
}

func (h *CharacterHandler) List(c *gin.Context) {
	accountID := middleware.GetAccountID(c)

	chars, err := h.svc.List(c.Request.Context(), accountID)
	if err != nil {
		response.InternalError(c, "获取角色列表失败")
		return
	}

	type CharWithRace struct {
		model.Character
		RaceName string `json:"race_name"`
	}
	enriched := make([]CharWithRace, len(chars))
	for i, ch := range chars {
		enriched[i] = CharWithRace{Character: ch, RaceName: model.RaceNames[ch.RaceID]}
	}

	response.OK(c, gin.H{
		"characters": enriched,
		"count":      len(enriched),
	})
}

func (h *CharacterHandler) Get(c *gin.Context) {
	accountID := middleware.GetAccountID(c)
	charID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的角色ID")
		return
	}

	char, err := h.svc.Get(c.Request.Context(), accountID, charID)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrCharNotFound):
			response.NotFound(c, err.Error())
		case errors.Is(err, service.ErrNotOwner):
			response.Forbidden(c, err.Error())
		default:
			response.InternalError(c, "获取角色失败")
		}
		return
	}

	response.OK(c, char)
}

func (h *CharacterHandler) Delete(c *gin.Context) {
	accountID := middleware.GetAccountID(c)
	charID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的角色ID")
		return
	}

	if err := h.svc.Delete(c.Request.Context(), accountID, charID); err != nil {
		switch {
		case errors.Is(err, service.ErrCharNotFound):
			response.NotFound(c, err.Error())
		case errors.Is(err, service.ErrNotOwner):
			response.Forbidden(c, err.Error())
		default:
			response.InternalError(c, "删除角色失败")
		}
		return
	}

	response.OK(c, gin.H{"message": "角色已删除"})
}
