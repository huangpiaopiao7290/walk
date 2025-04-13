package utils_test

import (
	"log"
	"os"
	"strings"
	"testing"
	"time"

	user_config "walk/configs/user"
	utils "walk/pkg/utils"
)

func initConfig() *user_config.JWT {
	filePath := "/home/pp/programs/program_go/timeTrack/walk/configs/user-service.yml"
	// Ensure the file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		log.Printf("The file %s does not exist", filePath)
		return nil
	}
	cfg, err := user_config.InitUserConfig(filePath) // Initialize user configuration
	if err != nil {
		log.Printf("Failed to initialize user configuration: %v", err)
		return nil
	}

	return &cfg.JWT
}

func TestJWT(t *testing.T) {
	// 加载user配置
	JWTcfg := initConfig()

	// 用户唯一标识
	var uuid string = "c2754f44-1079-11f0-89d9-00155d6c6ae0"

	// 获取jwt配置
	jwtConf := &user_config.JWT{
		Secret:           JWTcfg.Secret,
		Access_token_ttl: JWTcfg.Access_token_ttl,
		Refresh_token_ttl: JWTcfg.Refresh_token_ttl,
	}

	// 生成access token
	accessToken, err := utils.GenerateAccessToken(jwtConf, uuid)
	if err != nil {
		log.Fatalf("Error generating access token: %v", err)
	}

	log.Printf("Access Token: %s", accessToken)

	// 验证token
	var parseToken *utils.Claim
	parseToken, err = utils.ParseToken(accessToken, jwtConf)
	if err != nil {
		log.Fatalf("Error parsing token: %v", err)
	}
	log.Printf("Parsed Token:\n  UUID: %s\n  ExpiresAt: %v\n  NotBefore: %v\n  IssuedAt: %v\n",
    parseToken.UUID, parseToken.ExpiresAt, parseToken.NotBefore, parseToken.IssuedAt)

}

// 测试过期token
func TestExpiredToken(t *testing.T) {
	// 加载user配置
	JWTcfg := initConfig()

	// 用户唯一标识
	var uuid string = "c2754f44-1079-11f0-89d9-00155d6c6ae0"

	// 获取jwt配置
	jwtConf := &user_config.JWT{
		Secret:           JWTcfg.Secret,
		Access_token_ttl: 1,		// 1s
		// Refresh_token_ttl: 2,		// 2s
	}

	// 生成access token
	accessToken, err := utils.GenerateAccessToken(jwtConf, uuid)
	if err != nil {
		log.Fatalf("Error generating access token: %v", err)
	}

	log.Printf("Access Token: %s", accessToken)

	// 等到token过期
	time.Sleep(2 * time.Second)

	// 解析验证token
	_, err = utils.ParseToken(accessToken, jwtConf)
	if err == nil {
		log.Fatal("Expected error for expired token, but got nil")
	}

	// 检查错误信息
	if !strings.Contains(err.Error(), "token has expired") {
		t.Fatalf("Unexpected error: %v", err)
	}

	t.Logf("Token expired as expected: %v", err)
} 

// 测试劫持token
func TestHijackToken(t *testing.T) {
	// 加载user配置
	JWTcfg := initConfig()

	// 用户唯一标识
	var uuid string = "c2754f44-1079-11f0-89d9-00155d6c6ae0"

	// 获取jwt配置
	jwtConf := &user_config.JWT{
		Secret:           JWTcfg.Secret,
		Access_token_ttl: 1,		// 1s
		// Refresh_token_ttl: 2,		// 2s
	}

	// 生成access token
	accessToken, err := utils.GenerateAccessToken(jwtConf, uuid)
	if err != nil {
		log.Fatalf("Error generating access token: %v", err)
	}

	// 模拟劫持篡改token
	parts := strings.Split(accessToken, ".")
	if len(parts) != 3 {
		t.Fatalf("Invalid token format")
	}
	// 修改token的payload部分
	parts[1] = "test_hijack"
	tamperedToken := strings.Join(parts, ".")
	log.Printf("Tampered Token: %s", tamperedToken)

	// 解析验证token
	_, err = utils.ParseToken(tamperedToken, jwtConf)
	if err == nil {
		log.Fatal("Expected error for tampered token, but got nil")
	}
	// 检查错误信息
	if !strings.Contains(err.Error(), "token is invalid") {
		t.Fatalf("Unexpected error: %v", err)
	}
	t.Logf("Token tampering detected as expected: %v", err)


	// 解析原始token
	var parseToken *utils.Claim
	parseToken, err = utils.ParseToken(accessToken, jwtConf)
	if err != nil {
		log.Fatalf("Error parsing original token: %v", err)
	}
	log.Printf("Parsed Original Token:\n  UUID: %s\n  ExpiresAt: %v\n  NotBefore: %v\n  IssuedAt: %v\n",
		parseToken.UUID, parseToken.ExpiresAt, parseToken.NotBefore, parseToken.IssuedAt)
}
