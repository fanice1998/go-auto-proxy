package system

import (
    "io"
    "net"
    "net/http"
    "os"
    "runtime"
    "strings"
)

type SystemInfo struct {
    OS           string `json:"os"`
    Version      string `json:"version"`
    Architecture string `json:"architecture"`
    ExternalIP   string `json:"external_ip"`
    InternalIP   string `json:"internal_ip"`
}

// ReadFileFunc 定義讀取檔案的函數類型
type ReadFileFunc func(string) ([]byte, error)

// DefaultReadFile 使用標準的 os.ReadFile
var DefaultReadFile ReadFileFunc = os.ReadFile

func GetSystemInfo() SystemInfo {
    info := SystemInfo{
        OS:           runtime.GOOS,
        Architecture: runtime.GOARCH,
    }

    // 系統版本 (Linux 範例，從 /etc/os-release 獲取發行版本)
    if info.OS == "linux" {
        data, err := DefaultReadFile("/etc/os-release")
        if err == nil {
            lines := strings.Split(string(data), "\n")
            for _, line := range lines {
                if strings.HasPrefix(line, "PRETTY_NAME=") {
                    info.Version = strings.Trim(strings.TrimPrefix(line, "PRETTY_NAME="), `"`)
                    break
                }
            }
        }
        if info.Version == "" {
            info.Version = "unknown"
        }
    }

    // 內部 IP
    addrs, err := net.InterfaceAddrs()
    if err == nil {
        for _, addr := range addrs {
            if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
                if ipnet.IP.To4() != nil {
                    info.InternalIP = ipnet.IP.String()
                    break
                }
            }
        }
    }
    if info.InternalIP == "" {
        info.InternalIP = "unknown"
    }

    // 對外 IP
    resp, err := http.Get("https://api.ipify.org")
    if err == nil {
        defer resp.Body.Close()
        body, err := io.ReadAll(resp.Body)
        if err == nil {
            info.ExternalIP = string(body)
        }
    }
    if info.ExternalIP == "" {
        resp, err := http.Get("http://ifconfig.me")
        if err == nil {
            defer resp.Body.Close()
            body, err := io.ReadAll(resp.Body)
            if err == nil {
                info.ExternalIP = string(body)
            }
        }
    }
    if info.ExternalIP == "" {
        info.ExternalIP = "unknown"
    }

    return info
}