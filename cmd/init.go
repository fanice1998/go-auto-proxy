package cmd

import (
    "fmt"
    "go-auto-proxy/internal/config"
    "go-auto-proxy/internal/installer"
    "go-auto-proxy/internal/system"
    "io"
    "log"
    "os"
    "os/exec"
    "os/user"
    "path/filepath"
    "strings"

    "github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
    Use:   "init",
    Short: "Initialize go-auto-proxy and install dependencies",
    Run: func(cmd *cobra.Command, args []string) {
        // 初始化日誌
        logFile, err := os.OpenFile("go-auto-proxy.log", os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
        if err != nil {
            fmt.Printf("Failed to open log file: %v\n", err)
            return
        }
        defer logFile.Close()

        mw := io.MultiWriter(os.Stdout, logFile)
        log.SetOutput(mw)
        log.SetFlags(log.Ldate | log.Ltime)

        log.Println("Initializing go-auto-proxy...")

        // 檢查當前目錄權限
        if err := checkDirPermissions(); err != nil {
            log.Println(err)
            return
        }

        // 檢測當前用戶與 sudo 權限
        if err := checkUserAndSudo(); err != nil {
            log.Println(err)
            return
        }

        // 收集系統資訊
        sysInfo := system.GetSystemInfo()
        log.Printf("System Info: %+v", sysInfo)

        // 檢查並處理 trojan-go 目錄
        if err := handleTrojanGoDir(); err != nil {
            log.Println(err)
            return
        }

        // 寫入 config.json
        err = config.WriteConfig(sysInfo)
        if err != nil {
            log.Println("Error writing config:", err)
            return
        }

        // 安裝依賴
        if err := installer.InstallDependencies(); err != nil {
            log.Println("Error installing dependencies:", err)
            return
        }

        // 測試已安裝的工具
        if err := testInstalledTools(); err != nil {
            log.Println("Some tools failed verification:", err)
            // 不中止程式，僅記錄警告
        }

        log.Println("Initialization completed.")
    },
}

// checkDirPermissions 檢查當前目錄的寫入與刪除權限
func checkDirPermissions() error {
    testFile := "test-permission-file"
    if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
        return fmt.Errorf("no write permission in current directory: %v (try running with sudo or change to a writable directory)", err)
    }
    if err := os.Remove(testFile); err != nil {
        return fmt.Errorf("no delete permission in current directory: %v (try running with sudo or change to a writable directory)", err)
    }

    testDir := "test-permission-dir"
    if err := os.Mkdir(testDir, 0755); err != nil {
        return fmt.Errorf("no permission to create directories in current directory: %v (try running with sudo or change to a writable directory)", err)
    }
    if err := os.Remove(testDir); err != nil {
        return fmt.Errorf("no permission to delete directories in current directory: %v (try running with sudo or change to a writable directory)", err)
    }

    log.Println("Current directory permissions verified.")
    return nil
}

// handleTrojanGoDir 檢查並處理 trojan-go 目錄
func handleTrojanGoDir() error {
    trojanDir := "trojan-go"
    if _, err := os.Stat(trojanDir); err == nil {
        log.Printf("Directory %s already exists.", trojanDir)
        fmt.Printf("Directory %s already exists. Remove and recreate? (y/n): ", trojanDir)
        var input string
        if _, err := fmt.Scanln(&input); err != nil {
            return fmt.Errorf("failed to read input: %v", err)
        }
        input = strings.ToLower(strings.TrimSpace(input))
        if input == "y" {
            log.Printf("Removing existing %s directory...", trojanDir)
            if err := os.RemoveAll(trojanDir); err != nil {
                return fmt.Errorf("failed to remove %s: %v", trojanDir, err)
            }
            log.Printf("%s directory removed.", trojanDir)
        } else {
            log.Printf("Keeping existing %s directory.", trojanDir)
        }
    }
    return nil
}

// checkUserAndSudo 檢測當前用戶與 sudo 權限
func checkUserAndSudo() error {
    currentUser, err := user.Current()
    if err != nil {
        return fmt.Errorf("failed to get current user: %v", err)
    }

    log.Printf("Current user: %s (UID: %s)", currentUser.Username, currentUser.Uid)

    if currentUser.Uid == "0" {
        log.Println("Running as root, proceeding...")
        return nil
    }

    cmd := exec.Command("sudo", "-n", "true")
    if err := cmd.Run(); err != nil {
        if strings.Contains(err.Error(), "sudo: a password is required") {
            log.Println("Warning: sudo requires a password, which may cause issues during installation.")
            log.Println("To enable passwordless sudo, edit /etc/sudoers with 'sudo visudo' and add:")
            log.Printf("  %s ALL=(ALL) NOPASSWD: ALL", currentUser.Username)
            log.Println("Alternatively, run this program as root with 'sudo go run main.go init'.")
            return fmt.Errorf("sudo password required")
        }
        return fmt.Errorf("user %s does not have sudo privileges: %v", currentUser.Username, err)
    }

    log.Println("User has passwordless sudo privileges, proceeding...")
    return nil
}

// testInstalledTools 測試已安裝的工具是否可用
func testInstalledTools() error {
    var errors []string

    // 測試 trojan-go
    trojanPath := filepath.Join("trojan-go", "trojan-go")
    if _, err := os.Stat(trojanPath); err != nil {
        errors = append(errors, fmt.Sprintf("trojan-go not found at %s: %v", trojanPath, err))
    } else if err := exec.Command(trojanPath, "--version").Run(); err != nil {
        errors = append(errors, fmt.Sprintf("trojan-go failed to run: %v", err))
    } else {
        log.Println("trojan-go verified successfully.")
    }

    // 測試 acme.sh
    homeDir, _ := os.UserHomeDir()
    acmePath := filepath.Join(homeDir, ".acme.sh", "acme.sh")
    if err := exec.Command(acmePath, "--version").Run(); err != nil {
        errors = append(errors, fmt.Sprintf("acme.sh failed to run: %v (ensure it’s installed correctly at %s)", err, acmePath))
    } else {
        log.Println("acme.sh verified successfully.")
    }

    // 測試 nginx
    if err := exec.Command("nginx", "-v").Run(); err != nil {
        errors = append(errors, fmt.Sprintf("nginx failed to run: %v (ensure it’s installed correctly)", err))
    } else {
        log.Println("nginx verified successfully.")
    }

    // 測試 fail2ban
    if err := exec.Command("fail2ban-client", "version").Run(); err != nil {
        errors = append(errors, fmt.Sprintf("fail2ban failed to run: %v (ensure it’s installed correctly)", err))
    } else {
        log.Println("fail2ban verified successfully.")
    }

    // 測試 zerotier
    if err := exec.Command("zerotier-cli", "-v").Run(); err != nil {
        errors = append(errors, fmt.Sprintf("zerotier failed to run: %v (ensure it’s installed correctly)", err))
    } else {
        log.Println("zerotier verified successfully.")
    }

    if len(errors) > 0 {
        return fmt.Errorf("verification errors: %s", strings.Join(errors, "; "))
    }
    return nil
}

func init() {
    rootCmd.AddCommand(initCmd)
}