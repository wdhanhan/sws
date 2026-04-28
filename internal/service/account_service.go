package service

import (
	"context"
	"errors"
	"regexp"

	"github.com/starfall-warsong/sws/internal/model"
	"github.com/starfall-warsong/sws/internal/repository"
	"github.com/starfall-warsong/sws/pkg/auth"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrPhoneInvalid    = errors.New("手机号格式不正确")
	ErrPhoneExists     = errors.New("手机号已注册")
	ErrPasswordTooWeak = errors.New("密码长度至少8位")
	ErrLoginFailed     = errors.New("手机号或密码错误")
)

var phoneRegex = regexp.MustCompile(`^1[3-9]\d{9}$`)

type AccountService struct {
	repo *repository.AccountRepo
	jwt  *auth.JWTManager
}

func NewAccountService(repo *repository.AccountRepo, jwt *auth.JWTManager) *AccountService {
	return &AccountService{repo: repo, jwt: jwt}
}

type RegisterRequest struct {
	Phone    string `json:"phone" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type LoginRequest struct {
	Phone    string `json:"phone" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type TokenResponse struct {
	AccessToken  string         `json:"access_token"`
	RefreshToken string         `json:"refresh_token"`
	Account      *model.Account `json:"account"`
}

func (s *AccountService) Register(ctx context.Context, req *RegisterRequest) (*TokenResponse, error) {
	if !phoneRegex.MatchString(req.Phone) {
		return nil, ErrPhoneInvalid
	}
	if len(req.Password) < 8 {
		return nil, ErrPasswordTooWeak
	}

	existing, err := s.repo.GetByPhone(ctx, req.Phone)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, ErrPhoneExists
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	account, err := s.repo.Create(ctx, req.Phone, string(hash))
	if err != nil {
		return nil, err
	}

	return s.generateTokens(account)
}

func (s *AccountService) Login(ctx context.Context, req *LoginRequest) (*TokenResponse, error) {
	account, err := s.repo.GetByPhone(ctx, req.Phone)
	if err != nil {
		return nil, err
	}
	if account == nil {
		return nil, ErrLoginFailed
	}

	if err := bcrypt.CompareHashAndPassword([]byte(account.PasswordHash), []byte(req.Password)); err != nil {
		return nil, ErrLoginFailed
	}

	return s.generateTokens(account)
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
