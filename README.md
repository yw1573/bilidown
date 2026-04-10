# Bilidown

[![GitHub Release](https://img.shields.io/github/v/release/iuroc/bilidown)](https://github.com/iuroc/bilidown/releases)

哔哩哔哩视频解析下载工具，支持 8K 视频、Hi-Res 音频、杜比视界下载、批量解析，可扫码登录。

## 支持解析的链接类型

- 【单个视频】https://www.bilibili.com/video/BV1LLDCYJEU3/
- 【番剧和影视剧】https://www.bilibili.com/bangumi/play/ss48831
- 【视频合集】https://space.bilibili.com/282565107/channel/collectiondetail?sid=1427135
- 【收藏夹】https://space.bilibili.com/1176277996/favlist?fid=1234122612
- 【UP 主空间地址】等待 3.x 版本支持

## 使用说明

1. 从 [Releases](https://github.com/iuroc/bilidown/releases) 下载适合您系统版本的安装包
2. 安装 [FFmpeg](https://www.ffmpeg.org/) 并添加到系统 PATH
3. 运行程序，浏览器将自动打开操作界面

## 软件特色

- 前端采用 [Bootstrap](https://github.com/twbs/bootstrap) 和 [VanJS](https://github.com/vanjs-org/van) 构建，轻量美观
- 后端使用 Go 语言开发，SQLite 数据库，单文件部署
- 静态资源嵌入二进制，无需额外配置
- 通过 [p-queue](https://github.com/sindresorhus/p-queue) 控制并发请求，加快批量解析速度

## 本地构建

### 环境要求

- Go 1.22+
- Node.js 18+ & pnpm
- FFmpeg（系统 PATH）
- GCC（CGO 编译）

### 使用 Makefile

```bash
# 克隆项目
git clone https://github.com/iuroc/bilidown
cd bilidown

# 安装依赖并构建
make build

# 输出: build/bilidown.exe (Windows) 或 build/bilidown (Linux/macOS)
```

### 其他命令

```bash
make install     # 安装依赖
make dev         # 开发模式（前端热更新 + 后端运行）
make run         # 构建并运行
make clean       # 清理构建产物
make fmt         # 格式化 Go 代码
make test        # 运行测试
```

## 发布流程

项目使用 GitHub Actions + GoReleaser 自动构建发布：

1. 推送 tag 触发构建：`git tag v2.1.0 && git push origin v2.1.0`
2. GitHub Actions 自动构建多平台产物
3. 产物命名格式：`bilidown_v2.1.0_windows_x86_64.zip`

## 开发环境

```bash
# 前端开发
cd ui
pnpm install
pnpm dev

# 后端开发
cd server
go run ./cmd/bilidown
```

## 项目结构

```
bilidown/
├── ui/                     # 前端源码
│   └── src/
├── server/                 # 后端源码
│   ├── cmd/bilidown/       # 程序入口
│   └── internal/           # 内部模块
│       ├── app/            # 应用启动
│       ├── bilibili/       # B站 API 客户端
│       ├── handler/        # HTTP 处理器
│       ├── service/        # 业务逻辑
│       ├── store/          # 数据存储
│       ├── static/         # 嵌入的静态资源
│       └── util/           # 工具函数
├── build/                  # 构建输出
├── Makefile                # 构建脚本
└── .github/workflows/      # GitHub Actions
```

## 特别感谢

- [twbs/bootstrap](https://github.com/twbs/bootstrap) - 响应式前端框架
- [vanjs-org/van](https://github.com/vanjs-org/van) - 轻量级前端框架
- [vitejs/vite](https://github.com/vitejs/vite) - 快速前端构建工具
- [SocialSisterYi/bilibili-API-collect](https://github.com/SocialSisterYi/bilibili-API-collect) - B站 API 集合
- [sindresorhus/p-queue](https://github.com/sindresorhus/p-queue) - 并发控制队列
- [iuroc/vanjs-router](https://github.com/iuroc/vanjs-router) - Van.js 路由
- [modernc.org/sqlite](https://pkg.go.dev/modernc.org/sqlite) - Go SQLite 驱动
- [skip2/go-qrcode](https://github.com/skip2/go-qrcode) - QR 码生成

## 软件界面

![](./docs/2024-11-05_090604.png)

## Star History

[![Star History Chart](https://api.star-history.com/svg?repos=iuroc/bilidown&type=Date)](https://www.star-history.com/#iuroc/bilidown&Date)