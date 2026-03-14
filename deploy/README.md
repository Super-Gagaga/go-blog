# Docker 部署

## 1. 准备环境变量

在项目根目录执行：

```powershell
Copy-Item deploy/.env.example deploy/.env
```

部署到服务器前，请修改 `deploy/.env` 中的 MySQL 和 JWT 密钥。

## 2. 构建并启动

```powershell
docker compose --env-file deploy/.env -f deploy/docker-compose.yml up -d --build
```

启动后访问：

- 网站：<http://localhost:8080>
- 健康检查：<http://localhost:8080/health>

查看服务状态和日志：

```powershell
docker compose --env-file deploy/.env -f deploy/docker-compose.yml ps
docker compose --env-file deploy/.env -f deploy/docker-compose.yml logs -f app
```

## 3. 停止或更新

停止服务但保留数据库和上传文件：

```powershell
docker compose --env-file deploy/.env -f deploy/docker-compose.yml down
```

拉取代码后重新构建：

```powershell
docker compose --env-file deploy/.env -f deploy/docker-compose.yml up -d --build
```

删除全部容器和持久化数据：

```powershell
docker compose --env-file deploy/.env -f deploy/docker-compose.yml down -v
```

> `sql/init.sql` 只会在 MySQL 数据卷首次创建时执行。修改初始化 SQL 后，如需重新初始化数据库，必须先备份数据，再执行 `down -v`。
