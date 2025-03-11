package config

import (
    "encoding/json"
    "go-auto-proxy/internal/system"
    "os"
    "testing"

    "github.com/stretchr/testify/assert"
)

func TestWriteConfig(t *testing.T) {
    // 準備測試資料
    info := system.SystemInfo{
        OS:           "linux",
        Version:      "Ubuntu 22.04.3 LTS",
        Architecture: "amd64",
        ExternalIP:   "35.185.174.224",
        InternalIP:   "10.140.0.5",
    }

    // 執行 WriteConfig
    err := WriteConfig(info)
    assert.NoError(t, err)
    defer os.Remove("config.json")

    // 讀取並驗證檔案內容
    data, err := os.ReadFile("config.json")
    assert.NoError(t, err)

    var result map[string]system.SystemInfo
    err = json.Unmarshal(data, &result)
    assert.NoError(t, err)

    sys := result["system"]
    assert.Equal(t, info.OS, sys.OS, "OS should match")
    assert.Equal(t, info.Version, sys.Version, "Version should match")
    assert.Equal(t, info.Architecture, sys.Architecture, "Architecture should match")
    assert.Equal(t, info.ExternalIP, sys.ExternalIP, "ExternalIP should match")
    assert.Equal(t, info.InternalIP, sys.InternalIP, "InternalIP should match")
}