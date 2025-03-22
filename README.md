# Star-Go 项目

一个基于 Gin 框架的 Go 语言 Web 应用，提供完整的用户认证和管理功能。

## 项目简介

Star-Go 是一个使用 Go 语言和 Gin 框架开发的 Web 应用骨架，集成了用户认证、权限管理、数据库操作等常用功能。

## 技术栈

- **Go**: 1.23.0+ （不能低于1.23）
- **Gin**: Web 框架
- **GORM**: ORM 数据库操作
- **MySQL**: 数据库
- **Redis**: 缓存系统（可选）
- **JWT**: 用户认证
- **Zap**: 日志系统
- **Viper**: 配置管理

## 项目结构

```
star-go/
├── api/                    # API 相关代码
│   ├── handlers/           # 处理器
│   └── routes/             # 路由定义
├── docs/                   # 文档
├── internal/               # 内部应用代码
│   ├── controllers/        # 控制器
│   ├── models/             # 数据模型
│   ├── repository/         # 数据仓库
│   └── services/           # 业务逻辑
├── logs/                   # 日志文件
├── pkg/                    # 可重用的包
│   ├── cache/              # 缓存系统
│   ├── config/             # 配置管理
│   ├── core/               # 应用核心
│   ├── database/           # 数据库连接
│   ├── logger/             # 日志系统
│   ├── middleware/         # 中间件
│   └── utils/              # 工具函数
├── config.yaml             # 配置文件
├── go.mod                  # Go 模块定义
├── go.sum                  # 依赖版本锁定
├── main.go                 # 应用入口
└── README.md               # 项目说明
```

## 功能特性

### 用户管理
- 用户注册、登录、退出
- 用户信息查询与更新
- 密码修改
- 用户状态管理

### 权限控制
- 基于 JWT 的认证系统
- 角色权限管理（管理员、普通用户、访客）
- 接口访问控制

### 系统功能
- 完善的日志系统
- 请求限流
- 跨域支持
- 全局异常处理
- 数据库自动迁移
- 缓存系统支持（Redis/内存）
- 优雅启动和关闭

## 核心组件说明

### 应用核心 (pkg/core)
- 应用程序初始化和生命周期管理
- 优雅启动和关闭
- 组件集成

### 配置管理 (pkg/config)
- 基于 Viper 的配置管理
- 支持 YAML 格式配置文件
- 环境变量覆盖
- 服务器、数据库、缓存、JWT等配置

### 数据库 (pkg/database)
- 基于 GORM 的 MySQL 数据库连接
- 连接池配置
- 慢查询日志
- 数据库迁移

### 缓存系统 (pkg/cache)
- 支持 Redis 和内存缓存
- 通用缓存接口
- 可配置的过期时间
- 键前缀管理
- 连接池优化

### 日志系统 (pkg/logger)
- 基于 Zap 的高性能日志系统
- 日志分级
- 日志轮转
- 支持开发和生产环境配置
- Gin框架集成

### 中间件 (pkg/middleware)
- JWT 认证
- 角色验证
- 请求限流
- 跨域处理
- 异常恢复

### API 路由 (api/routes)
- RESTful API 设计
- 版本化 API
- 公开和私有路由分组

### 控制器 (internal/controllers)
- 用户认证控制器
- 用户管理控制器

### 数据模型 (internal/models)
- 用户模型
- 数据验证
- 密码加密

## 数据架构设计

Star-Go 采用了简化的 RBAC (基于角色的访问控制) 模型，主要由用户表和角色表两个核心表组成，而不是传统 RBAC 实现中的五到六个表（用户、角色、权限以及它们的关联表）。

### 数据模型关系

```
┌─────────┐       ┌─────────┐
│  User   │       │  Role   │
├─────────┤       ├─────────┤
│ ID      │       │ ID      │
│ Username│       │ Name    │
│ Password│       │ Code    │
│ Email   │       │ Desc    │
│ RoleID  │──────>│ Perms   │ (JSON Array)
└─────────┘       └─────────┘
```

### 核心表结构

#### 用户表 (users)
```go
type User struct {
    BaseModel
    Username     string    `gorm:"size:50;not null;uniqueIndex" json:"username"`
    Password     string    `gorm:"size:255;not null" json:"-"`
    Email        string    `gorm:"size:100;uniqueIndex" json:"email"`
    Nickname     string    `gorm:"size:50" json:"nickname"`
    Avatar       string    `gorm:"size:255" json:"avatar"`
    Status       int       `gorm:"default:1" json:"status"` // 1:正常 0:禁用
    LastLogin    time.Time `json:"lastLogin"`
    RoleID       uint      `gorm:"default:2" json:"roleId"` // 默认为普通用户角色
    Role         *Role     `gorm:"foreignKey:RoleID" json:"role"`
}
```

#### 角色表 (roles)
```go
type Role struct {
    BaseModel
    Name        string      `gorm:"size:50;not null;uniqueIndex" json:"name"`
    Code        string      `gorm:"size:50;not null;uniqueIndex" json:"code"`
    Description string      `gorm:"size:200" json:"description"`
    Permissions Permissions `gorm:"type:json" json:"permissions"` // 使用JSON存储权限列表
}

type Permissions []string
```

### 两表设计的优势

1. **简化的数据结构**：
   - 减少了表的数量，简化了数据库设计和维护
   - 降低了查询复杂度，减少了多表连接操作
   - 提高了查询性能，特别是在用户验证和权限检查时

2. **更高效的权限检查**：
   - 角色和权限信息可以在一次查询中获取
   - 权限检查可以在内存中完成，无需额外的数据库查询
   - 减少了数据库访问次数，提高了响应速度

3. **更简单的代码实现**：
   - 减少了ORM关联配置的复杂性
   - 简化了权限验证的逻辑实现
   - 降低了开发和维护成本

4. **更好的缓存友好性**：
   - 角色和权限数据可以更容易地缓存
   - 减少了缓存失效的情况
   - 提高了系统整体性能

5. **适合中小型应用**：
   - 对于大多数中小型应用来说，这种设计已经足够
   - 实现了权限控制的核心功能，同时保持了简单性
   - 降低了系统复杂度，提高了可维护性

6. **JSON存储权限的灵活性**：
   - 使用JSON存储权限列表提供了灵活性
   - 可以轻松添加、删除和修改权限，而无需更改表结构
   - 支持复杂的权限结构，如分组和层次结构

### 设计的局限性

1. **角色分配的限制**：
   - 每个用户只能分配一个角色，不支持多角色
   - 对于需要更复杂角色分配的场景可能不够灵活

2. **权限管理的复杂性**：
   - 当权限数量很大时，JSON字段可能变得难以管理
   - 对特定权限的查询可能不如关系表高效

3. **扩展性考虑**：
   - 如果未来需要支持多角色，需要重新设计数据结构
   - 对于非常大型的应用，可能需要转向更传统的多表RBAC模型

### 适用场景

这种两表设计特别适合：

- 中小型应用和项目
- 权限结构相对简单的系统
- 对性能有较高要求的应用
- 快速开发和迭代的项目
- 用户角色相对固定的系统

对于大型企业应用或需要非常复杂权限控制的系统，可以考虑扩展为更传统的多表RBAC模型。

## 启动项目

1. 配置服务
   - 修改 `config.yaml` 中的配置信息
   - 服务器配置：监听地址、端口等
   - 数据库配置：连接信息、连接池等
   - 缓存配置：类型、连接信息等（可选）
   - JWT配置：密钥、过期时间等
   - 日志配置：级别、输出方式等

2. 启动服务
   ```bash
   go run main.go
   ```

3. 访问 API
   - 默认地址: http://localhost:8080
   - 可通过配置文件修改监听地址和端口

## 配置说明

### 服务器配置
```yaml
server:
  host: "0.0.0.0"           # 监听地址，0.0.0.0表示所有地址，localhost仅本地访问
  port: 8080                # 监听端口
  mode: debug               # 服务器模式（debug/release/test）
  readTimeout: 10           # 读取超时（秒）
  writeTimeout: 10          # 写入超时（秒）
  disableDebug: false       # 是否禁用调试输出
```

### 数据库配置
```yaml
database:
  type: mysql               # 数据库类型
  host: localhost           # 数据库主机
  port: 3306                # 数据库端口
  username: root            # 用户名
  password: password        # 密码
  database: star_go         # 数据库名
  maxIdleConns: 10          # 最大空闲连接数
  maxOpenConns: 100         # 最大打开连接数
  connMaxLifetime: 3600     # 连接最大生命周期（秒）
  logLevel: info            # 日志级别
  slowThreshold: 200        # 慢查询阈值（毫秒）
```

### 缓存配置
```yaml
cache:
  type: redis               # 缓存类型（redis/memory）
  host: localhost           # Redis主机
  port: 6379                # Redis端口
  password: ""              # Redis密码
  db: 0                     # Redis数据库索引
  poolSize: 10              # 连接池大小
  defaultTTL: 3600          # 默认过期时间（秒）
  prefix: "star-go:"        # 键前缀
```

### JWT配置
```yaml
jwt:
  secret: "your-secret-key" # JWT密钥
  accessTokenExp: 15        # 访问令牌过期时间（分钟）
  refreshTokenExp: 10080    # 刷新令牌过期时间（分钟）
  tokenIssuer: "star-go"    # 令牌颁发者
```

## 认证与授权框架使用案例

Star-Go 提供了灵活而强大的认证与授权框架，以下是几个常见的使用案例：

### 1. 基本的 JWT 认证

```go
// 路由设置
func setupRoutes(router *gin.Engine) {
    // 公开路由
    public := router.Group("/api")
    {
        public.POST("/auth/login", authController.Login)
        public.POST("/auth/register", authController.Register)
    }
    
    // 需要认证的路由
    protected := router.Group("/api")
    protected.Use(middleware.JWTAuth()) // 添加JWT认证中间件
    {
        protected.GET("/user/profile", userController.GetProfile)
        protected.PUT("/user/profile", userController.UpdateProfile)
    }
}
```

### 2. 基于角色的访问控制 (RBAC)

```go
// 只允许管理员访问的路由
adminRoutes := router.Group("/api/admin")
adminRoutes.Use(middleware.JWTAuth())
adminRoutes.Use(middleware.RoleAuth("admin")) // 只允许admin角色访问
{
    adminRoutes.GET("/users", adminController.ListUsers)
    adminRoutes.DELETE("/users/:id", adminController.DeleteUser)
}

// 允许管理员或编辑者访问的路由
contentRoutes := router.Group("/api/content")
contentRoutes.Use(middleware.JWTAuth())
contentRoutes.Use(middleware.RoleOrPermissionAuth("admin", "content:edit")) // 允许admin角色或有content:edit权限的用户访问
{
    contentRoutes.POST("/articles", contentController.CreateArticle)
    contentRoutes.PUT("/articles/:id", contentController.UpdateArticle)
}
```

### 3. 基于权限的细粒度控制

```go
// 检查单个权限
userRoutes := router.Group("/api/users")
userRoutes.Use(middleware.JWTAuth())
{
    // 查看用户列表需要"user:list"权限
    userRoutes.GET("", middleware.PermissionAuth("user:list"), userController.ListUsers)
    
    // 创建用户需要"user:create"权限
    userRoutes.POST("", middleware.PermissionAuth("user:create"), userController.CreateUser)
    
    // 更新用户需要"user:update"权限
    userRoutes.PUT("/:id", middleware.PermissionAuth("user:update"), userController.UpdateUser)
    
    // 删除用户需要"user:delete"权限
    userRoutes.DELETE("/:id", middleware.PermissionAuth("user:delete"), userController.DeleteUser)
}
```

### 4. 组合多个权限检查

```go
// 需要同时拥有多个权限
reportRoutes := router.Group("/api/reports")
reportRoutes.Use(middleware.JWTAuth())
{
    // 生成财务报表需要同时拥有"report:generate"和"finance:view"权限
    reportRoutes.POST("/finance", 
        middleware.AllPermissionsAuth([]string{"report:generate", "finance:view"}), 
        reportController.GenerateFinanceReport)
}

// 需要拥有任意一个权限
dataRoutes := router.Group("/api/data")
dataRoutes.Use(middleware.JWTAuth())
{
    // 访问数据分析需要拥有"data:view"或"data:analyze"或"admin:data"中的任意一个权限
    dataRoutes.GET("/analytics", 
        middleware.AnyPermissionAuth([]string{"data:view", "data:analyze", "admin:data"}), 
        dataController.GetAnalytics)
}
```

### 5. 角色和权限组合控制

```go
// 同时检查角色和权限
settingsRoutes := router.Group("/api/settings")
settingsRoutes.Use(middleware.JWTAuth())
{
    // 系统设置需要是管理员角色并且拥有"settings:manage"权限
    settingsRoutes.PUT("/system", 
        middleware.RoleAndPermissionAuth("admin", "settings:manage"), 
        settingsController.UpdateSystemSettings)
    
    // 用户设置需要是管理员角色或拥有"settings:user"权限
    settingsRoutes.PUT("/user/:id", 
        middleware.RoleOrPermissionAuth("admin", "settings:user"), 
        settingsController.UpdateUserSettings)
}
```


## 许可证

本项目采用 [MIT 许可证](https://opensource.org/licenses/MIT)。

```
MIT License

Copyright (c) 2025 Star-Go 项目贡献者

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
```

MIT 许可证是一个宽松的许可证，允许任何人自由使用、修改、分发和商业化本项目，唯一的要求是在所有副本中包含原始版权声明和许可证声明。

## 贡献指南

我们欢迎任何形式的贡献，包括但不限于：

- 报告问题和提出建议
- 提交代码改进
- 完善文档
- 添加新功能
- 修复错误

请通过 GitHub Issues 和 Pull Requests 参与项目贡献。

## 联系方式

如有任何问题或建议，请通过以下方式联系我们：

- GitHub Issues: [提交问题](https://github.com/RyoLena/Star_Go/issues)
- Email: 2690373236@qq.com 

感谢您对 Star-Go 项目的关注和支持！
