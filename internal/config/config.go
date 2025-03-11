package config

import (
    "encoding/json"
    "go-auto-proxy/internal/system"
    "os"
)

func WriteConfig(info system.SystemInfo) error {
    file, err := os.Create("config.json")
    if err != nil {
        return err
    }
    defer file.Close()

    encoder := json.NewEncoder(file)
    encoder.SetIndent("", "  ")
    return encoder.Encode(map[string]interface{}{
        "system": info,
    })
}