# HTTP方法测试说明

## 新增的HTTP方法

在<mcfile name="route.go" path="/Users/chris/Works/project/qiao/Http/route.go"></mcfile>中新增了以下HTTP方法：

- `Put(path string, handler http.HandlerFunc)` - 处理HTTP PUT请求
- `Delete(path string, handler http.HandlerFunc)` - 处理HTTP DELETE请求
- `Patch(path string, handler http.HandlerFunc)` - 处理HTTP PATCH请求
- `Head(path string, handler http.HandlerFunc)` - 处理HTTP HEAD请求
- `Options(path string, handler http.HandlerFunc)` - 处理HTTP OPTIONS请求

## 测试文件说明

### 1. http_methods_test.go

这个文件包含了完整的HTTP方法测试代码：

- `Test_HttpMethods()` - 测试各个HTTP方法的独立路由
- `Test_AllHttpMethods()` - 测试统一资源的多方法路由
- 每个HTTP方法都有对应的处理函数

### 2. test.http

这个文件包含了HTTP请求测试用例，可以使用HTTP客户端工具（如VS Code的REST Client插件）进行测试：

- 端口8081：独立方法测试服务器
- 端口8082：统一资源测试服务器

## 运行测试

### 方法1：运行Go测试

```bash
# 运行所有测试
cd /Users/chris/Works/project/qiao
go test ./test/...

# 只运行HTTP方法测试
go test -v -run "Test_HttpMethods" ./test/
```

### 方法2：使用HTTP客户端测试

1. 首先启动测试服务器：
```bash
cd /Users/chris/Works/project/qiao
go test -v -run "Test_HttpMethods" ./test/ -timeout=30s
```

2. 然后使用HTTP客户端工具发送请求到：
   - `http://127.0.0.1:8081` - 独立方法测试
   - `http://127.0.0.1:8082` - 统一资源测试

## 测试用例说明

### PUT方法
- 用途：更新完整资源
- 示例：`PUT /test/put/123`
- 需要请求体包含完整资源数据

### DELETE方法
- 用途：删除资源
- 示例：`DELETE /test/delete/123`
- 通常不需要请求体

### PATCH方法
- 用途：部分更新资源
- 示例：`PATCH /test/patch/123`
- 请求体包含需要更新的字段

### HEAD方法
- 用途：获取资源头部信息
- 示例：`HEAD /test/head`
- 不返回响应体，只返回头部

### OPTIONS方法
- 用途：查询资源支持的HTTP方法
- 示例：`OPTIONS /test/options`
- 在Allow头部返回支持的方法列表

## 注意事项

1. 测试服务器使用不同的端口（8081和8082）以避免冲突
2. 每个测试都包含完整的错误处理
3. 响应格式统一使用JSON
4. 支持路径参数（如`{id}`）

这些测试文件可以帮助你验证新添加的HTTP方法是否正常工作。