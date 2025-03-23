package config_test

import (
	"fmt"
	"os"
	"testing"

	"walk/configs/user"
)

func TestUserConfig(t *testing.T) {
	filePath := "/home/pp/programs/program_go/timeTrack/walk/configs/user-service.yml"
	// Ensure the file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		t.Fatalf("Config file does not exist at path: %s", filePath)
	}
	cfg, err := user_config.InitUserConfig(filePath)
	if err != nil {
		t.Errorf("InitUserConfig returned unexpected error: %v", err)
	}

	// Output the config object
	fmt.Printf("config: %+v\n", cfg)
}

