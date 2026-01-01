# WinWithIOS

# 📱 Win-iOS USB Link

> 打破 Windows 与 iOS 的生殖隔离。基于 Go 语言和 iOS 快捷指令实现的“物理外挂”。
> 

## ✨ 项目简介

作为一名同时使用 iPhone 和 Windows 10 的用户，我深受跨设备体验割裂的困扰。为了解决 **“验证码跨屏同步”** 和 **“无网环境极速传文件”** 的痛点，我开发了这个轻量级的工具。

它利用 iPhone USB 共享网络的物理链路，配合 Go 服务端，实现了：

- **📨 验证码自动同步**：手机收到短信 -> 电脑右下角弹窗 -> 自动写入电脑剪贴板。
- **⚡ 极速文件传输**：基于 HTTP 的局域网直连，照片/文件秒传，不消耗流量。
- **🤫 无感运行**：无控制台窗口，开机静默自启。

## 🛠️ 技术架构

- **服务端 (Windows)**: Golang (`net/http`, `go-toast`, `atotto/clipboard`)
- **客户端 (iOS)**: iOS 快捷指令 (Shortcuts Automation)
- **网络层**: USB Tethering (虚拟网卡 IP `172.20.10.x`)

## 🚀 快速开始

### 1. 编译服务端

需要安装 [Go](https://go.dev/dl/) 环境。

```
# 克隆仓库
git clone [https://github.com/lymangos/WinWithIOS.git](https://github.com/lymangos/WinWithIOS.git)
cd WinWithIOS

# 初始化依赖
go mod tidy

# 编译为无窗口后台应用
go build -ldflags "-H windowsgui" -o usb-link.exe main.go

```

### 2. 运行与部署

1. **直接运行**：双击 `usb-link.exe`（没有任何窗口弹出是正常的，它已在后台运行）。
2. **开机自启**：
    - 按 `Win + R`，输入 `shell:startup`。
    - 将 `usb-link.exe` 的快捷方式放入该文件夹。

### 3. iOS 端设置 (关键)

### A. 网络连接

使用数据线将 iPhone 连接至 PC，并开启“个人热点” (仅 USB)。确保电脑获取到以太网 IP (通常为 `172.20.10.x`)。

### B. 快捷指令自动化

新建一个 iOS 自动化：

1. **触发器**：当收到“信息”包含“验证码”时。
2. **动作**：获取 URL 内容。
    - **URL**: `http://172.20.10.x:8080/api/sms` (IP 替换为你电脑的 USB 网卡 IP)
    - **Method**: `POST`
    - **JSON Body**: `{"sender": 发件人, "content": 短信内容}`
3. **设置**：关闭“运行前询问”和“运行时通知”。

## 📝 实现过程与日志

### Phase 1: 网络方案的选择

最初尝试使用 **Tailscale** 组网。虽然方便，但在文件传输测试中发现，Tailscale 经常走 Relay 中继模式，传输一张 100KB 的图片耗时数分钟。
**优化**：回归物理链路。利用 iPhone USB 热点形成的虚拟局域网，延迟 <1ms，带宽跑满 USB 2.0/3.0，实现真正的“秒传”。

### Phase 2: 服务端开发 (Go)

选择 Go 语言是因为其对系统 API (Windows API) 的良好支持和极小的编译体积。

- **短信同步**：通过简单的 HTTP POST 接口接收 JSON，正则提取 4-6 位数字验证码，直接写入剪贴板。
- **文件上传**：实现了一个简单的 Multipart 表单上传接口。为了安全，增加了 `filepath.Base` 清洗文件名，防止路径穿越攻击。

### Phase 3: 部署与运维 (DevOps)

为了达到“原生级”体验，不能忍受每次开机都要手动运行一个黑框终端。

- **Session 0 隔离问题**：如果使用 NSSM 注册为标准 Windows 服务，程序将无法与用户桌面的剪贴板和通知中心交互。
- **解决方案**：采用 User Session 后台应用方案。使用 `ldflags "-H windowsgui"` 编译参数隐藏控制台窗口，并放入 Startup 文件夹实现用户登录级自启。

## ⚠️ 注意事项

- USB 热点分配的 IP (`172.20.10.x`) 可能会变动，连接失败时请检查电脑 IP。
- 使用前请确保 Windows 防火墙允许程序通过（专用/公用网络）。

## 📄 License

MIT