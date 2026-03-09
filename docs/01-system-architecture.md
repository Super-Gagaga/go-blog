# Go 博客社区系统 — 系统架构文档

> 版本：v1.0 | 日期：2026-06-21 | 作者：架构组

---

## 目录

1. [系统概述](#1-系统概述)
2. [技术选型](#2-技术选型)
3. [架构总览](#3-架构总览)
4. [模块划分](#4-模块划分)
5. [数据库设计](#5-数据库设计)
6. [API 设计规范](#6-api-设计规范)
7. [安全设计](#7-安全设计)
8. [部署架构](#8-部署架构)
9. [附录](#9-附录)

---

## 1. 系统概述

### 1.1 项目定位

一个面向技术群体的轻量级博客社区平台。用户可以撰写 Markdown 文章、参与讨论、关注感兴趣的作者，形成以内容为中心的技术交流社区。

### 1.2 核心目标

| 目标 | 描述 |
|------|------|
| **可维护** | 清晰的模块边界，遵循 Go 社区惯用项目布局 |
| **高性能** | 接口响应 < 100ms（P99），支持万级 QPS |
| **可扩展** | 预留搜索、推荐、消息等能力的扩展点 |
| **易部署** | 单二进制 + 配置文件即可启动，支持容器化 |

### 1.3 功能全景

```
┌─────────────────────────────────────────────────────┐
│                    博客社区系统                        │
├─────────────┬─────────────┬──────────────┬──────────┤
│   用户系统   │   内容系统   │   互动系统    │  基础设施  │
├─────────────┼─────────────┼──────────────┼──────────┤
│ · 注册/登录  │ · 文章 CRUD │ · 评论/回复   │ · 配置管理 │
│ · JWT 认证   │ · Markdown  │ · 点赞/收藏   │ · 日志系统 │
│ · 个人资料   │ · 分类/标签  │ · 关注/粉丝   │ · 中间件   │
│ · 角色权限   │ · 草稿/发布  │ · 通知系统    │ · 文件上传 │
│ · 头像上传   │ · 文章搜索  │ · 浏览统计    │ · 限流熔断 │
└─────────────┴─────────────┴──────────────┴──────────┘
```

---

## 2. 技术选型

### 2.1 核心栈

| 层次 | 技术 | 版本 | 选型理由 |
|------|------|------|----------|
| **语言** | Go | 1.22+ | 高性能、并发原生、部署简单 |
| **Web 框架** | Gin | v1.9+ | 社区最活跃的 Go HTTP 框架，性能优秀 |
| **ORM** | GORM | v2 | 功能完善，迁移方案成熟（MySQL 驱动） |
| **数据库** | MySQL | 8.0+ | 成熟稳定、社区广泛、支持 JSON、全文索引、窗口函数 |
| **缓存** | Redis | 7 | 会话存储、热点数据、分布式锁 |
| **认证** | JWT (golang-jwt) | v5 | 无状态认证，适合 API 服务 |
| **配置** | Viper | v1.18+ | 多格式支持，环境变量绑定 |
| **日志** | Zap | v1.26+ | 零分配、结构化、高性能 |
| **验证** | go-playground/validator | v10 | 结构体标签式校验 |
| **迁移** | golang-migrate | v4 | 纯 SQL 迁移，版本控制 |

### 2.2 辅助工具

| 工具 | 用途 |
|------|------|
| Docker / Docker Compose | 本地开发环境 & 部署 |
| Makefile | 构建、测试、迁移任务编排 |
| Swagger / OpenAPI | API 文档自动生成 |
| Air | 热重载开发 |
| GitHub Actions | CI/CD |

### 2.3 项目布局 (Go 社区标准)

```
go-blog/
├── cmd/
│   └── server/
│       └── main.go            # 入口
├── internal/
│   ├── config/                # 配置加载
│   ├── handler/               # HTTP 处理器 (按模块)
│   ├── service/               # 业务逻辑层
│   ├── repository/            # 数据访问层
│   ├── model/                 # 数据模型
│   ├── middleware/            # 中间件
│   ├── router/                # 路由注册
│   └── pkg/                   # 内部公共工具
├── migrations/                # 数据库迁移脚本
├── docs/                      # 文档 & Swagger
├── deploy/                    # 部署配置
│   └── docker-compose.yml
├── config.yaml                # 默认配置
├── Makefile
├── Dockerfile
└── go.mod
```

---

## 3. 架构总览

### 3.1 分层架构

```
┌──────────────────────────────────────────┐
│            HTTP / Middleware              │  ← Gin Router + Middleware
│     (认证、日志、限流、CORS、恢复)          │
├──────────────────────────────────────────┤
│              Handler 层                   │  ← 请求绑定、参数校验、响应封装
│  user.go | article.go | comment.go | ... │
├──────────────────────────────────────────┤
│              Service 层                   │  ← 业务逻辑、事务编排、缓存策略
│  user.go | article.go | comment.go | ... │
├──────────────────────────────────────────┤
│            Repository 层                  │  ← 数据访问、SQL 构建、查询优化
│  user.go | article.go | comment.go | ... │
├──────────────────────────────────────────┤
│          Model / Domain 层               │  ← 结构体定义、枚举、DTO
│  user.go | article.go | comment.go | ... │
└──────────────────────────────────────────┘

横向关注点：
┌─────────┬─────────┬──────────┬──────────┬─────────┐
│   Logger │  Error   │  Config  │  Cache   │  Auth   │
│  (Zap)   │ Handler  │ (Viper)  │ (Redis)  │ (JWT)   │
└─────────┴─────────┴──────────┴──────────┴─────────┘
```

### 3.2 请求生命周期

```
Client Request
     │
     ▼
┌──────────┐
│  Gin Router   │ ── 路由匹配
└──────┬───────┘
       ▼
┌──────────┐
│ Middleware Chain │ ── Recovery → Logger → CORS → Auth → RateLimit
└──────┬───────┘
       ▼
┌──────────┐
│  Handler       │ ── 绑定请求体 → 参数校验 → 调用 Service
└──────┬───────┘
       ▼
┌──────────┐
│  Service       │ ── 业务逻辑 → 缓存检查 → 事务管理 → 调用 Repository
└──────┬───────┘
       ▼
┌──────────┐
│  Repository    │ ── GORM 查询 → MySQL / Redis
└──────┬───────┘
       ▼
┌──────────┐
│  Response      │ ── 统一 JSON 格式返回
└───────────────┘
```

### 3.3 统一响应格式

```json
{
  "code": 0,
  "message": "success",
  "data": {}
}
```

| code | 含义 |
|------|------|
| 0 | 成功 |
| 400xx | 参数错误 |
| 401xx | 认证错误 |
| 403xx | 权限不足 |
| 404xx | 资源不存在 |
| 500xx | 服务端错误 |

---

## 4. 模块划分

### 4.1 用户模块 (User)

```
功能：
├── POST   /api/v1/auth/register    注册
├── POST   /api/v1/auth/login       登录
├── GET    /api/v1/user/profile     获取个人信息
├── PUT    /api/v1/user/profile     更新个人信息
├── POST   /api/v1/user/avatar      上传头像
├── GET    /api/v1/user/:id         查看用户主页
└── GET    /api/v1/user/:id/articles 查看用户文章列表
```

**关键设计：**
- 密码 bcrypt 加盐哈希
- JWT Token（Access Token 15min + Refresh Token 7d）
- Refresh Token 存储在 Redis，支持主动失效

### 4.2 文章模块 (Article)

```
功能：
├── POST   /api/v1/articles                   创建文章（草稿/直接提交审核）
├── PUT    /api/v1/articles/:id               更新文章
├── DELETE /api/v1/articles/:id               删除文章 (软删除)
├── GET    /api/v1/articles/:id               文章详情
├── GET    /api/v1/articles                   文章列表 (分页+筛选，仅已发布)
├── PATCH  /api/v1/articles/:id/status        变更状态 (作者)
│            draft → pending_review（提交审核）
│            draft/pending_review → draft（撤回）
│            published → archived（归档）
└── POST   /api/v1/articles/:id/upload        文章内图片上传

管理员审核 (Admin)：
├── GET    /api/v1/admin/articles/pending     待审核文章列表
├── POST   /api/v1/admin/articles/:id/approve  通过审核 → published
└── POST   /api/v1/admin/articles/:id/reject   驳回 → draft（附带驳回原因）
```

**关键设计：**
- 文章状态机：`draft → pending_review → published ↔ archived`（驳回则退回 `draft`）
- 用户发布文章时，状态变为 `pending_review` 而非 `published`，等待管理员审核
- 审核通过后文章变为 `published` 才对公众可见
- 驳回时需填写 `review_comment` 告知作者原因，状态回到 `draft` 供作者修改后重新提交
- Markdown 渲染（服务端存原文 + 渲染后的 HTML）
- 分类 & 标签多对多关系
- 软删除（deleted_at）

### 4.3 评论模块 (Comment)

```
功能：
├── POST   /api/v1/articles/:id/comments        发表评论
├── DELETE /api/v1/comments/:id                 删除评论
├── GET    /api/v1/articles/:id/comments        文章评论列表 (树形)
└── POST   /api/v1/comments/:id/reply           回复评论
```

**关键设计：**
- 两级评论（评论 + 回复），通过 `parent_id` 实现
- 评论计数缓存（Redis），避免 COUNT 查询
- 软删除 + 已删除占位提示

### 4.4 互动模块 (Interaction)

```
功能：
├── POST   /api/v1/articles/:id/like        点赞/取消点赞
├── POST   /api/v1/articles/:id/favorite    收藏/取消收藏
├── GET    /api/v1/user/favorites           我的收藏列表
├── POST   /api/v1/users/:id/follow         关注/取消关注
├── GET    /api/v1/user/following           我的关注列表
├── GET    /api/v1/user/followers           我的粉丝列表
└── GET    /api/v1/user/notifications       通知列表
```

**关键设计：**
- 点赞/收藏/关注使用 Redis Set 做幂等去重，异步落库
- 通知系统基于"事件 + 订阅者"模式，异步生成

### 4.5 分类 & 标签模块 (Category & Tag)

```
功能：
├── GET    /api/v1/categories          分类列表
├── POST   /api/v1/categories          创建分类 (管理员)
├── GET    /api/v1/tags                标签列表
└── GET    /api/v1/tags?keyword=xxx    标签搜索
```

### 4.6 管理员模块 (Admin)

```
功能（全部需要 admin 角色）：
├── GET    /api/v1/admin/dashboard           仪表盘统计
├── GET    /api/v1/admin/articles/pending    待审核文章列表
├── POST   /api/v1/admin/articles/:id/approve  通过审核
├── POST   /api/v1/admin/articles/:id/reject   驳回文章
├── GET    /api/v1/admin/articles             所有文章列表
├── DELETE /api/v1/admin/articles/:id         强制删除文章
├── GET    /api/v1/admin/comments             所有评论列表
├── DELETE /api/v1/admin/comments/:id         强制删除评论
├── GET    /api/v1/admin/users                用户列表
├── PATCH  /api/v1/admin/users/:id/status     封禁/解封用户
├── PATCH  /api/v1/admin/users/:id/role       变更用户角色
├── GET    /api/v1/admin/audit-logs           审计日志
├── POST   /api/v1/admin/categories           创建分类
├── PUT    /api/v1/admin/categories/:id       更新分类
├── DELETE /api/v1/admin/categories/:id       删除分类
├── POST   /api/v1/admin/tags                 创建标签
├── PUT    /api/v1/admin/tags/:id             更新标签
└── DELETE /api/v1/admin/tags/:id             删除标签
```

**审核流程：**
1. 用户将文章状态设为 `published` → 后端自动转为 `pending_review`
2. 管理员在后台看到待审核列表，查看文章内容
3. 通过 → 文章状态变为 `published`，公开可见，通知作者"审核通过"
4. 驳回 → 文章状态回到 `draft`，填写 `review_comment` 驳回原因，通知作者"需修改后重新提交"
5. 作者修改后可重新提交审核

```

---

## 5. 数据库设计

### 5.1 ER 图（文字版）

```
┌───────────┐       ┌───────────────┐       ┌───────────┐
│   users   │       │   articles    │       │   comments │
├───────────┤       ├───────────────┤       ├───────────┤
│ id (PK)   │──┐    │ id (PK)       │──┐    │ id (PK)   │
│ username  │  │    │ title         │  │    │ content   │
│ email     │  │    │ content       │  │    │ article_id│
│ password  │  │    │ content_html  │  │    │ user_id   │
│ avatar    │  │    │ status        │  │    │ parent_id │
│ bio       │  │    │ view_count    │  │    │ created_at│
│ role      │  │    │ user_id (FK)  │◄─┤    │ deleted_at│
│ created_at│  │    │ category_id   │  │    └───────────┘
│ updated_at│  │    │ created_at    │  │
│ deleted_at│  │    │ updated_at    │  │
└───────────┘  │    │ deleted_at    │  │
               │    └───────────────┘  │
               │          │            │
               │          │ (M:N)      │
               │    ┌─────┴──────┐     │
               │    │ article_tag│     │
               │    │ article_id │     │
               │    │ tag_id     │     │
               │    └────────────┘     │
               │          │            │
               │    ┌─────┴──────┐     │
               │    │    tags    │     │
               │    │ id (PK)    │     │
               │    │ name       │     │
               │    └────────────┘     │
               │                       │
               │    ┌───────────────┐  │
               │    │  categories   │  │
               │    │ id (PK)       │◄─┘
               │    │ name          │
               │    │ slug          │
               │    └───────────────┘
               │
               │    ┌───────────────────┐
               │    │    likes          │
               │    │ user_id +         │
               │    │ article_id (PK)   │
               │    └───────────────────┘
               │
               │    ┌───────────────────┐
               │    │   favorites       │
               │    │ user_id +         │
               │    │ article_id (PK)   │
               │    └───────────────────┘
               │
               │    ┌───────────────────┐
               │    │   follows         │
               │    │ follower_id +     │
               │    │ followee_id (PK)  │
               │    └───────────────────┘
               │
               │    ┌───────────────────┐
               │    │  notifications    │
               │    │ id (PK)           │
               │    │ user_id (FK) ─────┘
               │    │ type              │
               │    │ content           │
               │    │ is_read           │
               │    └───────────────────┘
               │
               ▼
    (all FK references to users.id)
```

### 5.2 核心建表 SQL

<details>
<summary>展开查看完整 DDL</summary>

```sql
-- 用户表
CREATE TABLE users (
    id          BIGINT       NOT NULL AUTO_INCREMENT PRIMARY KEY,
    username    VARCHAR(50)  NOT NULL UNIQUE,
    email       VARCHAR(255) NOT NULL UNIQUE,
    password    VARCHAR(255) NOT NULL,
    avatar      VARCHAR(500),
    bio         TEXT,
    role        VARCHAR(20)  NOT NULL DEFAULT 'user',  -- user | admin
    created_at  DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at  DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at  DATETIME
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_username ON users(username);

-- 分类表
CREATE TABLE categories (
    id         BIGINT       NOT NULL AUTO_INCREMENT PRIMARY KEY,
    name       VARCHAR(100) NOT NULL,
    slug       VARCHAR(100) NOT NULL UNIQUE,
    created_at DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 标签表
CREATE TABLE tags (
    id         BIGINT      NOT NULL AUTO_INCREMENT PRIMARY KEY,
    name       VARCHAR(50) NOT NULL UNIQUE,
    created_at DATETIME    NOT NULL DEFAULT CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 文章表
CREATE TABLE articles (
    id           BIGINT       NOT NULL AUTO_INCREMENT PRIMARY KEY,
    title        VARCHAR(255) NOT NULL,
    slug         VARCHAR(300) NOT NULL UNIQUE,
    content      TEXT         NOT NULL,
    content_html TEXT         NOT NULL,
    summary      VARCHAR(500),
    cover_image  VARCHAR(500),
    status       VARCHAR(20)  NOT NULL DEFAULT 'draft',  -- draft | pending_review | published | archived
    review_comment VARCHAR(500),                              -- 审核驳回原因
    view_count   BIGINT       NOT NULL DEFAULT 0,
    user_id      BIGINT       NOT NULL,
    category_id  BIGINT,
    created_at   DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at   DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at   DATETIME,
    FOREIGN KEY (user_id)     REFERENCES users(id),
    FOREIGN KEY (category_id) REFERENCES categories(id) ON DELETE SET NULL,
    FULLTEXT INDEX ft_articles_search (title, content)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
CREATE INDEX idx_articles_user_id ON articles(user_id);
CREATE INDEX idx_articles_status  ON articles(status);
CREATE INDEX idx_articles_category ON articles(category_id);
CREATE INDEX idx_articles_created_at ON articles(created_at);

-- 文章-标签关联表
CREATE TABLE article_tags (
    article_id BIGINT NOT NULL,
    tag_id     BIGINT NOT NULL,
    PRIMARY KEY (article_id, tag_id),
    FOREIGN KEY (article_id) REFERENCES articles(id) ON DELETE CASCADE,
    FOREIGN KEY (tag_id)     REFERENCES tags(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 评论表
CREATE TABLE comments (
    id         BIGINT   NOT NULL AUTO_INCREMENT PRIMARY KEY,
    content    TEXT     NOT NULL,
    article_id BIGINT   NOT NULL,
    user_id    BIGINT   NOT NULL,
    parent_id  BIGINT,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at DATETIME,
    FOREIGN KEY (article_id) REFERENCES articles(id) ON DELETE CASCADE,
    FOREIGN KEY (user_id)    REFERENCES users(id),
    FOREIGN KEY (parent_id)  REFERENCES comments(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
CREATE INDEX idx_comments_article ON comments(article_id);
CREATE INDEX idx_comments_parent  ON comments(parent_id);

-- 点赞表
CREATE TABLE likes (
    user_id    BIGINT   NOT NULL,
    article_id BIGINT   NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (user_id, article_id),
    FOREIGN KEY (user_id)    REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (article_id) REFERENCES articles(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 收藏表
CREATE TABLE favorites (
    user_id    BIGINT   NOT NULL,
    article_id BIGINT   NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (user_id, article_id),
    FOREIGN KEY (user_id)    REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (article_id) REFERENCES articles(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 关注表
CREATE TABLE follows (
    follower_id BIGINT   NOT NULL,
    followee_id BIGINT   NOT NULL,
    created_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (follower_id, followee_id),
    FOREIGN KEY (follower_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (followee_id) REFERENCES users(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 通知表
CREATE TABLE notifications (
    id         BIGINT      NOT NULL AUTO_INCREMENT PRIMARY KEY,
    user_id    BIGINT      NOT NULL,
    type       VARCHAR(30) NOT NULL,  -- like | comment | follow | system | review_approved | review_rejected
    content    JSON        NOT NULL,
    is_read    BOOLEAN     NOT NULL DEFAULT FALSE,
    created_at DATETIME    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
CREATE INDEX idx_notifications_user ON notifications(user_id, is_read);
```
</details>

---

## 6. API 设计规范

### 6.1 URL 设计

```
Base URL: /api/v1

RESTful 风格：
GET    /api/v1/articles          → 列表
POST   /api/v1/articles          → 创建
GET    /api/v1/articles/:id      → 详情
PUT    /api/v1/articles/:id      → 全量更新
PATCH  /api/v1/articles/:id      → 部分更新
DELETE /api/v1/articles/:id      → 删除

子资源：
GET    /api/v1/articles/:id/comments     → 文章下的评论
POST   /api/v1/articles/:id/like         → 点赞文章
```

### 6.2 分页规范

```
Query: ?page=1&page_size=20

Response:
{
  "code": 0,
  "message": "success",
  "data": {
    "list": [...],
    "pagination": {
      "page": 1,
      "page_size": 20,
      "total": 150,
      "total_pages": 8
    }
  }
}
```

### 6.3 错误码

| 范围 | 含义 | 示例 |
|------|------|------|
| 0 | 成功 | — |
| 40001-40099 | 请求参数错误 | 40001: 缺少必填字段 |
| 40101-40199 | 认证失败 | 40101: Token 过期 |
| 40301-40399 | 权限不足 | 40301: 非文章作者 |
| 40401-40499 | 资源不存在 | 40401: 文章不存在 |
| 50001-50099 | 服务端错误 | 50001: 数据库异常 |

---

## 7. 安全设计

| 安全措施 | 实现方式 |
|----------|----------|
| **密码存储** | bcrypt (cost=12) |
| **JWT 签名** | HS256 / RS256 |
| **SQL 注入防护** | GORM 参数化查询 |
| **XSS 防护** | 输出 HTML 时使用 bluemonday 清洗 |
| **CSRF** | API 使用 Token 认证，不依赖 Cookie |
| **速率限制** | Redis 令牌桶 (每 IP / 每用户) |
| **文件上传** | 校验 MIME 类型 + 文件大小限制 (5MB) |
| **HTTPS** | 生产环境强制 TLS |

---

## 8. 部署架构

### 8.1 开发环境

```
┌──────────────────────────────────────────┐
│                Docker Compose             │
│                                           │
│  ┌──────────┐  ┌──────────┐  ┌─────────┐ │
│  │  Go App  │  │  MySQL   │  │  Redis  │ │
│  │  :8080   │  │  :3306   │  │ :6379   │ │
│  └──────────┘  └──────────┘  └─────────┘ │
└──────────────────────────────────────────┘
```

### 8.2 生产环境（推荐）

```
                    ┌──────────┐
                    │   CDN    │  ← 静态资源 & 头像
                    └─────┬────┘
                          │
                    ┌─────▼────┐
                    │  Nginx   │  ← 反向代理 + 负载均衡
                    └─────┬────┘
                          │
              ┌───────────┼───────────┐
              │           │           │
        ┌─────▼────┐ ┌───▼─────┐ ┌───▼─────┐
        │ Go App 1 │ │Go App 2 │ │Go App 3 │  ← 多实例
        └─────┬────┘ └───┬─────┘ └───┬─────┘
              │           │           │
              └───────────┼───────────┘
                          │
              ┌───────────┼───────────┐
              │           │           │
        ┌─────▼────┐ ┌───▼─────┐     │
        │  MySQL   │ │  Redis  │     │
        │ (主+从)  │ │ Cluster │     │
        └──────────┘ └─────────┘     │
                                     │
                          ┌──────────▼──────────┐
                          │    Prometheus +      │
                          │    Grafana           │  ← 监控 & 告警
                          └─────────────────────┘
```

---

## 9. 附录

### 9.1 缓存策略

| 数据 | 策略 | TTL | 失效时机 |
|------|------|-----|----------|
| 文章详情 | Cache-Aside | 10min | 更新时删除 |
| 热门文章列表 | 定时刷新 | 5min | — |
| 用户 Session | Refresh Token | 7d | 登出时删除 |
| 点赞/收藏状态 | Redis Set | 永久 | 取消时删除 |
| 评论计数 | Redis String | 永久 | 评论增删时 incr/decr |
| 浏览计数 | Redis HyperLogLog | — | 定时同步到 DB |

### 9.2 技术债务清单

- [ ] 引入 Elasticsearch 替代 MySQL 全文搜索
- [ ] 消息队列（RabbitMQ/Kafka）替换 channel 异步通知
- [ ] WebSocket 实现实时通知推送
- [ ] 单元测试 & 集成测试覆盖
- [ ] 分布式追踪（OpenTelemetry）
