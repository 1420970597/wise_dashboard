# Wise Dashboard

基于 [Nezha](https://github.com/nezhahq/nezha) 的增强版服务器监控系统，添加了终端审计、命令拦截和AutoSSH等企业级功能。

## 项目结构

```
wise_dashboard/
├── dashboard/          # Dashboard 服务端
│   ├── cmd/           # 命令行入口
│   ├── model/         # 数据模型
│   ├── service/       # 业务逻辑
│   └── proto/         # gRPC 协议定义
├── agent/             # Agent 客户端
│   ├── cmd/           # 命令行入口
│   ├── model/         # 数据模型
│   ├── pkg/           # 核心功能包
│   └── service/       # 业务逻辑
├── admin-frontend/    # 管理员前端 (Vite + React)
│   ├── src/           # 源代码
│   ├── public/        # 静态资源
│   └── dist/          # 编译输出
└── user-frontend/     # 用户仪表盘 (Next.js)
    ├── app/           # Next.js 应用目录
    ├── components/    # React 组件
    └── lib/           # 工具库
```

## 核心功能

### 1. 终端审计系统

完整的终端会话审计功能，支持命令拦截、警告和录制：

- **命令拦截**: 基于正则表达式的黑名单规则，实时拦截危险命令
- **命令警告**: 对敏感操作显示警告并要求验证码确认
- **会话录制**: 自动录制终端会话为 Asciinema 格式，支持回放
- **审计日志**: 完整记录所有命令执行历史、工作目录和退出码
- **实时监控**: Dashboard 实时查看所有终端会话和命令执行

#### 技术实现

- 使用 bash `DEBUG trap` + `extdebug` 模式实现命令拦截
- Agent 本地运行审计服务器，wrapper 脚本通过 HTTP API 检查命令
- Dashboard 提供黑名单规则管理和审计日志查询接口
- 支持三种拦截动作：`block`（阻止）、`warn`（警告）、`log`（仅记录）

### 2. AutoSSH 隧道管理

内网穿透和端口转发管理：

- 支持本地端口转发、远程端口转发和动态端口转发
- 自动重连和健康检查
- 多隧道管理
- 实时状态监控

### 3. 服务器监控

继承 Nezha 的核心监控功能：

- 实时监控 CPU、内存、磁盘、网络等系统资源
- 服务监控和告警
- 流量统计
- 在线终端

## 快速开始

### Dashboard 部署

```bash
cd dashboard

# 编译
CGO_ENABLED=1 go build -o dashboard ./cmd/dashboard

# 运行
./dashboard -c config.yaml -db sqlite.db
```

### 前端部署

#### 管理员前端 (admin-frontend)

```bash
cd admin-frontend

# 安装依赖
npm install

# 开发模式
npm run dev

# 生成 API 类型
npx swagger-typescript-api -p http://localhost:8008/swagger/doc.json -o ./src/types -n api.ts --no-client --union-enums

# 编译生产版本
npm run build

# 编译输出在 dist/ 目录
```

#### 用户仪表盘 (user-frontend)

```bash
cd user-frontend

# 安装依赖（推荐使用 bun，也可使用 npm）
bun install
# 或
npm install

# 配置环境变量（参考 .env.example）
cp .env.example .env.local
# 编辑 .env.local 配置 Dashboard API 地址等

# 开发模式
bun dev
# 或
npm run dev

# 编译生产版本
bun run build
# 或
npm run build

# 支持多种部署方式：
# - Vercel: 推送到 GitHub 后在 Vercel 导入项目
# - Cloudflare: 使用 @cloudflare/next-on-pages 适配器
# - Docker: 使用项目提供的 Dockerfile
```

### Agent 部署

```bash
cd agent

# 编译
CGO_ENABLED=0 go build -o nezha-agent ./cmd/agent

# 配置
cat > config.yml <<EOF
server: "dashboard.example.com:5555"
secret: "your-secret-key"
audit_enabled: true
audit_dashboard_url: "http://dashboard.example.com:8008"
audit_token: "your-audit-token"
EOF

# 运行
./nezha-agent -c config.yml
```

## 配置说明

### Dashboard 配置

```yaml
# HTTP 服务端口
http_port: 8008

# gRPC 服务端口
grpc_port: 5555

# 数据库配置
database:
  type: sqlite
  path: data/sqlite.db

# JWT 密钥
jwt_secret: "your-jwt-secret"

# 审计功能
audit:
  enabled: true
  recording_enabled: true
  recording_path: "data/recordings"
```

### Agent 配置

```yaml
# Dashboard 地址
server: "dashboard.example.com:5555"

# 连接密钥
secret: "your-secret-key"

# 终端审计
audit_enabled: true
audit_dashboard_url: "http://dashboard.example.com:8008"
audit_token: "your-audit-token"
```

## API 文档

### 终端审计 API

#### 检查命令

```http
POST /api/v1/terminal/check-command
Content-Type: application/json

{
  "stream_id": "session-id",
  "command": "rm -rf /",
  "working_dir": "/root"
}
```

响应：

```json
{
  "success": true,
  "data": {
    "blocked": true,
    "reason": "禁止删除",
    "action": "block"
  }
}
```

#### 记录命令

```http
POST /api/v1/terminal/record-command
Content-Type: application/json

{
  "stream_id": "session-id",
  "command": "ls -la",
  "working_dir": "/root",
  "exit_code": 0
}
```

#### 查询会话列表

```http
GET /api/v1/terminal/sessions?page=1&page_size=20
Authorization: Bearer <token>
```

#### 查询命令历史

```http
GET /api/v1/terminal/commands?session_id=123&page=1&page_size=50
Authorization: Bearer <token>
```

#### 黑名单管理

```http
# 获取黑名单列表
GET /api/v1/terminal/blacklist
Authorization: Bearer <token>

# 创建黑名单规则
POST /api/v1/terminal/blacklist
Authorization: Bearer <token>
Content-Type: application/json

{
  "pattern": "rm\\s+-rf",
  "description": "禁止删除",
  "action": "block",
  "enabled": true
}

# 更新黑名单规则
PATCH /api/v1/terminal/blacklist/:id
Authorization: Bearer <token>

# 删除黑名单规则
DELETE /api/v1/terminal/blacklist/:id
Authorization: Bearer <token>
```

## 数据库模型

### 终端会话表 (terminal_sessions)

| 字段 | 类型 | 说明 |
|------|------|------|
| id | uint64 | 主键 |
| stream_id | string | 会话唯一标识 |
| user_id | uint64 | 用户ID |
| server_id | uint64 | 服务器ID |
| started_at | time | 开始时间 |
| ended_at | time | 结束时间 |
| command_count | int | 命令数量 |
| recording_enabled | bool | 是否录制 |
| recording_path | string | 录制文件路径 |

### 终端命令表 (terminal_commands)

| 字段 | 类型 | 说明 |
|------|------|------|
| id | uint64 | 主键 |
| session_id | uint64 | 会话ID |
| user_id | uint64 | 用户ID |
| server_id | uint64 | 服务器ID |
| command | string | 命令内容 |
| working_dir | string | 工作目录 |
| executed_at | time | 执行时间 |
| exit_code | int | 退出码 |
| blocked | bool | 是否被拦截 |
| block_reason | string | 拦截原因 |

### 黑名单表 (terminal_blacklist)

| 字段 | 类型 | 说明 |
|------|------|------|
| id | uint64 | 主键 |
| pattern | string | 正则表达式 |
| description | string | 规则描述 |
| action | string | 动作: block/warn/log |
| enabled | bool | 是否启用 |
| created_by | uint64 | 创建者ID |
| created_at | time | 创建时间 |

## 技术栈

- **后端**: Go 1.24+
- **管理员前端**: React + Vite + TypeScript + Tailwind CSS
- **用户前端**: Next.js 15 + React + TypeScript + shadcn/ui
- **数据库**: SQLite / MySQL / PostgreSQL
- **通信**: gRPC + HTTP
- **终端**: PTY + WebSocket
- **录制**: Asciinema v2 格式

## 开发

### 环境要求

- Go 1.24+
- Node.js 18+ (或 Bun)
- SQLite3 / MySQL / PostgreSQL

### 编译前端

#### 管理员前端

```bash
cd admin-frontend

# 安装依赖
npm install

# 开发模式（热重载）
npm run dev

# 生成 TypeScript API 类型定义
npx swagger-typescript-api -p http://localhost:8008/swagger/doc.json -o ./src/types -n api.ts --no-client --union-enums

# 编译生产版本
npm run build

# 预览生产构建
npm run preview
```

#### 用户仪表盘

```bash
cd user-frontend

# 安装依赖
bun install
# 或
npm install

# 配置环境变量
cp .env.example .env.local
# 编辑 .env.local 设置 API 地址

# 开发模式
bun dev

# 编译生产版本
bun run build

# 启动生产服务器
bun start
```

### 生成 Proto 文件

```bash
cd dashboard/proto
protoc --go_out=. --go_opt=paths=source_relative \
       --go-grpc_out=. --go-grpc_opt=paths=source_relative \
       nezha.proto
```

## 安全建议

1. **终端审计**
   - 合理配置黑名单规则，避免过度拦截
   - 定期审查审计日志
   - 保护审计 token 不被泄露

2. **网络安全**
   - 使用 HTTPS/TLS 加密通信
   - 限制 Dashboard 访问 IP
   - 定期更新密钥

3. **权限管理**
   - 使用最小权限原则
   - 定期审查用户权限
   - 启用双因素认证

## 许可证

本项目基于 Apache License 2.0 开源协议。

## 致谢

本项目基于 [Nezha](https://github.com/nezhahq/nezha) 开发，感谢 Nezha 团队的优秀工作。

## 联系方式

- Issues: [GitHub Issues](https://github.com/yourusername/wise_dashboard/issues)
- Discussions: [GitHub Discussions](https://github.com/yourusername/wise_dashboard/discussions)
