package system

import (
    "net/http"
    "net/http/httptest"
    "os"
    "runtime"
    "testing"

    "github.com/stretchr/testify/assert"
)

func TestGetSystemInfo(t *testing.T) {
    // 模擬 /etc/os-release 檔案
    err := os.WriteFile("test-os-release", []byte(`PRETTY_NAME="Ubuntu 22.04.3 LTS"`), 0644)
    assert.NoError(t, err)
    defer os.Remove("test-os-release")

    // 模擬外部 IP 的 HTTP 伺服器
    ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.Write([]byte("35.185.174.224"))
    }))
    defer ts.Close()

    // 模擬 ReadFile
    originalReadFile := DefaultReadFile
    DefaultReadFile = func(path string) ([]byte, error) {
        return os.ReadFile("test-os-release")
    }
    defer func() { DefaultReadFile = originalReadFile }()

    // 測試 GetSystemInfo
    info := GetSystemInfo()

    assert.Equal(t, runtime.GOOS, info.OS, "OS should match runtime.GOOS")
    assert.Equal(t, "Ubuntu 22.04.3 LTS", info.Version, "Version should match /etc/os-release")
    assert.Equal(t, runtime.GOARCH, info.Architecture, "Architecture should match runtime.GOARCH")
    assert.NotEmpty(t, info.InternalIP, "InternalIP should not be empty")
    assert.Contains(t, info.ExternalIP, "35.185.174.224", "ExternalIP should match mock server response")
}