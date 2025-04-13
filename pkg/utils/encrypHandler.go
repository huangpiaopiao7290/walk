// Auth: pp
// Created: 2025/03/27
// Description: encryption handler in data
// including:
// 1. 加密密码
// 2. 比较密码
// 3. 生成jwt access token
// 4. 生成jwt refresh token
// 5. 解析验证jwt token

package utils

import (
	"time"

	"golang.org/x/crypto/bcrypt"
	jwt "github.com/golang-jwt/jwt/v5"
	user_config "walk/configs/user"
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


// 解析验证token
func ParseToken(tokenStr string, jwtCfg *user_config.JWT) (*Claim, error) {
	// 解析token
	token, err := jwt.ParseWithClaims(tokenStr, &Claim{}, func(token *jwt.Token) (any, error) {
		return []byte(jwtCfg.Secret), nil
	}, jwt.WithValidMethods([]string{"HS256"}))
	if err != nil {
		return nil, err
	}

	// 类型断言
	if claims, ok := token.Claims.(*Claim); ok && token.Valid {
		return claims, nil
	} else {
		return nil, err
	}
}


