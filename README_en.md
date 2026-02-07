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
  <a href="README.md">English</a> | <a href="README_zh.md">中文</a>
</p>

---

## ⚡ Quick Install

```bash
# One-line install (macOS/Linux)
curl -fsSL https://raw.githubusercontent.com/xsddz/whozere/main/install.sh | bash

# Or with Go
go install github.com/xsddz/whozere/cmd/whozere@latest
```

## ✨ Features

- 🖥️ **Cross-platform**: macOS, Linux, Windows
- 📡 **Multiple notification channels**: 
  - Generic Webhook
  - DingTalk (钉钉)
  - WeCom (企业微信)
  - Telegram
  - Slack
  - Email (SMTP)
- 🔍 **Detects various login types**:
  - SSH logins
  - Console/TTY logins
  - Remote Desktop (RDP)
  - Screen Sharing (VNC)
- ⚡ **Real-time monitoring**: Instant notifications when someone logs in
- 🛡️ **Lightweight**: Minimal resource usage

## Installation

### From Source

```bash
# Clone the repository
git clone https://github.com/xsddz/whozere.git
cd whozere

# Build
go build -o whozere ./cmd/whozere

# Or install to $GOPATH/bin
go install ./cmd/whozere
```

### Cross-compilation

```bash
# Linux
GOOS=linux GOARCH=amd64 go build -o whozere-linux-amd64 ./cmd/whozere

# Windows
GOOS=windows GOARCH=amd64 go build -o whozere-windows-amd64.exe ./cmd/whozere

# macOS Intel
GOOS=darwin GOARCH=amd64 go build -o whozere-darwin-amd64 ./cmd/whozere

# macOS Apple Silicon
GOOS=darwin GOARCH=arm64 go build -o whozere-darwin-arm64 ./cmd/whozere
```

## Configuration

Copy `config.example.yaml` to `config.yaml` and edit it:

```bash
cp config.example.yaml config.yaml
```

### Example Configuration

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
      secret: ""  # optional, for signed mode

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

## Usage

```bash
# Run with default config (config.yaml)
./whozere

# Specify config file
./whozere -config /path/to/config.yaml

# Test notifications
./whozere -test

# Show version
./whozere -version
```

## Running as a Service

### macOS (launchd)

Create `~/Library/LaunchAgents/com.whozere.plist`:

```xml
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
        <string>/etc/whozere/config.yaml</string>
    </array>
    <key>RunAtLoad</key>
    <true/>
    <key>KeepAlive</key>
    <true/>
</dict>
</plist>
```

Load the service:

```bash
launchctl load ~/Library/LaunchAgents/com.whozere.plist
```

### Linux (systemd)

Create `/etc/systemd/system/whozere.service`:

```ini
[Unit]
Description=whozere - Login Detection & Notification
After=network.target

[Service]
Type=simple
ExecStart=/usr/local/bin/whozere -config /etc/whozere/config.yaml
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
```

Enable and start:

```bash
sudo systemctl enable whozere
sudo systemctl start whozere
```

### Windows (Task Scheduler or NSSM)

Use Task Scheduler to run at startup, or use [NSSM](https://nssm.cc/) to install as a Windows service:

```cmd
nssm install whozere C:\path\to\whozere.exe -config C:\path\to\config.yaml
nssm start whozere
```

## Platform-Specific Notes

### macOS

- Uses `log stream` to monitor system logs
- Requires no special permissions for basic usage
- Detects: loginwindow, sshd, screensharingd events

### Linux

- Monitors `/var/log/auth.log` (Debian/Ubuntu) or `/var/log/secure` (RHEL/CentOS)
- May require read access to log files:
  ```bash
  sudo usermod -a -G adm $USER  # Debian/Ubuntu
  ```

### Windows

- Uses Windows Event Log (Security Log, Event ID 4624)
- May require running as Administrator for full access

## Development

```bash
# Run tests
go test ./...

# Run with race detector
go run -race ./cmd/whozere

# Build for all platforms
make build-all
```

## License

MIT License

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.
