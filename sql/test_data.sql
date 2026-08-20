-- ============================================================================
-- go-blog 测试数据
--
-- 目标数据库：MySQL 8.4+
-- 使用方式：
--   mysql -u root -p go_blog < sql/test_data.sql
--
-- 说明：
--   1. 执行本脚本前，先运行 sql/init.sql 或执行迁移。
--   2. 本脚本可重复执行。种子数据使用 1000+ / 2000+ / 3000+ 范围的 ID，
--      并在重新插入前仅删除该范围内的行。
--   3. 所有测试账号的密码均为：password123
-- ============================================================================

SET NAMES utf8mb4;
SET FOREIGN_KEY_CHECKS = 0;

-- ----------------------------------------------------------------------------
-- 清理上一次的种子数据。为了兼容性，先删除子表记录。
-- ----------------------------------------------------------------------------
DELETE FROM audit_logs
WHERE id BETWEEN 6001 AND 6999
   OR admin_id BETWEEN 1001 AND 1099;
DELETE FROM notifications WHERE id BETWEEN 5001 AND 5999;
DELETE FROM favorites WHERE user_id BETWEEN 1001 AND 1099 OR article_id BETWEEN 2001 AND 2099;
DELETE FROM likes WHERE user_id BETWEEN 1001 AND 1099 OR article_id BETWEEN 2001 AND 2099;
DELETE FROM follows WHERE follower_id BETWEEN 1001 AND 1099 OR followee_id BETWEEN 1001 AND 1099;
DELETE FROM comments WHERE id BETWEEN 3001 AND 3999;
DELETE FROM article_tags WHERE article_id BETWEEN 2001 AND 2099 OR tag_id BETWEEN 1001 AND 1099;
DELETE FROM articles WHERE id BETWEEN 2001 AND 2099;
DELETE FROM tags WHERE id BETWEEN 1001 AND 1099;
DELETE FROM categories WHERE id BETWEEN 1001 AND 1099;
DELETE FROM users WHERE id BETWEEN 1001 AND 1099;

SET FOREIGN_KEY_CHECKS = 1;

START TRANSACTION;

-- ----------------------------------------------------------------------------
-- 用户
-- 密码哈希值为 bcrypt(123456)。
-- article_count 统计已发布（status='published'）的文章数量。
-- ----------------------------------------------------------------------------
INSERT INTO users (
    id, username, email, password, avatar, bio, role, status,
    article_count, follower_count, following_count, created_at, updated_at
) VALUES
    (
        1001, 'admin', 'admin@example.com',
        '$2b$10$zXc2E/7tSgBpze6cTb6JZ.oZCVuL64xJAhlJBOg.qEhMvhPJ2q3N2',
        'https://api.dicebear.com/7.x/initials/svg?seed=Admin',
        '平台管理员，负责内容发布审核。', 'admin', 'active',
        1, 1, 2, '2026-06-15 00:00:00.000', '2026-06-22 02:14:43.000'
    ),
    (
        1002, 'alice', 'alice@example.com',
        '$2b$10$zXc2E/7tSgBpze6cTb6JZ.oZCVuL64xJAhlJBOg.qEhMvhPJ2q3N2',
        'https://api.dicebear.com/7.x/initials/svg?seed=Alice',
        'Go 后端工程师，专注于 API 与服务设计。', 'user', 'active',
        2, 4, 2, '2026-06-15 02:18:49.000', '2026-06-22 21:56:40.000'
    ),
    (
        1003, 'bob', 'bob@example.com',
        '$2a$10$KIXQXfl6ZVIwkV3GEPvqKuoeUry6RFSP4ePLiEhlG8cu35fATVmoa',
        'https://api.dicebear.com/7.x/initials/svg?seed=Bob',
        '数据库爱好者，专注 MySQL 与查询性能优化。', 'user', 'active',
        2, 3, 2, '2026-06-15 07:01:23.000', '2026-06-22 20:24:35.000'
    ),
    (
        1004, 'carol', 'carol@example.com',
        '$2a$10$KIXQXfl6ZVIwkV3GEPvqKuoeUry6RFSP4ePLiEhlG8cu35fATVmoa',
        'https://api.dicebear.com/7.x/initials/svg?seed=Carol',
        '云原生开发者，热衷于可观测性与 Redis。', 'user', 'active',
        2, 3, 3, '2026-06-15 12:23:24.000', '2026-06-22 23:59:59.000'
    ),
    (
        1005, 'dave', 'dave@example.com',
        '$2a$10$KIXQXfl6ZVIwkV3GEPvqKuoeUry6RFSP4ePLiEhlG8cu35fATVmoa',
        'https://api.dicebear.com/7.x/initials/svg?seed=Dave',
        '偏前端方向的全栈开发者，负责测试编辑器流程。', 'user', 'active',
        1, 1, 3, '2026-06-16 11:05:31.000', '2026-06-22 18:27:08.000'
    ),
    (
        1006, 'eve', 'eve@example.com',
        '$2a$10$KIXQXfl6ZVIwkV3GEPvqKuoeUry6RFSP4ePLiEhlG8cu35fATVmoa',
        'https://api.dicebear.com/7.x/initials/svg?seed=Eve',
        '已被封禁的测试账号，用于管理员状态管理测试。', 'user', 'banned',
        0, 0, 0, '2026-06-16 21:29:50.000', '2026-06-22 12:47:30.000'
    );

-- ----------------------------------------------------------------------------
-- 分类和标签
-- ----------------------------------------------------------------------------
INSERT INTO categories (id, name, slug, created_at) VALUES
    (1001, '测试',         'testing',        '2026-06-15 00:00:28.000'),
    (1002, 'DevOps',       'devops',         '2026-06-15 00:00:31.000'),
    (1003, '可观测性',     'observability',  '2026-06-15 00:00:33.000');

INSERT INTO tags (id, name, created_at) VALUES
    (1001, '测试',       '2026-06-15 00:00:56.000'),
    (1002, 'API',        '2026-06-15 00:00:59.000'),
    (1003, '性能',       '2026-06-15 00:01:02.000'),
    (1004, 'Docker',     '2026-06-15 00:01:04.000'),
    (1005, '可观测性',   '2026-06-15 00:01:07.000'),
    (1006, 'JWT',        '2026-06-15 00:01:10.000'),
    (1007, '中间件',     '2026-06-15 00:01:13.000'),
    (1008, '缓存',       '2026-06-15 00:01:16.000');

-- ============================================================================
-- 文章
-- 评论数、点赞数、收藏数与后续插入的 comments / likes / favorites 一致。
-- ============================================================================
INSERT INTO articles (
    id, title, slug, content, content_html, summary, cover_image, status,
    view_count, comment_count, like_count, favorite_count,
    user_id, category_id, created_at, updated_at
) VALUES
    -- =========================================================================
    -- 2001 · 已发布 — alice
    -- =========================================================================
    (
        2001,
        '构建可组合的 Gin 中间件',
        'building-composable-gin-middleware',
        '# 构建可组合的 Gin 中间件\n\n'
        '中间件是 Gin 的灵魂。每一个 HTTP 请求在到达业务处理器之前，都要穿过一层层中间件的"洋葱圈"——'
        '日志、认证、限流、恢复，每一层都只做一件事，却能像乐高积木一样任意组合。\n\n'
        '## 从请求追踪开始\n\n'
        '我们做的第一个中间件是 RequestID。它在请求进入时生成一个 UUID，注入到 `context.Context` 和 '
        '响应头 `X-Request-ID` 中。前端拿到这个 ID 之后，报 Bug 的时候顺手带上，后端就能在海量日志里秒级定位到那次请求。'
        '实现起来不到 30 行代码，但它几乎是所有后端服务的\"第一层\"。\n\n'
        '## 认证守卫的陷阱\n\n'
        'JWT 认证中间件看起来简单——解析 Token、查库、塞进 context——但很多人在错误处理上栽了跟头。'
        '中间件里 `c.Abort()` 之后如果不 `return`，后面的代码还会继续执行；如果只 `return` 不 `Abort`，'
        '请求照样会流到业务处理器。两者缺一不可。另外，把用户对象挂在 context 上虽然方便，'
        '但记得用自定义类型做 key，避免和其他中间件冲突。\n\n'
        '## 恢复层不只是 recover\n\n'
        'Gin 自带的 `gin.Recovery()` 在 panic 发生时打印堆栈、返回 500 就完事了。'
        '我们在此基础上加了三个增强：1) 用结构化日志替代裸字符串输出；2) 上报 panic 次数到 Prometheus 计数器；'
        '3) 对 `json: cannot unmarshal` 这类已知错误返回 400 而非 500，减少告警噪音。'
        '这样一来，panic 从\"服务炸了\"变成\"可以追溯、可以量化、可以分级的信号\"。\n\n'
        '## 组合优于排列\n\n'
        '最终我们的中间件栈长这样：`Recovery → TraceID → Logger → CORS → RateLimit → Auth → handler`。'
        '每个中间件只依赖 `*gin.Context`，彼此之间零耦合。加新功能不需要改旧代码——这是 compose 的魅力。',
        '<h1>构建可组合的 Gin 中间件</h1>'
        '<p>中间件是 Gin 的灵魂。每一个 HTTP 请求在到达业务处理器之前，都要穿过一层层中间件的"洋葱圈"——'
        '日志、认证、限流、恢复，每一层都只做一件事，却能像乐高积木一样任意组合。</p>'
        '<h2>从请求追踪开始</h2>'
        '<p>我们做的第一个中间件是 RequestID。它在请求进入时生成一个 UUID，注入到 context 和响应头中。'
        '实现起来不到 30 行代码，但它几乎是所有后端服务的"第一层"。</p>'
        '<h2>认证守卫的陷阱</h2>'
        '<p>JWT 认证中间件看似简单，但 <code>c.Abort()</code> 之后必须 <code>return</code>，两者缺一不可。</p>'
        '<h2>恢复层不只是 recover</h2>'
        '<p>我们在 Gin 自带的 Recovery 基础上增加了结构化日志、Prometheus 指标和已知错误分级。</p>'
        '<h2>组合优于排列</h2>'
        '<p>每个中间件只依赖 <code>*gin.Context</code>，彼此之间零耦合——这是 compose 的魅力。</p>',
        '深入 Gin 中间件栈：从 RequestID、JWT 认证到增强版 Recovery，'
               '探讨如何设计零耦合、可组合的中间件体系。',
        'https://images.unsplash.com/photo-1515879218367-8466d910aaa4',
        'published', 1280, 3, 3, 2,
        1002, 1001, '2026-06-20 15:20:28.000', '2026-06-21 12:45:46.000'
    ),
    -- =========================================================================
    -- 2002 · 已发布 — bob
    -- =========================================================================
    (
        2002,
        '真正能帮到博客 Feed 的 MySQL 索引',
        'mysql-indexes-that-help-blog-feeds',
        '# 真正能帮到博客 Feed 的 MySQL 索引\n\n'
        '博客系统的 Feed 流查询有一个共同特征：按时间倒序、多条件组合、分页取前 N 条。'
        '这类查询看似简单，一旦数据量上了百万级，没有正确的索引，每次刷新 Feed 都能把数据库打穿。\n\n'
        '## 复合索引的列顺序至关重要\n\n'
        '假设 Feed 主查询是 `WHERE status = ''published'' AND deleted_at IS NULL ORDER BY created_at DESC`。'
        '如果你建了三个单列索引，MySQL 只会挑其中一个（通常是选择性最高的 `status`），剩下的过滤 '
        '仍然要回表逐行扫描。正确的做法是建一个复合索引 `(status, deleted_at, created_at)`，'
        '让排序和过滤都在索引内部完成。\n\n'
        '## 覆盖索引：让 COUNT(*) 也飞起来\n\n'
        'Feed 接口通常要返回 `total` 做分页。`SELECT COUNT(*) FROM articles WHERE status = ''published''` '
        '在没有覆盖索引时，InnoDB 会走索引扫描然后回表——实际上它根本不需要行数据，只需要数行数。'
        '在 `(status, deleted_at)` 上建覆盖索引后，`EXPLAIN` 里看到 `Using index`，'
        'COUNT 从 800ms 降到了 12ms。\n\n'
        '## 关注者 Feed 的索引设计\n\n'
        '关注 Feed 是\"我所关注的用户发表的文章按时间倒序\"。SQL 长这样：'
        '`WHERE user_id IN (2,3,4) AND status = ''published'' ORDER BY created_at DESC LIMIT 20`。'
        '`IN` 子句让索引选择变得复杂——MySQL 可能选择 `user_id` 索引回表过滤，也可能全表扫描。'
        '我们最终用了 `(status, user_id, created_at)` 复合索引，并且把 `LIMIT 20` 下推到了索引扫描阶段。\n\n'
        '## 别忘了 EXPLAIN\n\n'
        '索引不是建完就完事的。数据分布变了、MySQL 版本升级了、查询条件改了——'
        '每次变更后跑一遍 `EXPLAIN FORMAT=TREE`，确认 `rows`、`filtered` 和 Extra 栏里的 `Using filesort` '
        '还在可接受范围内。养成习惯，把 EXPLAIN 输出贴在 Code Review 里。',
        '<h1>真正能帮到博客 Feed 的 MySQL 索引</h1>'
        '<p>博客系统 Feed 流查询的共同特征：按时间倒序、多条件组合、分页取前 N 条。没有正确的索引，百万级数据量下每次刷新都能打穿数据库。</p>'
        '<h2>复合索引的列顺序至关重要</h2>'
        '<p>建一个 <code>(status, deleted_at, created_at)</code> 复合索引，让排序和过滤都在索引内部完成。</p>'
        '<h2>覆盖索引：让 COUNT(*) 也飞起来</h2>'
        '<p><code>Using index</code> 的 COUNT 从 800ms 降到 12ms。</p>'
        '<h2>关注者 Feed 的索引设计</h2>'
        '<p><code>IN</code> 子句让索引选择变复杂，我们最终用了 <code>(status, user_id, created_at)</code> 复合索引。</p>'
        '<h2>别忘了 EXPLAIN</h2>'
        '<p>每次变更后跑一遍 <code>EXPLAIN FORMAT=TREE</code>，养成习惯。</p>',
        '从复合索引列顺序、覆盖索引到关注者 Feed 的实战索引设计，'
               '附 EXPLAIN 输出分析与优化验证思路。',
        'https://images.unsplash.com/photo-1558494949-ef010cbdcc31',
        'published', 860, 2, 2, 2,
        1003, 1001, '2026-06-20 23:29:30.000', '2026-06-21 15:16:34.000'
    ),
    -- =========================================================================
    -- 2003 · 已发布 — carol
    -- =========================================================================
    (
        2003,
        '小型内容平台的 Redis 缓存键设计',
        'redis-cache-keys-small-content-platform',
        '# 小型内容平台的 Redis 缓存键设计\n\n'
        '缓存不难，难的是\"什么时候失效\"和\"键名叫什么\"。一个 3 万行的小项目，'
        '缓存策略设计得当，能扛住 10 倍的读流量；设计不好，缓存反而变成延迟炸弹。\n\n'
        '## 键名即文档\n\n'
        '我们定了一条铁律：所有缓存键必须能被人一眼看懂。格式是 `{领域}:{实体}:{标识}`，比如 '
        '`article:123:detail`、`user:45:following`。键名本身就是活文档——'
        '新人入职看了键名就能理解缓存了什么，不需要翻三篇 Wiki。\n\n'
        '## 批量失效的学问\n\n'
        '文章列表缓存最难搞。一篇文章改了标题，哪些列表缓存受影响？作者分类页？标签筛选页？'
        '搜索页？如果你不知道有多少个 key 需要删，就只能设短 TTL 靠过期硬扛。'
        '我们的方案是用一个 Redis SET `articles:list:keys` 记录所有活跃的列表缓存键。'
        '文章变更时，取出这个 SET 里的所有 key，`Pipeline` 一把删掉。代价是一次写入多一条 `SADD`，'
        '收益是删得干干净净，不留脏缓存。\n\n'
        '## TTL 阶梯策略\n\n'
        '不是所有数据都配同样的 TTL。文章详情 10 分钟（改了马上生效），列表缓存 5 分钟（允许短暂不一致），'
        '点赞/收藏集合永久保留（因为它们是真相的副本，不是缓存）。'
        '最容易被忽视的是未读通知计数——我们用了 Redis 原子增删，但在用户标记已读后直接回源 MySQL 重算，'
        '杜绝计数漂移。\n\n'
        '## 避坑清单\n\n'
        '1) 永远不要用 `KEYS *` 生产环境扫 key——用 SET 或 SCAN；'
        '2) 缓存空值防止穿透——文章 ID 不存在也塞一个 `null` 占位，TTL 设短一点就行；'
        '3) 热 Key 问题——如果一个 key 被 1000 QPS 同时访问，加本地 `sync.Map` 做 L1 缓存。',
        '<h1>小型内容平台的 Redis 缓存键设计</h1>'
        '<p>缓存不难，难的是"什么时候失效"和"键名叫什么"。设计得当，能扛住 10 倍读流量。</p>'
        '<h2>键名即文档</h2>'
        '<p>格式 <code>{领域}:{实体}:{标识}</code>，键名本身就是活文档。</p>'
        '<h2>批量失效的学问</h2>'
        '<p>用 Redis SET 记录所有活跃列表缓存键，变更时 Pipeline 批量删除。</p>'
        '<h2>TTL 阶梯策略</h2>'
        '<p>详情 10 分钟，列表 5 分钟，点赞集永久。未读计数回源重算防漂移。</p>'
        '<h2>避坑清单</h2>'
        '<p>禁用 KEYS *、缓存空值防穿透、热 Key 加本地 L1。</p>',
        '系统化梳理博客平台的 Redis 键名规范、批量失效、TTL 阶梯策略和常见踩坑点。',
        'https://images.unsplash.com/photo-1551288049-bebda4e38f71',
        'published', 640, 1, 3, 1,
        1004, 1003, '2026-06-21 10:28:36.000', '2026-06-22 02:15:39.000'
    ),
    -- =========================================================================
    -- 2004 · 待审核 — alice
    -- =========================================================================
    (
        2004,
        '草稿：JWT Refresh Token 轮换笔记',
        'draft-jwt-refresh-token-rotation-notes',
        '# 草稿：JWT Refresh Token 轮换笔记\n\n'
        '> ⚠️ 本文为工作笔记，结构尚未整理，暂时不要外发。\n\n'
        '## 现状\n\n'
        '当前实现：Access Token 15 分钟过期，Refresh Token 7 天过期，存储在 Redis '
        '`auth:refresh:{token}` → `userID`。\n\n'
        '## 问题\n\n'
        '- Refresh Token 泄漏后，攻击者可以无限刷新，直到 7 天后过期——窗口太大。\n'
        '- 没有 Refresh Token Rotation：每次 refresh 仍返回同一个 token，泄漏检测无从做起。\n'
        '- 登出时只删了 Redis 里的 key，但如果 token 已经被"偷走"且 Redis 中尚未过期，'
        '删 key 并不触发轮换。\n\n'
        '## 方案草稿\n\n'
        '1. **Rotation**：每次用 Refresh Token 换取新 Access Token 时，同时签发新 Refresh Token，'
        '旧 Token 立即作废。\n'
        '2. **Reuse Detection**：如果有人拿着已作废的旧 Refresh Token 来请求，'
        '说明可能发生了泄漏——立即作废该用户的所有 Token 并发出安全告警。\n'
        '3. **Family 机制**：每次 rotation 属于同一个 token family，通过 `family_id` 串联。'
        '单点 reused 即可识别整个 family 的泄漏。\n\n'
        '## TODO\n\n'
        '- [ ] 确认 Redis Lua 脚本实现 rotation + reuse detection 的原子性\n'
        '- [ ] 评估并发 refresh 场景：同一用户同时发两个 refresh 请求会怎样\n'
        '- [ ] 补充测试用例：正常 rotation、reuse 触发、过期 refresh、并发刷新',
        '<h1>草稿：JWT Refresh Token 轮换笔记</h1>'
        '<blockquote>本文为工作笔记，结构尚未整理，暂时不要外发。</blockquote>'
        '<h2>现状</h2><p>Access Token 15 分钟过期，Refresh Token 7 天过期。</p>'
        '<h2>问题</h2>'
        '<ul><li>Refresh Token 泄漏窗口太大</li>'
        '<li>无 Rotation 机制</li><li>登出逻辑不完善</li></ul>'
        '<h2>方案草稿</h2>'
        '<p>Rotation + Reuse Detection + Family 机制，三位一体。</p>'
        '<h2>TODO</h2>'
        '<ul><li>Lua 原子化</li><li>并发 refresh 评估</li><li>测试用例</li></ul>',
        '关于 Access Token 与 Refresh Token 轮换机制的工作笔记：Rotation、Reuse Detection 和 Family 方案草稿。',
        NULL,
        'pending_review', 0, 0, 0, 0,
        1002, 1002, '2026-06-22 05:03:50.000', '2026-06-22 05:03:50.000'
    ),
    -- =========================================================================
    -- 2005 · 已归档 — dave
    -- =========================================================================
    (
        2005,
        '已归档的部署检查清单',
        'archived-deployment-checklist',
        '# 已归档的部署检查清单\n\n'
        '> 🗄️ 本文已归档，仅保留用于 UI 测试，内容可能已过时。\n\n'
        '一份旧版部署流程 Checklist，涵盖：数据库迁移回滚验证、静态文件 CDN 预热、'
        '健康检查端点 `/healthz` 的 readiness/liveness 区分、以及蓝绿部署的流量切换顺序。'
        '随着 CI/CD 流水线升级为 GitHub Actions + ArgoCD，此清单中的手动步骤已全部自动化，不再需要人工对照执行。',
        '<h1>已归档的部署检查清单</h1>'
        '<blockquote>本文已归档，仅保留用于 UI 测试，内容可能已过时。</blockquote>'
        '<p>旧版部署 Checklist：DB 迁移、CDN 预热、健康检查、蓝绿部署。已被 CI/CD 自动化替代。</p>',
        '一份已过时的部署 Checklist，保留用于归档文章 UI 测试。',
        NULL,
        'archived', 96, 0, 0, 0,
        1005, 1002, '2026-06-19 17:58:28.000', '2026-06-20 19:54:21.000'
    ),
    -- =========================================================================
    -- 2006 · 已发布 — admin
    -- =========================================================================
    (
        2006,
        '项目路线图：社区功能',
        'project-roadmap-community-features',
        '# 项目路线图：社区功能\n\n'
        '技术博客最怕"单向输出"——作者写、读者看，没有互动回路。这篇路线图勾勒了 go-blog 社区功能的完整规划，'
        '从最简单的点赞到最复杂的 Feed 流，分三个阶段交付。\n\n'
        '## Phase 1：基础互动（已完成）\n\n'
        '点赞和收藏是整个社区系统的基本粒子。技术上用联合主键 `(user_id, article_id)` 保证幂等——'
        '点两次赞不会产生两行数据。Redis 侧维护对应 SET，查询点赞状态时先查 SET（O(1)），'
        '不要每次都去 MySQL 扫表。文章计数器 `like_count` / `favorite_count` 用 `gorm.Expr` 原子更新，'
        '配合事务保证与 Redis 的一致性。\n\n'
        '## Phase 2：社交关系（已完成）\n\n'
        '关注关系是单向的——我关注你，不代表你关注我。Follows 表用 `(follower_id, followee_id)` '
        '联合主键。通知系统在 MySQL 里用 JSON 列存储消息体，宽松灵活：'
        '`{"from_user_id": 1003, "article_id": 2001, "type": "comment"}`。'
        '通知发送走异步 goroutine，不阻塞主请求，发送失败不影响核心操作。\n\n'
        '## Phase 3：Feed 与推荐（规划中）\n\n'
        'Feed 是社区功能里最吃查询性能的部分。目前的方案是 pull-based——'
        '`SELECT * FROM articles WHERE user_id IN (关注列表) ORDER BY created_at DESC LIMIT 20`，'
        '配合复合索引扛住十万级用户。未来数据量上去后考虑 push-based 的 fan-out on write，'
        '在文章发表时就写进每个粉丝的 Redis Timeline。\n\n'
        '## 一个原则\n\n'
        '做社区功能不要贪全。\"关注→评论→点赞→收藏→通知→Feed\"，'
        '每一步做完、上线、观察数据，再做下一步。每一环需要的表结构和索引设计，这篇文章里都有。',
        '<h1>项目路线图：社区功能</h1>'
        '<p>从最简单的点赞到最复杂的 Feed 流，分三个阶段交付 go-blog 社区功能的完整规划。</p>'
        '<h2>Phase 1：基础互动</h2>'
        '<p>点赞和收藏用联合主键保证幂等，Redis SET 做状态查询，<code>gorm.Expr</code> 原子更新计数器。</p>'
        '<h2>Phase 2：社交关系</h2>'
        '<p>关注关系单向，通知用 MySQL JSON 列存储，异步 goroutine 发送不阻塞主请求。</p>'
        '<h2>Phase 3：Feed 与推荐</h2>'
        '<p>当前 pull-based + 复合索引，未来考虑 push-based fan-out on write。</p>'
        '<h2>一个原则</h2>'
        '<p>不要贪全——每一步做完、上线、观察、迭代。</p>',
        'go-blog 社区功能的完整 Roadmap：第一阶段基础互动、第二阶段社交关系、第三阶段 Feed 与推荐。',
        'https://images.unsplash.com/photo-1497366754035-f200968a6e72',
        'published', 430, 1, 1, 1,
        1001, 1002, '2026-06-22 03:35:45.000', '2026-06-22 10:21:38.000'
    ),
    -- =========================================================================
    -- 2007 · 已发布 — bob · NEW
    -- =========================================================================
    (
        2007,
        'Go 并发模式实战：从 goroutine 到 errgroup',
        'go-concurrency-patterns-goroutine-errgroup',
        '# Go 并发模式实战：从 goroutine 到 errgroup\n\n'
        'Go 的并发原语简洁到令人不安——`go` 关键字一加，函数就跑起来了。但生产环境里，'
        '"跑起来"和"跑对了"之间隔着 goroutine 泄漏、channel 死锁、错误吞没三座大山。\n\n'
        '## 不要裸起 goroutine\n\n'
        '裸 `go func()` 最大的问题是：你没法知道它什么时候结束，更没法知道它有没有出错。'
        '除非是 `go func() { s.ListenAndServe() }()` 这种贯穿整个生命周期的 goroutine，'
        '否则任何并发任务都应该有明确的"完成"信号。\n\n'
        '## context 是第一公民\n\n'
        '每一条 goroutine 的第一参数都应该是 `context.Context`。它有三重职责：'
        '1) 超时控制——`context.WithTimeout` 防止下游服务 hang 死；'
        '2) 级联取消——上游放弃后，下游 goroutine 通过 `ctx.Done()` 立刻收手；'
        '3) 值传递——trace ID、user ID 等元数据在调用链中透传。\n\n'
        '## errgroup：并发的安全带\n\n'
        '`golang.org/x/sync/errgroup` 解决了并发中最常见的三个问题：'
        '1) 任一 goroutine 出错时取消所有兄弟 goroutine；'
        '2) Wait 时收集第一个非 nil 错误；'
        '3) 通过 `SetLimit` 控制并发度，防止打爆下游连接池。'
        '在 go-blog 的通知发送模块中，我们用 `errgroup` 并发向多个用户推送——'
        '单条失败不影响其他推送，整体超时 5 秒，超时后丢弃而非无限重试。\n\n'
        '## channel 的正确姿势\n\n'
        '一个简单法则：谁写谁关。生产者关闭 channel 通知消费者"没数据了"，消费者永远不要关 channel。'
        '如果多个 goroutine 都在写同一个 channel，用 `sync.WaitGroup` + 单独的关闭 goroutine 来处理关闭逻辑。\n\n'
        '## 一句话总结\n\n'
        'Go 的并发模型给了一把锋利的刀，但用之前先想清楚：这个 goroutine 什么时候结束？出错了谁兜底？',
        '<h1>Go 并发模式实战：从 goroutine 到 errgroup</h1>'
        '<p>Go 的并发原语简洁，但生产环境中"跑起来"和"跑对了"之间隔着 goroutine 泄漏、channel 死锁、错误吞没三座大山。</p>'
        '<h2>不要裸起 goroutine</h2>'
        '<p>任何并发任务都应该有明确的"完成"信号。</p>'
        '<h2>context 是第一公民</h2>'
        '<p>超时控制、级联取消、值传递——三条职责，缺一不可。</p>'
        '<h2>errgroup：并发的安全带</h2>'
        '<p>任一出错取消全部、Wait 收集错误、SetLimit 控并发度。</p>'
        '<h2>channel 的正确姿势</h2>'
        '<p>谁写谁关。多写单关用 WaitGroup + 独立关闭 goroutine。</p>'
        '<h2>一句话总结</h2>'
        '<p>起 goroutine 前先想清楚：它什么时候结束？出错了谁兜底？</p>',
        '从 goroutine 泄漏、channel 死锁到 errgroup 最佳实践，'
               '系统讲解 Go 并发编程中的常见陷阱与工程化应对方案。',
        'https://images.unsplash.com/photo-1516116216624-53e697fedbea',
        'published', 520, 2, 2, 1,
        1003, 1001, '2026-06-22 06:46:43.000', '2026-06-22 14:40:15.000'
    ),
    -- =========================================================================
    -- 2008 · 已发布 — carol · NEW
    -- =========================================================================
    (
        2008,
        'Go 应用的 Docker 多阶段构建',
        'go-docker-multi-stage-builds',
        '# Go 应用的 Docker 多阶段构建\n\n'
        'Go 编译出来的是静态二进制，理论上只需要一个裸 `scratch` 镜像就能跑。'
        '但生产环境远不止"能跑"——还要有 CA 证书、时区文件、健康检查工具，有时还要带配置文件。'
        '多阶段构建就是为此而生的。\n\n'
        '## 三阶段模式\n\n'
        '我们的 Dockerfile 分三层：\n'
        '1) **builder**：`golang:1.26-alpine`，带 CGO、git、ca-certificates，'
        '跑 `go mod download` + `go build`。注意这一步要用 `--mount=type=cache` 缓存 Go module 和 build cache，'
        '否则每次 CI 都在重新下载依赖。\n'
        '2) **runner-prep**：从 builder 复制二进制，从 alpine 复制时区数据和 CA 证书。'
        '这个中间层把"需要什么文件"的问题彻底解决掉。\n'
        '3) **runner**：`scratch` 镜像，只包含最终的二进制 + tzdata + CA。最终镜像 14MB，'
        'golang:alpine 的 1/30。\n\n'
        '## 为什么不用 alpine 直接跑\n\n'
        'alpine 有 musl libc 和 busybox，虽然不依赖它们也能跑 Go 二进制，'
        '但 DNS 解析、TLS 握手在极端情况下会踩坑（比如某些环境 `/etc/resolv.conf` 为空时 musl 和 glibc 行为不同）。'
        '`scratch` + 拷贝的 CA 证书是最干净、最可预测的方案。\n\n'
        '## BuildKit 缓存加速\n\n'
        '`DOCKER_BUILDKIT=1` 配合 `--mount=type=cache`，能把 CI 构建时间从 4 分钟压到 40 秒。'
        '核心是缓存 Go module 目录和 `GOCACHE`，在两个 cache mount 之间复用。',
        '<h1>Go 应用的 Docker 多阶段构建</h1>'
        '<p>Go 编译静态二进制，理论上 <code>scratch</code> 就能跑——但生产还要 CA 证书、时区、健康检查。</p>'
        '<h2>三阶段模式</h2>'
        '<p>builder → runner-prep → runner，最终镜像 14MB。</p>'
        '<h2>为什么不用 alpine 直接跑</h2>'
        '<p>musl DNS 行为与 glibc 不同，scratch + 拷贝 CA 最可靠。</p>'
        '<h2>BuildKit 缓存加速</h2>'
        '<p><code>--mount=type=cache</code> 让 CI 构建从 4 分钟降到 40 秒。</p>',
        '详解 Go 应用的三阶段 Docker 构建流程：14MB 的 scratch 镜像，'
               'BuildKit 缓存加速，以及不用 alpine 直接跑的原因。',
        'https://images.unsplash.com/photo-1605745341112-85968b19335b',
        'published', 380, 1, 2, 1,
        1004, 1002, '2026-06-22 12:37:38.000', '2026-06-22 17:08:13.000'
    ),
    -- =========================================================================
    -- 2009 · 已发布 — alice · NEW
    -- =========================================================================
    (
        2009,
        '从 log.Println 到 slog：Go 结构化日志入门',
        'from-log-println-to-slog',
        '# 从 log.Println 到 slog：Go 结构化日志入门\n\n'
        'Go 1.21 引入了 `log/slog` 标准库，终结了 Go 社区长达十年的日志库混战。'
        '从此不再需要纠结 zap、zerolog、logrus 的选择——标准库已经足够好。\n\n'
        '## 三行代码开始\n\n'
        '```go\n'
        'logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))\n'
        'slog.SetDefault(logger)\n'
        '```\n'
        '把全局 logger 换成 JSON Handler，所有 `slog.Info("msg")` 自动变成 JSON 输出。'
        '不需要改业务代码，不需要传 logger 参数——虽然传参是更好的实践。\n\n'
        '## 结构化就是加 Key\n\n'
        '从 `slog.Info("user login")` 变成 `slog.Info("user login", "user_id", 1002, "ip", r.RemoteAddr)`，'
        '多出来的 key-value 在 JSON 输出里自动变成结构化字段。'
        '在 Grafana Loki 或 ELK 里可以直接按 `user_id=1002` 过滤，不用写正则抓取。\n\n'
        '## Log Level 的纪律\n\n'
        '我们定了三条规则：\n'
        '- `Debug`：开发调试，生产默认关闭；\n'
        '- `Info`：关键业务节点（登录、发表文章、支付）；\n'
        '- `Error`：需要人工介入的异常（数据库连接失败、第三方 API 超时）。\n'
        '`Warn` 我们基本不用——要么够重要升到 Error，要么只是 Info 级别。保持日志数量少而精，告警才不会变成噪音。\n\n'
        '## 和 gin 集成\n\n'
        '通过 gin 中间件把 `*slog.Logger` 注入 context，'
        '后续 handler / service / repository 都从 context 取——这样每一行日志都自动带上 `trace_id` 和 `request_path`。',
        '<h1>从 log.Println 到 slog：Go 结构化日志入门</h1>'
        '<p>Go 1.21 引入了 <code>log/slog</code> 标准库，终结了日志库混战。</p>'
        '<h2>三行代码开始</h2>'
        '<p>把全局 logger 换成 JSON Handler，所有输出自动 JSON 化。</p>'
        '<h2>结构化就是加 Key</h2>'
        '<p>多出来的 key-value 在 Loki 里可直接按字段过滤，不用写正则。</p>'
        '<h2>Log Level 的纪律</h2>'
        '<p>Debug 开发用、Info 关键节点、Error 需人工介入。Warn 不用。</p>'
        '<h2>和 gin 集成</h2>'
        '<p>中间件注入 context，每行日志自动带 trace_id。</p>',
        '从标准库 log 迁移到 slog 的实操指南：JSON 输出、结构化字段、'
               'Log Level 纪律以及与 Gin 中间件的集成方式。',
        'https://images.unsplash.com/photo-1562813733-b31f71025d54',
        'published', 290, 1, 1, 0,
        1002, 1003, '2026-06-22 17:58:57.000', '2026-06-22 17:58:57.000'
    ),
    -- =========================================================================
    -- 2010 · 已发布 — dave · NEW
    -- =========================================================================
    (
        2010,
        '手写一个令牌桶限流器',
        'handwritten-token-bucket-rate-limiter',
        '# 手写一个令牌桶限流器\n\n'
        '用别人写好的限流库只需要引入一行 import，但理解它为什么这样设计，'
        '能让你在面试和排障时从容得多。这篇文章用 60 行 Go 代码实现一个 Redis 令牌桶，并解释每一个设计决策。\n\n'
        '## 令牌桶 vs 固定窗口\n\n'
        '固定窗口的问题是边界突刺——用户在窗口末尾把配额刷光，下一秒新窗口开始又能刷一批，'
        '短短 2 秒内能发出 2 倍的请求。令牌桶通过持续填充令牌规避了这个问题：'
        '桶容量是 burst 上限，填充速率是稳态 QPS，不会出现窗口边界的瞬时翻倍。\n\n'
        '## Redis 实现要点\n\n'
        '我们用 Lua 脚本保证"检查令牌数 → 扣减 → 更新最后填充时间"三步的原子性。'
        '脚本接收三个参数：当前时间戳、填充速率、桶容量。返回值是 `{剩余令牌数, 是否被限流}`。'
        '返回剩余令牌数的目的是设置 `X-RateLimit-Remaining` 响应头——前端拿到这个值可以做退避重试。\n\n'
        '## 并发安全问题\n\n'
        'Redis 单线程执行 Lua 脚本天然保证了原子性，不用加分布式锁。'
        '但要注意 clock drift——如果多台服务器时间不同步，填充逻辑的计算会偏移。'
        '我们的解决方案是以 Redis 服务器时间 (`TIME` 命令) 而非应用服务器时间作为时间源。\n\n'
        '## 代码只有 60 行\n\n'
        '完整的 Go 实现：`internal/middleware/ratelimit.go`。核心逻辑就一个 Lua 脚本 + '
        '`redis.Eval()` 调用，加上一个 Gin middleware 函数包装。'
        '读一遍不超过 5 分钟，但能帮你理解几乎所有限流库的底层原理。',
        '<h1>手写一个令牌桶限流器</h1>'
        '<p>用 60 行 Go 代码实现 Redis 令牌桶限流，解释每一个设计决策。</p>'
        '<h2>令牌桶 vs 固定窗口</h2>'
        '<p>令牌桶消除了固定窗口的边界突刺问题——burst 是上限，稳态速率恒定。</p>'
        '<h2>Redis 实现要点</h2>'
        '<p>Lua 脚本保证三步原子性，返回剩余令牌数用于设置响应头。</p>'
        '<h2>并发安全问题</h2>'
        '<p>Redis 单线程执行 Lua 天然原子性，但 clock drift 要以 Redis TIME 为准。</p>'
        '<h2>代码只有 60 行</h2>'
        '<p>一个 Lua 脚本 + <code>redis.Eval()</code> + Gin 中间件，读完不超过 5 分钟。</p>',
        '从原理到实现：60 行 Go 代码手写 Redis 令牌桶限流器，覆盖 Lua 原子性、'
               '固定窗口 vs 令牌桶对比、clock drift 处理等关键设计点。',
        'https://images.unsplash.com/photo-1558494949-ef010cbdcc31',
        'published', 175, 0, 1, 0,
        1005, 1001, '2026-06-22 21:44:27.000', '2026-06-22 21:44:27.000'
    );

-- ----------------------------------------------------------------------------
-- 文章 ⇔ 标签关联
-- ----------------------------------------------------------------------------
INSERT INTO article_tags (article_id, tag_id) VALUES
    -- 2001: Gin 中间件
    (2001, 1002), (2001, 1006), (2001, 1007),
    -- 2002: MySQL 索引
    (2002, 1001), (2002, 1003),
    -- 2003: Redis 缓存键
    (2003, 1005), (2003, 1008),
    -- 2004: JWT 草稿
    (2004, 1006),
    -- 2005: 部署清单
    (2005, 1004),
    -- 2006: 社区路线图
    (2006, 1002), (2006, 1004), (2006, 1005),
    -- 2007: Go 并发
    (2007, 1001), (2007, 1003),
    -- 2008: Docker 构建
    (2008, 1004), (2008, 1003),
    -- 2009: slog 日志
    (2009, 1005),
    -- 2010: 令牌桶限流
    (2010, 1002), (2010, 1007), (2010, 1008);

-- ----------------------------------------------------------------------------
-- 评论。回复通过 parent_id 引用同篇文章下的父评论。
-- ----------------------------------------------------------------------------
INSERT INTO comments (id, content, article_id, user_id, parent_id, created_at) VALUES
    -- 2001: Gin 中间件
    (3001, '请求追踪中间件的例子很容易适配。我会在响应头中也加上请求 ID。', 2001, 1003, NULL, '2026-06-20 16:25:18.000'),
    (3002, '说得好。我通常会暴露 X-Request-ID，这样前端的 Bug 报告更容易追踪。', 2001, 1002, 3001, '2026-06-20 16:25:52.000'),
    (3003, '恢复层的部分帮我清理了 panic 日志输出。结构化日志 + Prometheus 计数器这个组合太实用了。', 2001, 1004, NULL, '2026-06-20 17:59:43.000'),
    -- 2002: MySQL 索引
    (3004, '能在一篇后续文章中加上 EXPLAIN 的输出示例吗？覆盖索引那段我照着试了一下，COUNT 从 1.2s 降到了 8ms。', 2002, 1002, NULL, '2026-06-21 00:21:53.000'),
    (3005, '对的，尤其是关注者 Feed 查询。`(status, user_id, created_at)` 这个复合索引的顺序能不能展开讲讲？', 2002, 1004, 3004, '2026-06-21 00:22:58.000'),
    -- 2003: Redis 缓存键
    (3006, '键名命名规范和我们在通知计数器中用的一致。`articles:list:keys` 这个 SET 方案比设短 TTL 优雅多了。', 2003, 1001, NULL, '2026-06-21 12:05:22.000'),
    -- 2006: 社区路线图
    (3007, '这份路线图和当前前端页面的规划完全吻合。期待 Phase 3 的 push-based Feed。', 2006, 1005, NULL, '2026-06-22 04:27:54.000'),
    -- 2007: Go 并发
    (3008, 'errgroup 那段写得太好了。我之前一直用 sync.WaitGroup + 手动建 error channel，换成 errgroup 后代码少了 30%。', 2007, 1002, NULL, '2026-06-22 07:55:46.000'),
    (3009, '没错，SetLimit 尤其好用——我们通知模块用 errgroup 并发推 50 个用户，limit 设为 10，连接池再也没爆过。', 2007, 1003, 3008, '2026-06-22 07:56:57.000'),
    -- 2008: Docker 多阶段构建
    (3010, '三阶段模式学到了。我之前一直 builder → scratch 两步走，中间缺了 runner-prep 那层，每次都要手动调时区问题。', 2008, 1005, NULL, '2026-06-22 13:48:06.000'),
    -- 2009: slog 日志
    (3011, '从 logrus 迁到 slog 之后，依赖少了一个，编译快了两秒——标准库真香。', 2009, 1003, NULL, '2026-06-22 19:09:25.000');

-- ----------------------------------------------------------------------------
-- 点赞和收藏
-- ----------------------------------------------------------------------------
INSERT INTO likes (user_id, article_id, created_at) VALUES
    -- 2001: Gin 中间件 — 3 赞
    (1001, 2001, '2026-06-20 17:30:08.000'),
    (1003, 2001, '2026-06-20 17:32:57.000'),
    (1004, 2001, '2026-06-20 17:35:46.000'),
    -- 2002: MySQL 索引 — 2 赞
    (1002, 2002, '2026-06-21 01:34:56.000'),
    (1004, 2002, '2026-06-21 01:37:45.000'),
    -- 2003: Redis 缓存键 — 3 赞
    (1001, 2003, '2026-06-21 13:02:41.000'),
    (1002, 2003, '2026-06-21 13:05:30.000'),
    (1003, 2003, '2026-06-21 13:08:19.000'),
    -- 2006: 社区路线图 — 1 赞
    (1002, 2006, '2026-06-22 05:40:28.000'),
    -- 2007: Go 并发 — 2 赞
    (1001, 2007, '2026-06-22 08:57:47.000'),
    (1004, 2007, '2026-06-22 09:06:14.000'),
    -- 2008: Docker 多阶段 — 2 赞
    (1002, 2008, '2026-06-22 14:38:50.000'),
    (1005, 2008, '2026-06-22 14:58:34.000'),
    -- 2009: slog 日志 — 1 赞
    (1003, 2009, '2026-06-22 19:23:31.000'),
    -- 2010: 令牌桶限流 — 1 赞
    (1004, 2010, '2026-06-22 22:35:11.000');

INSERT INTO favorites (user_id, article_id, created_at) VALUES
    -- 2001: Gin 中间件 — 2 收藏
    (1003, 2001, '2026-06-20 18:40:36.000'),
    (1005, 2001, '2026-06-20 18:46:14.000'),
    -- 2002: MySQL 索引 — 2 收藏
    (1002, 2002, '2026-06-21 02:48:13.000'),
    (1004, 2002, '2026-06-21 02:51:02.000'),
    -- 2003: Redis 缓存键 — 1 收藏
    (1001, 2003, '2026-06-21 14:21:36.000'),
    -- 2006: 社区路线图 — 1 收藏
    (1002, 2006, '2026-06-22 06:45:18.000'),
    -- 2007: Go 并发 — 1 收藏
    (1004, 2007, '2026-06-22 10:11:04.000'),
    -- 2008: Docker 多阶段 — 1 收藏
    (1003, 2008, '2026-06-22 15:54:56.000');

-- ----------------------------------------------------------------------------
-- 关注关系。用户的 follower/following 计数与以下数据一一对应。
-- ----------------------------------------------------------------------------
INSERT INTO follows (follower_id, followee_id, created_at) VALUES
    (1001, 1002, '2026-06-19 05:25:25.000'),
    (1001, 1003, '2026-06-19 05:25:39.000'),
    (1002, 1003, '2026-06-19 06:35:53.000'),
    (1002, 1004, '2026-06-19 06:36:07.000'),
    (1003, 1002, '2026-06-19 07:46:21.000'),
    (1003, 1004, '2026-06-19 07:46:35.000'),
    (1004, 1001, '2026-06-19 08:56:49.000'),
    (1004, 1002, '2026-06-19 08:57:03.000'),
    (1004, 1005, '2026-06-19 08:57:17.000'),
    (1005, 1002, '2026-06-19 10:07:17.000'),
    (1005, 1003, '2026-06-19 10:07:31.000'),
    (1005, 1004, '2026-06-19 10:07:45.000');

-- ----------------------------------------------------------------------------
-- 通知。JSON 内容引用了已有的用户、文章和评论。
-- ----------------------------------------------------------------------------
INSERT INTO notifications (id, user_id, type, content, is_read, created_at) VALUES
    (
        5001, 1002, 'comment',
        JSON_OBJECT('from_user_id', 1003, 'article_id', 2001, 'comment_id', 3001,
                    'message', 'bob 评论了你的文章《构建可组合的 Gin 中间件》'),
        FALSE, '2026-06-20 16:25:18.000'
    ),
    (
        5002, 1003, 'reply',
        JSON_OBJECT('from_user_id', 1002, 'article_id', 2001, 'comment_id', 3002, 'parent_id', 3001,
                    'message', 'alice 回复了你在《Gin 中间件》下的评论'),
        TRUE, '2026-06-20 16:25:52.000'
    ),
    (
        5003, 1002, 'like',
        JSON_OBJECT('from_user_id', 1004, 'article_id', 2001,
                    'message', 'carol 赞了你的文章《构建可组合的 Gin 中间件》'),
        FALSE, '2026-06-20 17:35:46.000'
    ),
    (
        5004, 1002, 'follow',
        JSON_OBJECT('from_user_id', 1005,
                    'message', 'dave 关注了你'),
        FALSE, '2026-06-19 10:07:17.000'
    ),
    (
        5005, 1003, 'comment',
        JSON_OBJECT('from_user_id', 1002, 'article_id', 2002, 'comment_id', 3004,
                    'message', 'alice 评论了你的文章《真正能帮到博客 Feed 的 MySQL 索引》'),
        TRUE, '2026-06-21 00:21:53.000'
    ),
    (
        5006, 1001, 'system',
        JSON_OBJECT('message', '测试数据已加载成功'),
        TRUE, '2026-06-22 02:16:07.000'
    ),
    -- NEW: 新文章的互动通知
    (
        5007, 1003, 'comment',
        JSON_OBJECT('from_user_id', 1002, 'article_id', 2007, 'comment_id', 3008,
                    'message', 'alice 评论了你的文章《Go 并发模式实战》'),
        FALSE, '2026-06-22 07:55:46.000'
    ),
    (
        5008, 1002, 'reply',
        JSON_OBJECT('from_user_id', 1003, 'article_id', 2007, 'comment_id', 3009, 'parent_id', 3008,
                    'message', 'bob 回复了你在《Go 并发模式实战》下的评论'),
        TRUE, '2026-06-22 07:56:57.000'
    ),
    (
        5009, 1004, 'comment',
        JSON_OBJECT('from_user_id', 1005, 'article_id', 2008, 'comment_id', 3010,
                    'message', 'dave 评论了你的文章《Go 应用的 Docker 多阶段构建》'),
        FALSE, '2026-06-22 13:48:06.000'
    ),
    (
        5010, 1002, 'comment',
        JSON_OBJECT('from_user_id', 1003, 'article_id', 2009, 'comment_id', 3011,
                    'message', 'bob 评论了你的文章《从 log.Println 到 slog》'),
        FALSE, '2026-06-22 19:09:25.000'
    );

-- ----------------------------------------------------------------------------
-- 管理员审计日志。target_id 指向现有用户或文章，detail 保留操作时的上下文。
-- ----------------------------------------------------------------------------
INSERT INTO audit_logs (
    id, admin_id, action, target_type, target_id, detail, ip, created_at
) VALUES
    (
        6001, 1001, 'ban_user', 'user', 1006,
        JSON_OBJECT('status', 'banned', 'username', 'eve'),
        '127.0.0.1', '2026-06-22 12:47:30.000'
    ),
    (
        6002, 1001, 'approve_article', 'article', 2009,
        JSON_OBJECT('title', '从 log.Println 到 slog：Go 结构化日志入门'),
        '127.0.0.1', '2026-06-22 17:58:57.000'
    ),
    (
        6003, 1001, 'change_role', 'user', 1005,
        JSON_OBJECT('role', 'user'),
        '127.0.0.1', '2026-06-22 18:27:08.000'
    );

COMMIT;

-- 将自增值保持在高位，避免与种子数据的固定 ID 冲突。
ALTER TABLE users AUTO_INCREMENT = 1100;
ALTER TABLE categories AUTO_INCREMENT = 1100;
ALTER TABLE tags AUTO_INCREMENT = 1100;
ALTER TABLE articles AUTO_INCREMENT = 2200;
ALTER TABLE comments AUTO_INCREMENT = 4000;
ALTER TABLE notifications AUTO_INCREMENT = 6000;
ALTER TABLE audit_logs AUTO_INCREMENT = 7000;
