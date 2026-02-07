<p align="center">
  <img src="docs/logo.svg" alt="whozere" width="120">
</p>

<h1 align="center">whozere</h1>

<p align="center">
  <strong>Who's here? 🔔</strong> — 跨平台登录检测与通知工具
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
- 📡 **Multiple notification channels**: Webhook, DingTalk, WeCom, Telegram, Slack, Email
- 🔍 **Detects various login types**: SSH, Console, RDP, VNC
- ⚡ **Real-time monitoring**: Instant notifications
- 🛡️ **Lightweight**: Minimal resource usage

## 🚀 Quick Start

```bash
# 1. Download and install
curl -fsSL https://raw.githubusercontent.com/xsddz/whozere/main/install.sh | bash

# 2. Configure
cp /usr/local/etc/whozere/config.example.yaml /usr/local/etc/whozere/config.yaml
# Edit config.yaml with your notification settings

# 3. Test notification
whozere -test

# 4. Run
whozere
```

## 📖 Documentation

See [README_en.md](README_en.md) for detailed documentation.

查看 [README_zh.md](README_zh.md) 获取详细中文文档。

## 📜 License

[MIT License](LICENSE)
