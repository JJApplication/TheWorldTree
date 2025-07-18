# TwT - GitHub Repository Information Service

一个使用Go语言开发的GitHub仓库信息获取和展示服务，提供HTTP REST API和gRPC接口，使用SQLite3存储同步的仓库信息。

## 功能特性

- 🚀 **双协议支持**: 同时提供HTTP REST API和gRPC接口
- 📊 **数据持久化**: 使用SQLite3数据库存储GitHub仓库信息
- 🔄 **自动同步**: 从GitHub API获取最新的仓库信息
- 🎨 **Web界面**: 提供简洁美观的Web Dashboard
- ⚙️ **配置管理**: 支持TOML配置文件
- 🔐 **Token支持**: 支持GitHub Personal Access Token

## 技术栈

- **Web框架**: Gin
- **RPC框架**: gRPC + Protocol Buffers
- **数据库**: SQLite3
- **Git操作**: go-git
- **配置管理**: TOML
- **HTTP客户端**: 标准库net/http

## 项目结构

```
twt/
├── config/           # 配置管理
│   └── config.go
├── models/           # 数据模型
│   └── repository.go
├── proto/            # gRPC协议定义
│   ├── repository.proto
│   ├── repository.pb.go
│   └── repository_grpc.pb.go
├── server/           # 服务器实现
│   ├── http_server.go
│   └── grpc_server.go
├── services/         # 业务逻辑
│   └── github.go
├── web/              # Web界面
│   └── templates/
│       └── index.html
├── config.toml       # 配置文件
├── go.mod           # Go模块定义
├── main.go          # 程序入口
└── README.md        # 项目说明
```

## 快速开始

### 1. 安装依赖

```bash
go mod tidy
```

### 2. 配置GitHub Token

编辑 `config.toml` 文件，添加你的GitHub Personal Access Token：

```toml
[github]
repositories = [
    "https://github.com/gin-gonic/gin",
    "https://github.com/go-git/go-git"
]
token = "your_github_token_here"  # 替换为你的GitHub Token
```

### 3. 运行服务

```bash
# 同时启动HTTP和gRPC服务
go run main.go

# 仅启动HTTP服务
go run main.go -server=http

# 仅启动gRPC服务
go run main.go -server=grpc

# 使用自定义配置文件
go run main.go -config=custom-config.toml
```

### 4. 访问服务

- **Web界面**: http://localhost:8080
- **HTTP API**: http://localhost:8080/api/v1
- **gRPC服务**: localhost:9090

## API 接口

### HTTP REST API

#### 获取所有仓库
```bash
GET /api/v1/repositories
```

#### 获取特定仓库
```bash
GET /api/v1/repositories/{owner}/{name}
```

#### 同步仓库信息
```bash
POST /api/v1/sync
Content-Type: application/json

{
  "repository_urls": [
    "https://github.com/gin-gonic/gin",
    "https://github.com/go-git/go-git"
  ]
}
```

#### 健康检查
```bash
GET /api/v1/health
```

### gRPC API

gRPC服务定义在 `proto/repository.proto` 文件中，提供以下方法：

- `GetRepositories`: 获取所有仓库列表
- `GetRepository`: 获取特定仓库信息
- `SyncRepositories`: 同步仓库信息

## 配置说明

`config.toml` 配置文件说明：

```toml
[server]
http_port = 8080      # HTTP服务端口
grpc_port = 9090      # gRPC服务端口

[github]
repositories = [      # 要同步的GitHub仓库列表
    "https://github.com/gin-gonic/gin",
    "https://github.com/go-git/go-git"
]
token = "your_token"   # GitHub Personal Access Token

[database]
path = "./data/repositories.db"  # SQLite数据库文件路径

[log]
level = "info"         # 日志级别
```

## GitHub Token 获取

1. 访问 GitHub Settings > Developer settings > Personal access tokens
2. 点击 "Generate new token"
3. 选择适当的权限（建议选择 `public_repo` 权限）
4. 复制生成的token到配置文件中

## 开发说明

### 重新生成gRPC代码

如果修改了 `proto/repository.proto` 文件，需要重新生成Go代码：

```bash
protoc --go_out=. --go-grpc_out=. proto/repository.proto
```

### 数据库结构

项目使用SQLite3数据库，表结构如下：

```sql
CREATE TABLE repositories (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    full_name TEXT UNIQUE NOT NULL,
    description TEXT,
    url TEXT NOT NULL,
    language TEXT,
    stars INTEGER DEFAULT 0,
    forks INTEGER DEFAULT 0,
    created_at DATETIME,
    updated_at DATETIME,
    synced_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
```

## 许可证

MIT License

## 贡献

欢迎提交Issue和Pull Request！