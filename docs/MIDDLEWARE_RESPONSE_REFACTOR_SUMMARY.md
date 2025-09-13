# 中间件响应方法重构总结

## 重构目标
将 `internal/middleware/middleware.go` 文件中所有直接的 `c.JSON()` 调用替换为使用 `internal/common/response.go` 中的专用响应方法，实现统一的错误响应管理。

## 重构内容

### 1. 新增响应方法
在 `internal/common/response.go` 中添加了：
```go
// TooManyRequestsResponse 429 请求过多响应
func TooManyRequestsResponse(c *gin.Context, message string) {
    c.JSON(http.StatusTooManyRequests, web.ErrorResponse{
        Code:    http.StatusTooManyRequests,
        Message: message,
    })
}
```

### 2. 导入更新
更新了 `middleware.go` 的导入：
```go
import (
    // 移除了：
    // "net/http"
    // "github.com/zy84338719/filecodebox/internal/models/web"
    
    // 新增了：
    "github.com/zy84338719/filecodebox/internal/common"
    
    // 保留其他必要导入...
)
```

### 3. 响应方法替换详情

#### RateLimit 中间件
**上传频率限制 (HTTP 429):**
```go
// 之前
c.JSON(http.StatusTooManyRequests, web.ErrorResponse{
    Code:    429,
    Message: "上传频率过快，请稍后再试",
})

// 现在
common.TooManyRequestsResponse(c, "上传频率过快，请稍后再试")
```

**请求频率限制 (HTTP 429):**
```go
// 之前
c.JSON(http.StatusTooManyRequests, web.ErrorResponse{
    Code:    429,
    Message: "请求频率过快，请稍后再试",
})

// 现在
common.TooManyRequestsResponse(c, "请求频率过快，请稍后再试")
```

#### AdminAuth 中间件
**缺少认证信息 (HTTP 401):**
```go
// 之前
c.JSON(http.StatusUnauthorized, web.ErrorResponse{
    Code:    401,
    Message: "缺少认证信息",
})

// 现在
common.UnauthorizedResponse(c, "缺少认证信息")
```

**认证格式错误 (HTTP 401):**
```go
// 之前
c.JSON(http.StatusUnauthorized, web.ErrorResponse{
    Code:    401,
    Message: "认证格式错误",
})

// 现在
common.UnauthorizedResponse(c, "认证格式错误")
```

**认证失败 (HTTP 401):**
```go
// 之前
c.JSON(http.StatusUnauthorized, web.ErrorResponse{
    Code:    401,
    Message: "认证失败",
})

// 现在
common.UnauthorizedResponse(c, "认证失败")
```

**权限不足 (HTTP 401):**
```go
// 之前
c.JSON(http.StatusUnauthorized, web.ErrorResponse{
    Code:    401,
    Message: "权限不足",
})

// 现在
common.UnauthorizedResponse(c, "权限不足")
```

**Token格式错误 (HTTP 401):**
```go
// 之前
c.JSON(http.StatusUnauthorized, web.ErrorResponse{
    Code:    401,
    Message: "token格式错误",
})

// 现在
common.UnauthorizedResponse(c, "token格式错误")
```

#### ShareAuth 中间件
**上传功能关闭 (HTTP 403):**
```go
// 之前
c.JSON(http.StatusForbidden, web.ErrorResponse{
    Code:    403,
    Message: "上传功能已关闭",
})

// 现在
common.ForbiddenResponse(c, "上传功能已关闭")
```

#### UserAuth 中间件
**缺少认证信息 (HTTP 401):**
```go
// 之前
c.JSON(http.StatusUnauthorized, web.ErrorResponse{
    Code:    401,
    Message: "缺少认证信息",
})

// 现在
common.UnauthorizedResponse(c, "缺少认证信息")
```

**认证格式错误 (HTTP 401):**
```go
// 之前
c.JSON(http.StatusUnauthorized, web.ErrorResponse{
    Code:    401,
    Message: "认证格式错误",
})

// 现在
common.UnauthorizedResponse(c, "认证格式错误")
```

**认证失败 (HTTP 401):**
```go
// 之前
c.JSON(http.StatusUnauthorized, web.ErrorResponse{
    Code:    401,
    Message: "认证失败: " + err.Error(),
})

// 现在
common.UnauthorizedResponse(c, "认证失败: " + err.Error())
```

**Token格式错误 (HTTP 401):**
```go
// 之前
c.JSON(http.StatusUnauthorized, web.ErrorResponse{
    Code:    401,
    Message: "token格式错误",
})

// 现在
common.UnauthorizedResponse(c, "token格式错误")
```

## 重构优势

### 1. 统一响应管理
- 所有错误响应现在通过 `common` 包的统一方法处理
- 便于全局修改响应格式或添加新的响应字段
- 响应逻辑集中管理，避免重复代码

### 2. 简化代码
- 中间件代码更加简洁，可读性更好
- 减少了样板代码，每个错误响应只需一行调用
- 移除了对 `web.ErrorResponse` 结构体的直接依赖

### 3. 类型安全和一致性
- 使用专门的响应方法确保状态码和结构体的一致性
- 避免了手动设置状态码可能出现的错误
- 与项目其他部分的响应方式保持一致

### 4. 可维护性
- 如果需要修改特定类型的错误响应格式，只需修改 `common` 包中的对应方法
- 新增响应类型时，可以在 `common` 包中统一添加
- 便于添加日志、监控等横切关注点

## 使用的响应方法

### 现有方法
- `common.UnauthorizedResponse(c, message)` - 401 未授权错误
- `common.ForbiddenResponse(c, message)` - 403 禁止访问错误

### 新增方法
- `common.TooManyRequestsResponse(c, message)` - 429 请求过多错误

## 验证状态
- ✅ 编译成功
- ✅ 所有 `c.JSON()` 调用已替换为 common 方法
- ✅ 移除了不必要的导入
- ✅ 新增了 `TooManyRequestsResponse` 方法
- ✅ 保持了原有的 HTTP 状态码和错误消息

## 影响范围
- **修改文件**: 
  - `internal/middleware/middleware.go` - 替换所有响应调用
  - `internal/common/response.go` - 新增 429 响应方法
- **依赖关系**: 中间件现在依赖 `common` 包而不是直接依赖 `web` 包
- **兼容性**: 客户端 API 响应格式完全保持不变

## 代码统计
- **替换的响应调用**: 11 处
- **减少的代码行数**: 约 33 行（每个响应从 4-5 行减少到 1 行）
- **新增的响应方法**: 1 个（`TooManyRequestsResponse`）

这次重构进一步完善了项目的响应管理体系，实现了从底层到中间件层的统一响应处理，为后续的维护和扩展提供了良好的基础。
