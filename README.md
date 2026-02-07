<p align="center">
  <img src="docs/logo.svg" alt="whozere" width="120">
</p>

<h1 align="center">whozere</h1>

<p align="center">
  <strong>Who's here? 🔔</strong> — A cross-platform login detection and notification tool
</p>

<p align="center">
  <a href="https://github.com/xsddz/whozere/releases"><img src="https://img.shields.io/github/v/release/xsddz/whozere?style=flat-square" alt="Release"></a>
  <a href="https://github.com/xsddz/whozere/blob/main/LICENSE"><img src="https://img.shields.io/github/license/xsddz/whozere?style=flat-square" alt="License"></a>
  <a href="https://goreportcard.com/report/github.com/xsddz/whozere"><img src="https://goreportcard.com/badge/github.com/xsddz/whozere?style=flat-square" alt="Go Report Card"></a>
</p>

<p align="center">
  English | <a href="README.zh-CN.md">中文</a>
</p>

---

## ✨ Features

- 🖥️ **Cross-platform**: macOS, Linux, Windows
- 📡 **Multiple notification channels**: Webhook, DingTalk, WeCom, Telegram, Slack, Email
- 🔍 **Detects various login types**: SSH, Console/TTY, RDP, VNC
- ⚡ **Real-time monitoring**: Instant notifications when someone logs in
- 🛡️ **Lightweight**: Minimal resource usage

## 🚀 Quick Start

```bash
# 1. Install (one-line for macOS/Linux)
curl -fsSL https://raw.githubusercontent.com/xsddz/whozere/main/scripts/install.sh | bash

# 2. Configure
sudo cp /usr/local/etc/whozere/config.example.yaml /usr/local/etc/whozere/config.yaml
sudo vim /usr/local/etc/whozere/config.yaml  # Edit your notification settings

# 3. Test notification
whozere -config /usr/local/etc/whozere/config.yaml -test

# 4. Run
whozere -config /usr/local/etc/whozere/config.yaml
```

## 📋 Requirements

- Go 1.21+ (for building from source)
- macOS 10.15+ / Linux / Windows 10+
- Network access to notification services

## 📦 Installation

### From Source

```bash
git clone https://github.com/xsddz/whozere.git
cd whozere
go build -o whozere ./cmd/whozere
cp config.example.yaml config.yaml  # Then edit config.yaml
```

### Cross-compilation

```bash
# Linux
GOOS=linux GOARCH=amd64 go build -o whozere-linux-amd64 ./cmd/whozere

# Windows
GOOS=windows GOARCH=amd64 go build -o whozere-windows-amd64.exe ./cmd/whozere

# macOS
GOOS=darwin GOARCH=arm64 go build -o whozere-darwin-arm64 ./cmd/whozere
```

## ⚙️ Configuration

Copy `config.example.yaml` to `config.yaml`:

```bash
cp config.example.yaml config.yaml
```

### Example

```yaml
notifiers:
  # Generic Webhook
  - type: webhook
    name: "My Webhook"
    enabled: true
    config:
      url: "https://example.com/webhook"

  # DingTalk Robot
  - type: dingtalk
    name: "DingTalk Alert"
    enabled: false
    config:
      webhook: "https://oapi.dingtalk.com/robot/send?access_token=YOUR_TOKEN"
      secret: ""  # optional

  # WeCom (企业微信)
  - type: wecom
    enabled: false
    config:
      webhook: "https://qyapi.weixin.qq.com/cgi-bin/webhook/send?key=YOUR_KEY"

  # Telegram
  - type: telegram
    enabled: false
    config:
      token: "YOUR_BOT_TOKEN"
      chat_id: "YOUR_CHAT_ID"

  # Slack
  - type: slack
    enabled: false
    config:
      webhook: "https://hooks.slack.com/services/YOUR/WEBHOOK/URL"

  # Email
  - type: email
    enabled: false
    config:
      smtp_host: "smtp.example.com"
      smtp_port: "587"
      username: "your@email.com"
      password: "your_password"
      from: "whozere@example.com"
      to: "admin@example.com"
```

## 📖 Usage

```bash
./whozere                           # Run with default config
./whozere -config /path/config.yaml # Specify config file
./whozere -test                     # Send test notification
./whozere -version                  # Show version
```

## 📬 Notification Format

When a login is detected, you'll receive a notification like this:

**Text Message:**
```
🔔 Login Alert

User: alice
Host: my-server
Time: 2026-02-07 20:45:30
OS: linux
IP: 192.168.1.100
Terminal: ssh
```

**Webhook JSON Payload:**
```json
{
  "event": "login",
  "username": "alice",
  "hostname": "my-server",
  "ip": "192.168.1.100",
  "terminal": "ssh",
  "timestamp": "2026-02-07T20:45:30+08:00",
  "os": "linux",
  "message": "🔔 Login Alert\n\nUser: alice\n..."
}
```

## 🔧 Running as a Service

Use the `scripts/service.sh` helper or configure manually.

### Quick Setup (Recommended)

```bash
# Download service script
curl -fsSL https://raw.githubusercontent.com/xsddz/whozere/main/scripts/service.sh -o service.sh
chmod +x service.sh

# Install and start service
./service.sh install   # Auto-detects macOS/Linux
./service.sh start
./service.sh status

# Other commands: stop, restart, uninstall
```

### Manual Setup

<details>
<summary>macOS (launchd)</summary>

```bash
cat > ~/Library/LaunchAgents/com.whozere.agent.plist << 'EOF'
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>Label</key><string>com.whozere.agent</string>
    <key>ProgramArguments</key>
    <array>
        <string>/usr/local/bin/whozere</string>
        <string>-config</string>
        <string>/usr/local/etc/whozere/config.yaml</string>
    </array>
    <key>RunAtLoad</key><true/>
    <key>KeepAlive</key><true/>
</dict>
</plist>
EOF

launchctl load ~/Library/LaunchAgents/com.whozere.agent.plist
```
</details>

<details>
<summary>Linux (systemd)</summary>

```bash
sudo tee /etc/systemd/system/whozere.service << 'EOF'
[Unit]
Description=whozere - Login Detection & Notification
After=network.target

[Service]
Type=simple
ExecStart=/usr/local/bin/whozere -config /usr/local/etc/whozere/config.yaml
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
EOF

sudo systemctl enable --now whozere
```
</details>

### Windows

Use [NSSM](https://nssm.cc/):

```cmd
nssm install whozere C:\whozere\whozere.exe -config C:\whozere\config.yaml
nssm start whozere
```

## 🖥️ Platform Notes

| Platform | Method | Notes |
|----------|--------|-------|
| **macOS** | `log stream` | Monitors loginwindow, sshd, screensharingd |
| **Linux** | Log files | `/var/log/auth.log` or `/var/log/secure` |
| **Windows** | Event Log | Security Log, Event ID 4624 |

## 🗑️ Uninstall

```bash
# One-line uninstall
curl -fsSL https://raw.githubusercontent.com/xsddz/whozere/main/scripts/uninstall.sh | bash

# Or manually
sudo rm /usr/local/bin/whozere
sudo rm -rf /usr/local/etc/whozere
./scripts/service.sh uninstall  # Remove service
```

## 🛠️ Development

```bash
go test ./...        # Run tests
make build           # Build binary
make build-all       # Cross-platform build
```

## 📜 License

[MIT License](LICENSE)

## 🤝 Contributing

Contributions are welcome! Please feel free to submit a Pull Request.
