package system

import (
    "crypto/rand"
    "encoding/hex"
    "io"
    "net"
    "net/http"
    "os"
    "path/filepath"
    "runtime"
    "strings"
)

type SystemInfo struct {
    OS           string `json:"os"`
    Version      string `json:"version"`
    Architecture string `json:"architecture"`
    ExternalIP   string `json:"external_ip"`
    InternalIP   string `json:"internal_ip"`
    ZeroTier     struct {
        NetworkID string `json:"network_id"`
    } `json:"zerotier"`
    AcmeSH struct {
        Path     string `json:"path"`
        Provider string `json:"provider"`
    } `json:"acme_sh"`
    TrojanGo struct {
        Port     int    `json:"port"`
        Password string `json:"password"`
    } `json:"trojan_go"`
    Fail2Ban struct {
        MonitoredItems []string `json:"monitored_items"`
    } `json:"fail2ban"`
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

    // 系統版本
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

    // ZeroTier 預設值
    info.ZeroTier.NetworkID = "" // 留空，待用戶配置

    // acme.sh 預設值
    homeDir, _ := os.UserHomeDir()
    info.AcmeSH.Path = filepath.Join(homeDir, ".acme.sh", "acme.sh")
    info.AcmeSH.Provider = "letsencrypt" // 預設使用 Let’s Encrypt

    // trojan-go 預設值
    info.TrojanGo.Port = 443 // 預設 HTTPS 端口
    info.TrojanGo.Password = generateRandomPassword(16) // 隨機生成 16 字節密碼

    // fail2ban 預設值
    info.Fail2Ban.MonitoredItems = []string{"ssh"} // 預設監控 SSH

    return info
}

// generateRandomPassword 生成隨機密碼
func generateRandomPassword(length int) string {
    bytes := make([]byte, length)
    if _, err := rand.Read(bytes); err != nil {
        return "defaultpassword" // 若生成失敗，使用預設值
    }
    return hex.EncodeToString(bytes)
}