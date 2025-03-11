go-auto-proxy/
├── cmd/                # CLI 命令實作
│   └── init.go         # init 命令邏輯
├── internal/           # 內部邏輯
│   ├── config/         # 配置相關
│   │   └── config.go   # 處理 config.json
│   ├── system/         # 系統資訊收集
│   │   └── system.go   # 獲取系統資訊
│   └── installer/      # 軟體安裝邏輯
│       └── install.go  # 安裝 trojan-go 等
├── main.go             # 程式入口
├── go.mod              # Go 模組定義
├── go.sum              # 依賴檢查
└── README.md           # 專案說明
