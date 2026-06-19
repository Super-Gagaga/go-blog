# Web 静态页后端待办清单

当前 `web/index.html` 已接入现有后端能力：

- `GET /api/v1/articles`：首页文章列表、搜索、分类筛选、文章计数。
- `GET /api/v1/categories`：左侧专题导航和顶部分类标签。
- `GET /api/v1/tags`：右侧热门话题候选。
- `GET /api/v1/articles/:id`：文章详情接口，供 `article-detail.html` 页面消费。
- `GET /api/v1/articles/:id/comments`：文章评论列表。
- `GET /api/v1/articles/ranking?period=week`：本周精选/排行榜。

以下能力前端已有入口或展示位，但后端还缺少面向页面的完整接口，需要后续补齐。

## 1. 文章详情页面

- [x] 新增可渲染 HTML 的文章详情页（`web/article-detail.html`），消费 `GET /api/v1/articles/:id`。
- [x] 详情页展示 `content_html`、作者信息、分类、标签、阅读数、评论数、点赞数、收藏数。
- [x] 详情页接入评论列表 `GET /api/v1/articles/:id/comments`（树形展示评论与回复）。
- [x] 首页文章卡片点击跳转到详情页而非 JSON 接口。
- [x] 路由注册 `/article-detail.html` 静态文件。

## 2. 推荐作者

- [x] 新增公开接口 `GET /api/v1/recommendations/authors`。
- [x] 返回推荐作者列表：`id`、`username`、`avatar`、`bio`、`article_count`、`follower_count`。
- [x] 按粉丝数、发文数降序排列（活跃度优先）。
- [x] 首页右侧通过 `renderAuthors()` 异步加载推荐作者数据。
- [x] 接口支持 `limit` 参数控制返回数量（默认 5，最大 50）。

## 3. 热门话题

- [x] 新增公开接口 `GET /api/v1/topics/trending`。
- [x] 返回话题名、关联 tag_id、文章数（article_count）、近期热度分数（hotness_score）。
- [x] 热度分数综合最近30天文章数（权重0.7）与总文章数（权重0.3）。
- [x] 点击话题后按 `tag_id` 精准过滤文章（通过 `GET /api/v1/articles?tag_id=...`），不再仅用 keyword 模糊搜索。
- [x] 接口支持 `limit` 参数控制返回数量（默认 12，最大 50）。

## 4. 本周精选/排行榜 ✅

- [x] 新增公开接口 `GET /api/v1/articles/ranking?period=week`。
- [x] 排行规则综合阅读数(0.3)、点赞数(2)、收藏数(3)、评论数(2.5)和时间衰减(POWER(hours+2, 1.5))。
- [x] 返回文章基础信息（含作者、分类、标签）、阅读数、收藏数、热度分数。
- [x] 支持 `period=day|week|month` 参数和 `limit` 参数（默认10，最大50）。
- [x] 首页右侧"本周精选"通过 `renderDigest()` 异步加载排行榜数据。

## 5. 关注与个人入口 ✅

- [x] 首页”关注”入口链接到 `following.html` 页面（需登录态），不再直接打开 JSON 接口。
- [x] 推荐作者卡片的”关注”按钮接入 `POST /api/v1/users/:id/follow`，未登录时提示登录。
- [x] 消息入口链接到 `notifications.html` 页面，消费 `GET /api/v1/user/notifications`。
- [x] `following.html` 支持”我关注的”/”关注我的”两个标签切换。
- [x] `notifications.html` 展示通知列表，支持按类型图标区分、标记单条/全部已读。
- [x] 通知接口返回富化的 `from_user` 和 `article` 信息，前端直接渲染。

## 6. 写文章入口 ✅

- [x] 新增文章编辑器页面 `web/editor.html`。
- [x] 编辑器接入 `POST /api/v1/articles`（新建）和 `PUT /api/v1/articles/:id`（编辑）。
- [x] 接入 `POST /api/v1/articles/upload` 上传封面图。
- [x] 提供分类下拉选择器和标签多选芯片。
- [x] 支持草稿/发布/归档三种状态切换。
- [x] Ctrl+S 快捷键快速保存，预览按钮可弹窗预览 Markdown 渲染效果。
- [x] 路由注册 `/editor.html` 并通过 URL 参数 `?id=` 支持编辑已有文章。

## 7. 首页聚合接口 ✅

- [x] 新增 `GET /api/v1/home`，一次性返回首页所需数据。
- [x] 响应包含：`articles`（文章列表+分页）、`categories`、`trending_topics`、`recommended_authors`、`weekly_picks`。
- [x] 4 个子查询通过 goroutine 并发执行，减少响应延迟。
- [x] 后台各子模块的 Redis 缓存（文章列表、作者推荐、话题）仍然生效。
- [x] 接口可选参数：`page`、`page_size`。
