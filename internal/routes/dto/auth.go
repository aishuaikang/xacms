package dto

import (
	"github.com/google/uuid"
)

// LoginRequest 登录请求结构
type LoginRequest struct {
	Username string `json:"username" validate:"required,min=3,max=64"`
	Password string `json:"password" validate:"required,min=6,max=128"`
	Domain   string `json:"domain" validate:"omitempty,fqdn"`
}

// LoginResponse 登录响应结构
type LoginResponse struct {
	AccessToken  string     `json:"access_token"`
	RefreshToken string     `json:"refresh_token"`
	ExpiresIn    int64      `json:"expires_in"`
	TokenType    string     `json:"token_type"`
	UserID       uuid.UUID  `json:"user_id"`
	Username     string     `json:"username"`
	Nickname     string     `json:"nickname"`
	TenantID     *uuid.UUID `json:"tenant_id,omitempty"`
}

// RefreshTokenRequest 刷新令牌请求结构
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

// LogoutRequest 退出登录请求结构
type LogoutRequest struct {
	RefreshToken string `json:"refresh_token" validate:"omitempty"`
}

// RegisterRequest 注册请求结构
type RegisterRequest struct {
	Username string `json:"username" validate:"required,min=3,max=64"`
	Password string `json:"password" validate:"required,min=6,max=128"`
	Email    string `json:"email" validate:"required,email,max=128"`
	Phone    string `json:"phone" validate:"required,phone"`
	Nickname string `json:"nickname" validate:"required,min=2,max=64"`
	Domain   string `json:"domain" validate:"omitempty,fqdn"`
}

// ForgotPasswordRequest 忘记密码请求结构
type ForgotPasswordRequest struct {
	Email  string `json:"email" validate:"required,email"`
	Domain string `json:"domain" validate:"omitempty,fqdn"`
}
