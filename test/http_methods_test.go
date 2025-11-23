/*
 * @Author: Chris
 * @Date: 2025-03-22 16:30:00
 * @LastEditors: Strong
 * @LastEditTime: 2025-03-22 16:30:00
 * @Description: HTTP方法测试 - 测试PUT、DELETE、PATCH、HEAD、OPTIONS方法
 */
package qiao

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"testing"

	"github.com/chris-liu-zh/qiao/Http"
)

// 测试数据结构
type TestData struct {
	ID      int    `json:"id"`
	Name    string `json:"name"`
	Message string `json:"message"`
}

// 测试HTTP方法
func Test_HttpMethods(t *testing.T) {
	r := Http.NewRouter()

	// 注册各种HTTP方法的路由
	r.Get("/test/get", HandleGet)
	r.Post("/test/post", HandlePost)
	r.Put("/test/put/{id}", HandlePut)
	r.Delete("/test/delete/{id}", HandleDelete)
	r.Patch("/test/patch/{id}", HandlePatch)
	r.Head("/test/head", HandleHead)
	r.Options("/test/options", HandleOptions)

	// 启动测试服务器
	Http.NewHttpServer(":8081", r).Start()
}

// GET请求处理
func HandleGet(w http.ResponseWriter, r *http.Request) {
	data := TestData{
		ID:      1,
		Name:    "GET Test",
		Message: "This is a GET request test",
	}
	Http.Success(w, data)
}

// POST请求处理
func HandlePost(w http.ResponseWriter, r *http.Request) {
	var requestData TestData
	if err := json.NewDecoder(r.Body).Decode(&requestData); err != nil {
		Http.BadRequest(w)
		return
	}

	responseData := TestData{
		ID:      requestData.ID,
		Name:    requestData.Name,
		Message: "Data created successfully",
	}
	Http.Success(w, responseData)
}

// PUT请求处理 - 更新资源
func HandlePut(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	var requestData TestData
	if err := json.NewDecoder(r.Body).Decode(&requestData); err != nil {
		Http.BadRequest(w)
		return
	}

	responseData := TestData{
		ID:      requestData.ID,
		Name:    requestData.Name,
		Message: fmt.Sprintf("Resource with ID %s updated successfully", id),
	}
	Http.Success(w, responseData)
}

// DELETE请求处理 - 删除资源
func HandleDelete(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	responseData := map[string]string{
		"message": fmt.Sprintf("Resource with ID %s deleted successfully", id),
	}
	Http.Success(w, responseData)
}

// PATCH请求处理 - 部分更新资源
func HandlePatch(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	var requestData map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&requestData); err != nil {
		Http.BadRequest(w)
		return
	}

	responseData := map[string]interface{}{
		"id":             id,
		"message":        "Resource partially updated successfully",
		"updated_fields": requestData,
	}
	Http.Success(w, responseData)
}

// HEAD请求处理 - 返回头部信息
func HandleHead(w http.ResponseWriter, r *http.Request) {
	// HEAD请求通常只返回头部信息，不返回body
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Test-Header", "HEAD Request Test")
	w.WriteHeader(http.StatusOK)
}

// OPTIONS请求处理 - 返回支持的HTTP方法
func HandleOptions(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Allow", "GET, POST, PUT, DELETE, PATCH, HEAD, OPTIONS")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, PATCH, HEAD, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
	w.WriteHeader(http.StatusOK)
}

// 测试所有HTTP方法的集成测试
func Test_AllHttpMethods(t *testing.T) {
	r := Http.NewRouter()

	// 注册一个统一的路由，测试所有方法
	r.Get("/api/resource", HandleResource)
	r.Post("/api/resource", HandleResource)
	r.Put("/api/resource/{id}", HandleResource)
	r.Delete("/api/resource/{id}", HandleResource)
	r.Patch("/api/resource/{id}", HandleResource)
	r.Head("/api/resource", HandleResource)
	r.Options("/api/resource", HandleResource)

	log.Println("HTTP方法测试服务器启动在 :8082")
	Http.NewHttpServer(":8082", r).Start()
}

// 统一资源处理函数
func HandleResource(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"method":  r.Method,
		"path":    r.URL.Path,
		"id":      r.PathValue("id"),
		"message": fmt.Sprintf("Handled %s request", r.Method),
	}

	// 对于HEAD请求，只设置头部
	if r.Method == "HEAD" {
		w.Header().Set("X-Method", "HEAD")
		w.WriteHeader(http.StatusOK)
		return
	}

	// 对于OPTIONS请求，返回支持的HTTP方法
	if r.Method == "OPTIONS" {
		w.Header().Set("Allow", "GET, POST, PUT, DELETE, PATCH, HEAD, OPTIONS")
		w.WriteHeader(http.StatusOK)
		return
	}

	Http.Success(w, response)
}
