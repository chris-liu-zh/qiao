# qiao 框架

`qiao` 是一个用 Go 语言编写的综合性框架，旨在为开发者提供便捷、高效的工具集，用于构建各类 Go 应用程序。该框架包含了多个功能模块，涵盖错误处理、日志记录、数据库操作、HTTP 客户端以及任务处理等多个方面。

## 核心功能模块

### 1. 错误处理
框架定义了 `qiaoError` 结构体，用于封装详细的错误信息，包括错误本身、出错的文件名、行号、函数名以及其他额外信息。同时提供了 `Err` 函数，方便开发者生成和处理错误，避免重复的错误处理逻辑。

### 2. 日志记录
提供了 `QiaoLogger` 结构体，支持日志的切割和轮转功能。通过 `SetLog` 函数，开发者可以轻松设置日志记录器，指定日志文件的路径、大小、备份数量以及是否进行压缩等参数。

### 3. 数据库操作
`Mapper` 结构体用于构建 SQL 查询语句，`QiaoDB` 函数可创建 `Mapper` 实例，并支持开启调试模式，方便开发者进行数据库操作的调试。

### 4. HTTP 客户端
`httpClient` 结构体可用于发起 HTTP 请求，支持 GET、POST、DELETE、PUT 等常见的 HTTP 方法。同时，支持设置请求头、Cookie 和请求体，以及对响应结果进行处理。

### 5. 任务处理
`WG` 结构体用于管理任务的并发执行，`TaskStart` 函数可启动任务处理，支持添加任务和获取任务结果，方便开发者进行多任务处理。

## 测试模块
框架提供了多个测试文件，如 `http_test.go`、`cache_test.go`、`db_test.go` 等，用于对框架的各个功能模块进行测试，确保框架的稳定性和可靠性。

## 代码示例
以下是一个使用 `qiao` 框架的 HTTP 服务器示例：

```go
package qiao

import (
	"log"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/chris-liu-zh/qiao/Http"
)

const (
	ATExp = 1 * time.Hour
	RTExp = 72 * time.Hour
)

var newAuth = Http.DefaultAuth(ATExp, RTExp, "1D4JWUEGWWFK94JB74W1YGP9OF4L205F")

func Test_Http(t *testing.T) {
	if err := Http.NewTemplates("template/*.html", "template/**/*.html"); err != nil {
		log.Println(err)
	}

	r := Http.NewRouter()
	r.SetOnEvicted(onEvicted)
	r.SetTimeout(10 * time.Second)
	r.SetHeader(Http.DefaultHeader)
	r.SetContextSetter(setContest)
	r.SetSign("/api/", sign)
	r.SetAuth("/api/user/", auth)
	r.Get("/version", GetVersion)
	r.Get("/users/{id}", GetUserByID)
	r.Get("/", home)
	r.FileServer("/static/", "./template")
	Http.NewHttpServer(":8080", r).Start()
}

func home(w http.ResponseWriter, r *http.Request) {
	var data struct {
		Public string
	}
	data.Public = "public"
	Http.Html(w, "template/index", data)
}

func onEvicted(w http.ResponseWriter, r *http.Request) {
	h := Http.GetHeader(r)
	if r.URL.Path == "/logout" {
		if token, ok := h["Authorization"]; ok {
			newAuth.SetInvalidToken(token)
		}
	}
}

func setContest(r *http.Request) *http.Request {
	return Http.SetContext(r, "user", "admin")
}

func GetUserByID(w http.ResponseWriter, r *http.Request) {
	user, ok := Http.GetContext(r, "user").(string)
	if ok {
		log.Printf("获取到的用户值: %s", user)
	} else {
		log.Println("未获取到用户值")
	}

	id := r.PathValue("id")
	response := map[string]string{
		"id":   id,
		"name": user,
	}
	Http.Success(response).Json(w)
}

func GetVersion(w http.ResponseWriter, r *http.Request) {
	var ver struct {
		Version string `json:"version"`
		Ip      string `json:"ip"`
	}
	ver.Version = "1.0.0"
	ver.Ip = strings.Split(r.RemoteAddr, ":")[0]
	Http.Html(w, "template/version/index", ver)
}

func sign(header map[string]string) error {
	const (
		key    = "ALYDDNS"
		secret = "1D4JWUEGWWFK94JB74W1YGP9OF4L205F"
	)
	return Http.DefaultSign(header, key, secret)
}

func auth(header map[string]string) (contextKey Http.CtxKey, data any, err error) {
	if data, err = newAuth.CheckToken(header); err != nil {
		return "", nil, err
	}
	return "userinfo", data, nil
}
```