package qiao

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"
)

type HttpClient struct {
	req     *http.Request
	Url     string
	Method  string
	Cookies []*http.Cookie
	err     error
}

// 全局HTTP客户端，带连接池和超时控制
var (
	defaultClient *http.Client
	clientOnce    sync.Once
)

// 获取默认HTTP客户端（带连接池）
func getDefaultClient() *http.Client {
	clientOnce.Do(func() {
		transport := &http.Transport{
			TLSClientConfig:     &tls.Config{InsecureSkipVerify: true},
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 10,
			IdleConnTimeout:     90 * time.Second,
		}
		defaultClient = &http.Client{
			Transport: transport,
			Timeout:   30 * time.Second,
		}
	})
	return defaultClient
}

func NewHttpClient(url string) *HttpClient {
	return &HttpClient{
		Url: url,
	}
}

func (client *HttpClient) Cookie(cookies []*http.Cookie) *HttpClient {
	client.Cookies = cookies
	return client
}

func (client *HttpClient) Post() *HttpClient {
	client.Method = "POST"
	return client
}

func (client *HttpClient) Get() *HttpClient {
	if client.req, client.err = http.NewRequest("GET", client.Url, nil); client.err != nil {
		return client
	}
	return client
}

func (client *HttpClient) Delete() *HttpClient {
	client.Method = "DELETE"
	return client
}

func (client *HttpClient) Put() *HttpClient {
	client.Method = "PUT"
	return client
}

func (client *HttpClient) DoJson(bodyJson []byte) *HttpClient {
	reqBody := io.NopCloser(bytes.NewReader(bodyJson))
	if client.req, client.err = http.NewRequest(client.Method, client.Url, reqBody); client.err != nil {
		return client
	}
	for _, c := range client.Cookies {
		client.req.AddCookie(&http.Cookie{Name: c.Name, Value: c.Value, HttpOnly: c.HttpOnly})
	}
	return client
}

func (client *HttpClient) HeaderAdd(header map[string]string) *HttpClient {
	if client.err != nil {
		return client
	}
	for k, v := range header {
		client.req.Header.Add(k, v)
	}
	return client
}

func (client *HttpClient) Respond() (body []byte, cookies []*http.Cookie, err error) {
	if client.err != nil {
		err = client.err
		return
	}

	// 使用带连接池的默认客户端
	c := getDefaultClient()

	var resp *http.Response
	if resp, err = c.Do(client.req); err != nil {
		return
	}
	cookies = resp.Cookies()
	defer DeferErr(&err, resp.Body.Close)
	if resp.StatusCode != http.StatusOK {
		err = fmt.Errorf("method:%v url: %v code:%d", client.Method, client.Url, resp.StatusCode)
		return
	}
	if body, err = io.ReadAll(resp.Body); err != nil {
		return
	}
	return
}

// 带自定义超时的HTTP客户端
func NewHttpClientWithTimeout(url string, timeout time.Duration) *HttpClient {
	client := &HttpClient{
		Url: url,
	}

	// 为这个特定客户端创建专用的传输层
	transport := &http.Transport{
		TLSClientConfig:     &tls.Config{InsecureSkipVerify: true},
		MaxIdleConns:        50,
		MaxIdleConnsPerHost: 5,
		IdleConnTimeout:     60 * time.Second,
	}

	clientOnce.Do(func() {
		defaultClient = &http.Client{
			Transport: transport,
			Timeout:   timeout,
		}
	})

	return client
}
