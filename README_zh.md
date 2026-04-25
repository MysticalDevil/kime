<h1 align="center">kime</h1>

<p align="center">
  <a href="https://go.dev"><img src="https://img.shields.io/badge/Go-1.26.2%2B-00ADD8?logo=go" alt="Go"></a>
  <a href="https://www.kimi.com/code"><img src="https://img.shields.io/badge/Kimi-Code%20Console-5B5B5B" alt="Kimi"></a>
  <a href="./LICENSE"><img src="https://img.shields.io/badge/License-BSD--3--Clause-blue" alt="License"></a>
</p>

> [!IMPORTANT]
> 本项目使用了 `encoding/json/v2`，编译时**必须**携带 `GOEXPERIMENT=jsonv2`。

一个在终端中展示 **Kimi Code 控制台** 数据的精美 CLI 工具。

---

## 功能

- **本周用量** – 每次实时请求 API
- **频限明细** – 每次实时请求 API
- **额度使用** – 每次实时请求 API
- **当前套餐 & 有效期** – 本地缓存到套餐有效期截止日
- **模型权限** – 本地缓存到套餐有效期截止日
- 使用 [Lipgloss](https://github.com/charmbracelet/lipgloss) 绘制的 Unicode 圆角边框与彩色 UI
- 在非现代终端中自动降级为 ASCII 渲染，并支持环境变量覆盖
- 多语言输出：**简体中文（默认）**、繁体中文、英文和日文
- Mock 模式，测试时不触发真实 API 请求

---

## 安装

### 通过 `go install`

```bash
GOEXPERIMENT=jsonv2 go install github.com/MysticalDevil/kime@latest
```

### 通过 `mise`

```bash
# 使用 Go backend
mise use -g go:github.com/MysticalDevil/kime@latest

# 或使用 GitHub backend（预编译二进制）
mise use -g github:MysticalDevil/kime@latest
```

### 源码构建

```bash
git clone https://github.com/MysticalDevil/kime.git
cd kime
go mod tidy
GOEXPERIMENT=jsonv2 go build -o kime
```

将二进制文件移动到 `$PATH` 中的目录：

```bash
mv kime ~/.local/bin/
```

---

## 配置

`kime` 从 `~/.config/kime/config.json` 读取凭证（可手动创建，也可通过浏览器自动提取）。

### 交互式配置

最简单的配置方式是使用内置的交互式向导：

```bash
kime init
```

向导会提示你输入 token，并自动从 JWT 中解析 `device_id`、`session_id` 和 `user_id`。你还可以设置偏好语言等选项。

### 如何获取凭证（开发者工具）

1. 打开 [https://www.kimi.com/code/console?from=kfc_overview_topbar](https://www.kimi.com/code/console?from=kfc_overview_topbar) 并登录。
2. 按 `F12` 或 `Ctrl+Shift+I` 打开**开发者工具**。
3. 切换到 **Console（控制台）** 标签页，执行：

   ```javascript
   copy(localStorage.getItem('access_token'))
   ```

   这会将你的 JWT token 复制到剪贴板，粘贴到配置文件的 `token` 字段即可。

4.（可选）如果你想手动填写其余字段，可以将 token 粘贴到 [jwt.io](https://jwt.io) 解码，或在控制台执行：

   ```javascript
   const parts = localStorage.getItem('access_token').split('.');
   const payload = JSON.parse(atob(parts[1].replace(/-/g, '+').replace(/_/g, '/')));
   console.log('device_id:', payload.device_id);
   console.log('session_id (ssid):', payload.ssid);
   console.log('user_id (sub):', payload.sub);
   ```

   `kime` 会自动从 JWT 中解析 `device_id`、`session_id` 和 `user_id`，因此通常只需提供 `token` 即可。

### 配置文件示例

```json
{
  "token": "eyJhbGciOiJIUzUxMiIs...",
  "device_id": "1234567890123456789",
  "session_id": "9876543210987654321",
  "user_id": "your_user_id_here",
  "language": "zh",
  "show_progress": false
}
```

| 字段 | 说明 |
|------|------|
| `token` | JWT access token（`kimi-auth` Cookie 或 LocalStorage 中的 `access_token`） |
| `device_id` | `x-msh-device-id` 请求头值（若省略，自动从 JWT 提取） |
| `session_id` | `x-msh-session-id` 请求头值（若省略，自动从 JWT 提取） |
| `user_id` | `x-traffic-id` 请求头值，即用户 ID（若省略，自动从 JWT 提取） |
| `language` | 界面语言：`"zh"`（默认）、`"zh_TW"`、`"en"`、`"ja"` |
| `show_progress` | 设为 `true` 将用量卡片显示为进度条，而非纯数字 |

### 环境变量（覆盖配置文件）

| 变量 | 说明 |
|------|------|
| `KIME_TOKEN` | JWT token |
| `KIME_DEVICE_ID` | 设备 ID |
| `KIME_SESSION_ID` | 会话 ID |
| `KIME_USER_ID` | 用户 ID |
| `KIME_LANG` | 界面语言：`zh`、`zh_TW`、`en`、`ja` |
| `KIME_RENDER_MODE` | 渲染模式：`auto`（默认）、`unicode` 或 `ascii` |
| `KIME_MOCK` | 设为 `1` 开启 Mock 模式（不请求真实 API） |
| `KIME_FORCE_REFRESH` | 设为 `1` 强制刷新全部内容并更新缓存 |

如果 `device_id` 或 `user_id` 缺失，`kime` 会自动尝试从 JWT payload 中解码提取。

ASCII 渲染会使用英文标签和纯 ASCII 装饰，以便在非 UTF-8 终端中保持可读。

当配置文件和环境变量同时存在时，环境变量优先。

---

## 使用

```bash
# 查看帮助
kime
kime --help

# 查询数据（中文界面，默认）
kime check

# 英文界面
KIME_LANG=en kime check

# Mock 模式（不发起网络请求）
KIME_MOCK=1 kime check

# 强制 ASCII 渲染
KIME_RENDER_MODE=ascii kime check

# 强制刷新（跳过缓存并重新写入）
KIME_FORCE_REFRESH=1 kime check
```

---

## 缓存

- **缓存文件**: `~/.cache/kime/membership.json`
- **有效期**: 到 `subscription.currentEndTime`（当前套餐有效期截止日）
- "当前套餐"、"有效期" 和 "模型权限" 在套餐有效期内直接读取本地缓存。
- "本周用量"、"频限明细" 和 "额度使用" 始终实时请求。
- 设置 `KIME_FORCE_REFRESH=1` 可跳过缓存，强制全量更新。

---

## 许可证

BSD 3-Clause License。详见 [LICENSE](./LICENSE)。
