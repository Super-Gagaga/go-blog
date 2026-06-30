# go-blog

基于 Gin + GORM + Redis 的个人博客与社区系统:文章发布与审核、评论/点赞/收藏、关注与通知、
首页推荐流,以及一套泛型缓存管理器驱动的通用查询/写入层。前端为零依赖的原生 HTML/JS 单页集。

## 技术栈

- Go 1.26 + Gin + GORM(MySQL 8) + Redis 7
- 认证:JWT 双令牌(access/refresh),首用户自动成为管理员(命名锁保证并发安全)
- 缓存:泛型 `cachemanager`(单条/列表回源、异步回填、版本键),详见 `docs/cachemanager.md`
- 通用层:泛型 query/write 路由,详见 `docs/generic-handlers.md`
- 部署:多阶段 Dockerfile + docker-compose(mysql/redis/app 健康检查齐备)

## 快速开始

```bash
cp config.example.yaml config.yaml   # 按需修改数据库/Redis/JWT 密钥
make migrate-up                      # 应用 migrations/ 下的迁移
make dev                             # 本地运行,默认监听 :8080
```

或直接用 Docker 起整套栈(首次启动会用 `sql/init.sql` 初始化表结构):

```bash
make docker-up
```

前端页面由服务端直接托管,访问 `http://127.0.0.1:8080/` 即可;管理后台在 `/admin-login`。

## 测试

```bash
make test        # go test -race ./...
```

缓存管理器与通用路由的测试基于 miniredis + sqlmock,无需外部依赖。

## 目录结构

```
cmd/server          程序入口(优雅启停)
internal/config     viper 配置加载(环境变量覆盖)
internal/router     路由装配
internal/middleware 认证/限流/追踪/日志/恢复/CORS
internal/handler    HTTP 处理器与通用路由
internal/service    业务逻辑
internal/repository 数据访问
internal/model      数据模型与统一响应
internal/pkg        jwt/slug/markdown/upload/errcode/cachemanager 等基础件
migrations          golang-migrate 迁移
sql                 init 结构与测试种子数据
web                 原生前端
docs                架构与模块文档
```

## 文档

- [系统架构](docs/01-system-architecture.md)
- [实现路线图](docs/02-implementation-roadmap.md)
- [缓存管理器](docs/cachemanager.md)
- [通用处理器](docs/generic-handlers.md)
- [部署说明](deploy/README.md)
