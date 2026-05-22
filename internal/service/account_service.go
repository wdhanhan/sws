package service

import (
	"context"
	"errors"
	"strings"
	"unicode/utf8"

	"github.com/starfall-warsong/sws/internal/model"
	"github.com/starfall-warsong/sws/internal/repository"
	"github.com/starfall-warsong/sws/pkg/auth"
)

var (
	ErrAccountInvalid = errors.New("账号长度为2-20个字符")
)

type AccountService struct {
	repo *repository.AccountRepo
	jwt  *auth.JWTManager
}

func NewAccountService(repo *repository.AccountRepo, jwt *auth.JWTManager) *AccountService {
	return &AccountService{repo: repo, jwt: jwt}
}

type LoginRequest struct {
	Account string `json:"account" binding:"required"`
}

type TokenResponse struct {
	AccessToken  string         `json:"access_token"`
	RefreshToken string         `json:"refresh_token"`
	Account      *model.Account `json:"account"`
	IsNew        bool           `json:"is_new"`
}

func (s *AccountService) LoginOrRegister(ctx context.Context, req *LoginRequest) (*TokenResponse, error) {
	accountName := strings.TrimSpace(req.Account)
	if utf8.RuneCountInString(accountName) < 2 || utf8.RuneCountInString(accountName) > 20 {
		return nil, ErrAccountInvalid
	}

	account, err := s.repo.GetByPhone(ctx, accountName)
	if err != nil {
		return nil, err
	}

	isNew := false
	if account == nil {
		isNew = true
		account, err = s.repo.Create(ctx, accountName, "")
		if err != nil {
			return nil, err
		}
	}

	resp, err := s.generateTokens(account)
	if err != nil {
		return nil, err
	}
	resp.IsNew = isNew
	return resp, nil
}

func (s *AccountService) generateTokens(account *model.Account) (*TokenResponse, error) {
	access, err := s.jwt.GenerateAccessToken(account.ID, account.Phone)
	if err != nil {
		return nil, err
	}

	refresh, err := s.jwt.GenerateRefreshToken(account.ID, account.Phone)
	if err != nil {
		return nil, err
	}

	return &TokenResponse{
		AccessToken:  access,
		RefreshToken: refresh,
		Account:      account,
	}, nil
}
