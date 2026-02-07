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

## ✨ 特性

- 🖥️ **跨平台支持**：macOS、Linux、Windows
- 📡 **多种通知渠道**：Webhook、钉钉、飞书、企业微信、Telegram、Slack、邮件
- 🔍 **检测多种登录方式**：SSH、控制台、远程桌面、屏幕共享
- ⚡ **实时监控**：登录即推送
- 🛡️ **轻量级**：资源占用极低

## 🚀 快速开始

```bash
# 1. 安装 (macOS/Linux 一键安装)
curl -fsSL https://raw.githubusercontent.com/xsddz/whozere/main/scripts/install.sh | bash

# 2. 配置
sudo cp /usr/local/etc/whozere/config.example.yaml /usr/local/etc/whozere/config.yaml
sudo vim /usr/local/etc/whozere/config.yaml  # 编辑通知设置

# 3. 测试通知
whozere -config /usr/local/etc/whozere/config.yaml -test

# 4. 安装为服务 (开机自启)
whozere-service install
whozere-service start
```

## 📋 环境要求

- Go 1.21+ (仅源码编译需要)
- macOS 10.15+ / Linux / Windows 10+
- 网络访问权限 (用于发送通知)

## 📦 安装

### 源码编译

```bash
git clone https://github.com/xsddz/whozere.git
cd whozere
go build -o whozere ./cmd/whozere
cp config.example.yaml config.yaml  # 然后编辑 config.yaml
```

### 交叉编译

```bash
# Linux
GOOS=linux GOARCH=amd64 go build -o whozere-linux-amd64 ./cmd/whozere

# Windows
GOOS=windows GOARCH=amd64 go build -o whozere-windows-amd64.exe ./cmd/whozere

# macOS
GOOS=darwin GOARCH=arm64 go build -o whozere-darwin-arm64 ./cmd/whozere
```

## ⚙️ 配置

复制示例配置文件：

```bash
cp config.example.yaml config.yaml
```

### 配置示例

```yaml
notifiers:
  # 通用 Webhook
  - type: webhook
    name: "我的 Webhook"
    enabled: true
    config:
      url: "https://example.com/webhook"

  # 邮件通知
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

> 📝 查看 [config.example.yaml](config.example.yaml) 了解所有通知渠道：
> 钉钉、飞书、企业微信、Telegram、Slack 等。

## 📖 使用方法

```bash
./whozere                           # 使用默认配置运行
./whozere -config /path/config.yaml # 指定配置文件
./whozere -since 1h                 # 检查过去 1 小时的登录 + 监听新登录
./whozere -test                     # 发送测试通知
./whozere -version                  # 显示版本
./whozere -help                     # 显示所有选项
```

<details>
<summary>完整帮助信息</summary>

```
Usage of whozere:
  -config string
        配置文件路径 (默认 "config.yaml")
  -integrity
        启用日志完整性监控 (检测篡改) (默认 true)
  -since duration
        检查指定时间前的登录事件 (如 1h, 30m)
  -test
        发送测试通知后退出
  -version
        显示版本信息
```
</details>

## 📬 通知格式

当检测到登录时，你会收到类似这样的通知：

**文本消息：**
```
🔔 Login Alert

User: alice
Host: my-server
Time: 2026-02-07 20:45:30
Zone: CST (UTC+8)
OS: linux
IP: 192.168.1.100
Terminal: ssh
```

**Webhook JSON 格式：**
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

## 🔧 作为服务运行

安装脚本会自动安装 `whozere-service` 命令。

### 快速配置（推荐）

```bash
whozere-service install   # 自动检测 macOS/Linux
whozere-service start
whozere-service status

# 其他命令
whozere-service stop      # 停止服务
whozere-service restart   # 重启服务
whozere-service uninstall # 删除服务
```

### 手动配置

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
Description=whozere - 登录检测与通知服务
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

<details>
<summary>Windows (NSSM)</summary>

使用任务计划程序或 [NSSM](https://nssm.cc/) 安装为 Windows 服务：

```cmd
nssm install whozere C:\whozere\whozere.exe -config C:\whozere\config.yaml
nssm start whozere
```
</details>

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

## 🔐 安全与检测原理

### 检测流程

```
┌──────────┐     ┌──────────────┐     ┌─────────┐     ┌──────────────┐
│  Login   │ ──▶ │ System Logs  │ ──▶ │ whozere │ ──▶ │ Notification │
│  Event   │     │ (auth/event) │     │ watcher │     │   Channel    │
└──────────┘     └──────────────┘     └─────────┘     └──────────────┘
```

### 日志完整性监控 (仅 Linux)

在 Linux 平台，whozere 会监控认证日志文件是否被篡改：

- **截断检测**：日志文件大小显著减少 (50%+) 会触发告警
- **删除检测**：日志文件被删除会触发告警
- **替换检测**：文件 inode 变化 (文件被替换) 会触发告警
- **权限变更**：文件权限被修改会触发告警

这有助于检测攻击者试图清除入侵痕迹的行为。

### 检测能力边界

whozere 依赖系统日志进行检测，以下情况无法检测：

- 内核级 rootkit（在系统调用层面拦截）
- 攻击者在登录前已禁用日志
- 绕过标准认证的攻击（如内核漏洞利用）

## 🗑️ 卸载

```bash
# 快速卸载（通过安装脚本安装的）
whozere-uninstall

# 或一键远程卸载
curl -fsSL https://raw.githubusercontent.com/xsddz/whozere/main/scripts/uninstall.sh | bash
```

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

## � 许可证

[MIT License](LICENSE)

## 🤝 贡献

欢迎提交 Issue 和 Pull Request！
