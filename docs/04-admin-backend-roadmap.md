# 管理员后台 — 分步推进文档

> 版本：v2.0 | 日期：2026-06-22 | 新增：文章审核系统

---

## 当前状态

后端已有管理员基础能力，前端页面空白。

### 已具备

| 层级 | 文件 | 能力 |
|------|------|------|
| Middleware | `internal/middleware/admin.go` | JWT 解析 + role 校验（只放行 `admin`） |
| Service | `internal/service/admin.go` | `ListUsers` / `UpdateUserStatus` / `UpdateUserRole` |
| Handler | `internal/handler/admin.go` | 对应用户管理 3 个 HTTP 端点 |
| Router | `internal/router/router.go` | `/api/v1/admin/*` 路由组，已注册用户管理 + 分类 CRUD + 标签 CRUD |

### 已有 API

```
POST   /api/v1/admin/categories       创建分类
PUT    /api/v1/admin/categories/:id    更新分类
DELETE /api/v1/admin/categories/:id    删除分类
POST   /api/v1/admin/tags             创建标签
PUT    /api/v1/admin/tags/:id          更新标签
DELETE /api/v1/admin/tags/:id          删除标签
GET    /api/v1/admin/users             用户列表 (分页+搜索)
PATCH  /api/v1/admin/users/:id/status  封禁/解封用户
PATCH  /api/v1/admin/users/:id/role    变更用户角色
```

### 缺失

- ❌ 无管理后台前端页面（`web/admin.html`）
- ❌ 无仪表盘概览统计
- ❌ **无文章审核系统**（用户发布后需管理员审核才能公开）
- ❌ 无文章管理（列表、强制删除、状态变更）
- ❌ 无评论管理（列表、强制删除）
- ❌ 无操作日志 / 审计

---

## 推进总览

分 **6 步** 推进，每一步产出可用的增量：

```
Step 1  ████░░░░░░░░░░░░░░░░  仪表盘 (~1天)
Step 2  ████████░░░░░░░░░░░░  文章审核系统 (~2天) ★ 核心
Step 3  ████████████░░░░░░░░  文章管理 (~1天)
Step 4  ██████████████░░░░░░  评论管理 (~1天)
Step 5  ████████████████░░░░  管理前端页面 (~2天)
Step 6  ████████████████████  审计日志 & 收尾 (~1天)
        ───────────────────
        总计约 8 个工作日
```

---

## Step 1：仪表盘 Dashboard

**目标：** 一个接口返回全站核心数据概览。

### 1.1 后端

- [ ] `internal/model/stats.go` — `DashboardStats` 响应结构体
- [ ] `internal/repository/stats.go`（或在现有 repo 中新增统计方法）
  - `CountUsers() (int64, error)` — 总用户数
  - `CountUsersByRole(role string) (int64, error)` — 按角色计数
  - `CountArticles(status string) (int64, error)` — 按状态统计文章数
  - `CountComments() (int64, error)` — 总评论数
  - `CountLikes() (int64, error)` — 总点赞数
  - `CountFavorites() (int64, error)` — 总收藏数
  - `CountFollows() (int64, error)` — 总关注关系数
  - `RecentUsers(limit int) ([]model.User, error)` — 最新注册用户
  - `RecentArticles(limit int) ([]model.Article, error)` — 最新文章
  - `TopAuthors(limit int) ([]AuthorStat, error)` — 高产作者排行
- [ ] `internal/service/admin.go` 新增 `GetDashboard(ctx) (*DashboardStats, error)`
  - 使用 goroutine 并发执行多个统计查询
- [ ] `internal/handler/admin.go` 新增 `GetDashboard(c *gin.Context)`
- [ ] 路由注册 `GET /api/v1/admin/dashboard`

### 1.2 响应格式

```
GET /api/v1/admin/dashboard
Authorization: Bearer <admin_token>

Response:
{
  "code": 0,
  "data": {
    "user_count": 1280,
    "admin_count": 3,
    "article_count": 456,
    "published_count": 398,
    "pending_review_count": 12,      // 待审核
    "draft_count": 42,
    "comment_count": 2300,
    "like_count": 15000,
    "favorite_count": 8900,
    "follow_count": 3200,
    "recent_users": [...],
    "recent_articles": [...],
    "top_authors": [...]
  }
}
```

### 1.3 验证标准

```bash
curl :8080/api/v1/admin/dashboard -H "Authorization: Bearer <admin_token>"
# 非 admin 用户返回 403
```

<details>
<summary>📁 Step 1 涉及文件</summary>

```
internal/
├── model/stats.go                  # 新增 DashboardStats
├── repository/stats.go             # 新增
├── service/admin.go               # 修改 新增 GetDashboard
├── handler/admin.go               # 修改 新增 GetDashboard
└── router/router.go               # 修改 新增路由
```
</details>

---

## Step 2：文章审核系统 ★

**目标：** 用户发布文章后进入审核队列，管理员审核通过后才对外公开。

这是后台最核心的功能——控制内容质量的门禁。

### 2.1 文章状态变更（基础改造）

**影响范围：用户侧 + 模型层**

- [ ] `internal/model/article.go` 新增常量
  ```go
  ArticleStatusPendingReview = "pending_review"
  ```
- [ ] `internal/model/article.go` Article 结构体新增字段
  ```go
  ReviewComment *string `gorm:"size:500" json:"review_comment,omitempty"`
  ```
- [ ] 数据库迁移：articles 表新增 `review_comment` 列
  ```sql
  ALTER TABLE articles ADD COLUMN review_comment VARCHAR(500) AFTER status;
  ```
- [ ] `internal/service/article.go` 修改发布逻辑
  - 用户将文章状态设为 `published` 时 → 后端**自动转为 `pending_review`**
  - 新增 `SubmitForReview(ctx, userID, articleID int64) error` — 提交审核
  - 新增 `WithdrawReview(ctx, userID, articleID int64) error` — 撤回审核（回到 draft）
  - 状态校验函数新增 `pending_review`
- [ ] `internal/handler/article.go` 修改 `ChangeStatus` 逻辑
  - `draft → pending_review`（提交审核）
  - `pending_review → draft`（撤回）
  - `published → archived`（归档，已发布文章可用）

**状态机流转图：**

```
          用户提交审核
   draft ─────────────→ pending_review
     ↑                       │
     │   ┌─── 管理员通过 ────→ published
     │   │                       │
     │   │   ┌─── 作者归档 ──────┘
     │   │   │
     └───┤   │
   管理员驳回（带原因）
         │   │
         作者修改后重新提交
```

### 2.2 待审核文章列表 API

- [ ] `internal/repository/article.go` 新增
  - `ListPending(page, pageSize int, keyword string) ([]Article, int64, error)`
    - 只查 `status = 'pending_review'`
    - 预加载 `User`、`Category`、`Tags`
    - 支持 keyword 搜索（标题）
    - 按 `created_at` 正序（早提交的优先审核）
- [ ] `internal/service/admin.go` 新增
  - `ListPendingArticles(ctx, page, pageSize int, keyword string) (*PageResult, error)`
- [ ] `internal/handler/admin.go` 新增
  - `ListPendingArticles(c *gin.Context)`
- [ ] 路由注册 `GET /api/v1/admin/articles/pending`

### 2.3 审核通过

- [ ] `internal/service/admin.go` 新增
  - `ApproveArticle(ctx, adminID, articleID int64, clientIP string) error`
    1. 查文章是否存在且状态为 `pending_review`
    2. 状态改为 `published`，清空 `review_comment`
    3. 记录审计日志：`approve_article`
    4. 清除文章缓存
    5. **发送通知给作者**：类型 `review_approved`，内容含文章标题 + 链接
- [ ] `internal/handler/admin.go` 新增
  - `ApproveArticle(c *gin.Context)`
- [ ] 路由注册 `POST /api/v1/admin/articles/:id/approve`

### 2.4 审核驳回

- [ ] `internal/service/admin.go` 新增
  - `RejectArticle(ctx, adminID, articleID int64, reason string, clientIP string) error`
    1. 校验 reason 不为空（必填驳回原因）
    2. 状态改为 `draft`
    3. `review_comment` 写入驳回原因
    4. 记录审计日志：`reject_article`
    5. 清除文章缓存
    6. **发送通知给作者**：类型 `review_rejected`，内容含文章标题 + 驳回原因
- [ ] 请求体 `RejectArticleReq`
  ```go
  type RejectArticleReq struct {
      Reason string `json:"reason" binding:"required,min=1,max=500"`
  }
  ```
- [ ] `internal/handler/admin.go` 新增
  - `RejectArticle(c *gin.Context)`
- [ ] 路由注册 `POST /api/v1/admin/articles/:id/reject`

### 2.5 通知系统扩展

- [ ] `internal/model/notification.go` 新增通知类型常量
  ```go
  NotificationTypeReviewApproved = "review_approved"
  NotificationTypeReviewRejected = "review_rejected"
  ```
- [ ] `internal/service/notification.go` 新增
  - `NotifyReviewApproved(ctx, userID int64, articleTitle, articleSlug string)`
  - `NotifyReviewRejected(ctx, userID int64, articleTitle, reason string)`
  - 通知 content JSON 示例：
    ```json
    // 通过
    { "type": "review_approved", "title": "文章审核通过", "article_title": "Go并发入门", "article_slug": "go-concurrency-101" }
    // 驳回
    { "type": "review_rejected", "title": "文章审核未通过", "article_title": "Go并发入门", "reason": "内容过于简单，建议补充示例代码" }
    ```

### 2.6 用户侧：查看审核状态 & 驳回原因

- [ ] `GET /api/v1/user/articles?status=pending_review` — 作者查看自己待审核的文章
- [ ] 文章详情返回 `review_comment` 字段（仅作者本人和管理员可见）
- [ ] 作者修改文章后可重新提交审核

### 2.7 验证标准

```bash
# === 用户侧 ===
# 用户提交文章审核（status 设为 published → 后端自动转为 pending_review）
curl -X PATCH :8080/api/v1/articles/1/status \
  -H "Authorization: Bearer <user_token>" \
  -d '{"status": "published"}'
# → 实际变为 pending_review

# 用户查看自己的待审核文章
curl ":8080/api/v1/user/articles?status=pending_review" \
  -H "Authorization: Bearer <user_token>"

# === 管理员侧 ===
# 查看待审核文章列表
curl ":8080/api/v1/admin/articles/pending?page=1" \
  -H "Authorization: Bearer <admin_token>"

# 审核通过
curl -X POST :8080/api/v1/admin/articles/1/approve \
  -H "Authorization: Bearer <admin_token>"

# 审核驳回（必须带 reason）
curl -X POST :8080/api/v1/admin/articles/2/reject \
  -H "Authorization: Bearer <admin_token>" \
  -H "Content-Type: application/json" \
  -d '{"reason":"文章内容与标题不符，请修改后重新提交"}'

# === 用户收到通知 ===
# 作者登录后 GET /api/v1/user/notifications
# 看到 review_approved 或 review_rejected 类型通知

# === 作者修改后重新提交 ===
# 编辑文章 → 再次 PATCH status=published → 再次进入 pending_review
```

<details>
<summary>📁 Step 2 涉及文件</summary>

```
migrations/
└── XXXXXX_add_review_comment.up.sql    # 新增 review_comment 列
internal/
├── model/article.go                    # 修改 新增常量 + ReviewComment 字段
├── model/notification.go               # 修改 新增通知类型
├── repository/article.go               # 修改 新增 ListPending
├── service/article.go                  # 修改 发布逻辑 + SubmitForReview / WithdrawReview
├── service/admin.go                    # 修改 新增 ApproveArticle / RejectArticle / ListPendingArticles
├── service/notification.go             # 修改 新增审核通知方法
├── handler/article.go                  # 修改 ChangeStatus 逻辑
├── handler/admin.go                    # 修改 新增审核端点
└── router/router.go                    # 修改 新增路由
```
</details>

---

## Step 3：文章管理

**目标：** 管理员可查看、搜索、强制删除全站文章（审核之外的通用管理）。

### 3.1 后端

- [ ] `internal/repository/article.go` 新增
  - `ListAll(page, pageSize int, filters AdminArticleFilter) ([]Article, int64, error)`
    - 过滤条件：`keyword`（标题搜索）、`status`（含 `pending_review`）、`user_id`（按作者）、`category_id`（按分类）、`date_from` / `date_to`（时间范围）
    - 默认按 `created_at` 倒序
  - `ForceDelete(id int64) error` — 硬删除文章（仅管理员）
- [ ] `internal/service/admin.go` 新增
  - `ListArticles(ctx, page, pageSize int, filters) (*PageResult, error)`
  - `DeleteArticle(ctx, adminID, articleID int64, clientIP string) error` — 硬删除 + 清理关联数据 + 记录审计日志
- [ ] `internal/handler/admin.go` 新增
  - `ListArticles(c *gin.Context)`
  - `DeleteArticle(c *gin.Context)`

### 3.2 接口说明

```
GET    /api/v1/admin/articles?page=1&page_size=20&keyword=xxx&status=pending_review&user_id=5&category_id=2
DELETE /api/v1/admin/articles/:id             强制删除文章（硬删除）
```

### 3.3 验证标准

```bash
curl ":8080/api/v1/admin/articles?page=1&status=pending_review" \
  -H "Authorization: Bearer <admin_token>"

curl -X DELETE :8080/api/v1/admin/articles/123 \
  -H "Authorization: Bearer <admin_token>"
```

<details>
<summary>📁 Step 3 涉及文件</summary>

```
internal/
├── repository/article.go         # 修改 新增 ListAll / ForceDelete
├── service/admin.go              # 修改 新增文章管理方法
├── handler/admin.go              # 修改 新增文章管理端点
└── router/router.go              # 修改 新增路由
```
</details>

---

## Step 4：评论管理

**目标：** 管理员可查看、搜索、强制删除评论。

### 4.1 后端

- [ ] `internal/repository/comment.go` 新增
  - `ListAll(page, pageSize int64, filters AdminCommentFilter) ([]Comment, int64, error)`
    - 过滤条件：`keyword`（内容模糊搜索）、`user_id`（按评论者）、`article_id`（按文章）、`date_from` / `date_to`
    - 预加载 `User`（评论者昵称 + 头像）和 `Article`（文章标题）
  - `ForceDelete(id int64) error` — 硬删除
- [ ] `internal/service/admin.go` 新增
  - `ListComments(ctx, page, pageSize int, filters) (*PageResult, error)`
  - `DeleteComment(ctx, adminID, commentID int64, clientIP string) error`
    - 有子回复时：保留子回复，将被删评论内容替换为 `[该评论已被管理员删除]`
    - 记录审计日志
- [ ] `internal/handler/admin.go` 新增
  - `ListComments(c *gin.Context)`
  - `DeleteComment(c *gin.Context)`

### 4.2 接口说明

```
GET    /api/v1/admin/comments?page=1&page_size=20&keyword=xxx&user_id=5&article_id=123
DELETE /api/v1/admin/comments/:id             强制删除评论
```

### 4.3 验证标准

```bash
curl ":8080/api/v1/admin/comments?keyword=广告" \
  -H "Authorization: Bearer <admin_token>"

curl -X DELETE :8080/api/v1/admin/comments/456 \
  -H "Authorization: Bearer <admin_token>"
```

<details>
<summary>📁 Step 4 涉及文件</summary>

```
internal/
├── repository/comment.go         # 修改 新增 ListAll / ForceDelete
├── service/admin.go              # 修改 新增评论管理方法
├── handler/admin.go              # 修改 新增评论管理端点
└── router/router.go              # 修改 新增路由
```
</details>

---

## Step 5：管理后台前端页面

**目标：** 一个完整可用的 SPA 管理后台页面，审核队列为第一优先级。

### 5.1 页面结构

```
web/admin.html
├── 顶部栏
│   ├── 站点标题 + "管理后台"
│   ├── 待审核数量 Badge（醒目提示）
│   └── 管理员头像 + 退出按钮
├── 左侧导航
│   ├── 📋 待审核 (12)      ← 默认首页，带未审数量角标
│   ├── 📊 仪表盘
│   ├── 📝 文章管理
│   ├── 💬 评论管理
│   ├── 👥 用户管理
│   ├── 📂 分类管理
│   ├── 🏷️  标签管理
│   └── 📋 审计日志
└── 右侧内容区（动态切换）
```

### 5.2 待审核 Tab（默认首页，最高优先级）

- [ ] 数据表格：ID | 标题 | 作者 | 提交时间 | 操作
- [ ] 每行显示文章摘要（截断 100 字）
- [ ] 操作按钮：
  - **👁️ 预览** — 弹窗展示文章完整内容（含渲染后的 HTML）
  - **✅ 通过** — 二次确认 + Toast 提示
  - **❌ 驳回** — 弹出驳回原因输入框（必填，最大 500 字）+ 确认
- [ ] 通过/驳回后该行从列表消失（带动画）
- [ ] 顶部待审核数量实时更新
- [ ] 空状态："暂无待审核文章 🎉"

### 5.3 仪表盘 Tab

- [ ] 统计卡片行：用户数 | 文章数 | 待审核数 | 评论数
  - 待审核卡片突出显示，点击跳转到审核 Tab
  - 每个卡片带增长趋势（本周新增 vs 上周）
- [ ] 最新用户表格（Top 10）
- [ ] 最新文章列表（Top 10）
- [ ] 高产作者排行

### 5.4 文章管理 Tab

- [ ] 数据表格：ID | 标题 | 作者 | 状态 Tag（不同颜色）| 分类 | 阅读量 | 发布时间 | 操作
  - `pending_review` — 橙色 Tag
  - `published` — 绿色 Tag
  - `draft` — 灰色 Tag
  - `archived` — 蓝色 Tag
- [ ] 筛选栏：状态下拉（含 `pending_review`）| 分类下拉 | 关键词搜索 | 时间范围
- [ ] 操作按钮：查看（新窗口）| 强制删除（二次确认）

### 5.5 评论管理 Tab

- [ ] 数据表格：ID | 评论内容（截断）| 评论者 | 所属文章 | 发布时间 | 操作
- [ ] 筛选栏：关键词搜索 | 时间范围
- [ ] 操作按钮：查看原文章 | 强制删除（二次确认）

### 5.6 用户管理 Tab

- [ ] 数据表格：ID | 头像 | 用户名 | 邮箱 | 角色 | 状态 | 注册时间 | 操作
- [ ] 筛选栏：关键词搜索（用户名/邮箱）
- [ ] 操作按钮：封禁/解封 | 设为管理员/取消管理员
  - 所有操作二次确认，不可操作自己

### 5.7 分类 & 标签管理 Tab

- [ ] 分类：列表 + 新建/编辑弹窗 + 删除（有文章的不可删）
- [ ] 标签：列表 + 新建/编辑弹窗 + 删除

### 5.8 登录态

- [ ] 普通用户访问时展示 403 "无管理权限"
- [ ] 管理员 token 过期时跳转到首页登录
- [ ] 退出按钮清除 token 返回首页

### 5.9 路由注册

- [ ] 注册 `GET /admin.html` 指向 `web/admin.html`
- [ ] `GET /admin*` → 静态文件回退（支持 SPA 内刷新）

<details>
<summary>📁 Step 5 涉及文件</summary>

```
web/
├── admin.html                     # 新增 管理后台 SPA
├── css/admin.css                  # 新增 管理后台样式
└── js/admin.js                    # 新增 管理后台逻辑
internal/
└── router/router.go               # 修改 注册 /admin.html 路由
```
</details>

---

## Step 6：审计日志 & 收尾

**目标：** 关键管理操作可追溯，后台系统收尾完善。

### 6.1 审计日志

- [ ] `internal/model/audit_log.go` — 审计日志模型
  ```go
  type AuditLog struct {
      ID         int64     `gorm:"primaryKey;autoIncrement"`
      AdminID    int64     `gorm:"index;not null"`
      Action     string    `gorm:"index;not null"`
      TargetType string                              // user / article / comment / category / tag
      TargetID   int64
      Detail     string    `gorm:"type:text"`
      IP         string
      CreatedAt  time.Time
  }
  ```
- [ ] 数据库迁移脚本 — 创建 `audit_logs` 表
- [ ] `internal/repository/audit_log.go`
  - `Create(log *AuditLog) error`
  - `List(page, pageSize int, filters AuditLogFilter) ([]AuditLog, int64, error)`
    - 过滤：`admin_id` / `action` / `target_type` / `date_from` / `date_to`
    - 预加载 Admin 用户信息
- [ ] `internal/service/admin.go` 新增
  - `LogAudit(ctx, adminID int64, action, targetType string, targetID int64, detail, ip string)`
  - `GetAuditLogs(ctx, page, pageSize int, filters) (*PageResult, error)`
- [ ] 在已有管理操作中嵌入日志记录：
  - 文章审核通过 → `LogAudit(ctx, adminID, "approve_article", "article", targetID, ...)`
  - 文章审核驳回 → `LogAudit(ctx, adminID, "reject_article", "article", targetID, ...)`
  - 用户状态变更 → `LogAudit(ctx, adminID, "ban_user" / "unban_user", "user", targetID, ...)`
  - 用户角色变更 → `LogAudit(ctx, adminID, "change_role", "user", targetID, ...)`
  - 文章删除 → `LogAudit(ctx, adminID, "delete_article", "article", targetID, ...)`
  - 评论删除 → `LogAudit(ctx, adminID, "delete_comment", "comment", targetID, ...)`
  - 分类/标签创建/编辑/删除 → 同理

### 6.2 接口

```
GET /api/v1/admin/audit-logs?page=1&page_size=20&action=approve_article&date_from=2026-06-01

Response:
{
  "code": 0,
  "data": {
    "list": [
      {
        "id": 1001,
        "admin": { "id": 1, "username": "superadmin" },
        "action": "approve_article",
        "target_type": "article",
        "target_id": 456,
        "detail": "{\"title\":\"Go并发入门\"}",
        "ip": "192.168.1.100",
        "created_at": "2026-06-22T10:30:00+08:00"
      }
    ],
    "pagination": { ... }
  }
}
```

### 6.3 管理后台前端 — 审计日志 Tab

- [ ] 在左侧导航新增"📋 审计日志"
- [ ] 数据表格：操作者 | 操作类型（中文映射）| 目标类型 | 目标 ID | 详情摘要 | 操作 IP | 时间
- [ ] 筛选栏：操作者下拉 | 操作类型下拉 | 目标类型下拉 | 时间范围
- [ ] 分页

### 6.4 收尾

- [ ] 管理员后台 Route Guard：前端加载时调 `GET /api/v1/user/profile`，role 非 `admin` 则 403
- [ ] 管理员不能操作自己（封禁/降级按钮置灰）
- [ ] 删除/驳回操作统一二次确认弹窗（含被删内容摘要）
- [ ] 移动端响应式适配（最小 768px 可用）
- [ ] 非管理员访问 `/admin*` 时后端也返回 403
- [ ] Swagger 补充 admin 分组注解（含审核端点）

### 6.5 验证标准

```bash
# E2E 流程
1. 用户 A 登录 → 创建文章 → 提交审核（status=published → pending_review）
2. 管理员登录 → 打开 admin.html
3. 首页"待审核"Tab 看到文章 → 点击预览 → 觉得有问题 → 驳回（填原因）
4. 用户 A 收到通知 "文章审核未通过: 内容过于简单" → 修改文章 → 重新提交
5. 管理员再次审核 → 通过
6. 用户 A 收到通知 "文章审核通过"
7. 首页文章列表出现该文章（仅 published 对公众可见）
8. 审计日志可见 approve_article 和 reject_article 各 1 条
```

<details>
<summary>📁 Step 6 涉及文件</summary>

```
migrations/
├── XXXXXX_create_audit_logs.up.sql    # 新增
└── XXXXXX_create_audit_logs.down.sql  # 新增
internal/
├── model/audit_log.go                 # 新增
├── repository/audit_log.go            # 新增
├── service/admin.go                   # 修改 新增日志相关方法
├── handler/admin.go                   # 修改 新增审计日志端点
└── router/router.go                   # 修改 新增路由
web/
├── admin.html                         # 修改 新增审计日志 Tab
├── css/admin.css                      # 修改 新增审计日志样式
└── js/admin.js                        # 修改 新增审计日志逻辑
```
</details>

---

## 最终 API 汇总

完成全部 6 步后，管理后台 API 全景：

```
┌───────────────────────────────────────────────┐
│                 Admin API v1                    │
├──────────────┬─────────────────────────────────┤
│   Dashboard   │ GET    /admin/dashboard         │
├──────────────┼─────────────────────────────────┤
│  文章审核 ★   │ GET    /admin/articles/pending  │
│              │ POST   /admin/articles/:id/approve │
│              │ POST   /admin/articles/:id/reject  │
├──────────────┼─────────────────────────────────┤
│  文章管理     │ GET    /admin/articles           │
│              │ DELETE /admin/articles/:id        │
├──────────────┼─────────────────────────────────┤
│  评论管理     │ GET    /admin/comments           │
│              │ DELETE /admin/comments/:id        │
├──────────────┼─────────────────────────────────┤
│  用户管理     │ GET    /admin/users              │
│              │ PATCH  /admin/users/:id/status    │
│              │ PATCH  /admin/users/:id/role      │
├──────────────┼─────────────────────────────────┤
│  分类管理     │ POST   /admin/categories         │
│              │ PUT    /admin/categories/:id      │
│              │ DELETE /admin/categories/:id      │
├──────────────┼─────────────────────────────────┤
│  标签管理     │ POST   /admin/tags               │
│              │ PUT    /admin/tags/:id            │
│              │ DELETE /admin/tags/:id            │
├──────────────┼─────────────────────────────────┤
│  审计日志     │ GET    /admin/audit-logs          │
└──────────────┴─────────────────────────────────┘
```

---

## 前端页面路由

```
/admin.html          → 管理后台 SPA（需 admin 角色）
                        默认展示"待审核"队列
```

所有管理 API 均需要 `Authorization: Bearer <token>` + `role=admin`。

---

## 用户侧配套改动总览

审核系统不仅仅是后台加接口，用户侧也需要配合改动：

| 改动点 | 说明 |
|--------|------|
| 发布按钮 | 用户点"发布"后，状态变为 `pending_review`，前端提示"已提交审核，请耐心等待" |
| 我的文章列表 | 新增"待审核"筛选项，显示审核中文章 |
| 文章详情页 | 被驳回的文章展示驳回原因（`review_comment`）和"重新提交"按钮 |
| 通知中心 | 新增 `review_approved` / `review_rejected` 两种通知展示 |
| 公开文章列表 | `GET /api/v1/articles` 只返回 `published` 状态文章 |

---

## Commit 建议

```bash
git commit -m "feat(admin): step 1 - dashboard stats API"
git commit -m "feat(admin): step 2 - article review system (approve/reject/notify)"
git commit -m "feat(admin): step 3 - article management for admin"
git commit -m "feat(admin): step 4 - comment management for admin"
git commit -m "feat(admin): step 5 - admin frontend SPA page"
git commit -m "feat(admin): step 6 - audit log & final polish"
```
