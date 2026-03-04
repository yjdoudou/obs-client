# <div align="center">
  <img src="frontend/src/assets/images/logo-universal.png" alt="OBS Client Logo" width="120" height="120">
  
  # 华为 OBS 客户端工具
  
  **一款现代化的华为云 OBS 对象存储管理工具**
  
  [![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go)](https://golang.org/)
  [![Vue Version](https://img.shields.io/badge/Vue-3.0+-4FC08D?style=flat&logo=vue.js)](https://vuejs.org/)
  [![Wails](https://img.shields.io/badge/Wails-2.0+-3498db?style=flat)](https://wails.io/)
  [![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)
  [![Platform](https://img.shields.io/badge/Platform-Windows%20%7C%20macOS%20%7C%20Linux-lightgrey.svg)](https://wails.io/docs/introduction/installation/windows)
  
  [English](README_EN.md) | 简体中文
</div>

---

## <div align="center">✨ 项目简介</div>

华为 OBS 客户端工具是一款基于 Wails 框架开发的跨平台桌面应用，专为华为云 OBS（对象存储服务）用户打造。它提供了直观的图形化界面，让您能够轻松管理 OBS 资源，无需依赖命令行工具或 Web 控制台。

### 🎯 核心特性

- 🚀 **跨平台支持** - 支持 Windows、macOS 和 Linux 三大操作系统
- 🎨 **现代化 UI** - 采用玻璃拟态设计，支持深色/浅色主题切换
- 📁 **虚拟文件夹** - 基于分隔符自动构建层级文件结构，像管理本地文件一样管理 OBS 对象
- 🔗 **多连接管理** - 支持管理多个 OBS 连接，快速切换不同账户
- ⚡ **高性能** - 基于 Go 语言和 Wails 框架，轻量级、启动快、响应迅速
- 🔐 **安全可靠** - 本地加密存储敏感信息，支持断点续传
- 📦 **批量操作** - 支持批量上传、下载、删除文件
- 🔍 **智能搜索** - 快速搜索桶和对象
- 🎯 **自定义窗口** - Frameless 窗口设计，自定义标题栏，完美融入系统

---

## <div align="center">📸 功能展示</div>

### 连接管理
- ✅ 创建、编辑、删除 OBS 连接
- ✅ 测试连接有效性
- ✅ 实时显示连接状态
- ✅ 支持自定义 Endpoint 和分隔符

### 资源管理
- ✅ 桶的查看、创建、删除
- ✅ 对象的上传、下载、删除、复制、移动
- ✅ 虚拟文件夹层级浏览
- ✅ 文件拖拽上传
- ✅ 批量操作支持
- ✅ 地址栏直接访问

### 传输管理
- ✅ 实时显示传输进度
- ✅ 支持断点续传
- ✅ 传输任务队列管理
- ✅ 传输历史记录

### 用户体验
- ✅ 深色/浅色主题切换
- ✅ 自适应布局
- ✅ 快捷键支持
- ✅ 右键菜单操作
- ✅ 文件详情查看

---

## <div align="center">🛠️ 技术栈</div>

### 后端
- **Go 1.21+** - 高性能后端语言
- **Wails 2.0+** - 跨平台桌面应用框架
- **华为云 OBS Go SDK** - OBS 服务集成
- **SQLite** - 本地数据存储

### 前端
- **Vue 3** - 渐进式 JavaScript 框架
- **TypeScript** - 类型安全的 JavaScript 超集
- **Element Plus** - 基于 Vue 3 的组件库
- **TailwindCSS** - 实用优先的 CSS 框架
- **Pinia** - Vue 3 状态管理
- **Vue Router** - 官方路由管理器

### 构建工具
- **Vite 5.0+** - 下一代前端构建工具
- **PostCSS** - CSS 转换工具

---

## <div align="center">📦 安装指南</div>

### 前置要求

在安装本应用之前，请确保您的系统已安装以下依赖：

#### Windows
- Windows 10 或更高版本（64 位）
- [Go 1.21+](https://golang.org/dl/)
- [Node.js 18+](https://nodejs.org/)
- [Git](https://git-scm.com/)

#### macOS
- macOS 10.15 (Catalina) 或更高版本
- [Go 1.21+](https://golang.org/dl/)
- [Node.js 18+](https://nodejs.org/)
- [Xcode Command Line Tools](https://developer.apple.com/xcode/)

#### Linux
- Ubuntu 20.04+ / CentOS 8+ 或其他主流发行版（64 位）
- [Go 1.21+](https://golang.org/dl/)
- [Node.js 18+](https://nodejs.org/)
- [GTK3 开发库](https://www.gtk.org/)

### 从源码构建

#### 1. 克隆项目

```bash
git clone https://github.com/yourusername/obs-client.git
cd obs-client
```

#### 2. 安装 Wails CLI

```bash
go install github.com/wailsapp/wails/v2/cmd/wails@latest
```

#### 3. 安装前端依赖

```bash
cd frontend
npm install
cd ..
```

#### 4. 开发模式运行

```bash
wails dev
```

#### 5. 构建生产版本

```bash
wails build
```

构建完成后，可执行文件将位于项目根目录下：
- Windows: `obs-client.exe`
- macOS: `obs-client.app`
- Linux: `obs-client`

### 下载预编译版本

我们提供预编译的安装包，您可以从 [Releases](https://github.com/yourusername/obs-client/releases) 页面下载对应平台的安装包。

---

## <div align="center">🚀 快速开始</div>

### 1. 创建连接

1. 点击左侧边栏的"新建连接"按钮
2. 填写连接信息：
   - **连接名称**：为您的连接起一个易记的名称
   - **Access Key ID**：华为云 OBS 的 Access Key ID
   - **Secret Access Key**：华为云 OBS 的 Secret Access Key
   - **区域**：选择您的 OBS 服务所在区域
   - **Endpoint**（可选）：自定义 OBS 服务地址
3. 点击"测试连接"验证配置是否正确
4. 点击"保存"完成连接创建

### 2. 管理桶和对象

- **浏览桶**：在左侧连接树中展开连接，查看所有桶
- **浏览对象**：点击桶或文件夹，在中间区域查看对象列表
- **上传文件**：点击"上传文件"按钮或拖拽文件到文件列表区域
- **下载文件**：右键点击文件，选择"下载"
- **删除对象**：选中文件后点击"删除"按钮或使用右键菜单

### 3. 使用虚拟文件夹

OBS 对象存储使用扁平结构，但本工具通过分隔符（默认为 `/`）自动构建虚拟文件夹层级，让您可以像管理本地文件一样管理 OBS 对象。

- **导航**：点击文件夹进入下一级目录
- **返回**：使用工具栏的返回按钮或地址栏导航
- **直接访问**：在地址栏输入完整的对象 Key 直接访问

---

## <div align="center">📖 项目结构</div>

```
obs-client/
├── cmd/                      # 主入口
│   └── obs-client/           # 应用主入口
│       └── main.go           # Wails 应用主函数
├── internal/                 # Go 后端内部代码
│   ├── app/                  # Wails 应用结构
│   │   └── app.go            # 应用核心逻辑
│   ├── connection/          # 连接管理模块
│   │   ├── manager.go        # 连接管理器
│   │   └── model.go          # 连接数据模型
│   ├── db/                   # 数据库模块
│   │   ├── connection.go     # 连接数据库操作
│   │   └── models.go         # 数据库模型
│   └── obs/                  # OBS 客户端模块
│       └── client.go         # OBS 客户端封装
├── frontend/                 # 前端代码
│   ├── src/
│   │   ├── assets/           # 静态资源
│   │   │   ├── fonts/       # 字体文件
│   │   │   └── images/       # 图片资源
│   │   ├── components/       # Vue 组件
│   │   │   ├── ConnectionDialog.vue  # 连接对话框
│   │   │   ├── TitleBar.vue          # 自定义标题栏
│   │   │   └── TransferPanel.vue     # 传输面板
│   │   ├── router/           # 路由配置
│   │   ├── store/            # 状态管理
│   │   │   ├── connection.ts # 连接状态
│   │   │   └── transfer.ts   # 传输状态
│   │   ├── views/            # 页面视图
│   │   │   └── Home.vue      # 主页
│   │   ├── App.vue           # 根组件
│   │   ├── main.ts           # 入口文件
│   │   └── style.css         # 全局样式
│   ├── wailsjs/              # Wails 生成的绑定代码
│   │   ├── go/               # Go 方法绑定
│   │   │   └── app/
│   │   │       ├── App.d.ts  # TypeScript 类型定义
│   │   │       └── App.js    # JavaScript 绑定
│   │   ├── models.ts         # 数据模型类型定义
│   │   └── runtime/          # Wails 运行时绑定
│   ├── index.html            # HTML 模板
│   ├── package.json          # Node.js 依赖
│   ├── tailwind.config.js    # TailwindCSS 配置
│   ├── tsconfig.json         # TypeScript 配置
│   └── vite.config.ts        # Vite 配置
├── build/                    # 构建配置
│   ├── appicon.png           # 应用图标
│   ├── darwin/               # macOS 构建配置
│   └── windows/              # Windows 构建配置
├── main.go                   # Wails 应用入口
├── wails.json                # Wails 项目配置
├── go.mod                    # Go 模块定义
├── go.sum                    # Go 依赖锁定文件
├── .gitignore                # Git 忽略文件
└── README.md                 # 项目说明文档
```

---

## <div align="center">⚙️ 配置说明</div>

### Wails 配置

项目配置文件为 `wails.json`，包含以下主要配置项：

```json
{
  "name": "obs-client",
  "outputfilename": "obs-client",
  "frontend:install": "npm install",
  "frontend:build": "npm run build",
  "frontend:dev:watcher": "npm run dev",
  "frontend:dev:serverUrl": "auto",
  "author": {
    "name": "Your Name",
    "email": "your.email@example.com"
  }
}
```

### 窗口配置

窗口配置位于 `main.go` 文件中：

```go
err := wails.Run(&options.App{
    Title:  "obs-client",
    Width:  1200,
    Height: 800,
    Frameless: true,
    BackgroundColour: &options.RGBA{R: 27, G: 38, B: 54, A: 0},
    Windows: &windows.Options{
        WebviewIsTransparent:              true,
        WindowIsTranslucent:               true,
        BackdropType:                      windows.Mica,
        DisableWindowIcon:                 false,
        DisableFramelessWindowDecorations: false,
        DisableResize:                     false,
    },
})
```

---

## <div align="center">🤝 贡献指南</div>

我们欢迎所有形式的贡献！如果您想为项目做出贡献，请遵循以下步骤：

### 1. Fork 项目

点击 GitHub 页面右上角的 "Fork" 按钮，将项目 Fork 到您的账户下。

### 2. 创建分支

```bash
git checkout -b feature/your-feature-name
```

### 3. 提交更改

```bash
git add .
git commit -m "Add some feature"
```

### 4. 推送到分支

```bash
git push origin feature/your-feature-name
```

### 5. 提交 Pull Request

在 GitHub 上创建 Pull Request，描述您的更改内容。

### 代码规范

- **Go 代码**：遵循 [Effective Go](https://golang.org/doc/effective_go) 和 [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- **Vue/TypeScript 代码**：遵循 [Vue 风格指南](https://vuejs.org/style-guide/)
- **提交信息**：使用清晰的提交信息，遵循 [Conventional Commits](https://www.conventionalcommits.org/) 规范

---

## <div align="center">📝 开发计划</div>

### 已完成 ✅
- [x] 连接管理功能
- [x] 桶和对象基本操作
- [x] 虚拟文件夹层级管理
- [x] 地址栏路径直接访问
- [x] 基本的上传下载功能
- [x] 自定义 TitleBar（Frameless 窗口）
- [x] 深色/浅色主题切换
- [x] 传输进度显示
- [x] 批量操作功能

### 进行中 🚧
- [ ] 断点续传功能
- [ ] 高级搜索功能
- [ ] 文件预览功能
- [ ] 操作历史记录

### 计划中 📋
- [ ] 性能优化
- [ ] 多语言支持（i18n）
- [ ] 快捷键支持
- [ ] 插件系统
- [ ] 支持多云存储（AWS S3、阿里云 OSS 等）
- [ ] 版本管理功能
- [ ] 生命周期管理
- [ ] 数据迁移工具

---

## <div align="center">🐛 问题反馈</div>

如果您在使用过程中遇到任何问题或有任何建议，请通过以下方式反馈：

- 📧 **邮件**：your.email@example.com
- 🐛 **GitHub Issues**：[提交问题](https://github.com/yourusername/obs-client/issues)
- 💬 **Discussions**：[参与讨论](https://github.com/yourusername/obs-client/discussions)

在提交问题时，请提供以下信息：
- 操作系统和版本
- 应用版本号
- 详细的错误描述
- 复现步骤
- 相关的日志或截图

---

## <div align="center">📄 许可证</div>

本项目采用 [MIT 许可证](LICENSE) 开源。

```
MIT License

Copyright (c) 2024 Your Name

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
```

---

## <div align="center">🙏 致谢</div>

感谢以下开源项目：

- [Wails](https://wails.io/) - 跨平台桌面应用框架
- [Vue.js](https://vuejs.org/) - 渐进式 JavaScript 框架
- [Element Plus](https://element-plus.org/) - Vue 3 组件库
- [TailwindCSS](https://tailwindcss.com/) - 实用优先的 CSS 框架
- [华为云 OBS Go SDK](https://support.huaweicloud.com/sdkreference-obs/obs_02_0001.html) - OBS 服务 SDK

---

## <div align="center">⭐ Star History</div>

如果这个项目对您有帮助，请给我们一个 Star ⭐️

[![Star History Chart](https://api.star-history.com/svg?repos=yourusername/obs-client&type=Date)](https://star-history.com/#yourusername/obs-client&Date)

---

## <div align="center">📮 联系我们</div>

- **作者**：Your Name
- **邮箱**：your.email@example.com
- **GitHub**：[https://github.com/yourusername](https://github.com/yourusername)
- **项目主页**：[https://github.com/yourusername/obs-client](https://github.com/yourusername/obs-client)

---

<div align="center">
  <sub>Built with ❤️ using Wails and Vue 3</sub>
</div>
