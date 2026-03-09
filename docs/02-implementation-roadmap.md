# Go 博客社区系统 — 分步推进文档

> 版本：v1.0 | 日期：2026-06-21

---

## 总览

本项目分 **7 个阶段** 推进，每阶段产出可运行的中间版本。建议按序推进，每完成一阶段进行 Code Review 后再进入下一阶段。

```
Phase 1  ████████░░░░░░░░░░  项目骨架 (~2天)
Phase 2  ████████░░░░░░░░░░  用户系统 (~3天)
Phase 3  ████████████░░░░░░  内容系统 (~4天)
Phase 4  ██████████████░░░░  互动系统 (~3天)
Phase 5  ████████████████░░  社区功能 (~3天)
Phase 6  █████████████████░  优化增强 (~3天)
Phase 7  ██████████████████  测试部署 (~2天)
        ───────────────────
        总计约 20 个工作日
```

---

## Phase 1：项目骨架 & 基础设施

**目标：** 搭建可运行的空服务，基础设施就位。

### 1.1 初始化项目

```bash
# 创建 Go module
mkdir go-blog && cd go-blog
go mod init github.com/yourname/go-blog

# 安装核心依赖
go get github.com/gin-gonic/gin
go get gorm.io/gorm gorm.io/driver/mysql
go get github.com/redis/go-redis/v9
go get github.com/golang-jwt/jwt/v5
go get github.com/spf13/viper
go get go.uber.org/zap
go get github.com/go-playground/validator/v10
go get github.com/golang-migrate/migrate/v4
go get github.com/swaggo/swag
```

### 1.2 项目目录结构

- [ ] 创建 `cmd/server/main.go` 入口文件
- [ ] 创建 `internal/config/` — Viper 配置加载
- [ ] 创建 `internal/model/` — 公共模型（分页、响应结构体）
- [ ] 创建 `internal/middleware/` — Recovery, Logger, CORS
- [ ] 创建 `internal/router/` — 路由骨架
- [ ] 创建 `internal/pkg/` — 公共工具（响应封装、错误码定义）

### 1.3 配置文件

- [ ] `config.yaml` — 数据库、Redis、JWT、应用配置
- [ ] 配置绑定环境变量（12-Factor App）

### 1.4 基础设施

- [ ] `deploy/docker-compose.yml` — MySQL + Redis 服务
- [ ] `migrations/` — golang-migrate 初始化
- [ ] `Makefile` — 常用命令（build, run, dev, migrate-up, migrate-down, test）
- [ ] `Dockerfile` — 多阶段构建

### 1.5 验证标准

```bash
make dev          # 服务启动
curl :8080/health # 返回 {"code":0,"message":"ok"}
curl :8080/api/v1/ping  # 返回 pong
```

<details>
<summary>📁 Phase 1 生产文件清单</summary>

```
go-blog/
├── cmd/server/main.go
├── internal/
│   ├── config/config.go
│   ├── model/
│   │   ├── response.go        # 统一响应结构
│   │   └── pagination.go      # 分页结构
│   ├── middleware/
│   │   ├── recovery.go
│   │   ├── logger.go
│   │   └── cors.go
│   ├── router/router.go
│   └── pkg/
│       ├── app/response.go    # 响应工具函数
│       └── errcode/errcode.go # 错误码定义
├── config.yaml
├── deploy/docker-compose.yml
├── migrations/
├── Makefile
├── Dockerfile
└── go.mod
```
</details>

---

## Phase 2：用户系统

**目标：** 完整的注册、登录、认证、个人信息管理。

### 2.1 数据模型

- [ ] `internal/model/user.go` — User 结构体（GORM）
- [ ] 创建 users 迁移脚本

### 2.2 仓库层

- [ ] `internal/repository/user.go`
  - `Create(user *User) error`
  - `FindByEmail(email string) (*User, error)`
  - `FindByUsername(username string) (*User, error)`
  - `FindByID(id int64) (*User, error)`
  - `Update(user *User) error`

### 2.3 业务层

- [ ] `internal/service/user.go`
  - `Register(req RegisterReq) (*User, error)` — 校验 + bcrypt + 创建
  - `Login(req LoginReq) (*TokenPair, error)` — 验证密码 + 签发 JWT
  - `GetProfile(userID int64) (*UserProfile, error)`
  - `UpdateProfile(userID int64, req UpdateProfileReq) error`
  - `UploadAvatar(userID int64, file io.Reader) (string, error)`

### 2.4 处理器层

- [ ] `internal/handler/user.go`
  - `/api/v1/auth/register` POST
  - `/api/v1/auth/login` POST
  - `/api/v1/auth/refresh` POST — 刷新 Token
  - `/api/v1/user/profile` GET + PUT
  - `/api/v1/user/avatar` POST
  - `/api/v1/user/:id` GET — 用户公开主页

### 2.5 中间件

- [ ] `internal/middleware/auth.go` — JWT 解析中间件
- [ ] 将 `?token=xxx` 或 `Authorization: Bearer xxx` 注入 `ctx`

### 2.6 JWT 工具

- [ ] `internal/pkg/jwt/jwt.go`
  - `GenerateAccessToken(userID int64) (string, error)`
  - `GenerateRefreshToken(userID int64) (string, error)`
  - `ParseToken(tokenStr string) (*Claims, error)`
- [ ] Refresh Token 存入 Redis (7天过期)
- [ ] 登出时删除 Redis 中的 Refresh Token

### 2.7 验证标准

```bash
# 注册
curl -X POST :8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"username":"alice","email":"alice@example.com","password":"Abc12345"}'

# 登录 → 拿到 token
curl -X POST :8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"alice@example.com","password":"Abc12345"}'

# 获取个人信息 (带 Token)
curl :8080/api/v1/user/profile -H "Authorization: Bearer <token>"
```

<details>
<summary>📁 Phase 2 新增文件</summary>

```
internal/
├── model/user.go
├── repository/user.go
├── service/user.go
├── handler/
│   └── user.go
├── middleware/auth.go
└── pkg/
    ├── jwt/jwt.go
    └── hash/password.go       # bcrypt 工具
```
</details>

---

## Phase 3：内容系统（文章 + 分类 + 标签）

**目标：** 完整的文章 CRUD，Markdown 渲染，分类 & 标签管理。

### 3.1 数据模型 & 迁移

- [ ] `internal/model/article.go`
- [ ] `internal/model/category.go`
- [ ] `internal/model/tag.go`
- [ ] 创建 articles, categories, tags, article_tags 迁移脚本
- [ ] 创建预置数据种子（几个分类和常用标签）

### 3.2 仓库层

- [ ] `internal/repository/article.go`
  - `Create(a *Article, tags []int64) error` — 事务创建文章 + 关联标签
  - `Update(a *Article, tags []int64) error`
  - `SoftDelete(id int64) error`
  - `FindByID(id int64) (*Article, error)` — 预加载 Category, Tags, User
  - `List(req ListArticleReq) ([]Article, int64, error)` — 分页 + 筛选
    - 支持按状态、分类、标签、关键词筛选
    - 默认按 created_at 倒序
  - `IncrementViewCount(id int64) error` — 原子自增
- [ ] `internal/repository/category.go`
- [ ] `internal/repository/tag.go`

### 3.3 业务层

- [ ] `internal/service/article.go`
  - `CreateArticle(userID int64, req CreateArticleReq) (*Article, error)`
    - 生成 slug（标题转拼音 + 唯一性校验）
    - Markdown → HTML（使用 goldmark）
    - 自动提取摘要（取前 200 字符纯文本）
  - `UpdateArticle(userID, articleID int64, req UpdateArticleReq) error`
    - 权限校验：仅作者可编辑
  - `DeleteArticle(userID, articleID int64) error`
  - `GetArticle(id int64) (*ArticleDetail, error)`
    - 增加浏览计数（异步）
  - `ListArticles(req ListArticleReq) (*PageResult, error)`
  - `ChangeStatus(userID, articleID int64, status string) error`
  - `UploadArticleImage(file io.Reader) (string, error)`

### 3.4 Markdown 渲染

- [ ] `internal/pkg/markdown/renderer.go`
  - goldmark 渲染 Markdown → HTML
  - 代码高亮 (chroma)
  - 自动生成目录 (TOC)

### 3.5 处理器层

- [ ] `internal/handler/article.go`
- [ ] `internal/handler/category.go`
- [ ] `internal/handler/tag.go`

### 3.6 文件上传

- [ ] `internal/pkg/upload/upload.go`
  - 本地存储实现（先做简单版）
  - MIME 校验 + 大小限制
  - 静态文件服务路由

### 3.7 验证标准

```bash
# 创建文章
curl -X POST :8080/api/v1/articles \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{
    "title":"Go 并发编程入门",
    "content":"# Go 并发\n\nGoroutine 是...",
    "category_id":1,
    "tag_ids":[1,2]
  }'

# 文章列表
curl :8080/api/v1/articles?page=1&page_size=10

# 文章详情
curl :8080/api/v1/articles/1
```

<details>
<summary>📁 Phase 3 新增文件</summary>

```
internal/
├── model/
│   ├── article.go
│   ├── category.go
│   └── tag.go
├── repository/
│   ├── article.go
│   ├── category.go
│   └── tag.go
├── service/
│   ├── article.go
│   ├── category.go
│   └── tag.go
├── handler/
│   ├── article.go
│   ├── category.go
│   └── tag.go
└── pkg/
    ├── markdown/renderer.go
    ├── upload/upload.go
    └── slug/slug.go
```
</details>

---

## Phase 4：互动系统（评论 + 点赞 + 收藏）

**目标：** 用户可以对文章进行评论、点赞、收藏。

### 4.1 评论系统

- [ ] 数据模型 `internal/model/comment.go` + 迁移
- [ ] `internal/repository/comment.go`
  - `Create(c *Comment) error`
  - `FindByArticleID(articleID int64) ([]Comment, error)`
  - 组装树形结构（parent_id 关联）
  - `SoftDelete(commentID, userID int64) error`
- [ ] `internal/service/comment.go`
  - `CreateComment(userID, articleID int64, req CreateCommentReq) (*Comment, error)`
  - `ReplyComment(userID, commentID int64, req ReplyReq) (*Comment, error)`
  - `DeleteComment(userID, commentID int64) error` — 作者或文章作者可删
  - `GetArticleComments(articleID int64) ([]CommentTreeNode, error)`
- [ ] 评论计数缓存 — Redis incr/decr，异步同步到文章表
- [ ] `internal/handler/comment.go`

### 4.2 点赞系统

- [ ] 数据模型 `internal/model/like.go` + 迁移
- [ ] `internal/service/like.go`
  - `ToggleLike(userID, articleID int64) (liked bool, count int64, error)`
  - 使用 Redis Set 防重 + 计数，异步落库
  - 返回当前点赞状态和总数
- [ ] 获取文章列表时填充用户点赞状态
- [ ] `internal/handler/like.go`

### 4.3 收藏系统

- [ ] 数据模型 `internal/model/favorite.go` + 迁移
- [ ] `internal/service/favorite.go`
  - `ToggleFavorite(userID, articleID int64) (favorited bool, error)`
  - `GetMyFavorites(userID int64, page int) (*PageResult, error)`
- [ ] `internal/handler/favorite.go`

### 4.4 验证标准

```bash
# 发表评论
curl -X POST :8080/api/v1/articles/1/comments \
  -H "Authorization: Bearer <token>" \
  -d '{"content":"写得很棒！"}'

# 获取评论列表 (树形)
curl :8080/api/v1/articles/1/comments

# 点赞
curl -X POST :8080/api/v1/articles/1/like \
  -H "Authorization: Bearer <token>"

# 收藏
curl -X POST :8080/api/v1/articles/1/favorite \
  -H "Authorization: Bearer <token>"

# 我的收藏
curl :8080/api/v1/user/favorites?page=1 \
  -H "Authorization: Bearer <token>"
```

<details>
<summary>📁 Phase 4 新增文件</summary>

```
internal/
├── model/
│   ├── comment.go
│   ├── like.go
│   └── favorite.go
├── repository/
│   ├── comment.go
│   ├── like.go
│   └── favorite.go
├── service/
│   ├── comment.go
│   ├── like.go
│   └── favorite.go
├── handler/
│   ├── comment.go
│   ├── like.go
│   └── favorite.go
└── pkg/
    └── cache/            # Redis 工具封装
        ├── redis.go
        └── keys.go       # 缓存 key 命名规范
```
</details>

---

## Phase 5：社区功能（关注 + 通知 + Feed）

**目标：** 用户之间可以关注，接收通知，查看关注者的动态。

### 5.1 关注系统

- [ ] 数据模型 `internal/model/follow.go` + 迁移
- [ ] `internal/service/follow.go`
  - `ToggleFollow(followerID, followeeID int64) (following bool, error)`
  - 禁止自己关注自己
  - `GetFollowing(userID int64, page int) (*PageResult, error)`
  - `GetFollowers(userID int64, page int) (*PageResult, error)`
  - 关注/粉丝计数缓存

### 5.2 通知系统

- [ ] 数据模型 `internal/model/notification.go` + 迁移
- [ ] `internal/service/notification.go`
  - 定义通知事件类型：
    - `comment` — 有人评论了你的文章
    - `reply` — 有人回复了你的评论
    - `like` — 有人点赞了你的文章
    - `follow` — 有人关注了你
    - `system` — 系统通知
  - `SendNotification(userID int64, nType string, content map[string]interface{})` — 异步写入
  - `GetNotifications(userID int64, page int) (*PageResult, error)`
  - `MarkAsRead(userID int64, notificationIDs []int64) error`
  - 未读计数缓存
- [ ] 在评论、点赞、关注服务中嵌入通知触发
- [ ] `internal/handler/notification.go`

### 5.3 动态 Feed

- [ ] `internal/service/feed.go`
  - `GetUserFeed(userID int64, page int) (*PageResult, error)`
  - 拉取关注用户的最新文章
  - 缓存热门 Feed (Redis Sorted Set)

### 5.4 公共主页增强

- [ ] 用户主页展示：文章数、粉丝数、关注数
- [ ] 关注/取消关注按钮状态

### 5.5 验证标准

```bash
# 关注用户
curl -X POST :8080/api/v1/users/2/follow \
  -H "Authorization: Bearer <token>"

# 获取通知
curl :8080/api/v1/user/notifications?page=1 \
  -H "Authorization: Bearer <token>"

# 标记已读
curl -X PATCH :8080/api/v1/user/notifications/read \
  -H "Authorization: Bearer <token>" \
  -d '{"ids":[1,2,3]}'

# 动态 Feed
curl :8080/api/v1/user/feed?page=1 \
  -H "Authorization: Bearer <token>"
```

<details>
<summary>📁 Phase 5 新增文件</summary>

```
internal/
├── model/
│   ├── follow.go
│   └── notification.go
├── repository/
│   ├── follow.go
│   └── notification.go
├── service/
│   ├── follow.go
│   ├── notification.go
│   └── feed.go
└── handler/
    ├── follow.go
    ├── notification.go
    └── feed.go
```
</details>

---

## Phase 6：优化增强

**目标：** 性能优化、全文搜索、限流、后台管理。

### 6.1 全文搜索

- [ ] MySQL `FULLTEXT INDEX` + `MATCH ... AGAINST` 实现全文搜索
- [ ] `GET /api/v1/articles?keyword=xxx` 支持标题 + 内容搜索
- [ ] 高亮搜索关键词

### 6.2 性能优化

- [ ] Redis 缓存热点文章详情（Cache-Aside 模式）
- [ ] 文章列表加入 Redis 缓存（按分类 + 页面）
- [ ] 数据库查询优化：避免 N+1、添加必要索引
- [ ] 引入 `sync.Pool` 复用临时对象
- [ ] 分页查询使用游标分页（cursor-based）替代 offset 分页（可选，用于 Feed 等高频场景）

### 6.3 速率限制

- [ ] `internal/middleware/ratelimit.go`
- [ ] Redis 令牌桶算法
- [ ] 对 `/auth/login` `/auth/register` 做 IP 级限制
- [ ] 对全局 API 做用户级限制

### 6.4 后台管理

- [ ] `internal/middleware/admin.go` — 管理员权限中间件
- [ ] 分类管理 CRUD
- [ ] 标签管理 CRUD
- [ ] 用户列表 & 封禁

### 6.5 Swagger 文档

- [ ] 使用 swaggo 生成注解
- [ ] `GET /swagger/index.html` — Swagger UI
- [ ] 至少覆盖所有用户和文章接口

### 6.6 日志增强

- [ ] 请求日志带 TraceID
- [ ] 结构化日志，关键操作记录（创建/删除文章等）
- [ ] 慢查询日志（>200ms 告警）

### 6.7 验证标准

```bash
# 搜索
curl ":8080/api/v1/articles?keyword=Go并发&page=1"

# Swagger
open http://localhost:8080/swagger/index.html

# 限流测试
for i in {1..20}; do
  curl -X POST :8080/api/v1/auth/login \
    -d '{"email":"test@test.com","password":"wrong"}'
done
# 第 6 次之后应返回 429
```

---

## Phase 7：测试 & 部署

**目标：** 完善的测试覆盖与生产级部署方案。

### 7.1 单元测试

- [ ] Service 层单元测试（Mock Repository）
- [ ] Repository 层测试（使用 test DB 或 SQLite 内存模式）
- [ ] Handler 层测试（httptest）
- [ ] JWT 工具测试
- [ ] 覆盖率 ≥ 70%

### 7.2 集成测试

- [ ] 核心用户流程 E2E：
  - 注册 → 登录 → 发文章 → 评论 → 点赞 → 收藏 → 关注
- [ ] Docker Compose 测试环境（MySQL + Redis）

### 7.3 压力测试

- [ ] 使用 `vegeta` 或 `wrk` 进行压测
- [ ] 目标：单实例 1000 QPS，P99 < 100ms
- [ ] 优化瓶颈点

### 7.4 CI/CD

- [ ] `.github/workflows/ci.yml`
  - lint (golangci-lint)
  - test (go test -race)
  - build (go build + Docker build)
- [ ] `.github/workflows/deploy.yml`
  - 推送到 Docker Registry
  - SSH 到服务器部署

### 7.5 生产部署清单

- [ ] Nginx 反向代理配置（TLS + gzip + 静态资源缓存）
- [ ] MySQL 主从复制配置
- [ ] Redis 持久化配置 (AOF + RDB)
- [ ] 环境变量管理 (.env.production，不提交 Git)
- [ ] 健康检查端点 `/health`
- [ ] 优雅关闭 (Graceful Shutdown)
- [ ] 数据库备份脚本 (cron)

### 7.6 Makefile 完整命令

```makefile
.PHONY: help build run dev test lint clean migrate-up migrate-down seed

help:
	@echo "build      - 编译"
	@echo "run         - 运行"
	@echo "dev         - 开发模式 (Air 热重载)"
	@echo "test        - 运行测试"
	@echo "lint        - 代码检查"
	@echo "migrate-up  - 数据库迁移"
	@echo "migrate-down- 回滚迁移"
	@echo "seed        - 填充测试数据"
	@echo "swagger    - 生成 API 文档"

build:
	go build -o bin/server cmd/server/main.go

dev:
	air

test:
	go test -v -race -cover ./...

lint:
	golangci-lint run

migrate-up:
	migrate -path migrations -database "$(DATABASE_URL)" up

migrate-down:
	migrate -path migrations -database "$(DATABASE_URL)" down 1

seed:
	go run cmd/seed/main.go

swagger:
	swag init -g cmd/server/main.go -o docs/swagger
```

### 7.7 验证标准

```bash
make lint        # 零告警
make test        # 全部通过
make build       # 编译成功
docker-compose up # 整套环境正常运行
```

---

## 里程碑总结

```
Phase 1 ──── 空服务跑通                  ██  Day 1-2
Phase 2 ──── 用户可注册/登录/管理资料     ██  Day 3-5
Phase 3 ──── 用户可写/发/管理文章         ██  Day 6-9
Phase 4 ──── 用户可评论/点赞/收藏         ██  Day 10-12
Phase 5 ──── 关注/通知/Feed 打通          ██  Day 13-15
Phase 6 ──── 搜索/缓存/限流/后台管理      ██  Day 16-18
Phase 7 ──── 测试/CI /部署上线            ██  Day 19-20
```

---

## 附录：每个 Phase 的 commit 建议

```bash
git commit -m "feat: phase 1 - project scaffolding"
git commit -m "feat: phase 2 - user authentication & profile"
git commit -m "feat: phase 3 - article CRUD with markdown & categories"
git commit -m "feat: phase 4 - comments, likes & favorites"
git commit -m "feat: phase 5 - follow, notifications & feed"
git commit -m "feat: phase 6 - search, caching & admin panel"
git commit -m "feat: phase 7 - tests, CI/CD & production deploy"
```
