# 通用数据/缓存 HTTP 路由使用说明

> 本文由原 http_router 文档(get/set/cache_get/cache_set 四份 readme)整合而来,代码已并入 `internal/handler`(query_router.go / write_router.go / cache_lookup_router.go / cache_writedown_router.go)。博客中挂载于 /api/v1/admin/generic/* (Auth+Admin 保护);WriteRouterGroup 仅保留代码未挂载。

## 查询路由(QueryRouterGroup)

# Query 模块使用文档

## 一、路由字段说明表

| 路由字段 | 对应 Service 方法 | 入参（示例） | 请求方式 | 出参（示例） |
|---------|------------------|------------|---------|------------|
| `POST /api/v1/{resource}/query` | `GetQuery` | `{"method":"list","page":1,"filters":[{"field":"age","operator":"gt","value":18}]}` | POST | `{"code":0,"message":"success","data":[{...}],"total":100,"page":1,"page_size":20,"total_pages":5}` |
| `GET /api/v1/{resource}/:id` | `GetSingleByID` | URL参数: `123` | GET | `{"code":0,"message":"success","data":{"id":123,"name":"John",...}}` |
| `POST /api/v1/{resource}/count` | `CountQuery` | `{"filters":[{"field":"status","operator":"eq","value":"active"}]}` | POST | `{"code":0,"message":"success","count":42}` |

## 二、支持的过滤器操作符

| 操作符 | 说明 | 示例 |
|-------|------|------|
| `eq` | 等于 | `{"field":"status","operator":"eq","value":"active"}` |
| `ne` | 不等于 | `{"field":"status","operator":"ne","value":"inactive"}` |
| `gt` | 大于 | `{"field":"age","operator":"gt","value":18}` |
| `gte` | 大于等于 | `{"field":"age","operator":"gte","value":18}` |
| `lt` | 小于 | `{"field":"price","operator":"lt","value":100}` |
| `lte` | 小于等于 | `{"field":"price","operator":"lte","value":100}` |
| `like` | 模糊匹配 | `{"field":"name","operator":"like","value":"John"}` |
| `in` | IN查询 | `{"field":"status","operator":"in","value":["active","pending"]}` |
| `between` | 范围查询 | `{"field":"age","operator":"between","value":[18,30]}` |
| `isnull` | 字段为空 | `{"field":"deleted_at","operator":"isnull"}` |
| `isnotnull` | 字段不为空 | `{"field":"email","operator":"isnotnull"}` |

## 三、完整使用示例

### 3.1 基础配置代码

```go
package main

import (
    "github.com/gin-gonic/gin"
    "gorm.io/gorm"
    
    "github.com/yourname/go-blog/internal/huancun/http_router"
    "github.com/yourname/go-blog/internal/huancun/service"
)

// 数据模型
type User struct {
    ID        uint      `json:"id" gorm:"primaryKey;autoIncrement"`
    Name      string    `json:"name" gorm:"size:100"`
    Age       int       `json:"age"`
    Status    string    `json:"status" gorm:"size:20"`
    Email     string    `json:"email" gorm:"size:100"`
    CreatedAt time.Time `json:"created_at"`
    DeletedAt *time.Time `json:"deleted_at" gorm:"index"`
}

func main() {
    r := gin.Default()
    v1 := r.Group("/api/v1")
    
    // 1. 创建 ServiceManager
    userService := service.NewServiceManager(User{})
    
    // 2. 创建 Query 路由组
    userRouter := http_router.NewQueryRouterGroup(v1, userService)
    
    // 3. 注册查询方法
    // 方式1: 使用便捷方法注册常用查询(list, active_list, search)
    userRouter.RegisterCommonMethods(20) // 每页20条
    
    // 方式2: 单独注册方法
    // 基础列表查询
    userRouter.RegisterListMethod(20)
    
    // 活跃用户列表
    userRouter.RegisterActiveListMethod(20)
    
    // 自定义查询方法
    userRouter.RegisterMethod(
        "vip_list",              // 方法名
        10,                      // 每页条数
        func(db *gorm.DB) *gorm.DB {
            return db.Where("age >= ?", 18)
        },
        "age",                   // 排序字段
        "DESC",                  // 排序方式
    )
    
    // 4. 注册路由
    userRouter.RegisterRoutes("/users")
    
    r.Run(":8080")
}
```

### 3.2 API 调用示例

#### 示例 1: 基础分页查询 - 获取所有用户

**请求:**
```bash
POST /api/v1/users/query
Content-Type: application/json

{
  "method": "list",
  "page": 1,
  "filters": []
}
```

**响应:**
```json
{
  "code": 0,
  "message": "success",
  "data": [
    {
      "id": 1,
      "name": "John Doe",
      "age": 25,
      "status": "active",
      "email": "john@example.com",
      "created_at": "2024-01-15T10:30:00Z"
    },
    {
      "id": 2,
      "name": "Jane Smith",
      "age": 30,
      "status": "active",
      "email": "jane@example.com",
      "created_at": "2024-01-16T11:20:00Z"
    }
  ],
  "total": 100,
  "page": 1,
  "page_size": 20,
  "total_pages": 5
}
```

#### 示例 2: 单条件过滤 - 获取活跃用户

**请求:**
```bash
POST /api/v1/users/query
Content-Type: application/json

{
  "method": "list",
  "page": 1,
  "filters": [
    {
      "field": "status",
      "operator": "eq",
      "value": "active"
    }
  ]
}
```

**响应:**
```json
{
  "code": 0,
  "message": "success",
  "data": [
    {
      "id": 1,
      "name": "John Doe",
      "age": 25,
      "status": "active",
      "email": "john@example.com",
      "created_at": "2024-01-15T10:30:00Z"
    }
  ],
  "total": 42,
  "page": 1,
  "page_size": 20,
  "total_pages": 3
}
```

#### 示例 3: 组合过滤 - 年龄大于18且名字包含"John"的活跃用户

**请求:**
```bash
POST /api/v1/users/query
Content-Type: application/json

{
  "method": "list",
  "page": 1,
  "filters": [
    {
      "field": "age",
      "operator": "gt",
      "value": 18
    },
    {
      "field": "name",
      "operator": "like",
      "value": "John"
    },
    {
      "field": "status",
      "operator": "eq",
      "value": "active"
    }
  ]
}
```

**响应:**
```json
{
  "code": 0,
  "message": "success",
  "data": [
    {
      "id": 1,
      "name": "John Doe",
      "age": 25,
      "status": "active",
      "email": "john@example.com",
      "created_at": "2024-01-15T10:30:00Z"
    },
    {
      "id": 5,
      "name": "Johnny Test",
      "age": 35,
      "status": "active",
      "email": "johnny@example.com",
      "created_at": "2024-01-18T09:15:00Z"
    }
  ],
  "total": 2,
  "page": 1,
  "page_size": 20,
  "total_pages": 1
}
```

#### 示例 4: 单个查询 - 根据ID获取用户

**请求:**
```bash
GET /api/v1/users/123
```

**响应:**
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "id": 123,
    "name": "John Doe",
    "age": 25,
    "status": "active",
    "email": "john@example.com",
    "created_at": "2024-01-15T10:30:00Z"
  }
}
```

**错误响应 (用户不存在):**
```json
{
  "code": 404,
  "message": "record not found"
}
```

#### 示例 5: 计数查询 - 统计活跃用户数量

**请求:**
```bash
POST /api/v1/users/count
Content-Type: application/json

{
  "filters": [
    {
      "field": "status",
      "operator": "eq",
      "value": "active"
    }
  ]
}
```

**响应:**
```json
{
  "code": 0,
  "message": "success",
  "count": 42
}
```

#### 示例 6: 范围查询 - 年龄在18-30之间的用户

**请求:**
```bash
POST /api/v1/users/query
Content-Type: application/json

{
  "method": "list",
  "page": 1,
  "filters": [
    {
      "field": "age",
      "operator": "between",
      "value": [18, 30]
    }
  ]
}
```

#### 示例 7: IN 查询 - 多状态查询

**请求:**
```bash
POST /api/v1/users/query
Content-Type: application/json

{
  "method": "list",
  "page": 1,
  "filters": [
    {
      "field": "status",
      "operator": "in",
      "value": ["active", "pending", "verified"]
    }
  ]
}
```

#### 示例 8: 使用自定义方法 - VIP用户列表

**请求:**
```bash
POST /api/v1/users/query
Content-Type: application/json

{
  "method": "vip_list",
  "page": 1,
  "filters": []
}
```

**响应:**
```json
{
  "code": 0,
  "message": "success",
  "data": [
    {
      "id": 1,
      "name": "John Doe",
      "age": 25,
      "status": "active",
      "email": "john@example.com"
    }
  ],
  "total": 15,
  "page": 1,
  "page_size": 10,
  "total_pages": 2
}
```

#### 示例 9: NULL 值过滤 - 查询未删除的用户

**请求:**
```bash
POST /api/v1/users/query
Content-Type: application/json

{
  "method": "list",
  "page": 1,
  "filters": [
    {
      "field": "deleted_at",
      "operator": "isnull"
    }
  ]
}
```

#### 示例 10: 分页查询第2页

**请求:**
```bash
POST /api/v1/users/query
Content-Type: application/json

{
  "method": "list",
  "page": 2,
  "filters": [
    {
      "field": "status",
      "operator": "eq",
      "value": "active"
    }
  ]
}
```

**响应:**
```json
{
  "code": 0,
  "message": "success",
  "data": [...],
  "total": 42,
  "page": 2,
  "page_size": 20,
  "total_pages": 3
}
```

## 四、代码讲解

### 4.1 核心组件说明

#### PaginatedQueryMethod - 分页查询方法定义
```go
type PaginatedQueryMethod[T any] struct {
    Name       string                    // 方法名称(如"list","vip_list")
    PageSize   int                       // 每页条数
    Service    *service.ServiceManager[T] // 服务管理器
    FilterFunc func(*gorm.DB) *gorm.DB   // 预定义过滤函数
    OrderBy    string                    // 排序字段
    Order      string                    // 排序方式(ASC/DESC)
}
```

**执行流程:**
1. 应用预定义的 FilterFunc (如果有)
2. 应用前端传入的 GORM 过滤器
3. 调用 ServiceManager 的 GetQuery 方法进行分页查询
4. 返回包含数据、总数、分页信息的结果

#### QueryRouterGroup - 路由组管理器
```go
type QueryRouterGroup[T any] struct {
    RouterGroup        *gin.RouterGroup              // Gin路由组
    Service            *service.ServiceManager[T]    // 服务管理器
    MethodRegistry     *QueryMethodRegistry[T]       // 方法注册表
    TranslatorRegistry *filter_translator.GormTranslatorRegistry // GORM过滤器翻译器
}
```

**主要功能:**
- 注册查询方法
- 注册 HTTP 路由
- 翻译前端过滤器为 GORM 过滤器
- 处理请求并返回响应

### 4.2 使用方式详解

#### 方式 1: 使用便捷方法 (推荐)

```go
// 一次性注册 list, active_list, search 三个方法
userRouter.RegisterCommonMethods(20) // 每页20条
```

#### 方式 2: 单独注册预定义方法

```go
// 基础列表 (所有数据)
userRouter.RegisterListMethod(20)

// 活跃列表 (status='active' 且未删除)
userRouter.RegisterActiveListMethod(20)

// 搜索方法 (预留接口)
userRouter.RegisterSearchMethod(20, []string{"name", "email"})
```

#### 方式 3: 完全自定义方法

```go
// 注册 VIP 用户查询
userRouter.RegisterMethod(
    "vip_list",              // 方法名
    10,                      // 每页10条
    func(db *gorm.DB) *gorm.DB {
        // 自定义过滤逻辑
        return db.Where("age >= ?", 18).
                  Where("vip_level > ?", 0)
    },
    "vip_level",             // 按VIP等级排序
    "DESC",                  // 降序
)

// 最近注册用户
userRouter.RegisterMethod(
    "recent_users",
    50,
    func(db *gorm.DB) *gorm.DB {
        return db.Where("created_at > ?", time.Now().AddDate(0, 0, -7))
    },
    "created_at",
    "DESC",
)

// 待审核用户
userRouter.RegisterMethod(
    "pending_users",
    15,
    func(db *gorm.DB) *gorm.DB {
        return db.Where("status = ?", "pending")
    },
    "created_at",
    "ASC",
)
```

### 4.3 过滤器工作原理

过滤器采用**两阶段过滤**:

1. **预定义过滤阶段** (可选)
   - 在方法注册时通过 `FilterFunc` 定义
   - 在 Execute 方法中首先应用
   - 适用于固定的业务逻辑过滤

2. **动态过滤阶段**
   - 前端传入的 filters 参数
   - 通过 `GormTranslatorRegistry` 翻译为 `GormFilter`
   - 支持 11 种标准操作符

**过滤执行顺序:**
```
GORM查询 → 预定义FilterFunc → 动态过滤器1 → 动态过滤器2 → ... → 分页 → 排序 → 结果
```

### 4.4 GORM 过滤器翻译流程

```go
// 1. 前端发送 FilterParam
{
  "field": "age",
  "operator": "gt",
  "value": 18
}

// 2. GormTranslatorRegistry 翻译为 GormFilter
filter := &GormGreaterThanFilter{
    GenericFilter: &GenericFilter{
        Field: "age",
        Operator: "gt",
        Value: 18,
    },
}

// 3. 应用到 GORM 查询
db = filter.ApplyGorm(db)
// 等价于: db.Where("age > ?", 18)
```

## 五、最佳实践

### 5.1 合理设置分页大小

```go
// 大数据量列表
userRouter.RegisterListMethod(50)

// 详细信息列表
orderRouter.RegisterListMethod(20)

// 精简选择列表
categoryRouter.RegisterListMethod(100)
```

### 5.2 组合使用多种查询方法

```go
// 全量查询
userRouter.RegisterListMethod(20)

// 业务场景1: 活跃用户
userRouter.RegisterActiveListMethod(20)

// 业务场景2: VIP用户
userRouter.RegisterMethod("vip_users", 10, vipFilterFunc, "vip_level", "DESC")

// 业务场景3: 新用户
userRouter.RegisterMethod("new_users", 30, newUserFilterFunc, "created_at", "DESC")
```

### 5.3 合理使用排序

```go
// 按创建时间倒序 (最新的在前)
userRouter.RegisterMethod("list", 20, nil, "created_at", "DESC")

// 按年龄升序
userRouter.RegisterMethod("age_list", 20, nil, "age", "ASC")

// 按价格降序 (热门商品)
productRouter.RegisterMethod("hot_products", 20, nil, "price", "DESC")
```

### 5.4 前端调用封装

```javascript
// 封装 API 调用函数
class UserAPI {
  static async query(method, page, filters) {
    const response = await fetch('/api/v1/users/query', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ method, page, filters })
    });
    return response.json();
  }
  
  static async getByID(id) {
    const response = await fetch(`/api/v1/users/${id}`);
    return response.json();
  }
  
  static async count(filters) {
    const response = await fetch('/api/v1/users/count', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ filters })
    });
    return response.json();
  }
}

// 使用示例
const result = await UserAPI.query('list', 1, [
  { field: 'age', operator: 'gte', value: 18 },
  { field: 'status', operator: 'eq', value: 'active' }
]);

const user = await UserAPI.getByID(123);

const count = await UserAPI.count([
  { field: 'status', operator: 'eq', value: 'active' }
]);
```

### 5.5 完整的多资源应用示例

```go
func SetupRoutes() *gin.Engine {
    r := gin.Default()
    v1 := r.Group("/api/v1")
    
    // 用户模块
    setupUserRoutes(v1)
    
    // 订单模块
    setupOrderRoutes(v1)
    
    // 产品模块
    setupProductRoutes(v1)
    
    return r
}

func setupUserRoutes(rg *gin.RouterGroup) {
    userService := service.NewServiceManager(User{})
    userRouter := http_router.NewQueryRouterGroup(rg, userService)
    
    // 注册常用方法
    userRouter.RegisterCommonMethods(20)
    
    // 注册自定义方法
    userRouter.RegisterMethod("vip_users", 10, func(db *gorm.DB) *gorm.DB {
        return db.Where("vip_level > ?", 0)
    }, "vip_level", "DESC")
    
    userRouter.RegisterRoutes("/users")
}

func setupOrderRoutes(rg *gin.RouterGroup) {
    orderService := service.NewServiceManager(Order{})
    orderRouter := http_router.NewQueryRouterGroup(rg, orderService)
    
    orderRouter.RegisterListMethod(30)
    
    // 待支付订单
    orderRouter.RegisterMethod("pending_payment", 20, func(db *gorm.DB) *gorm.DB {
        return db.Where("status = ?", "pending_payment")
    }, "created_at", "DESC")
    
    // 已完成订单
    orderRouter.RegisterMethod("completed", 20, func(db *gorm.DB) *gorm.DB {
        return db.Where("status = ?", "completed")
    }, "completed_at", "DESC")
    
    orderRouter.RegisterRoutes("/orders")
}

func setupProductRoutes(rg *gin.RouterGroup) {
    productService := service.NewServiceManager(Product{})
    productRouter := http_router.NewQueryRouterGroup(rg, productService)
    
    productRouter.RegisterListMethod(20)
    
    // 热销产品
    productRouter.RegisterMethod("hot_products", 20, func(db *gorm.DB) *gorm.DB {
        return db.Where("sales_count > ?", 100)
    }, "sales_count", "DESC")
    
    productRouter.RegisterRoutes("/products")
}
```

## 六、错误处理

常见错误及处理:

| 错误码 | 说明 | 处理方式 |
|-------|------|---------|
| 400 | 请求参数错误 | 检查 method, page, filters 格式 |
| 404 | 记录不存在 | 确认 ID 是否正确 |
| 500 | 服务器错误 | 检查数据库连接和查询语句 |

## 七、Query 与 Lookup 的区别

| 特性 | Query 模块 | Lookup 模块 |
|-----|-----------|------------|
| **数据源** | 数据库 (GORM) | Redis 缓存 |
| **分页** | 支持 (page, page_size) | 不支持 (一次性返回) |
| **过滤器** | GormFilter (SQL查询) | RedisFilter (内存过滤) |
| **性能** | 取决于数据库索引 | 快速 (内存操作) |
| **数据量** | 适合大数据量 | 适合中小数据量 |
| **使用场景** | 持久化数据查询 | 缓存数据快速查询 |
| **回源** | 直接查数据库 | 可选回源到数据库 |

---

以上就是 Query 模块的完整使用文档。该模块与 Lookup 模块配合使用,可以构建高性能的数据查询系统。

## 写入路由(WriteRouterGroup)

# set(为了方便下面称为write) 模块使用文档

## 一、路由字段说明表

### 1.1 单个操作路由

| 路由字段 | 对应 Service 方法 | 入参（示例） | 请求方式 | 出参（示例） |
|---------|------------------|------------|---------|------------|
| `POST /api/v1/{resource}/set` | `SetSingle` | `{"data":{...},"on_conflict_update":true}` | POST | `{"code":0,"message":"success"}` |
| `POST /api/v1/{resource}/insert` | `Insert` | `{"data":{...}}` | POST | `{"code":0,"message":"success"}` |
| `PUT /api/v1/{resource}/update` | `Update` / `UpdateByID` | `{"id":1,"updates":{"name":"John"}}` | PUT | `{"code":0,"message":"success"}` |
| `DELETE /api/v1/{resource}/delete` | `Delete` / `DeleteByID` | `{"id":1,"soft":false}` | DELETE | `{"code":0,"message":"success"}` |
| `POST /api/v1/{resource}/upsert` | `Upsert` | `{"data":{...},"conflict_columns":["id"]}` | POST | `{"code":0,"message":"success"}` |
| `POST /api/v1/{resource}/increment` | `Increment` / `Decrement` | `{"id":1,"column":"count","value":1}` | POST | `{"code":0,"message":"success"}` |

### 1.2 批量操作路由

| 路由字段 | 对应 Service 方法 | 入参（示例） | 请求方式 | 出参（示例） |
|---------|------------------|------------|---------|------------|
| `POST /api/v1/{resource}/batch/set` | `SetQuery` | `{"data":[{...}],"batch_size":100}` | POST | `{"code":0,"message":"success","rows_affected":100}` |
| `POST /api/v1/{resource}/batch/insert` | `BatchInsert` | `{"data":[{...}],"batch_size":100}` | POST | `{"code":0,"message":"success","rows_affected":100}` |
| `PUT /api/v1/{resource}/batch/update` | `BatchUpdate` | `{"updates":{"status":"active"}}` | PUT | `{"code":0,"message":"success","rows_affected":50}` |
| `DELETE /api/v1/{resource}/batch/delete` | `BatchDelete` / `BatchSoftDelete` | `{"ids":[1,2,3],"soft":false}` | DELETE | `{"code":0,"message":"success","rows_affected":3}` |
| `POST /api/v1/{resource}/batch/upsert` | `BatchUpsert` | `{"data":[{...}],"conflict_columns":["id"]}` | POST | `{"code":0,"message":"success","rows_affected":100}` |
| `POST /api/v1/{resource}/batch/increment` | `BatchIncrement` / `BatchDecrement` | `{"ids":[1,2,3],"column":"count","value":1}` | POST | `{"code":0,"message":"success","rows_affected":3}` |

## 二、完整使用示例

### 2.1 基础配置代码

```go
package main

import (
    "github.com/gin-gonic/gin"
    
    "github.com/yourname/go-blog/internal/huancun/http_router"
    "github.com/yourname/go-blog/internal/huancun/service"
)

// 数据模型
type User struct {
    ID     uint   `json:"id" gorm:"primaryKey;autoIncrement"`
    Name   string `json:"name" gorm:"size:100"`
    Age    int    `json:"age"`
    Status string `json:"status" gorm:"size:20"`
    Email  string `json:"email" gorm:"size:100;uniqueIndex"`
    Count  int    `json:"count" gorm:"default:0"`
}

func main() {
    r := gin.Default()
    v1 := r.Group("/api/v1")
    
    // 1. 创建 ServiceManager
    userService := service.NewServiceManager(User{})
    
    // 2. 创建 Write 路由组
    userWriter := http_router.NewWriteRouterGroup(v1, userService)
    
    // 3. 注册路由
    userWriter.RegisterRoutes("/users")
    
    r.Run(":8080")
}
```

### 2.2 单个操作 API 调用示例

#### 示例 1: Set - 设置单个用户(新增或更新)

**请求:**
```bash
POST /api/v1/users/set
Content-Type: application/json

{
  "data": {
    "id": 1,
    "name": "John Doe",
    "age": 25,
    "status": "active",
    "email": "john@example.com"
  },
  "on_conflict_update": true,
  "invalidate_cache": true
}
```

**响应:**
```json
{
  "code": 0,
  "message": "success"
}
```

**说明:**
- `on_conflict_update`: 当主键或唯一索引冲突时是否更新(默认 true)
- `invalidate_cache`: 是否使缓存失效(默认 true)

#### 示例 2: Insert - 插入用户(不处理冲突)

**请求:**
```bash
POST /api/v1/users/insert
Content-Type: application/json

{
  "data": {
    "name": "Jane Smith",
    "age": 30,
    "status": "active",
    "email": "jane@example.com"
  }
}
```

**响应:**
```json
{
  "code": 0,
  "message": "success"
}
```

**错误响应(唯一索引冲突):**
```json
{
  "code": 500,
  "message": "insert failed: UNIQUE constraint failed"
}
```

#### 示例 3: Update - 更新用户信息

**按ID更新:**
```bash
PUT /api/v1/users/update
Content-Type: application/json

{
  "id": 1,
  "updates": {
    "name": "John Smith",
    "age": 26,
    "status": "inactive"
  }
}
```

**全局更新(不推荐,会更新所有记录):**
```bash
PUT /api/v1/users/update
Content-Type: application/json

{
  "updates": {
    "status": "pending"
  }
}
```

**响应:**
```json
{
  "code": 0,
  "message": "success"
}
```

#### 示例 4: Delete - 删除用户

**硬删除:**
```bash
DELETE /api/v1/users/delete
Content-Type: application/json

{
  "id": 1,
  "soft": false
}
```

**软删除(设置 deleted_at):**
```bash
DELETE /api/v1/users/delete
Content-Type: application/json

{
  "id": 1,
  "soft": true
}
```

**响应:**
```json
{
  "code": 0,
  "message": "success"
}
```

#### 示例 5: Upsert - 插入或更新

**请求:**
```bash
POST /api/v1/users/upsert
Content-Type: application/json

{
  "data": {
    "id": 1,
    "name": "John Doe",
    "age": 25,
    "status": "active",
    "email": "john@example.com"
  },
  "conflict_columns": ["id"],
  "update_columns": ["name", "age"]
}
```

**响应:**
```json
{
  "code": 0,
  "message": "success"
}
```

**说明:**
- `conflict_columns`: 冲突检测字段(如 ["id"] 或 ["email"])
- `update_columns`: 冲突时要更新的字段(为空则更新所有字段)

#### 示例 6: Increment - 字段增量操作

**增加计数:**
```bash
POST /api/v1/users/increment
Content-Type: application/json

{
  "id": 1,
  "column": "count",
  "value": 5,
  "is_decr": false
}
```

**减少计数:**
```bash
POST /api/v1/users/increment
Content-Type: application/json

{
  "id": 1,
  "column": "count",
  "value": 3,
  "is_decr": true
}
```

**响应:**
```json
{
  "code": 0,
  "message": "success"
}
```

### 2.3 批量操作 API 调用示例

#### 示例 7: 批量设置用户

**请求:**
```bash
POST /api/v1/users/batch/set
Content-Type: application/json

{
  "data": [
    {
      "id": 1,
      "name": "John Doe",
      "age": 25,
      "status": "active",
      "email": "john@example.com"
    },
    {
      "id": 2,
      "name": "Jane Smith",
      "age": 30,
      "status": "active",
      "email": "jane@example.com"
    },
    {
      "id": 3,
      "name": "Bob Johnson",
      "age": 35,
      "status": "inactive",
      "email": "bob@example.com"
    }
  ],
  "batch_size": 100,
  "on_conflict_update": true,
  "invalidate_cache": true
}
```

**响应:**
```json
{
  "code": 0,
  "message": "success",
  "rows_affected": 3
}
```

#### 示例 8: 批量插入用户

**请求:**
```bash
POST /api/v1/users/batch/insert
Content-Type: application/json

{
  "data": [
    {
      "name": "User 1",
      "age": 20,
      "status": "active",
      "email": "user1@example.com"
    },
    {
      "name": "User 2",
      "age": 22,
      "status": "active",
      "email": "user2@example.com"
    }
  ],
  "batch_size": 50
}
```

**响应:**
```json
{
  "code": 0,
  "message": "success",
  "rows_affected": 2
}
```

#### 示例 9: 批量更新用户

**按条件批量更新(预留功能):**
```bash
PUT /api/v1/users/batch/update
Content-Type: application/json

{
  "updates": {
    "status": "verified",
    "count": 0
  }
}
```

**响应:**
```json
{
  "code": 0,
  "message": "success",
  "rows_affected": 15
}
```

**说明:** 当前版本会更新所有记录,后续可以添加 filters 支持条件更新

#### 示例 10: 批量删除用户

**硬删除:**
```bash
DELETE /api/v1/users/batch/delete
Content-Type: application/json

{
  "ids": [1, 2, 3, 4, 5],
  "soft": false
}
```

**软删除:**
```bash
DELETE /api/v1/users/batch/delete
Content-Type: application/json

{
  "ids": [1, 2, 3],
  "soft": true
}
```

**响应:**
```json
{
  "code": 0,
  "message": "success",
  "rows_affected": 3
}
```

#### 示例 11: 批量 Upsert

**请求:**
```bash
POST /api/v1/users/batch/upsert
Content-Type: application/json

{
  "data": [
    {
      "email": "john@example.com",
      "name": "John Doe",
      "age": 25,
      "status": "active"
    },
    {
      "email": "jane@example.com",
      "name": "Jane Smith",
      "age": 30,
      "status": "active"
    }
  ],
  "conflict_columns": ["email"],
  "update_columns": ["name", "age", "status"],
  "batch_size": 100
}
```

**响应:**
```json
{
  "code": 0,
  "message": "success",
  "rows_affected": 2
}
```

#### 示例 12: 批量增量操作

**批量增加计数:**
```bash
POST /api/v1/users/batch/increment
Content-Type: application/json

{
  "ids": [1, 2, 3, 4, 5],
  "column": "count",
  "value": 10,
  "is_decr": false
}
```

**批量减少计数:**
```bash
POST /api/v1/users/batch/increment
Content-Type: application/json

{
  "ids": [1, 2, 3],
  "column": "count",
  "value": 5,
  "is_decr": true
}
```

**响应:**
```json
{
  "code": 0,
  "message": "success",
  "rows_affected": 5
}
```

## 三、代码讲解

### 3.1 核心组件说明

#### WriteRouterGroup - 写操作路由组
```go
type WriteRouterGroup[T any] struct {
    RouterGroup *gin.RouterGroup           // Gin路由组
    Service     *service.ServiceManager[T] // 服务管理器
}
```

**设计特点:**
- 无需过滤器翻译器(写操作不涉及复杂查询)
- 直接调用 ServiceManager 的写操作方法
- 统一的请求/响应格式
- 完整的错误处理

### 3.2 操作分类

#### 单个操作
- **Set**: 智能新增或更新(根据冲突策略)
- **Insert**: 纯插入(冲突时报错)
- **Update**: 更新指定字段
- **Delete**: 硬删除或软删除
- **Upsert**: 精确控制冲突策略
- **Increment/Decrement**: 原子性增减操作

#### 批量操作
- **Batch Set**: 批量智能写入
- **Batch Insert**: 批量纯插入
- **Batch Update**: 批量更新
- **Batch Delete**: 批量删除
- **Batch Upsert**: 批量 Upsert
- **Batch Increment**: 批量增减

### 3.3 请求选项说明

#### SetSingle/SetQuery 选项
```go
{
  "on_conflict_update": true,  // 冲突时是否更新(默认true)
  "invalidate_cache": true,    // 是否使缓存失效(默认true)
  "batch_size": 100            // 批次大小(仅批量操作,默认100)
}
```

#### Upsert 选项
```go
{
  "conflict_columns": ["id"],              // 冲突检测字段(必填)
  "update_columns": ["name", "age"],       // 冲突时更新的字段(空=全部)
  "batch_size": 100                        // 批量操作的批次大小
}
```

### 3.4 事务处理

所有写操作都自动包裹在事务中:
- 隔离级别: `REPEATABLE READ`
- 自动提交: 操作成功时自动提交
- 自动回滚: 操作失败时自动回滚
- 批量操作: 按 batch_size 分批,每批独立事务

## 四、最佳实践

### 4.1 选择合适的操作类型

```go
// 场景1: 不确定数据是否存在,使用 Set
POST /users/set
{
  "data": {...},
  "on_conflict_update": true
}

// 场景2: 确定是新数据,使用 Insert
POST /users/insert
{
  "data": {...}
}

// 场景3: 需要精确控制更新字段,使用 Upsert
POST /users/upsert
{
  "data": {...},
  "conflict_columns": ["email"],
  "update_columns": ["name", "age"]
}
```

### 4.2 批量操作性能优化

```go
// 大数据量导入: 调整 batch_size
{
  "data": [...10000条数据...],
  "batch_size": 500,  // 每批500条
  "on_conflict_update": false  // 纯插入更快
}

// 中等数据量: 使用默认值
{
  "data": [...1000条数据...],
  "batch_size": 100  // 默认100
}
```

### 4.3 软删除 vs 硬删除

```go
// 推荐: 软删除(可恢复)
{
  "id": 1,
  "soft": true
}

// 谨慎: 硬删除(不可恢复)
{
  "id": 1,
  "soft": false
}
```

### 4.4 原子性计数操作

```go
// ✅ 正确: 使用 Increment(原子性)
POST /users/increment
{
  "id": 1,
  "column": "count",
  "value": 1
}

// ❌ 错误: 先查询再更新(非原子性,可能丢失更新)
// 1. GET /users/1  → count = 10
// 2. PUT /users/update {"id":1, "updates":{"count":11}}
```

### 4.5 前端封装示例

```javascript
class UserWriteAPI {
  // 单个操作
  static async set(data, options = {}) {
    return fetch('/api/v1/users/set', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        data,
        on_conflict_update: options.onConflictUpdate ?? true,
        invalidate_cache: options.invalidateCache ?? true
      })
    }).then(r => r.json());
  }
  
  static async insert(data) {
    return fetch('/api/v1/users/insert', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ data })
    }).then(r => r.json());
  }
  
  static async update(id, updates) {
    return fetch('/api/v1/users/update', {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ id, updates })
    }).then(r => r.json());
  }
  
  static async delete(id, soft = true) {
    return fetch('/api/v1/users/delete', {
      method: 'DELETE',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ id, soft })
    }).then(r => r.json());
  }
  
  static async increment(id, column, value, isDecr = false) {
    return fetch('/api/v1/users/increment', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ id, column, value, is_decr: isDecr })
    }).then(r => r.json());
  }
  
  // 批量操作
  static async batchSet(data, options = {}) {
    return fetch('/api/v1/users/batch/set', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        data,
        batch_size: options.batchSize ?? 100,
        on_conflict_update: options.onConflictUpdate ?? true,
        invalidate_cache: options.invalidateCache ?? true
      })
    }).then(r => r.json());
  }
  
  static async batchDelete(ids, soft = true) {
    return fetch('/api/v1/users/batch/delete', {
      method: 'DELETE',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ ids, soft })
    }).then(r => r.json());
  }
}

// 使用示例
await UserWriteAPI.set({ id: 1, name: 'John', age: 25 });
await UserWriteAPI.batchSet([...], { batchSize: 200 });
await UserWriteAPI.increment(1, 'view_count', 1);
```

### 4.6 完整应用示例

```go
func SetupUserRoutes(rg *gin.RouterGroup) {
    userService := service.NewServiceManager(User{})
    
    // 查询路由
    userQuery := http_router.NewQueryRouterGroup(rg, userService)
    userQuery.RegisterCommonMethods(20)
    userQuery.RegisterRoutes("/users")
    
    // 写入路由
    userWrite := http_router.NewWriteRouterGroup(rg, userService)
    userWrite.RegisterRoutes("/users")
    
    // Lookup路由(可选)
    userLookup := http_router.NewLookupRouterGroup(rg, userService)
    userLookup.RegisterListMethod("cache:user:*", 1*time.Hour)
    userLookup.RegisterRoutes("/users")
}
```

## 五、错误处理

| 错误码 | 说明 | 常见原因 |
|-------|------|---------|
| 400 | 请求参数错误 | data为空、字段缺失、类型错误 |
| 500 | 服务器错误 | 数据库连接失败、约束冲突、事务失败 |

**常见错误示例:**

```json
// 唯一索引冲突
{
  "code": 500,
  "message": "insert failed: UNIQUE constraint failed: users.email"
}

// 外键约束失败
{
  "code": 500,
  "message": "delete failed: FOREIGN KEY constraint failed"
}

// 数据为空
{
  "code": 400,
  "message": "data cannot be nil"
}
```

## 六、与其他模块的配合

### 6.1 Query + Write 完整 CRUD

```go
// 查询
GET  /api/v1/users/:id          → 获取单个
POST /api/v1/users/query        → 分页查询
POST /api/v1/users/count        → 计数

// 写入
POST   /api/v1/users/set        → 新增或更新
PUT    /api/v1/users/update     → 更新
DELETE /api/v1/users/delete     → 删除
```

### 6.2 Lookup + Write 缓存同步

```go
// 1. 写入数据并自动失效缓存
POST /api/v1/users/set
{
  "data": {...},
  "invalidate_cache": true  // 自动清除缓存
}

// 2. 从缓存查询最新数据
POST /api/v1/users/lookup
{
  "method": "list",
  "filters": [...]
}
```

### 6.3 三模块协同工作流

```
用户请求
  ↓
Write 模块(写入数据 + 失效缓存)
  ↓
Query 模块(查询数据库)
  ↓
Lookup 模块(更新缓存)
  ↓
返回结果
```

## 缓存查询路由(LookupRouterGroup)

# Lookup 模块使用文档

## 一、路由字段说明表

| 路由字段 | 对应 Service 方法 | 入参（示例） | 请求方式 | 出参（示例） |
|---------|------------------|------------|---------|------------|
| `POST /api/v1/{resource}/lookup` | `LookupQuery` | `{"method":"list","filters":[{"field":"age","operator":"gt","value":18}]}` | POST | `{"code":0,"message":"success","data":{"cache:user:1":{...}},"keys":["cache:user:1"],"count":1}` |
| `GET /api/v1/{resource}/:key` | `LookupSingle` | URL参数: `cache:user:1` | GET | `{"code":0,"message":"success","data":{"id":1,"name":"John",...}}` |
| `POST /api/v1/{resource}/count` | `LookupQuery` (仅计数) | `{"method":"list","filters":[{"field":"status","operator":"eq","value":"active"}]}` | POST | `{"code":0,"message":"success","count":42}` |
| `POST /api/v1/{resource}/invalidate` | `InvalidateCache` / `InvalidateCacheByPattern` | `{"keys":["cache:user:1","cache:user:2"]}` 或 `{"pattern":"cache:user:*"}` | POST | `{"code":0,"message":"success","count":2}` |

## 二、支持的过滤器操作符

| 操作符 | 说明 | 示例 |
|-------|------|------|
| `eq` | 等于 | `{"field":"status","operator":"eq","value":"active"}` |
| `ne` | 不等于 | `{"field":"status","operator":"ne","value":"inactive"}` |
| `gt` | 大于 | `{"field":"age","operator":"gt","value":18}` |
| `gte` | 大于等于 | `{"field":"age","operator":"gte","value":18}` |
| `lt` | 小于 | `{"field":"price","operator":"lt","value":100}` |
| `lte` | 小于等于 | `{"field":"price","operator":"lte","value":100}` |
| `like` | 模糊匹配 | `{"field":"name","operator":"like","value":"John"}` |
| `in` | IN查询 | `{"field":"status","operator":"in","value":["active","pending"]}` |
| `between` | 范围查询 | `{"field":"age","operator":"between","value":[18,30]}` |
| `isnull` | 字段不存在 | `{"field":"deleted_at","operator":"isnull"}` |
| `isnotnull` | 字段存在 | `{"field":"email","operator":"isnotnull"}` |

## 三、完整使用示例

### 3.1 基础配置代码

```go
package main

import (
    "context"
    "time"
    
    "github.com/yourname/go-blog/internal/huancun/http_router"
    "github.com/yourname/go-blog/internal/huancun/service"
    
    "github.com/gin-gonic/gin"
)

// 数据模型
type User struct {
    ID     uint   `json:"id"`
    Name   string `json:"name"`
    Age    int    `json:"age"`
    Status string `json:"status"`
    Email  string `json:"email"`
}

func main() {
    r := gin.Default()
    v1 := r.Group("/api/v1")
    
    // 1. 创建 ServiceManager
    userService := service.NewServiceManager(User{})
    userService.CacheKeyType = "cache"
    userService.CacheKeyName = "user"
    
    // 2. 创建 Lookup 路由组
    userLookup := http_router.NewLookupRouterGroup(v1, userService)
    
    // 3. 注册查询方法
    // 基础列表查询
    userLookup.RegisterListMethod("cache:user:*", 1*time.Hour)
    
    // 带数据库回源的查询
    userLookup.RegisterFallbackMethod("cache:user:*", 30*time.Minute)
    
    // 自定义查询方法（活跃用户）
    userLookup.RegisterActiveListMethod(
        "cache:user:*",
        1*time.Hour,
        func(ctx context.Context, client *redis.Client, keys []string) ([]string, error) {
            var activeKeys []string
            for _, key := range keys {
                status, err := client.HGet(ctx, key, "status").Result()
                if err == nil && status == "active" {
                    activeKeys = append(activeKeys, key)
                }
            }
            return activeKeys, nil
        },
    )
    
    // 4. 注册路由
    userLookup.RegisterRoutes("/users")
    
    r.Run(":8080")
}
```

### 3.2 API 调用示例

#### 示例 1: 基础查询 - 获取所有活跃用户

**请求:**
```bash
POST /api/v1/users/lookup
Content-Type: application/json

{
  "method": "list",
  "filters": [
    {
      "field": "status",
      "operator": "eq",
      "value": "active"
    }
  ]
}
```

**响应:**
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "cache:user:1": {
      "id": 1,
      "name": "John Doe",
      "age": 25,
      "status": "active",
      "email": "john@example.com"
    },
    "cache:user:2": {
      "id": 2,
      "name": "Jane Smith",
      "age": 30,
      "status": "active",
      "email": "jane@example.com"
    }
  },
  "keys": ["cache:user:1", "cache:user:2"],
  "count": 2
}
```

#### 示例 2: 组合过滤 - 年龄大于18且名字包含"John"的活跃用户

**请求:**
```bash
POST /api/v1/users/lookup
Content-Type: application/json

{
  "method": "list",
  "filters": [
    {
      "field": "age",
      "operator": "gt",
      "value": 18
    },
    {
      "field": "name",
      "operator": "like",
      "value": "John"
    },
    {
      "field": "status",
      "operator": "eq",
      "value": "active"
    }
  ]
}
```

**响应:**
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "cache:user:1": {
      "id": 1,
      "name": "John Doe",
      "age": 25,
      "status": "active",
      "email": "john@example.com"
    },
    "cache:user:5": {
      "id": 5,
      "name": "Johnny Test",
      "age": 35,
      "status": "active",
      "email": "johnny@example.com"
    }
  },
  "keys": ["cache:user:1", "cache:user:5"],
  "count": 2
}
```

#### 示例 3: 单个查询 - 根据键获取用户

**请求:**
```bash
GET /api/v1/users/cache:user:1
```

**响应:**
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "id": 1,
    "name": "John Doe",
    "age": 25,
    "status": "active",
    "email": "john@example.com"
  }
}
```

#### 示例 4: 计数查询 - 统计活跃用户数量

**请求:**
```bash
POST /api/v1/users/count
Content-Type: application/json

{
  "method": "list",
  "filters": [
    {
      "field": "status",
      "operator": "eq",
      "value": "active"
    }
  ]
}
```

**响应:**
```json
{
  "code": 0,
  "message": "success",
  "count": 42
}
```

#### 示例 5: 缓存失效 - 按键删除

**请求:**
```bash
POST /api/v1/users/invalidate
Content-Type: application/json

{
  "keys": ["cache:user:1", "cache:user:2", "cache:user:3"]
}
```

**响应:**
```json
{
  "code": 0,
  "message": "success",
  "count": 3
}
```

#### 示例 6: 缓存失效 - 按模式删除

**请求:**
```bash
POST /api/v1/users/invalidate
Content-Type: application/json

{
  "pattern": "cache:user:*"
}
```

**响应:**
```json
{
  "code": 0,
  "message": "success",
  "count": -1
}
```
*注: count 为 -1 表示按模式删除,无法精确统计数量*

#### 示例 7: 范围查询 - 年龄在18-30之间的用户

**请求:**
```bash
POST /api/v1/users/lookup
Content-Type: application/json

{
  "method": "list",
  "filters": [
    {
      "field": "age",
      "operator": "between",
      "value": [18, 30]
    }
  ]
}
```

#### 示例 8: IN 查询 - 多状态查询

**请求:**
```bash
POST /api/v1/users/lookup
Content-Type: application/json

{
  "method": "list",
  "filters": [
    {
      "field": "status",
      "operator": "in",
      "value": ["active", "pending", "verified"]
    }
  ]
}
```

## 四、代码讲解

### 4.1 核心组件说明

#### LookupQueryMethod - 查询方法定义
```go
type LookupQueryMethod[T any] struct {
    Name             string                    // 方法名称(如"list","active_list")
    Service          *service.ServiceManager[T] // 服务管理器
    KeyPattern       string                     // Redis键模式
    CacheExpire      time.Duration              // 缓存过期时间
    FallbackToDB     bool                       // 是否回源数据库
    CustomFilterFunc func(...)                  // 自定义过滤函数
}
```

**执行流程:**
1. 获取所有匹配 KeyPattern 的 Redis 键
2. 应用自定义过滤函数(如果有)
3. 应用前端传入的过滤器
4. 从缓存查询数据(必要时回源数据库)

#### LookupRouterGroup - 路由组管理器
```go
type LookupRouterGroup[T any] struct {
    RouterGroup        *gin.RouterGroup              // Gin路由组
    Service            *service.ServiceManager[T]    // 服务管理器
    MethodRegistry     *LookupMethodRegistry[T]      // 方法注册表
    TranslatorRegistry *filter_translator.RedisTranslatorRegistry // 过滤器翻译器
}
```

**主要功能:**
- 注册查询方法
- 注册 HTTP 路由
- 处理请求并返回响应

### 4.2 使用方式详解

#### 方式 1: 使用预定义方法

```go
// 最简单的方式 - 注册通用方法
userLookup.RegisterCommonMethods("cache:user:*", 1*time.Hour)
// 等价于:
// - RegisterListMethod: 基础列表查询
// - RegisterFallbackMethod: 带数据库回源查询
```

#### 方式 2: 单独注册方法

```go
// 基础列表
userLookup.RegisterListMethod("cache:user:*", 1*time.Hour)

// 活跃列表(带自定义过滤)
userLookup.RegisterActiveListMethod("cache:user:*", 1*time.Hour, customFunc)

// 数据库回源
userLookup.RegisterFallbackMethod("cache:user:*", 30*time.Minute)
```

#### 方式 3: 完全自定义方法

```go
userLookup.RegisterMethod(
    "vip_users",              // 方法名
    "cache:user:*",           // 键模式
    2*time.Hour,              // 缓存时间
    true,                     // 是否回源
    func(ctx context.Context, client *redis.Client, keys []string) ([]string, error) {
        // 自定义过滤逻辑
        var vipKeys []string
        for _, key := range keys {
            // 检查用户等级等...
            vipKeys = append(vipKeys, key)
        }
        return vipKeys, nil
    },
)
```

### 4.3 过滤器工作原理

过滤器采用**两阶段过滤**:

1. **自定义过滤阶段** (可选)
   - 在方法注册时通过 `CustomFilterFunc` 定义
   - 适用于复杂的业务逻辑过滤

2. **标准过滤阶段**
   - 前端传入的 filters 参数
   - 通过 `RedisTranslatorRegistry` 翻译为 `RedisFilter`
   - 支持 11 种标准操作符

**过滤执行顺序:**
```
所有键 → 自定义过滤 → 标准过滤器1 → 标准过滤器2 → ... → 最终结果
```

## 五、最佳实践

### 5.1 合理设置缓存时间

```go
// 频繁变动的数据
userLookup.RegisterListMethod("cache:session:*", 5*time.Minute)

// 相对稳定的数据
userLookup.RegisterListMethod("cache:user:*", 1*time.Hour)

// 很少变动的数据
configLookup.RegisterListMethod("cache:config:*", 24*time.Hour)
```

### 5.2 组合使用多种查询方法

```go
// 基础查询(不回源)
userLookup.RegisterListMethod("cache:user:*", 1*time.Hour)

// 带回源的查询(数据可能不完整时使用)
userLookup.RegisterFallbackMethod("cache:user:*", 30*time.Minute)

// 特定业务场景的查询
userLookup.RegisterMethod("premium_users", "cache:user:*", 2*time.Hour, false, premiumFilterFunc)
```

### 5.3 前端调用建议

```javascript
// 封装 API 调用函数
async function lookupUsers(filters) {
  const response = await fetch('/api/v1/users/lookup', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({
      method: 'list',
      filters: filters
    })
  });
  return response.json();
}

// 使用示例
const activeAdults = await lookupUsers([
  { field: 'age', operator: 'gte', value: 18 },
  { field: 'status', operator: 'eq', value: 'active' }
]);
```

## 六、错误处理

常见错误及处理:

| 错误码 | 说明 | 处理方式 |
|-------|------|---------|
| 400 | 请求参数错误 | 检查 method 和 filters 格式 |
| 404 | 键不存在 | 确认键名正确或使用带回源的方法 |
| 500 | 服务器错误 | 检查 Redis 连接和数据格式 |

---

以上就是 Lookup 模块的完整使用文档,包含了路由说明、代码示例和最佳实践。

## 缓存写入路由(WritedownRouterGroup)

# cache set (写入) 模块使用文档

## 一、路由字段说明表

| 路由字段 | 对应 Handler | 入参（示例） | 请求方式 | 出参（示例） |
|---------|--------------|------------|---------|------------|
| `POST /api/v1/{resource}/cache/write` | `HandleWritedownSingle` | `{"key":"cache:user:1","data":{...},"expiration":3600}` 或 `{"key":"cache:user:1","id":1}` | POST | `{"code":0,"message":"success","items_written":1}` |
| `POST /api/v1/{resource}/cache/write-lock` | `HandleWritedownWithLock` | `{"key":"cache:user:1","id":1,"expiration":3600}` | POST | `{"code":0,"message":"success","items_written":1,"data":{...}}` |
| `POST /api/v1/{resource}/cache/write-version` | `HandleWritedownWithVersion` | `{"key":"cache:user:1","data":{...},"version":2,"expiration":3600}` | POST | `{"code":0,"message":"success","items_written":1}` |
| `POST /api/v1/{resource}/cache/refresh` | `HandleRefreshCache` | `{"key":"cache:user:1","id":1,"expiration":3600}` | POST | `{"code":0,"message":"success","items_written":1}` |
| `POST /api/v1/{resource}/cache/batch-write` | `HandleWritedownQuery` | `{"key_template":"cache:user:{id}","data":[...],"expiration":3600}` 或 `{"key_template":"cache:user:{id}","ids":[1,2,3]}` | POST | `{"code":0,"message":"success","items_written":N}` |
| `POST /api/v1/{resource}/cache/warmup` | `HandleWarmupCache` | `{"key_template":"cache:user:{id}","limit":1000,"order_by":"access_count"}` | POST | `{"code":0,"message":"success"}` |

## 二、请求结构说明（简要）

- 单个写入 `WritedownSingleRequest`:
  - `key`: 必填，缓存键
  - `data`: 可选，直接提供要写入的数据
  - `id`: 可选，通过数据库加载数据（与 `data` 二选一）
  - `expiration`: 过期时间（秒），默认 3600
  - `overwrite`/`nx`/`xx`: 控制是否覆盖或仅在存在/不存在时写入
  - `async`: 是否异步写入

- 批量写入 `WritedownQueryRequest`:
  - `data` / `ids` / `load_all`：三选一（直接提供数据、指定 ID 列表或加载全量）
  - `key_template`: 必填，键模板，如 `cache:user:{id}`
  - `expiration`、`batch_size`、`use_pipeline`、`incremental` 等控制写入行为

- 带锁写入 `WritedownWithLockRequest`：用于并发场景，提供 `key` 和 `id`，可设置 `lock_timeout`
- 带版本写入 `WritedownWithVersionRequest`：提供 `data` 与 `version`，用于乐观并发控制
- 预热 `WarmupCacheRequest`：按 `key_template` + 排序/limit 预热缓存

## 三、快速示例代码（注册路由）

```go
package main

import (
    "github.com/gin-gonic/gin"
    "time"

    "github.com/yourname/go-blog/internal/huancun/http_router"
    "github.com/yourname/go-blog/internal/huancun/service"
)

// 假设已有模型 User
func main() {
    r := gin.Default()
    v1 := r.Group("/api/v1")

    userService := service.NewServiceManager(User{})
    // 创建缓存写入路由组
    userW := http_router.NewWritedownRouterGroup(v1, userService)
    // 可选：自定义 KeyBuilder
    // userW.SetKeyBuilder(customBuilder)

    // 注册路由（资源路径）
    userW.RegisterRoutes("/users")

    r.Run(":8080")
}
```

## 四、API 调用示例

### 4.1 单个写入（同步）

请求:
```bash
POST /api/v1/users/cache/write
Content-Type: application/json

{
  "key": "cache:user:1",
  "data": {
    "id": 1,
    "name": "John",
    "age": 30
  },
  "expiration": 3600,
  "overwrite": true
}
```

响应:
```json
{
  "code": 0,
  "message": "success",
  "items_written": 1
}
```

### 4.2 单个写入（通过 ID 从 DB 加载）

请求:
```bash
POST /api/v1/users/cache/write
Content-Type: application/json

{
  "key": "cache:user:1",
  "id": 1,
  "expiration": 3600
}
```

响应同上。

### 4.3 带锁写入（确保串行化）

请求:
```bash
POST /api/v1/users/cache/write-lock
Content-Type: application/json

{
  "key": "cache:user:1",
  "id": 1,
  "expiration": 3600,
  "lock_timeout": 5
}
```

响应:
```json
{ "code": 0, "message": "success", "items_written": 1, "data": { /* 返回写入的数据 */ } }
```

### 4.4 带版本写入（乐观并发）

请求:
```bash
POST /api/v1/users/cache/write-version
Content-Type: application/json

{
  "key": "cache:user:1",
  "data": { "id":1, "name":"John", "age":31 },
  "version": 2,
  "expiration": 3600
}
```

响应:
```json
{ "code": 0, "message": "success", "items_written": 1 }
```

### 4.5 刷新单个缓存（从 DB 重新加载）

请求:
```bash
POST /api/v1/users/cache/refresh
Content-Type: application/json

{
  "key": "cache:user:1",
  "id": 1,
  "expiration": 3600
}
```

响应:
```json
{ "code": 0, "message": "success", "items_written": 1 }
```

### 4.6 批量写入 / 预热

请求（按数据列表写入）:
```bash
POST /api/v1/users/cache/batch-write
Content-Type: application/json

{
  "key_template": "cache:user:{id}",
  "data": [ {"id":1,...}, {"id":2,...} ],
  "expiration": 3600,
  "batch_size": 100
}
```

请求（按 ID 列表写入）:
```bash
POST /api/v1/users/cache/batch-write
Content-Type: application/json

{
  "key_template": "cache:user:{id}",
  "ids": [1,2,3]
}
```

请求（预热）:
```bash
POST /api/v1/users/cache/warmup
Content-Type: application/json

{
  "key_template": "cache:user:{id}",
  "limit": 1000,
  "order_by": "access_count",
  "expiration": 3600
}
```

响应示例:
```json
{ "code": 0, "message": "success", "items_written": N }
```

## 五、注意事项

- `key_template` 必须与 `KeyBuilder`/模型字段匹配，例如 `cache:user:{id}`。
- 当使用 `nx`/`xx` 或 `overwrite` 时，请注意并发场景下的语义区别。
- 批量写入支持 `use_pipeline` 来提高大规模写入性能，但会占用更多 Redis 连接。
- 对于需要强一致性的场景，可使用带锁写入或带版本写入。


---

如需我把示例中的模型或请求体改成更贴合你项目的字段，告诉我模型结构即可，我会更新示例。