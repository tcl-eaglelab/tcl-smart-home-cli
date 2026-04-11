# TCL Smart Home CLI

[![CI](https://github.com/tcl-eaglelab/tcl-smart-home-cli/actions/workflows/ci.yml/badge.svg)](https://github.com/tcl-eaglelab/tcl-smart-home-cli/actions/workflows/ci.yml)
[![Go Version](https://img.shields.io/github/go-mod/go-version/tcl-eaglelab/tcl-smart-home-cli)](https://github.com/tcl-eaglelab/tcl-smart-home-cli)
[![License](https://img.shields.io/github/license/tcl-eaglelab/tcl-smart-home-cli)](LICENSE)

[English](./README.md) | [中文](./README_zh.md)

TCL 智能家居 CLI — 让人类和 AI Agent 都能在终端中控制 TCL 智能家居设备。

单个二进制文件，零依赖，基于 Go 构建。

[安装](#安装) · [AI Agent Skill](#ai-agent-skill) · [认证](#认证) · [命令](#命令) · [安全](#安全与风险提示使用前必读) · [贡献](#贡献)

## 功能

| 能力 | 说明 |
|------|------|
| 🔐 扫码授权 | TCL App 扫码，一次授权，30 天有效 |
| 📱 设备管理 | 查看设备列表、状态、物模型 |
| 🎮 设备控制 | 基于物模型的属性控制，自动过滤只读属性 |
| 🤖 AI 友好 | 所有命令支持 `--json` 结构化输出，支持 `NO_COLOR` |
| 🛡️ 安全校验 | 只读属性自动拦截，防止误操作 |
| ⚡ 零依赖 | 单个二进制文件，下载即用 |

## 安装

### 方式一 — npm 安装（推荐）

```bash
npm install -g @tcl-eaglelab-01/smart-home-cli
```

### 方式二 — 下载二进制

从 [GitHub Releases](https://github.com/tcl-eaglelab/tcl-smart-home-cli/releases) 下载对应平台的文件：

| 平台 | 文件 |
|------|------|
| Linux x64 | `tcl-<version>-linux-amd64.tar.gz` |
| Linux ARM64 | `tcl-<version>-linux-arm64.tar.gz` |
| macOS Intel | `tcl-<version>-darwin-amd64.tar.gz` |
| macOS Apple Silicon | `tcl-<version>-darwin-arm64.tar.gz` |
| Windows x64 | `tcl-<version>-windows-amd64.zip` |

### 方式三 — Go Install

```bash
go install github.com/tcl-eaglelab/tcl-smart-home-cli/cmd/tcl@latest
```

### 方式四 — 从源码编译

```bash
git clone https://github.com/tcl-eaglelab/tcl-smart-home-cli.git
cd tcl-smart-home-cli
go build -o tcl ./cmd/tcl
```

## AI Agent Skill

安装 [Skill](https://www.npmjs.com/package/skills) 以赋予 AI Agent 设备控制能力：

```bash
npx skills add tcl-eaglelab/tcl-smart-home-cli -y -g
```

安装后，AI Agent（OpenClaw、Claude Code、Cursor 等）将自动获得 TCL 设备控制能力。

## 认证

授权分两步，兼容人类和 AI Agent 使用：

**第一步 — 生成二维码链接：**
```bash
tcl init
```
输出二维码链接后立即退出。在浏览器中打开链接，使用 **TCL App** 扫码。

**第二步 — 扫码后验证授权：**
```bash
tcl init --verify <qr_code>
```
`<qr_code>` 值由第一步输出。轮询授权结果，成功后保存凭证。

Token 有效期 30 天，过期前自动刷新。

验证配置：

```bash
tcl doctor
```

## 命令

```
tcl <command> [subcommand] [flags]

命令:
  init           扫码授权配置账号
  doctor         检查配置和连通性
  device list    列出所有设备
  device detail  设备详情和当前状态
  device model   物模型（支持的能力）
  device control 控制设备
```

> 所有命令支持 `--json` 输出结构化数据。

---

### 设备列表

```bash
tcl device list
```

### 设备详情

```bash
tcl device detail <设备ID>
```

### 物模型查询

```bash
tcl device model <设备ID>
```

### 控制设备

```bash
tcl device control <设备ID> '{"powerSwitch":1}'
tcl device control <设备ID> '{"targetTemperature":26}'
tcl device control <设备ID> '{"powerSwitch":1,"workMode":1,"targetTemperature":26}'
```

**Windows PowerShell** — 双引号需要双写：

```powershell
tcl device control <设备ID> "{""powerSwitch"":1}"
```

> 不同设备的物模型完全不同，控制前务必先查看 `tcl device model`。

---

## 常用工作流

```bash
tcl device list                              # 1. 查找设备
tcl device model <id>                        # 2. 查看支持的属性
tcl device detail <id>                       # 3. 查看当前状态
tcl device control <id> '{"powerSwitch":1}'  # 4. 发送控制命令
```

## JSON 输出

```bash
tcl device list --json
tcl device list --json | jq '.[].deviceId'
```

## 环境变量

| 变量 | 说明 |
|------|------|
| `NO_COLOR` | 设置任意值禁用彩色输出（[no-color.org](https://no-color.org)） |

## 配置文件

凭证存储在 `~/.tcl-cli/config.json`：

```json
{
  "accessToken": "eyJhbG...",
  "refreshToken": "eyJhbG...",
  "tokenExpiresAt": 1234567890
}
```

## 常见问题

| 问题 | 解决方案 |
|------|---------|
| 配置文件不存在 | 运行 `tcl init` |
| Token 已过期 | 运行 `tcl init` 重新授权 |
| 设备未找到 | 用 `tcl device list` 确认设备 ID |
| 控制命令失败 | 用 `tcl device model` 确认属性名和访问模式 |
| 设备离线 | 检查设备 WiFi 连接 |

## 安全与风险提示（使用前必读）

本工具可供 AI Agent 调用以自动化控制 TCL 智能家居设备，存在模型幻觉、执行不可控、提示词注入等固有风险。扫码授权后，AI Agent 将以您的用户身份在授权范围内执行操作，可能导致非预期的设备操作。

为降低风险，本工具在发送控制命令前会自动过滤只读属性。但上述风险仍然存在，请确保设备处于安全状态后再允许自动化控制。

使用本工具即视为您已知悉并自愿承担所有相关风险和责任。

## 项目结构

```
tcl-smart-home-cli/
├── cmd/
│   ├── tcl/
│   │   └── main.go      # 入口文件
│   ├── root.go          # 根命令
│   ├── init.go          # 扫码授权
│   ├── doctor.go        # 诊断检查
│   └── device.go        # 设备管理
├── internal/
│   ├── api/             # TCL IoT API 客户端
│   ├── config/          # 配置管理
│   └── output/          # 输出格式化
├── scripts/
│   ├── install.js       # npm postinstall（下载二进制）
│   └── run.js           # npm bin 包装器
├── skills/              # AI Agent Skill
├── package.json         # npm 分发
├── .goreleaser.yml      # 发布自动化
├── Makefile             # 本地开发
└── .github/workflows/   # CI/CD
```

## 许可证

MIT

## 贡献

欢迎提交 Issue 和 Pull Request！请先阅读 [贡献指南](CONTRIBUTING.md)。
