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

func (h *AccountHandler) Login(c *gin.Context) {
	var req service.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请输入账号")
		return
	}

	result, err := h.svc.LoginOrRegister(c.Request.Context(), &req)
	if err != nil {
		if errors.Is(err, service.ErrAccountInvalid) {
			response.BadRequest(c, err.Error())
		} else {
			response.InternalError(c, "登录失败")
		}
		return
	}

	response.OK(c, result)
}
