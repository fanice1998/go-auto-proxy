package installer

import (
    "os"
    "os/exec"
    "testing"

    "github.com/stretchr/testify/assert"
)

// mockExecCommand 用於模擬 exec.Command
func mockExecCommand(command string, args ...string) *exec.Cmd {
    cs := []string{"-test.run=TestHelperProcess", "--", command}
    cs = append(cs, args...)
    cmd := exec.Command(os.Args[0], cs...)
    cmd.Env = []string{"GO_WANT_HELPER_PROCESS=1"}
    return cmd
}

func TestInstallDependenciesSuccess(t *testing.T) {
    // 保存原始函數
    originalCommand := DefaultCommand
    originalMkdirAll := DefaultMkdirAll
    originalRemove := DefaultRemove

    // 替換為 mock
    DefaultCommand = mockExecCommand
    DefaultMkdirAll = func(path string, perm os.FileMode) error { return nil }
    DefaultRemove = func(path string) error { return nil }

    defer func() {
        DefaultCommand = originalCommand
        DefaultMkdirAll = originalMkdirAll
        DefaultRemove = originalRemove
    }()

    // 執行並驗證
    err := InstallDependencies()
    assert.NoError(t, err, "InstallDependencies should succeed with mock commands")
}

func TestInstallDependenciesFailure(t *testing.T) {
    // 模擬命令失敗
    DefaultCommand = func(command string, args ...string) *exec.Cmd {
        return exec.Command("false") // 模擬失敗
    }
    defer func() { DefaultCommand = exec.Command }()

    err := InstallDependencies()
    assert.Error(t, err, "InstallDependencies should fail with mock failure")
    assert.Contains(t, err.Error(), "failed", "Error message should indicate failure")
}

// TestHelperProcess 模擬命令執行
func TestHelperProcess(t *testing.T) {
    if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
        return
    }
    os.Exit(0) // 模擬成功
}