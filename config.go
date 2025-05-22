package main

import (
	"log"
	"os"
	"encoding/json"
)

type Config struct {
    UploadKBps uint64 `json:"uploadKBps"`
    UploadBurstSizeKB uint64 `json:"uploadBurstSizeKB"`
	SeedDurationMinute uint64 `json:"seedDurationMinute"`
    SeedingPort uint `json:"seedingPort"`
}

var (
	UploadKBps uint64
	UploadBurstSizeKB uint64
	SeedDurationMinute uint64
	SeedingPort uint
)

func InitializeConfig(path string) {

	file, err := os.ReadFile(path)
    if err != nil {
		log.Fatalf("can't read config.json: %v", err)
    }

    var config Config
    if err := json.Unmarshal(file, &config); err != nil {
        log.Fatalf("failed to parse config: %v", err)
    }

	UploadKBps = config.UploadKBps 
	UploadBurstSizeKB = config.UploadBurstSizeKB 
	SeedDurationMinute = config.SeedDurationMinute
	SeedingPort = config.SeedingPort

}
