package Http

import (
	"net/http"
	"net/http/httputil"
	"net/url"
)

func NewStaticFileProxy(targetUrl string) (*httputil.ReverseProxy, error) {
	// 解析目标URL
	target, err := url.Parse(targetUrl)
	if err != nil {
		return nil, err
	}

	// 创建反向代理
	proxy := httputil.NewSingleHostReverseProxy(target)

	// 自定义Director函数，确保请求路径和头部正确转发
	originalDirector := proxy.Director
	proxy.Director = func(req *http.Request) {
		// 执行原始的Director逻辑
		originalDirector(req)

		// 1. 保留原始请求的路径（核心：确保/upload/image/aaaa.png这类路径完整转发）
		// 2. 重置Host为目标服务器的Host，避免跨域/服务器识别问题
		req.Host = target.Host
		// 3. 清除可能导致代理异常的头部
		req.Header.Del("Accept-Encoding")
		// 4. 可选：添加X-Forwarded-For头部，标识原始客户端IP
		req.Header.Set("X-Forwarded-For", req.RemoteAddr)
	}

	// 自定义错误处理（可选，增强容错性）
	proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		http.Error(w, "代理服务暂时不可用", http.StatusBadGateway)
	}

	return proxy, nil
}
