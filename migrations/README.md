# Migrations

Database migrations are managed by `golang-migrate`.

Create new migrations with:

```bash
migrate create -ext sql -dir migrations -seq create_users_table
```
