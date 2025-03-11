package installer

import (
    "bytes"
    "context"
    "fmt"
    "log"
    "os"
    "os/exec"
    "path/filepath"
    "strings"
    "time"
)

const (
    bashrcPath = ".bashrc"
    commandTimeout = 5 * time.Minute // 設置 5 分鐘超時
)

// CommandFunc 定義執行命令的函數類型
type CommandFunc func(string, ...string) *exec.Cmd

// MkdirAllFunc 定義建立目錄的函數類型
type MkdirAllFunc func(string, os.FileMode) error

// RemoveFunc 定義刪除檔案的函數類型
type RemoveFunc func(string) error

// DefaultCommand 使用標準的 exec.Command
var DefaultCommand CommandFunc = exec.Command

// DefaultMkdirAll 使用標準的 os.MkdirAll
var DefaultMkdirAll MkdirAllFunc = os.MkdirAll

// DefaultRemove 使用標準的 os.Remove
var DefaultRemove RemoveFunc = os.Remove

// InstallDependencies 使用指定的 CommandFunc 安裝依賴
func InstallDependencies() error {
    log.Println("Updating package index...")
    if err := runCommand("sudo", "apt", "update"); err != nil {
        return err
    }
    log.Println("Installing unzip...")
    if err := runCommand("sudo", "apt", "install", "-y", "unzip"); err != nil {
        return err
    }

    trojanDir := "trojan-go"
    log.Println("Creating trojan-go directory...")
    if err := DefaultMkdirAll(trojanDir, 0755); err != nil {
        return err
    }
    zipPath := filepath.Join(trojanDir, "trojan-go-linux-amd64.zip")
    log.Println("Downloading trojan-go...")
    if err := runCommand("wget", "-O", zipPath, "https://github.com/p4gefau1t/trojan-go/releases/download/v0.10.6/trojan-go-linux-amd64.zip"); err != nil {
        return err
    }
    log.Println("Unzipping trojan-go...")
    if err := runCommand("unzip", zipPath, "-d", trojanDir); err != nil {
        return err
    }
    log.Println("Removing trojan-go zip file...")
    if err := DefaultRemove(zipPath); err != nil {
        return err
    }

    log.Println("Installing acme.sh...")
    if err := runCommand("sh", "-c", "curl https://get.acme.sh | sh"); err != nil {
        return err
    }
    log.Println("Setting alias for acme.sh...")
    homeDir, err := os.UserHomeDir()
    if err != nil {
        return err
    }
    acmePath := filepath.Join(homeDir, ".acme.sh", "acme.sh")
    aliasCmd := fmt.Sprintf(`alias acme.sh="%s"`, acmePath)
    bashrc := filepath.Join(homeDir, bashrcPath)
    if err := appendToFile(bashrc, aliasCmd); err != nil {
        return err
    }
    log.Printf("Alias added to %s: %s", bashrc, aliasCmd)
    log.Println("Please run 'source ~/.bashrc' or restart your shell to apply the alias.")

    log.Println("Installing nginx...")
    if err := runCommand("sudo", "apt", "install", "-y", "nginx"); err != nil {
        return err
    }

    log.Println("Installing fail2ban...")
    if err := runCommand("sudo", "apt", "install", "-y", "fail2ban"); err != nil {
        return err
    }

    log.Println("Installing ZeroTier...")
    if err := installZeroTier(); err != nil {
        return err
    }

    log.Println("Dependencies installed successfully.")
    return nil
}

// runCommand 執行命令並記錄輸出，帶超時
func runCommand(name string, args ...string) error {
    ctx, cancel := context.WithTimeout(context.Background(), commandTimeout)
    defer cancel()

    cmd := DefaultCommand(name, args...)
    cmd.Env = os.Environ() // 繼承環境變數
    var out bytes.Buffer
    cmd.Stdout = &out
    cmd.Stderr = &out

    if err := cmd.Start(); err != nil {
        return fmt.Errorf("failed to start command: %v", err)
    }

    errChan := make(chan error, 1)
    go func() {
        errChan <- cmd.Wait()
    }()

    select {
    case err := <-errChan:
        if err != nil {
            log.Printf("Command failed: %s %v\nOutput: %s", name, args, out.String())
            return fmt.Errorf("command failed: %v", err)
        }
        log.Printf("Command output: %s", out.String())
        return nil
    case <-ctx.Done():
        cmd.Process.Kill() // 超時後殺死進程
        log.Printf("Command timed out after %v: %s %v\nPartial output: %s", commandTimeout, name, args, out.String())
        return fmt.Errorf("command timed out after %v", commandTimeout)
    }
}

// installZeroTier 使用官方推薦方式安裝 ZeroTier
func installZeroTier() error {
    log.Println("Adding ZeroTier GPG key...")
    cmd := exec.Command("sh", "-c", "curl -s https://raw.githubusercontent.com/zerotier/ZeroTierOne/master/doc/contact@zerotier.com.gpg | sudo gpg --dearmor -o /usr/share/keyrings/zerotier.gpg")
    var out bytes.Buffer
    cmd.Stdout = &out
    cmd.Stderr = &out
    if err := cmd.Run(); err != nil {
        log.Printf("Command output: %s", out.String())
        return fmt.Errorf("failed to add ZeroTier GPG key: %v", err)
    }
    log.Printf("GPG key added: %s", out.String())

    sourceLine := "deb [signed-by=/usr/share/keyrings/zerotier.gpg] http://download.zerotier.com/debian/jammy jammy main"
    log.Println("Adding ZeroTier repository...")
    if err := appendToFile("/etc/apt/sources.list.d/zerotier.list", sourceLine); err != nil {
        return fmt.Errorf("failed to add ZeroTier repository: %v", err)
    }

    log.Println("Updating package index for ZeroTier...")
    if err := runCommand("sudo", "apt", "update"); err != nil {
        return fmt.Errorf("failed to update package index for ZeroTier: %v", err)
    }
    log.Println("Installing zerotier-one...")
    if err := runCommand("sudo", "apt", "install", "-y", "zerotier-one"); err != nil {
        return fmt.Errorf("failed to install zerotier-one: %v", err)
    }

    return nil
}

// appendToFile 將內容追加到指定檔案，若內容已存在則跳過
func appendToFile(filename, content string) error {
    data, err := os.ReadFile(filename)
    if err == nil && len(data) > 0 {
        if bytes.Contains(data, []byte(content)) {
            log.Printf("Content already exists in %s, skipping...", filename)
            return nil
        }
    }

    if strings.HasPrefix(filename, "/etc/") {
        cmd := exec.Command("sudo", "tee", "-a", filename)
        cmd.Stdin = strings.NewReader(content + "\n")
        var out bytes.Buffer
        cmd.Stdout = &out
        cmd.Stderr = &out
        if err := cmd.Run(); err != nil {
            log.Printf("Command failed: sudo tee -a %s\nOutput: %s", filename, out.String())
            return fmt.Errorf("failed to write to %s: %v", filename, err)
        }
        log.Printf("Content appended to %s: %s", filename, content)
        return nil
    }

    f, err := os.OpenFile(filename, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
    if err != nil {
        return fmt.Errorf("failed to open %s: %v", filename, err)
    }
    defer f.Close()

    if _, err := f.WriteString("\n" + content + "\n"); err != nil {
        return fmt.Errorf("failed to write to %s: %v", filename, err)
    }
    log.Printf("Content appended to %s: %s", filename, content)
    return nil
}