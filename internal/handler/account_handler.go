package handler

import (
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/starfall-warsong/sws/internal/service"
	"github.com/starfall-warsong/sws/pkg/response"
)

type AccountHandler struct {
	svc *service.AccountService
}

func NewAccountHandler(svc *service.AccountService) *AccountHandler {
	return &AccountHandler{svc: svc}
}

func (h *AccountHandler) Register(c *gin.Context) {
	var req service.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}

	result, err := h.svc.Register(c.Request.Context(), &req)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrPhoneInvalid):
			response.BadRequest(c, err.Error())
		case errors.Is(err, service.ErrPhoneExists):
			response.Conflict(c, err.Error())
		case errors.Is(err, service.ErrPasswordTooWeak):
			response.BadRequest(c, err.Error())
		default:
			response.InternalError(c, "注册失败")
		}
		return
	}

	response.Created(c, result)
}

func (h *AccountHandler) Login(c *gin.Context) {
	var req service.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}

	result, err := h.svc.Login(c.Request.Context(), &req)
	if err != nil {
		if errors.Is(err, service.ErrLoginFailed) {
			response.Unauthorized(c, err.Error())
		} else {
			response.InternalError(c, "登录失败")
		}
		return
	}

	response.OK(c, result)
}
