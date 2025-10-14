package utils

import (
	"errors"
	"strconv"
	"strings"
	"time"
	"uav_defender/internal/models"
	"uav_defender/internal/pkg/config"

	jwt "github.com/golang-jwt/jwt/v4"
)

var (
	ErrTokenExpired     = errors.New("token is expired")
	ErrTokenNotValidYet = errors.New("token not active yet")
	ErrTokenMalformed   = errors.New("that's not even a token")
	ErrTokenInvalid     = errors.New("couldn't handle this token")
)

func GenerateToken(user *models.User) (string, error) {
	j := NewJWT()

	claims := j.CreateClaims(BaseClaims{
		UserID:   user.ID,
		Username: user.Username,
		Nickname: user.Nickname,
	})
	return j.CreateToken(claims)
}

type JWT struct {
	SigningKey []byte
}

func NewJWT() *JWT {
	return &JWT{
		[]byte(config.AppConfig.Jwt.SigningKey),
	}
}

func (j *JWT) CreateClaims(baseClaims BaseClaims) CustomClaims {
	// bf, _ := ParseDuration(config.AppConfig.Jwt.BufferTime)
	ep, err := ParseDuration(config.AppConfig.Jwt.ExpiresTime)
	if err != nil {
		// 如果解析失败，使用默认值7天
		ep = 7 * 24 * time.Hour
	}
	claims := CustomClaims{
		BaseClaims: baseClaims,
		// BufferTime: int64(bf / time.Second), // 缓冲时间1天 缓冲时间内会获得新的token刷新令牌 此时一个用户会存在两个有效令牌 但是前端只留一个 另一个会丢失
		RegisteredClaims: jwt.RegisteredClaims{
			Audience:  jwt.ClaimStrings{"sbpatc"},                // 受众
			NotBefore: jwt.NewNumericDate(time.Now().Add(-1000)), // 签名生效时间
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(ep)),    // 过期时间 7天  配置文件
			Issuer:    config.AppConfig.SN,                       // 签名的发行者
		},
	}
	return claims
}

func (j *JWT) CreateToken(claims CustomClaims) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(j.SigningKey)
}

func (j *JWT) ParseToken(tokenString string) (*CustomClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &CustomClaims{}, func(token *jwt.Token) (i any, e error) {
		return j.SigningKey, nil
	})
	if err != nil {
		if ve, ok := err.(*jwt.ValidationError); ok {
			if ve.Errors&jwt.ValidationErrorMalformed != 0 {
				return nil, ErrTokenMalformed
			} else if ve.Errors&jwt.ValidationErrorExpired != 0 {
				// Token is expired
				return nil, ErrTokenExpired
			} else if ve.Errors&jwt.ValidationErrorNotValidYet != 0 {
				return nil, ErrTokenNotValidYet
			} else {
				return nil, ErrTokenInvalid
			}
		}
	}
	if token != nil {
		if claims, ok := token.Claims.(*CustomClaims); ok && token.Valid {
			return claims, nil
		}
		return nil, ErrTokenInvalid

	} else {
		return nil, ErrTokenInvalid
	}
}

type BaseClaims struct {
	UserID   uint64
	Username string
	Nickname string
}

type CustomClaims struct {
	BaseClaims
	jwt.RegisteredClaims
}

func ParseDuration(d string) (time.Duration, error) {
	d = strings.TrimSpace(d)
	dr, err := time.ParseDuration(d)
	if err == nil {
		return dr, nil
	}
	if strings.Contains(d, "d") {
		index := strings.Index(d, "d")

		hour, err := strconv.Atoi(d[:index])
		if err != nil {
			return 0, err
		}
		dr = time.Hour * 24 * time.Duration(hour)

		// 如果"d"后面还有内容，继续解析
		if index+1 < len(d) {
			ndr, err := time.ParseDuration(d[index+1:])
			if err != nil {
				return dr, nil
			}
			return dr + ndr, nil
		}
		return dr, nil
	}

	dv, err := strconv.ParseInt(d, 10, 64)
	return time.Duration(dv), err
}
