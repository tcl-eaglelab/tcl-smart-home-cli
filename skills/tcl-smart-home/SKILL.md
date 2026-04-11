---
name: tcl-smart-home
version: 1.0.0
description: "TCL 智能家居 CLI — 通过终端控制空调、冰箱、洗衣机等设备。支持设备列表查询、状态查询、物模型查看和属性控制。当用户需要查看 TCL 设备、查询设备状态、控制智能家居设备时触发。"
metadata:
  requires:
    bins: ["tcl"]
  cliHelp: "tcl --help"
---

# TCL Smart Home CLI

通过终端控制 TCL 智能家居设备，支持人类和 AI Agent 使用。

## 前置条件

- 已安装 `tcl` CLI（推荐 `npm install -g @tcl-eaglelab-01/smart-home-cli`，或从 [GitHub Releases](https://github.com/tcl-eaglelab/tcl-smart-home-cli/releases) 下载对应平台二进制文件）
- TCL App 账号（首次需扫码授权）

## 首次配置（需用户交互）

授权分两步，AI Agent 全程不阻塞：

**第一步 — 生成二维码链接：**

```bash
tcl init
```

命令立即输出二维码 URL 和 `--verify` 所需的 QR code 值后退出。

**AI Agent 操作：**
1. 运行 `tcl init`，获取输出
2. 从输出中提取包含 `generateQRCodePic` 的 URL，发送给用户，告知在浏览器中打开
3. 同时提取输出中 `tcl init --verify <qr_code>` 里的 `<qr_code>` 值备用
4. 告知用户使用 **TCL App** 扫码，扫完后通知 AI

**第二步 — 用户扫码后验证授权：**

```bash
tcl init --verify <qr_code>
```

用户确认扫码后，AI Agent 运行此命令。轮询授权结果（最长 60 秒），输出 `Authorization successful` 表示成功。若输出 `Authorization timed out`，重新从第一步开始。

## 命令

### 列出所有设备

```bash
tcl device list --json
```

返回设备数组，包含 `deviceId`、`nickName`、`category`、`productKey`、`isOnline`、`locationName` 和 `family` 信息。

### 查看设备状态

```bash
tcl device detail <设备ID> --json
```

返回设备信息和所有当前属性值（如 `powerSwitch`、`targetTemperature`、`workMode`）。

### 查看设备物模型

```bash
tcl device model <设备ID> --json
```

返回所有属性，包含 `identifier`（标识符）、`name`（名称）、`accessMode`（`r` = 只读，`rw` = 可读写）和 `dataType`（类型、最小值、最大值、步长、单位）。

**控制设备前必须先查看物模型。** 不同品类的设备属性完全不同。

### 控制设备

```bash
# Linux/macOS
tcl device control <设备ID> '{"powerSwitch":1}'

# Windows (PowerShell)
tcl device control <设备ID> "{""powerSwitch"":1}"
```

只有可下发的属性可以控制，只读属性（`accessMode: r`）会被自动过滤。

## 工作流

控制设备的推荐流程：

```
1. tcl device list --json          → 查找目标设备（获取 deviceId）
2. tcl device model <id> --json    → 查看支持的属性和访问模式
3. tcl device detail <id> --json   → 查看当前属性值
4. tcl device control <id> '{...}' → 发送控制命令（只读属性会被自动过滤）
```

**注意：** Windows PowerShell 下 JSON 参数需要用双引号包裹并转义内部双引号：`tcl device control <id> "{""key"":value}"`

## 诊断

```bash
tcl doctor
```

检查配置文件、Token 有效期和 API 连通性。

## 错误处理

| 错误信息 | 原因 | AI Agent 应对 |
|---------|------|--------------|
| `not configured` | 未授权 | 运行 `tcl init` 获取二维码链接发给用户，用户扫码后运行 `tcl init --verify <qr_code>` |
| `token expired` / `token invalid` | Token 已失效 | 运行 `tcl init` 重新授权 |
| `device not found` | 设备 ID 不存在 | 运行 `tcl device list --json` 重新获取设备列表 |
| `Device is offline` | 设备离线 | 告知用户检查设备 WiFi 连接 |
| `No writable properties` | 所有属性均为只读 | 运行 `tcl device model <id> --json` 确认可写属性 |
| `failed to fetch Thing Model` | 网络异常 | 等待后重试，或告知用户检查网络 |
| 其他 API 错误 | 服务端异常 | 运行 `tcl doctor` 诊断，等待后重试 |

## 安全

- 凭证存储在本地 `~/.tcl-cli/config.json`
- 所有 API 请求使用 HTTPS
- Token 有效期 30 天，自动刷新
- 仅访问 TCL 官方 API 端点
