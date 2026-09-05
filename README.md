# LumaLogBackEnd

LumaLogBackEnd 是 LumaLog 习惯追踪系统的后端 API 服务，使用 Go、Gin 和 PostgreSQL 构建，为 Web 前端提供账户、习惯、分类、签到、热力图、统计、徽章及成果分享等数据能力。

项目围绕“点亮每一次坚持”设计。用户可以为阅读、健身、学习、健康管理等长期目标创建 habit，按照全天或指定时间段完成签到，并通过近一年的贡献热力图、连续天数、完成率和阶段徽章观察自己的积累。服务端统一处理签到约束和统计口径，使不同客户端能够获得一致的数据结果。

## 项目架构

项目采用清晰的分层架构，请求从 Gin 路由进入 Handler，经由 Service 执行业务规则和统计计算，再由 Repository 访问 PostgreSQL。数据库连接池和迁移在启动阶段完成初始化，认证中间件则负责保护需要登录的接口。

```text
客户端（LumaLogFrontEnd）
          │ HTTP / JSON
          ▼
Router（路由、CORS、认证中间件）
          ▼
Handler（参数解析、校验、响应组装）
          ▼
Service（签到状态、热力图、统计、徽章）
          ▼
Repository（数据查询与持久化）
          ▼
PostgreSQL（用户、分类、habit、签到记录）
```

### 分层职责

- `router`：注册 `/api` 路由，处理 CORS，并为受保护接口挂载身份认证中间件。
- `handler`：接收和校验 HTTP 请求，执行用户身份隔离，将业务结果转换为 JSON 响应。
- `service`：实现签到状态判断、近一年热力图生成、连续签到与完成率统计、徽章判定等核心规则。
- `repository`：封装用户、分类、habit 和签到记录的 PostgreSQL 数据访问。
- `model`：定义数据库实体、接口请求、接口响应以及统计和分享数据结构。
- `database`：使用 `pgxpool` 管理数据库连接，并在服务启动时按文件名顺序执行内嵌 SQL 迁移。
- `config`：从环境变量读取服务端口、数据库连接和令牌密钥，并提供开发环境默认值。
- `patch`：提供可重复执行的开发演示数据补丁，方便联调和界面展示。

### 项目结构

```text
LumaLogBackEnd
├─ main.go                          # 服务入口与 patch 命令分发
├─ config/config.go                 # 环境变量与运行配置
├─ database
│  ├─ database.go                   # PostgreSQL 连接池与迁移执行器
│  └─ migrations/001_init.sql       # 数据表、字段和索引定义
├─ router/router.go                 # Gin 路由与 CORS 中间件
├─ handler/*.go                     # 按业务模块组织的 HTTP Handler
├─ service/*.go                     # 签到、统计、热力图与徽章逻辑
├─ repository/*.go                  # PostgreSQL 数据访问
├─ model/model.go                   # 实体、请求与响应结构
├─ patch/patch.go                   # 开发演示数据补丁
├─ util/util.go                     # 日期、时间与通用转换工具
├─ go.mod
└─ README.md
```

## 项目特点

- 完整的 habit 生命周期：支持创建、编辑、软删除、归档、恢复、排序和首页展示控制，并通过分类组织不同类型的长期目标。
- 灵活的签到规则：支持每日目标次数、全天或指定时间段签到、额外签到、本月补签及每月补签次数限制。
- 统一的统计能力：按近一年记录生成贡献热力图，并计算当前连续、最长连续、累计次数、完成天数和完成率。
- 阶段成长反馈：根据累计签到、连续签到、活跃 habit 数和完成率自动判定个人及单项徽章。
- 面向分享的数据接口：聚合 habit、统计、热力图、今日进度和徽章，为前端生成成果分享图提供一次性数据载荷。
- 用户数据隔离：受保护接口统一从身份令牌读取用户 ID，所有核心数据查询均限定当前用户。
- 基础安全机制：密码使用 bcrypt 哈希保存；登录令牌通过 HMAC-SHA256 签名并设置 14 天有效期。
- 自动数据库初始化：SQL 迁移随程序内嵌，启动连接 PostgreSQL 后自动创建或补充所需表结构和索引。
- 偏好设置同步：服务端持久化主题、语言、首页视图模式和统计项显示开关，支持多端共享设置。
- 易于本地联调：配置均可通过环境变量覆盖，并提供健康检查和可重复执行的演示数据 patch 命令。

## Local Database

Create the database first:

```sql
CREATE DATABASE lumalogdb2026;
```

Default connection used by the server:

```text
host=localhost
port=5432
user=postgres
password=794859685
database=lumalogdb2026
```

The server creates tables automatically on startup.

## Run

```bash
go run .
```

## Patch Data

Run a development data patch:

```bash
go run . patch yu-hua-reading-246
```

This creates or updates:

```text
login email: demo@lumalog.local
password: 123456
item: 余华阅读
checkins: 246 consecutive daily checkins
```

Available patch names are shown when the patch name is missing or unknown.

Optional environment variables:

```text
PORT=8080
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=794859685
DB_NAME=lumalogdb2026
DATABASE_URL=postgres://postgres:794859685@localhost:5432/lumalogdb2026?sslmode=disable
JWT_SECRET=change-me
```
