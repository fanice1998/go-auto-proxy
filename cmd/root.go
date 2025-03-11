package cmd

import (
    "fmt"
    "os"

    "github.com/spf13/cobra"
)

// rootCmd 表示 CLI 的根命令
var rootCmd = &cobra.Command{
    Use:   "go-auto-proxy",
    Short: "A CLI tool for automating proxy setup",
    Long:  `go-auto-proxy is a command-line tool to automate proxy-related setup, including system info collection and dependency installation.`,
}

// Execute 執行根命令
func Execute() {
    if err := rootCmd.Execute(); err != nil {
        fmt.Println(err)
        os.Exit(1)
    }
}