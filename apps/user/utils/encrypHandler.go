// Auth: pp
// Created: 2025/03/27
// Description: encryption handler in data
// including:
// 1. 加密密码
// 2. 比较密码
// 3. 生成jwt access token
// 4. 生成jwt refresh token
// 5. 解析验证jwt token

package user_utils

import (
	"errors"
	"fmt"
	"time"

	user_config "walk/apps/user/config"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// 定义
type Claim struct {
	UUID string
	jwt.RegisteredClaims
}

// 加密密码
func HashPWD(originPwd string) (string, error) {
	hashedPwd, err := bcrypt.GenerateFromPassword([]byte(originPwd), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}

	return string(hashedPwd), nil
}

// 比较密码
func ComparePWD(hashedPwd string, originPwd string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPwd), []byte(originPwd))
}

// 生成jwt access token
// @param jwtCfg jwt配置
// @param uuid 用户id
// @return access token
func GenerateAccessToken(jwtCfg *user_config.JWT, uuid string) (string, error) {
	// 创建claims
	claims := &Claim{
		UUID: uuid,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(jwtCfg.Access_token_ttl) * time.Second)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	// 创建token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	// 签名
	signedStr, err := token.SignedString([]byte(jwtCfg.Secret))
	if err != nil {
		return "", err
	}

	return signedStr, err
}

// 生成jwt refresh token
// @param jwtCfg jwt配置
// @param uuid 用户id
// @return refresh token
func GenerateRefreshToken(jwtCfg *user_config.JWT, uuid string) (string, error) {

	// 创建claims
	claims := &Claim{
		UUID: uuid,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(jwtCfg.Refresh_token_ttl) * time.Second)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	// 创建token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	// 签名
	signedStr, err := token.SignedString([]byte(jwtCfg.Secret))
	if err != nil {
		return "", err
	}
	return signedStr, err
}

func ParseToken(tokenStr string, jwtCfg *user_config.JWT) (*Claim, error) {
	// 定义解析函数
	token, err := jwt.ParseWithClaims(tokenStr, &Claim{}, func(token *jwt.Token) (any, error) {
		// 检查签名方法是否为 HS256
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(jwtCfg.Secret), nil
	})

	// 检查解析错误
	if err != nil {
		// 检查错误类型
		if errors.Is(err, jwt.ErrTokenMalformed) {
			return nil, fmt.Errorf("token is malformed")
		} else if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, fmt.Errorf("token is expired")
		} else if errors.Is(err, jwt.ErrTokenNotValidYet) {
			return nil, fmt.Errorf("token is not valid yet")
		} else if errors.Is(err, jwt.ErrSignatureInvalid) {
			return nil, fmt.Errorf("signature is invalid")
		} else {
			return nil, fmt.Errorf("failed to parse token: %w", err)
		}
	}

	// 检查 Token 是否有效
	if claims, ok := token.Claims.(*Claim); ok && token.Valid {
		// 额外检查：Token 是否即将过期
		if time.Until(claims.ExpiresAt.Time) < time.Minute {
			fmt.Println("Warning: Token is about to expire")
		}
		return claims, nil
	}

	return nil, fmt.Errorf("invalid token claims")
}
