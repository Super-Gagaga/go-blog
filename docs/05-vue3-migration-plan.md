# Vue 3 迁移方案（go-blog 前端）

> 版本：v1.1 ｜ 修订日期：2026-10-14 ｜ 适用仓库：go-blog ｜ 关联文档：docs/03-frontend-backend-todo.md
>
> 状态：方案修订完成，工程实施尚未开始。本文中的依赖、配置开关、测试及发布流程均为待实现项，不代表已经验证通过。

---

## 0. 重要前提更正（请先读）

经对仓库实际代码的核查，**当前前端项目中不存在任何 Vue 代码**，因此这不是一次「Vue 2 → Vue 3」的版本升级，而是**从原生 JavaScript 多页应用到 Vue 3 的框架引入/重写**。

核查证据：

- 检查 `web/` 业务 HTML/JS 与项目配置，**无任何业务 Vue 代码**；第三方图标中的 `.ph-file-vue` 类名不代表使用了 Vue。文档中的 Vue 讨论不作为代码证据。
- **无 `package.json`、无 `node_modules`、无任何构建工具配置**（vite/webpack/rollup/tsconfig/eslint 均不存在）。
- 前端为 **9 个静态 HTML 页面 + 9 个原生 JS 文件 + 9 个手写 CSS**，由 Go (Gin) 直接按文件路径返回 `web/` 下的文件。
- 仓库无 CI 配置（无 `.github/`）。

因此原需求中「当前 Vue 版本、Vuex/Pinia 现状」等问题的正确答案是：**均不存在，相关对比项属于"引入选型"而非"升级映射"**。后文凡涉及无法从代码确认的信息，一律标注为【假设】。

---

## 1. 现状梳理

### 1.1 已确认事实（来自代码）

| 维度 | 现状 |
|---|---|
| 框架 | 无。原生 ES6+ JavaScript，多页应用（MPA） |
| 构建工具 | 无。源码即产物，浏览器直接加载 |
| 包管理 | 无 npm/yarn/pnpm；第三方库以文件形式内置 |
| 路由 | 服务端路由。Gin 注册 clean URL：`/`、`/article-detail`、`/editor`、`/following`、`/notifications`、`/favorites`、`/my-profile`、`/admin-console`、`/admin-login`（见 `internal/router/router.go:56-66`），**前端无客户端路由** |
| 状态管理 | 无集中式 store。页面各自维护模块变量，首页使用 `state` 对象 |
| UI 组件库 | 无。手写 CSS + 设计令牌（CSS 变量），共享样式 `web/css/page-common.css` + 每页独立 CSS |
| 图标 | 本地内置 Phosphor Icons（`web/lib/phosphor-icons/`，含 index.js 与 woff2 字体） |
| 字体 | Google Fonts CDN（Instrument Serif / Geist Mono / Noto Sans SC） |
| TypeScript | 无。纯 JS，无类型检查 |
| 测试 | 前端无任何测试。后端有 Go 测试（`make test`） |
| CI/CD | 无 CI。发布靠 `Dockerfile`（`COPY web /app/web`）与 `deploy/docker-compose.yml` |
| 后端 | Go 1.26 + Gin；静态文件**从文件系统读取**（非 embed），`webFilePath()` 见 `internal/router/router.go:279-290` |
| API 约定 | `/api/v1/*`，响应封装 `{code, message, data?}`；同时检查 HTTP 状态与 `code===0`。`data` 使用 `omitempty`，无返回值的成功响应可能不含该字段 |
| 鉴权 | JWT 存 `localStorage`（键 `jwt_token` / `refresh_token`），请求头 `Authorization: Bearer <token>` |

### 1.2 前端资产清单

页面与脚本（1:1 对应，除 `admin-login.html` 使用内联脚本）：

| 页面 | 脚本 | JS 行数/大小 | 说明 |
|---|---|---|---|
| index.html | js/index.js | 561 行 / 21.3KB | 首页 Feed、分类、话题、推荐作者、排行榜、登录/注册弹窗 |
| admin-login.html | 内联 `<script>` | — | 管理员登录 + 角色校验 |
| admin.html | js/admin.js | 146 行 / 11.7KB | 管理后台（已是视图切换式迷你 SPA：pending/dashboard/articles/comments/users/taxonomies/audit） |
| editor.html | js/editor.js | 281 行 / 11.5KB | 文章编辑器，含本地草稿防抖、离开提醒、封面上传及封面预览；无 Ctrl+S 和 Markdown 正文预览 |
| my-profile.html | js/my-profile.js | — / 7.6KB | 个人主页、我的文章、我的收藏、资料编辑和头像上传 |
| article-detail.html | js/article-detail.js | — / 4.8KB | 文章详情 + 树形评论 |
| favorites.html | js/favorites.js | — / 5.0KB | 收藏列表 |
| following.html | js/following.js | — / 2.3KB | 关注/粉丝 |
| notifications.html | js/notifications.js | — / 3.0KB | 通知列表、已读 |
| （共享） | js/theme.js | 51 行 | 主题切换，暴露 `window.JournalTheme` |

CSS：`page-common.css`(8.0KB) + 8 个页面级 CSS，合计约 42KB。

### 1.3 主要技术债 / 迁移动因

- **重复代码多**：`API_BASE`、`token()`、`esc()`、`api()/getJSON()`、`show()`、401 处理在各页面脚本中复制粘贴（`web/js/admin.js:1-14`、`web/js/editor.js:1-60`、`web/js/index.js:48-68` 等）。
- **命令式 DOM 拼接**：大量 `innerHTML` 字符串模板 + 手工 `escapeHTML` 防 XSS（`web/js/index.js:130-143, 248-280`、`web/js/admin.js:16-18`）。
- **事件绑定分散**：全局 `document.addEventListener('click')` 委托 + 各元素单独监听，逻辑难追踪（`web/js/index.js:431-558`）。
- **视图状态靠手写 class 切换**：`setActiveNav`、`active` class 手动维护，易与数据不同步（`web/js/index.js:410-420`）。
- **无类型**：API 字段拼写错误只能在运行时暴露。
- **无构建**：无法按需加载、无法 tree-shaking、无法用 SFC。

### 1.4 必须确认的假设（缺失信息）

以下信息**无法从仓库确认**，方案中相关部分据此假设，落地前需与团队确认：

- 【假设 A】前端由 1~2 名工程师维护，可接受引入 Node/npm 工具链。
- 【假设 B】部署环境允许在构建阶段运行 `npm ci && npm run build`（Docker 多阶段构建 + 本机开发）。
- 【假设 C】浏览器基线为现代常青浏览器（ES2020+、支持原生 ESM），无需兼容 IE。
- 【假设 D】现有 CSS 视觉规范需保留，不希望通过引入 UI 组件库推翻设计。
- 【假设 E】API 契约（`/api/v1/*` 与 `{code,message,data}`）在迁移期内保持稳定。
- 【假设 F】无 SEO/首屏 SSR 硬性要求（当前是纯客户端渲染的静态页）。

> Node 只需存在于构建环境，生产运行镜像无需安装 Node。本方案优先使用独立 Node Docker 构建阶段；若 CI 宿主没有 Node，仍可通过 Docker 构建。P1 必须验证实际构建环境可以安装锁定依赖。

### 1.5 本轮范围与已有问题

本轮终态为 **Vue 3 多页应用（MPA）+ 独立管理后台 hash SPA**。保留现有 clean URL、查询参数及视觉规范。全站 SPA、SSR、Ctrl+S 和 Markdown 正文预览属于后续独立需求，不计入本轮迁移验收或工期。

已确认的行为差异必须先登记，不能以“新旧完全一致”为由复制缺陷：

- 编辑器已有 token 刷新重试，首页、个人页和后台的失效处理不同，通知页等仅显示请求错误；共享 client 的新约定见 3.7。
- `internal/pkg/markdown/renderer.go` 开启 `html.WithUnsafe()`，文章 HTML 未经过统一清洗；详情页和管理预览直接插入该内容。3.8 是迁移前必须完成的配套修复。
- 通知接口通常返回 `data.notifications` 和 `data.unread_count`，未读计数失败时直接返回分页对象；旧通知页却直接读 `data.list`。首个真实页面必须覆盖两种返回结构并修正展示，不能沿用空列表作为正确基线。
- 旧编辑器仅在新文章且输入为空时恢复本地草稿；保存新文章时先更新 ID 再清草稿，会遗漏 `editor_draft_new`。迁移时补齐清理旧键、取消待执行防抖、编辑已有文章的草稿恢复选择，并测试；不静默覆盖服务端正文。

---

## 2. 迁移路径对比

### 2.1 路径 A：整体重写为 Vue 3 SPA

| 项 | 内容 |
|---|---|
| 做法 | 新建 Vite + Vue 3 + TS 工程，`vue-router` 接管全部路由，全部页面改为 SFC，Gin 只保留 API 与一个 SPA fallback |
| 适用 | 页面间跳转频繁、共享状态多、愿意一次性承担风险、有充足测试覆盖 |
| 成本 | 高。9 页全部重写，同时改动 Go 路由模型与 Docker 产物 |
| 风险 | 高。登录态、编辑器草稿、管理后台权限守卫等流程需全部回归；一次性上线，回滚粒度粗 |
| 收益 | 客户端导航与共享内存状态统一，适合已明确需要全站 SPA 的场景 |

### 2.2 路径 B：渐进式迁移（构建基建先行 + 按页替换）

| 项 | 内容 |
|---|---|
| 做法 | 保留 Go MPA 与 clean URL；引入 Vite 多页构建，**逐页**把某个 HTML 的脚本替换为 Vue 3 应用挂载；未迁移页面继续用原生 JS；共享层（API client / auth / 主题）先抽成库供两端复用 |
| 适用 | 页面相对独立、线上要持续可用、希望小步验证 |
| 成本 | 中。基建 + 共享层一次性投入后，每页独立、可并行 |
| 风险 | 低-中。可按页发布与回滚（恢复 Go 路由指向旧 HTML）；代价是过渡期双栈共存 |
| 收益 | 可按页交付；本轮保留 MPA，管理后台最后迁为独立 hash SPA |

### 2.3 路径 C：仅引入构建工具（不上框架）

仅用 Vite 打包现有原生 JS，不做框架化。**不推荐**：解决了工程化，但未解决 DOM 命令式、状态分散的根本问题，与本需求目标不符。

### 2.4 对比与推荐

| 维度 | A 整体重写 | B 渐进迁移 | C 仅构建 |
|---|---|---|---|
| 首次可用时间 | 长 | 短（首屏页可先上） | 短 |
| 总工作量 | 高 | 中 | 低 |
| 上线风险 | 高 | 低 | 低 |
| 回滚粒度 | 整站 | **单页** | 单页 |
| 过渡期双栈 | 无 | 有 | 有 |
| 终态 | 全站 SPA | Vue MPA + 管理后台 hash SPA | 原生 JS MPA + 构建工具 |

**推荐：路径 B（构建基建先行 + 按页渐进迁移）。**

理由：
1. 本项目前端仅约 78KB JS、9 个页面，页面各自维护状态，适合逐页替换。仍需处理登录态、主题及公共资源的新旧兼容。
2. 当前**最大的缺口不是 Vue，而是工程化**（无构建、无类型、无测试）。一次性 SPA 重写会把这些风险与框架迁移风险叠加，而 B 可以先补基建、再逐页迁移，风险分层。
3. Go 端 clean URL 已就绪，B 保留其路由模型，新增静态产物挂载与按页入口选择。回滚使用同一镜像修改配置并重建容器，详见 4.3。
4. 本轮完成后保持 Vue MPA；全站 SPA 会改变导航、状态生命周期和部署 fallback，另行评估，不承诺自动收敛为路径 A。

**推荐顺序**：构建与最小共享层 → notifications 真实试点（含标记已读）→ 共享层收敛 → following / favorites → article-detail / my-profile → 首页 → 编辑器 → 管理员登录与管理后台。只读列表页和包含写操作的通知页分别验收。

---

## 3. 关键技术改造点

### 3.1 构建工具与依赖

- **版本基线**：以 Node 22（不低于 22.12）、Vite 8.3、Vue 3、Pinia 3、TypeScript 5、Vue Router 4 作为 P1 的候选组合。Vite 5 已不在官方支持范围；候选组合须经安装、类型检查、测试及构建验证后才能称为已锁定，不能把本方案当作兼容性验证结果。
- **可复现安装**：P1 在 `package.json` 固定直接依赖精确版本，提交 `package-lock.json`，本机和 CI 统一 `npm ci`；记录精确 Node/npm 版本，Docker Node 镜像锁定相同版本及 digest。`@vitejs/plugin-vue`、`vue-tsc`、Vitest、Vue Test Utils、Playwright、ESLint 及 Vue/TS 插件、Prettier 按 peer dependencies 一并验证，不混装未限定版本的最新插件。普通 MPA 入口不导入 Vue Router。
- **多页构建**：`frontend/` 为 Vite root，HTML 入口放在该目录根层，以 Vite 8 的 `build.rolldownOptions.input` 声明多入口，输出到 `../web/dist/`，`base: '/'`。显式设置 `emptyOutDir: true`，清理范围只能是 `web/dist/`，不能指向 `web/`。HTML 输出名取决于源文件路径，不靠 input 对象的别名重命名。
- **路由映射**：`/` → `dist/index.html`；`/admin-console` → `dist/admin.html`；其他既有页面 URL → 同名 `dist/<page>.html`。按白名单配置启用，启动时校验启用入口及引用资源存在，缺失即启动失败并报告路径。新增 `r.Static("/assets", webFilePath("dist/assets"))`；迁移期间保留 `/css`、`/js`、`/lib` 和 `/uploads`。
- **开发访问**：以 Gin 同源入口作为跨新旧页面验收地址。P1 先用 `vite build --watch` 更新 dist；若另用 Vite dev server，必须配置 API、上传资源及 legacy 页面代理，避免 5173 与 8080 的 localStorage 隔离。HMR 开发方式不替代 Gin 下的验收。
- **Docker**：新增独立 Node 构建阶段，在该阶段执行 `npm ci && npm run build`。运行阶段先复制 legacy `web/`，再 `COPY --from=frontend-builder /src/web/dist /app/web/dist`（以构建阶段实际工作目录为准）；不得依赖宿主机预先生成的 dist。通过 `.dockerignore` 排除宿主 node_modules/dist，确保空工作目录也能构建。生产运行镜像只含 Go 二进制及静态资源。
- **浏览器基线**：P0 列出实际支持的浏览器最低版本，P1 据此设置 `build.target` 并验证；“ES2020+”不能替代浏览器兼容矩阵。

版本依据（2026-10-14 核对）：[Vite 支持范围](https://vite.dev/releases)、[Node 要求](https://vite.dev/guide/)、[多页构建](https://vite.dev/guide/build.html)、[Pinia 3 要求](https://pinia.vuejs.org/cookbook/migration-v2-v3.html)。

### 3.2 Composition API vs Options API

- **统一采用 Composition API + `<script setup lang="ts">`**（无历史包袱，可直接选终态）。
- 约束：Vue 响应式逻辑放入 `composables/useXxx.ts`；HTTP、存储和格式化使用不依赖 Vue 的普通模块。组件不复制请求逻辑，legacy 复用时不必初始化 Pinia。
- **不引入 Options API**（本项目无存量 Vue 代码，无需兼容层）。

### 3.3 路由

- **MPA 阶段**：仍由 Gin 提供页面路由，Vue 应用仅负责「页内」交互；页面间跳转沿用 `location.href`（后续可替换为集中封装的 `navigate()`）。
- **管理后台**：采用 `createWebHashHistory('/admin-console')`，子路由为 `/admin-console#/pending`、`#/dashboard` 等，登录入口仍为 `/admin-login`。默认进入 pending；测试复制深链接、刷新、前进后退及未知 hash 的页内 404。hash 不发往服务端，Gin 无需全站 fallback。
- **后续 history 方案**：仅在独立 SPA 决策中考虑；若后台改为 history，统一使用 `/admin-console/` 基路径并为其子路径提供入口响应。不能使用与现有入口无关的 `/admin`。API、缺失静态资源和未知页面须保持正确的 404，不一律返回 HTML。依据：[Vue Router history 模式](https://router.vuejs.org/guide/essentials/history-mode)。

### 3.4 状态管理

- **Pinia** 只管理同一页面应用内的共享状态；MPA 跳转会重建实例，不能依靠 store 内存跨页持久化：
  - `useAuthStore`：token、refresh token、profile、登录/登出/刷新、`isAdmin`。替代散落在各文件的 `localStorage` 读写。
  - `useThemeStore`：接管 `web/js/theme.js` 的逻辑，保留 `localStorage` 键 `journal_theme` 以兼容未迁移页面。
  - `useToastStore`：统一 `show()/alert()` 反馈。
- **页面局部状态**用 `ref/reactive` 留在组件内，不滥用全局 store。
- 每个页面启动时从存储恢复 token/主题，profile 重新请求。监听 `storage` 事件同步其他标签页的登出和主题；同页更新显式通知 store。主题在应用挂载前初始化，避免亮暗闪烁。登出后取消或忽略在途刷新结果，避免旧请求重新写回 token。

### 3.5 UI 组件库

| 方案 | 说明 | 建议 |
|---|---|---|
| 不引入组件库 | 沿用现有手写 CSS + 设计令牌，用 Vue 封装现有 DOM 结构为组件 | **推荐**（符合【假设 D】，保留既有视觉，成本最低） |
| Headless 组件 | 仅提供无样式的可访问性行为（弹窗、下拉、Tabs） | 按实际需求引入，P1 不预装 |
| 全量 UI 库（Element Plus / Naive UI） | 开箱即用但样式体系与现有设计冲突 | 不推荐，除非计划整体改版 |

图标沿用已内置的 **Phosphor Icons**，不新增图标依赖。

### 3.6 原生写法迁移映射

**必须澄清：这些是 Vue 2 的概念，本仓库不含任何 Vue 2 代码，因此不存在需要处理的 filters、`$listeners`、EventBus。** 对应的「遗留写法 → Vue 3 写法」映射如下：

| 现有原生写法 | 迁移后 Vue 3 写法 |
|---|---|
| `innerHTML = 字符串模板` | 文本改为模板插值，确认该路径不再拼接 HTML 后移除重复转义；legacy 转义保留 |
| `escapeHTML()` 手工防 XSS | 文本插值默认转义；正文 HTML 走 3.8 的清洗流程，`v-html` 不负责清洗 |
| `document.querySelector` + `addEventListener` | `ref` + `@click` / `@submit` 等模板监听 |
| `document.addEventListener('click')` 全局委托 | 组件级事件；跨组件用 store 或 `provide/inject` |
| 手工 `classList.toggle('active')` | `:class="{ active: cond }"` 绑定 |
| `localStorage` 直接读写 token | `useAuthStore` |
| `location.href = ...` | 封装 `navigate()`；SPA 阶段改 `router.push` |
| `window.alert / confirm` | Toast / 统一确认弹窗组件 |
| 命令式 `show()` 状态行 | 响应式 `statusMessage` ref |

### 3.7 共享层（迁移的杠杆点）

抽出不依赖 Vue 的 `api/client.ts`、`auth/session.ts`、`theme/storage.ts` 和格式化函数。Vue 直接导入源码；需要复用的 legacy 页面以单独 library 构建生成的稳定入口 `/assets/legacy-shared.js` 作为 ESM 导入，并将调用脚本改为 `type="module"`。明确模块执行时序及产物依赖，不同时引入 UMD 和 ESM 两套桥接。仅改造需要共享能力的 legacy 页；回滚版本也须通过共用模块回归。

构建顺序固定为 MPA 构建后再生成桥接；第二次构建设置 `emptyOutDir: false`，不得清掉前一步 HTML/assets。watch 模式也要验证桥接仍存在。桥接产物随同镜像发布，HTML 与桥接使用需重新验证的缓存策略。

- `api/client.ts`：`API_BASE='/api/v1'`，检查 HTTP 与业务 code，允许成功响应无 data；注入 Bearer，支持 JSON 与 FormData，后者不手动设置 Content-Type。保留 HTTP 状态、业务 code、message；非 JSON 错误响应提供可读错误，不能当作成功。
- `composables/usePagination.ts`：统一 `page/pageSize/total_pages/append` 逻辑（当前在 index.js 手写）。
- `utils/format.ts`：`formatDate`、`readMinutes`、`avatarFor`、`coverFor`（源自 `web/js/index.js:70-96`）。
- `components/`：`AppHeader`、`LoginModal`、`UserMenu`、`ArticleCard`、`StatusLine`、`ThemeToggle`——这些在多个页面重复出现。

**鉴权约定**：

1. 登录、注册、刷新接口不进入自动刷新逻辑。受保护请求遇到 401 时，先检查请求所用 token 是否已被同页并发请求刷新；否则复用同一在途 refresh Promise，成功后更新两个存储键，每个原请求最多重试一次。403 不刷新，网络失败和 5xx 不自动重放写请求，尤其点赞/收藏/关注 Toggle。
2. 缺少 refresh token、刷新返回明确的凭据失效或重试后仍为 401，清理登录态并抛出统一的登录失效错误。刷新网络失败、429 或 5xx 显示暂时失败并保留凭据，不无限重试，也不无条件跳转。页面决定提示/跳转位置，client 不直接修改 location。
3. 首页和公开详情页保持可读，写操作提示登录；个人页、关注、收藏和通知页失效后提示重新登录；后台返回 `/admin-login`。编辑器先立即保存草稿并取消待执行防抖，保留当前输入、提示登录，只有保存成功且用户主动离开时才跳转；localStorage 写入失败时明确提示，不能假称已保存。
4. 新旧页面共用 token 键，接入新 client 的 legacy 页采用上述约定；未接入页面保留现状并记录差异。测试“旧首页登录 → 新页面 → 旧页面”和跨标签页登出。页面级刷新合并不宣称实现跨标签页互斥。
5. 管理员标记只用于 UI；服务端 `Auth` 与 `Admin` 中间件继续作为权限边界。测试普通用户直接请求管理 API 得到拒绝，不能仅测试路由守卫。

### 3.8 文章 HTML 清洗（配套修复）

`content_html` 来自用户 Markdown，不能因其由后端返回就视为可信。模板转义仅保护文本插值，`v-html` 仍会插入原始 HTML，见 [Vue 安全说明](https://vuejs.org/guide/best-practices/security.html)。

- 在 Go 输出层建立统一 HTML 白名单清洗，覆盖详情、包含正文的列表、管理员待审列表/预览，以及缓存命中的返回结果，使历史记录和新记录均受保护。选择维护中的解析器型清洗库，不用正则清洗。仅修改 Markdown 新写入链路不能保护历史数据。
- 保留正文所需的标题、段落、列表、引用、代码块、表格与受限图片/链接；移除脚本、事件属性、iframe、危险 URL 协议及任意内联样式。代码高亮和 GFM 样式允许范围以实际输出样例验证，不宽泛放行所有属性。
- Vue 使用唯一 `ArticleBody` 组件消费经过该流程的 HTML；其他字段使用文本插值。配套修复先于详情页和后台迁移上线，新旧页面回滚均保留它。
- 测试包含事件属性、危险链接、畸形嵌套 HTML 的恶意输入，以及正常表格/代码/图片。覆盖“历史存量 + 缓存命中 + 管理员预览”，并在浏览器确认无脚本执行。
- 本轮不批量重写数据库中的正文或 HTML；输出清洗不改变字段结构，但会改变不安全 HTML 的展示，登记为有意修复并回归正常内容。

---

## 4. 兼容与稳定

### 4.1 新旧代码共存

- **目录共存**：保留 `web/*.html` + `web/js` + `web/css` 为 legacy；新增 `frontend/`（源码）与 `web/dist/`（产物）。迁移完成的页面在 Gin 中指向 dist 入口。
- **样式共存**：legacy 的 `page-common.css` 继续被未迁移页引用；Vue 端通过 `<style>` 复用同一套 CSS 变量，避免视觉漂移。
- **主题/登录态共存**：`journal_theme`、`jwt_token`、`refresh_token` 三个 localStorage 键在过渡期保持键名不变，保证新旧页面互相切换时状态不丢。
- **API 共存**：HTTP method、路径、字段以附录 B 和 Go handler 为准；401 行为统一遵循 3.7，不能假定 legacy 实现原本一致。HTML 清洗、通知响应适配和草稿修复属于有意差异，单独验收。

### 4.2 按页发布

- **按页发布**：一次只切一个页面，该页所有访问者看到同一版本；这是按页面分阶段发布，不包含按用户或流量比例的灰度系统。
- **入口开关**：新增 `web.migrated_pages` 配置及显式环境变量绑定 `GO_BLOG_WEB_MIGRATED_PAGES`，例如 `notifications,following`，默认空列表；合法名称为 index、article-detail、editor、following、notifications、favorites、my-profile、admin、admin-login。启动时解析和校验，未知名称报错；admin 对应 `/admin-console`。Compose 显式透传此变量。
- **可观测**：先在测试环境记录 legacy 的加载错误、接口失败和关键交互结果；新入口错误记录带页面名和发布版本，不能包含 token 或正文。若接入线上错误监控，P1 明确采集位置和查看方式，没有可读数据时不能宣称“错误率无异常”。
- **观察期**：每页建议 1~3 天；无足够流量时用完整回归补足，不能只以等待天数判定通过。新出现的阻断性错误、登录回归或草稿丢失应触发回滚。

### 4.3 回滚

- **单页回滚**：从 Compose 实际使用的环境配置中移除页面名，使用同一镜像执行 `docker compose -f deploy/docker-compose.yml up -d --no-deps --force-recreate --no-build app`（从仓库根目录执行）。配置只在启动时读取，因此需要重建 app 容器，单实例会短暂中断；无需重新构建镜像，但不承诺即时或零停机切换。
- **验证**：保留当前镜像、变更前后开关值和回滚记录；检查 `/health`、目标 clean URL、静态资源、登录态和关键写操作。确认浏览器收到 legacy 入口，而非仅看容器健康。普通进程部署需要用更新后的环境重启进程。
- **产物回退**：镜像使用不可变 tag/digest。HTML 与其引用资源随同镜像发布；测试打开旧页后部署新版本、资源加载失败后刷新恢复。P1 试点不引入懒加载 chunk，若后续启用则先实现旧 hash 资源保留或明确的失败恢复。HTML 不设置长期缓存；稳定名称的 legacy 共享桥接也须重新验证缓存。
- **配套修复保留**：P2 安全基线建立后，整体回退镜像不能早于 HTML 清洗等配套修复的可回退基线。legacy 若已接入共享模块，回切 HTML 不会撤销共享模块变更；共享模块故障需回退到已验证的安全镜像。
- **数据边界**：本轮不做数据库结构迁移或历史正文批量改写，API method/path 与字段结构保持稳定；输出清洗会改变不安全 HTML 的显示。localStorage 草稿格式保持旧字段兼容，扩展字段不能破坏回切后的读取。

---

## 5. 质量保障

| 层级 | 手段 |
|---|---|
| 类型检查 | `vue-tsc --noEmit`，纳入 CI 门禁；API 响应定义 TS 接口，集中在 `api/types.ts` |
| 代码规范 | ESLint（`eslint-plugin-vue` + `@typescript-eslint`）+ Prettier，CI 校验 |
| 单元测试 | Vitest + `@vue/test-utils`：重点覆盖刷新合并、重试上限、登出竞态、错误分类、FormData、主题恢复及分页边界；覆盖率仅作辅助，不替代行为断言 |
| 组件/视觉 | 登录弹窗、用户菜单做交互与焦点测试；固定数据下检查关键页面截图，动态时间/头像等单独处理。Storybook 不作为本轮前置依赖 |
| E2E | Playwright 覆盖 9 个页面的**冒烟路径**：首页加载与筛选、登录/登出、文章详情+评论、编辑保存、收藏/关注/通知、管理后台各视图与权限守卫 |
| API 契约 | MSW/fixture 用于可控错误和组件测试；关键 E2E 直接连接真实 Go + MySQL + Redis 测试环境，校验 method/path/body、HTTP/code 和真实 data 结构。fixture 取自验证过的响应，不能只靠手写 mock 宣称防字段漂移 |
| 迁移验证手段 | 新旧页面在隔离测试数据下对照；写操作每轮重置数据，避免 Toggle 重复执行改变基线。每页连接真实 API 跑冒烟，再观察发布结果；已登记的缺陷以修复目标验收 |
| 部署回归 | 从干净上下文构建 Docker；验证页开关、缺失入口启动失败、静态资源、后台 hash 深链接刷新、单页回滚；扩展现有 Gin 路由测试 |
| HTML 安全 | 3.8 清洗单测及浏览器回归，包含存量、缓存命中、后台预览和正常 Markdown |
| 可访问性 | 保留现有 ARIA（如登录弹窗 `aria-modal`、`aria-expanded`），迁移后用 axe 检查 |

真实 API 测试只使用测试账号和隔离数据库，预置普通用户、管理员、已发布文章、待审文章及通知；不得把 mock 响应注入该套测试。至少验证登录/续期、通知读写、资料/头像、文章保存/上传、评论与 Toggle、后台权限和审核。现有 Go 测试并不自动覆盖这些业务链路。

性能先记录同设备、同数据、同网络下的 legacy 基线，至少重复 3 次取中位数；P1 试点测量 LCP、交互耗时、首次 JS gzip 大小后确定预算并记录，P3 前固定验收阈值。不能用未定义的“Lighthouse 不低于 legacy”替代预算，也不承诺引入框架后包体积不增。

---

## 6. 实施计划

**人力假设**：1~2 名前端工程师（【假设 A】）。工期以「人日」估算，1 人日≈6 有效工时。

| 阶段 | 目标 | 主要任务 | 人日 | 里程碑 | 验收标准 |
|---|---|---|---|---|---|
| P0 范围与基线 | 明确本轮边界 | 记录假设 A~F 的结论；Vue MPA + 后台 hash SPA；浏览器矩阵；现有行为与缺陷清单 | 1~2 | 范围冻结 | 现状、新增需求和配套修复分清；首次试点验收项明确 |
| P1 构建与真实试点 | 跑通交付链路 | 锁定工具链；MPA/Gin/Docker/CI；最小 client、session、主题；迁移 notifications，适配两种响应并实现标记已读 | 6~9 | 首个真实页面可发布和回滚 | 干净安装与容器构建通过；旧首页登录 → 新通知页读写/续期 → 旧页面；同镜像回滚演练通过；记录性能基线 |
| P2 共享层与配套修复 | 固定复用约定 | 完善鉴权/主题/分页与基础组件；legacy ESM 桥接；Go HTML 输出清洗（含历史与缓存路径） | 6~9 | 共享行为和安全基线可验证 | 3.7/3.8 测试通过；旧页接入后无未登记回归；安全基线镜像可回退 |
| P3 其余页面迁移 | 逐页替换 | following→favorites→article-detail→my-profile→index→editor→admin-login→admin；编辑器草稿修复；后台 hash 路由 | 23~29 | 每页独立交付 | 每页真实 API E2E + 行为对照 + 回切检查；观察 1~3 天，缺陷有明确处理结果 |
| P4 质量加固 | 收敛性能与交互 | 补齐跨页 E2E；性能预算检查；a11y；发布期间资源恢复 | 4~6 | 验收记录完整 | 核心链路全绿；性能满足 P1/P2 固定预算；无未解决的阻断性或高优先级缺陷 |
| P5 收尾 | 结束过渡期 | 完成约定观察期后删除 legacy 与桥接；仅清理无引用样式；统一文档 | 2~3 | Vue MPA 迁移完成 | 每个 clean URL 对应构建产物；保留仍在使用的 CSS/图标；可回退镜像归档；不再承诺当前镜像内单页回切 |

**初估：约 42~58 人日**，仅用于容量规划；包含首次通知页、鉴权及 HTML 配套修复，不含全站 SPA、新编辑功能和新的监控平台建设。P1 后按实测重估剩余页面。观察期、基础设施等待和串行验收另计日历时间，不能将人日直接除以人数承诺完成日期。

**关键依赖**：P1 同时验证构建链路与最小共享能力，避免等全部组件完成才验证真实页面；P2 完成后才进入 P3。HTML 输出清洗必须先于详情页和后台切换。以下清单尚未执行，修订方案不等于完成对应里程碑。

**风险与缓解**：

| 风险 | 影响 | 缓解 |
|---|---|---|
| 构建环境无法安装锁定依赖 | 阻断 | P1 先验证独立 Node Docker 构建阶段及依赖网络 |
| 双栈共存导致样式/状态漂移 | 中 | 共用 CSS 变量与 localStorage 键名；对拍验证 |
| 管理后台权限回归 | 高 | 前端守卫与 Go 管理 API 拒绝普通用户都需测试 |
| 编辑器草稿/续期回归 | 高 | 保留键名与字段；取消待执行防抖；保存前后 ID 切换、恢复选择和 token 失效专项测试 |
| HTML 清洗遗漏历史或缓存路径 | 高 | 检查全部正文输出；新旧详情与后台预览共用测试 |
| 首页数据字段多、易漏 | 中 | 以 `/api/v1/home` 聚合契约为准，做字段清单核对 |

---

## 7. 分阶段执行清单

### P0 立项与决策

- [ ] 逐条确认 1.4 中的假设 A~F，记录结论与责任人
- [ ] 评审三条路径，正式采纳路径 B
- [ ] 确定不引入全量 UI 组件库（沿用现有 CSS 令牌），如需要则选用 Headless 方案
- [ ] 确定浏览器基线（【假设 C】）与是否需要 SEO/SSR（【假设 F】）
- [ ] 记录本轮终态与浏览器矩阵；把 Ctrl+S/正文预览列为后续需求
- [ ] 登记通知结构、草稿恢复与 HTML 的已有问题及修复目标
- [ ] 明确按页开关、重建容器回滚和测试数据准备方式

### P1 构建与真实试点

- [ ] 初始化 `frontend/`：Vite + Vue 3 + TypeScript
- [ ] 固定 Node/npm 和直接依赖版本，提交 lockfile，验证 peer dependencies
- [ ] 配置多页 HTML 入口（`rolldownOptions.input`），仅清理并输出 `web/dist/`
- [ ] Go：`/assets` + 按页开关、配置校验、入口/资源存在性检查
- [ ] Docker 多阶段构建接入 `npm ci && npm run build`
- [ ] CI：安装 Node、缓存依赖，执行 lint / typecheck / test / build
- [ ] 接入 ESLint + Prettier + `vue-tsc`
- [ ] 完成最小 client/session/theme 和 notifications 真实读写试点，覆盖两种返回结构
- [ ] 验证同源新旧页跳转、登录续期、Docker 构建、开关与同镜像回滚
- [ ] 记录实际性能及依赖兼容结果，P1 后重新估算工期

### P2 共享层与配套修复

- [ ] 完善 `api/client.ts`：可选 data、JSON/FormData、刷新合并、最多重试一次、失败分类
- [ ] 实现 `useAuthStore`（login/register/logout/refresh/isAdmin，沿用 `jwt_token`/`refresh_token` 键名）
- [ ] 实现 `useThemeStore`（沿用 `journal_theme` 键名，替代 `web/js/theme.js`）
- [ ] 实现 `usePagination`、`utils/format.ts`
- [ ] 实现基础组件：StatusLine、ThemeToggle、LoginModal、UserMenu、ArticleCard
- [ ] 为 store / client / utils 写 Vitest 单测
- [ ] 实现 legacy ESM 桥接并回归执行时序、主题与登录状态
- [ ] 实现 Go HTML 输出清洗，覆盖历史数据、缓存和管理预览；建立安全回退镜像
- [ ] 固定性能预算，记录 API 实测样例；鉴权边界、回滚和安全用例通过后进入 P3

### P3 页面迁移（每页独立完成下列步骤）

- [ ] 复核 P1 notifications 试点在 P2 共享层收敛后仍通过真实 API E2E
- [ ] following → 迁移 + E2E + 按页发布
- [ ] favorites → 迁移 + E2E + 按页发布
- [ ] article-detail → 迁移 + E2E + 按页发布
- [ ] my-profile → 迁移 + E2E + 按页发布
- [ ] index（首页）→ 迁移 + E2E + 按页发布
- [ ] editor → 迁移 + E2E + 按页发布（草稿防抖/恢复/清理、上传、续期与保存；不含新增快捷键和正文预览）
- [ ] admin-login → 迁移 + E2E + 按页发布
- [ ] admin → `/admin-console#/…` hash SPA；深链接刷新/历史导航/权限/审核验证
- [ ] 每页：真实 API 验证、重置测试数据后对照、有意修复登记、回切检查与观察记录

### P4 质量加固

- [ ] Playwright 覆盖 9 页冒烟路径
- [ ] 同条件性能复测，满足 P1/P2 固定的预算并记录包体积增量
- [ ] a11y（axe）检查并修复
- [ ] 真实 Go API 用例通过；mock 用例覆盖边界，不能替代真实契约验证
- [ ] 验证发布时已打开页面的静态资源失败恢复

### P5 收尾

- [ ] 观察期结束后清理 legacy HTML/JS 和桥接，CSS/图标按引用检查保留
- [ ] 确认 Gin 路由与构建产物一一对应
- [ ] 更新 README 与 docs（含本文档）
- [ ] 归档可回退镜像，记录单页开关退出后的回退方式
- [ ] 全站 SPA 仅记录为后续候选，不计入本轮完成条件

---

## 附录 A：迁移期目录结构建议

```
go-blog/
├── frontend/                 # 新增：Vue 3 源码
│   ├── src/
│   │   ├── api/              # client.ts, types.ts
│   │   ├── auth/             # session.ts，不依赖 Vue
│   │   ├── theme/            # storage.ts，不依赖 Vue
│   │   ├── stores/           # auth.ts, theme.ts
│   │   ├── composables/      # usePagination.ts
│   │   ├── components/       # 共享组件
│   │   ├── pages/            # 每页一个入口组件
│   │   └── utils/            # format.ts
│   ├── *.html               # MPA 入口，与 Gin 映射一致
│   ├── package.json
│   ├── package-lock.json
│   ├── vite.config.ts
│   └── tsconfig.json
├── web/                      # legacy 静态资源（过渡期保留）
│   ├── *.html                # 未迁移页面
│   ├── js/ css/ lib/         # 过渡期保留
│   └── dist/                 # 新增：Vite 构建产物
└── internal/router/router.go # 按页选择 legacy / dist 入口
```

## 附录 B：API 端点清单（迁移核对用）

以下路径均以 `/api/v1` 为前缀，依据 `internal/router/router.go`、`internal/handler/`、`internal/service/` 及 `internal/model/` 核对。只列业务页面相关接口，不为迁移新增未有 UI 的功能；通用管理缓存/查询端点和 `/ping` 不纳入页面迁移范围。

返回列表示解包后的 `data`；“无”表示成功但 data 可缺省。分页类型 `Page<T>` 为 `{list: T[], pagination: {page, page_size, total, total_pages}}`，迁移时兼容实际空数组/null 表现。`TokenPair` 为 `{access_token, refresh_token, token_type, expires_in}`。对象完整字段由对应 Go model/服务返回结构和真实接口样例共同校验，TS 不靠本表简写猜字段。

### B.1 公开读取

| Method | Path | 查询参数 | 返回 data |
|---|---|---|---|
| GET | `/articles` | page、page_size、status、category_id、tag_id、keyword、user_id | `Page<Article>` |
| GET | `/articles/ranking` | period、limit | Article 列表 |
| GET | `/articles/:id` | 无 | Article，含 content/content_html |
| GET | `/articles/:id/comments` | 无 | CommentTreeNode 列表，非分页对象 |
| GET | `/categories` | 无 | Category 列表 |
| GET | `/tags` | keyword | Tag 列表 |
| GET | `/topics/trending` | limit | TrendingTag 列表 |
| GET | `/recommendations/authors` | limit | RecommendedAuthor 列表 |
| GET | `/home` | page、page_size | articles、categories、trending_topics、recommended_authors、weekly_picks；articles 为分页对象，非关键聚合项可能为 null |
| GET | `/user/:id` | 无 | 公开 UserProfile |

### B.2 身份与个人数据

`/auth/*` 不要求 Bearer；下列 `/user/*` 和 `/users/*` 写操作及个人数据要求登录。

| Method | Path | 请求体或查询参数 | 返回 data |
|---|---|---|---|
| POST | `/auth/register` | JSON `{username,email,password}` | User；注册不直接返回 token，后续登录单独调用 |
| POST | `/auth/login` | JSON `{email,password}` | TokenPair |
| POST | `/auth/refresh` | JSON `{refresh_token}` | TokenPair |
| GET | `/user/profile` | 无 | 私有 UserProfile，含 role/status |
| PUT | `/user/profile` | JSON `{username?,bio?}` | 更新后的私有 UserProfile |
| POST | `/user/avatar` | multipart 字段 `avatar` | `{avatar: string}` |
| GET | `/user/articles` | page、page_size、status、category_id、tag_id、keyword | `Page<Article>` |
| GET | `/user/favorites` | page、page_size | 分页收藏结果；核对条目的 article 嵌套，不假定与文章列表相同 |
| GET | `/user/following` | page、page_size | 分页关注用户结果 |
| GET | `/user/followers` | page、page_size | 分页粉丝结果 |
| GET | `/user/feed` | page、page_size | 分页 Feed 结果 |
| GET | `/user/notifications` | page、page_size | 通常 `{notifications: Page<NotificationItem>, unread_count: number}`；未读计数失败时为 `Page<NotificationItem>` |
| PATCH | `/user/notifications/read` | JSON `{ids: number[]}`；空数组表示全部已读 | 无 |
| POST | `/users/:id/follow` | 无 | `{following: boolean}`；切换操作 |

通知页面使用端点专属适配器将 `data.notifications ?? data` 转为分页结果；缺失 unread_count 时标为未知，不直接推断为 0。两种结构都需用 Go handler 契约测试覆盖，正常读写链路连接真实服务验证。

### B.3 文章与互动（要求登录）

| Method | Path | 请求体 | 返回 data |
|---|---|---|---|
| POST | `/articles` | CreateArticleReq：title、content；可选 category_id、tag_ids、cover_image、status | Article |
| PUT | `/articles/:id` | UpdateArticleReq：title、content、category_id、tag_ids、cover_image、status；省略/null 语义按 Go 请求结构与实测验证 | Article |
| DELETE | `/articles/:id` | 无 | 无 |
| PATCH | `/articles/:id/status` | JSON `{status}` | 无 |
| POST | `/articles/upload` | multipart 字段 `image` | `{url: string}` |
| POST | `/articles/:id/upload` | multipart 字段 `image` | `{url: string}`；现有同 handler 路由别名，不推断额外绑定行为 |
| POST | `/articles/:id/comments` | JSON `{content}` | Comment |
| POST | `/comments/:id/reply` | JSON `{content}` | Comment |
| DELETE | `/comments/:id` | 无 | 无 |
| POST | `/articles/:id/like` | 无 | `{liked: boolean, count: number}` |
| POST | `/articles/:id/favorite` | 无 | `{favorited: boolean}` |
| POST | `/categories` | JSON `{name}` | Category；记录当前路由，不据此新增页面操作 |
| POST | `/tags` | JSON `{name}` | Tag；记录当前路由，不据此新增页面操作 |

文章保存允许的请求 status 为 draft/published/archived；普通用户提交 published 会经过服务层审核规则，页面显示和测试以实际返回状态为准，不自行把提交审核视为已发布。分类清空和标签清空的行为须验证后记录，不能只凭 TS 可选字段推断。

### B.4 管理接口（要求登录且通过 Admin 中间件）

| Method | Path | 请求体或查询参数 | 返回 data |
|---|---|---|---|
| GET | `/admin/dashboard` | 无 | 仪表盘统计对象，按 AdminService 返回结构核对 |
| GET | `/admin/users` | page、page_size、keyword | 分页用户结果 |
| PATCH | `/admin/users/:id/status` | JSON `{status: active/banned}` | 无 |
| PATCH | `/admin/users/:id/role` | JSON `{role: user/admin}` | 无 |
| GET | `/admin/articles/pending` | page、page_size、keyword | 分页待审文章结果 |
| POST | `/admin/articles/:id/approve` | 无 | 无 |
| POST | `/admin/articles/:id/reject` | JSON `{reason}` | 无 |
| GET | `/admin/articles` | page、page_size、keyword、status、user_id、category_id、date_from、date_to | 分页文章结果 |
| DELETE | `/admin/articles/:id` | 无 | 无 |
| GET | `/admin/comments` | page、page_size、keyword、user_id、article_id、date_from、date_to | 分页评论结果 |
| DELETE | `/admin/comments/:id` | 无 | 无 |
| GET | `/admin/audit-logs` | page、page_size、admin_id、action、target_type、date_from、date_to | 分页审计结果 |
| POST | `/admin/categories` | JSON `{name}` | Category |
| PUT | `/admin/categories/:id` | JSON `{name}` | Category |
| DELETE | `/admin/categories/:id` | 无 | 无 |
| POST | `/admin/tags` | JSON `{name}` | Tag |
| PUT | `/admin/tags/:id` | JSON `{name}` | Tag |
| DELETE | `/admin/tags/:id` | 无 | 无 |

### B.5 契约验收记录

每个迁移页面记录实际使用的接口、测试账号角色、请求样例、响应样例和对应 E2E 用例。测试同时断言 HTTP 状态与业务 code，检查无 data 成功、空列表、401、403、校验失败和必要的网络失败恢复。保留脱敏样例，禁止保存凭据。前端 fixture 与 Go 测试各自通过，不等于跨端契约已验证。
