package config_test

import (
	"fmt"
	"os"
	"testing"

	map_config "walk/apps/map/config"
)

func TestLoadMapConfig_ValidFile(t *testing.T) {
	filePath := "/home/pp/programs/program_go/timeTrack/walk/apps/map/config/map-service.yml"
	// Ensure the file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		t.Fatalf("Config file does not exist at path: %s", filePath)
	}
	cfg, err := map_config.GetMapConfig(filePath)
	if err != nil {
		t.Errorf("GetMapConfig returned unexpected error: %v", err)
	}

	// Output the config object
	fmt.Printf("config: %+v\n", cfg)

}
