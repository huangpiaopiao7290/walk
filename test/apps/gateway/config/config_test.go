package gateway_config

import (
	"fmt"
	"os"
	"testing"

	gateway_config "walk/apps/gateway/config"
)

func TestGatewayConfig(t *testing.T) {
	filePath := "/home/pp/programs/program_go/timeTrack/walk/apps/gateway/config/grpc_gateway_conf.yml"
	// Ensure the file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		t.Fatalf("Config file does not exist at path: %s", filePath)
	}
	cfg, err := gateway_config.InitGatewayConfig(filePath)
	if err != nil {
		t.Errorf("InitGatewayConfig returned unexpected error: %v", err)
	}

	// Output the config object
	fmt.Printf("config: %+v\n", cfg)
}
