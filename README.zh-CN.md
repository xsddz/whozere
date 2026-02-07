<p align="center">
  <img src="docs/logo.svg" alt="whozere" width="120">
</p>

<h1 align="center">whozere</h1>

<p align="center">
  <strong>谁来了？🔔</strong> — 跨平台登录检测与通知工具
</p>

<p align="center">
  <a href="https://github.com/xsddz/whozere/releases"><img src="https://img.shields.io/github/v/release/xsddz/whozere?style=flat-square" alt="Release"></a>
  <a href="https://github.com/xsddz/whozere/blob/main/LICENSE"><img src="https://img.shields.io/github/license/xsddz/whozere?style=flat-square" alt="License"></a>
  <a href="https://goreportcard.com/report/github.com/xsddz/whozere"><img src="https://goreportcard.com/badge/github.com/xsddz/whozere?style=flat-square" alt="Go Report Card"></a>
</p>

<p align="center">
  <a href="README.md">English</a> | 中文
</p>

---

## ⚡ 一键安装

```bash
# 一键安装 (macOS/Linux)
curl -fsSL https://raw.githubusercontent.com/xsddz/whozere/main/install.sh | bash

# 或使用 Go 安装
go install github.com/xsddz/whozere/cmd/whozere@latest
```

## ✨ 特性

- 🖥️ **跨平台支持**：macOS、Linux、Windows
- 📡 **多种通知渠道**：
  - 通用 Webhook
  - 钉钉机器人
  - 企业微信机器人
  - Telegram Bot
  - Slack
  - 邮件 (SMTP)
- 🔍 **检测多种登录方式**：SSH、控制台、远程桌面、屏幕共享
- ⚡ **实时监控**：登录即推送
- 🛡️ **轻量级**：资源占用极低

## 🚀 快速开始

### 安装

```bash
# 方式一：一键安装脚本
curl -fsSL https://raw.githubusercontent.com/xsddz/whozere/main/install.sh | bash

# 方式二：Go 安装
go install github.com/xsddz/whozere/cmd/whozere@latest

# 方式三：手动下载
# 从 https://github.com/xsddz/whozere/releases 下载对应平台的二进制文件
```

### 配置

```bash
# 复制示例配置
cp config.example.yaml config.yaml

# 编辑配置文件，启用并配置你需要的通知渠道
vim config.yaml
```

### 配置示例

```yaml
notifiers:
  # 钉钉机器人
  - type: dingtalk
    name: "钉钉告警"
    enabled: true
    config:
      webhook: "https://oapi.dingtalk.com/robot/send?access_token=你的TOKEN"
      secret: "你的加签密钥"  # 可选

  # 企业微信机器人
  - type: wecom
    name: "企微告警"  
    enabled: false
    config:
      webhook: "https://qyapi.weixin.qq.com/cgi-bin/webhook/send?key=你的KEY"

  # Telegram
  - type: telegram
    enabled: false
    config:
      token: "你的BOT_TOKEN"
      chat_id: "你的CHAT_ID"
```

### 运行

```bash
# 测试通知是否正常
whozere -test

# 前台运行
whozere

# 指定配置文件
whozere -config /path/to/config.yaml

# 查看版本
whozere -version
```

## 🔧 作为服务运行

### macOS (launchd)

```bash
# 创建 plist 文件
cat > ~/Library/LaunchAgents/com.whozere.plist << 'EOF'
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>Label</key>
    <string>com.whozere</string>
    <key>ProgramArguments</key>
    <array>
        <string>/usr/local/bin/whozere</string>
        <string>-config</string>
        <string>/usr/local/etc/whozere/config.yaml</string>
    </array>
    <key>RunAtLoad</key>
    <true/>
    <key>KeepAlive</key>
    <true/>
</dict>
</plist>
EOF

# 加载服务
launchctl load ~/Library/LaunchAgents/com.whozere.plist
```

### Linux (systemd)

```bash
# 创建 service 文件
sudo cat > /etc/systemd/system/whozere.service << 'EOF'
[Unit]
Description=whozere - 登录检测与通知服务
After=network.target

[Service]
Type=simple
ExecStart=/usr/local/bin/whozere -config /etc/whozere/config.yaml
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
EOF

# 启用并启动服务
sudo systemctl daemon-reload
sudo systemctl enable whozere
sudo systemctl start whozere
```

### Windows

使用任务计划程序或 [NSSM](https://nssm.cc/) 安装为 Windows 服务：

```cmd
nssm install whozere C:\whozere\whozere.exe -config C:\whozere\config.yaml
nssm start whozere
```

## 🖥️ 平台说明

### macOS

- 使用 `log stream` 监控系统日志
- 检测：loginwindow、sshd、screensharingd 事件
- 无需特殊权限

### Linux

- 监控 `/var/log/auth.log` (Debian/Ubuntu) 或 `/var/log/secure` (RHEL/CentOS)
- 可能需要日志文件读取权限：
  ```bash
  sudo usermod -a -G adm $USER  # Debian/Ubuntu
  ```

### Windows

- 使用 Windows 事件日志 (安全日志, 事件 ID 4624)
- 可能需要管理员权限运行

## 🛠️ 开发

```bash
# 克隆仓库
git clone https://github.com/xsddz/whozere.git
cd whozere

# 运行测试
go test ./...

# 本地构建
make build

# 跨平台构建
make build-all
```

## 📄 许可证

[MIT License](LICENSE)

## 🤝 贡献

欢迎提交 Issue 和 Pull Request！
