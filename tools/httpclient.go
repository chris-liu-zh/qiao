package tools

import (
	"bytes"
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"slices"
	"time"
)

type HttpClient struct {
	defaultClient *http.Client
	req           *http.Request
	Url           string
	Method        string
	Cookies       []*http.Cookie
	headers       map[string][]string
	ctx           context.Context
	err           error

	maxRetries     int
	retryDelay     time.Duration
	retryableCodes []int
	body           []byte
}

func NewHttpClient(url string) *HttpClient {
	return &HttpClient{
		Url:     url,
		headers: make(map[string][]string),
		defaultClient: &http.Client{
			Transport: &http.Transport{
				TLSClientConfig:     &tls.Config{InsecureSkipVerify: true},
				MaxIdleConns:        100,
				MaxIdleConnsPerHost: 10,
				IdleConnTimeout:     60 * time.Second,
			},
			Timeout: 10 * time.Second,
		},
		maxRetries:     0,
		retryDelay:     1 * time.Second,
		retryableCodes: []int{http.StatusInternalServerError, http.StatusBadGateway, http.StatusServiceUnavailable, http.StatusGatewayTimeout},
	}
}

func (client *HttpClient) SetCookie(cookies []*http.Cookie) *HttpClient {
	client.Cookies = cookies
	return client
}

// SetHeader 设置单个请求头
func (client *HttpClient) SetHeader(key, value string) *HttpClient {
	client.headers[key] = append(client.headers[key], value)
	return client
}

// SetHeaders 批量设置请求头
func (client *HttpClient) SetHeaders(headers map[string]string) *HttpClient {
	for k, v := range headers {
		client.headers[k] = append(client.headers[k], v)
	}
	return client
}

// SetContentType 设置 Content-Type 请求头
func (client *HttpClient) SetContentType(contentType string) *HttpClient {
	client.headers["Content-Type"] = []string{contentType}
	return client
}

// SetTimeout 设置请求超时时间
func (client *HttpClient) SetTimeout(timeout time.Duration) *HttpClient {
	client.defaultClient.Timeout = timeout
	return client
}

// SetTransport 设置自定义的 HTTP Transport
func (client *HttpClient) SetTransport(transport *http.Transport) *HttpClient {
	client.defaultClient.Transport = transport
	return client
}

// SetContext 设置请求上下文，用于取消请求或传递元数据
func (client *HttpClient) SetContext(ctx context.Context) *HttpClient {
	client.ctx = ctx
	return client
}

// SetMaxRetries 设置最大重试次数，默认 0（不重试）
func (client *HttpClient) SetMaxRetries(maxRetries int) *HttpClient {
	client.maxRetries = maxRetries
	return client
}

// SetRetryDelay 设置重试间隔时间，默认 1 秒
func (client *HttpClient) SetRetryDelay(retryDelay time.Duration) *HttpClient {
	client.retryDelay = retryDelay
	return client
}

// SetRetryableCodes 设置需要重试的 HTTP 状态码列表
func (client *HttpClient) SetRetryableCodes(codes []int) *HttpClient {
	client.retryableCodes = codes
	return client
}

func (client *HttpClient) Post(body io.Reader) *HttpClient {
	client.Method = "POST"
	return client.prepareRequest(body)
}

// PostCtx 带上下文的 POST 请求
func (client *HttpClient) PostCtx(ctx context.Context, body io.Reader) *HttpClient {
	client.Method = "POST"
	client.ctx = ctx
	return client.prepareRequest(body)
}

func (client *HttpClient) Get() *HttpClient {
	client.Method = "GET"
	return client.prepareRequest(nil)
}

// GetCtx 带上下文的 GET 请求
func (client *HttpClient) GetCtx(ctx context.Context) *HttpClient {
	client.Method = "GET"
	client.ctx = ctx
	return client.prepareRequest(nil)
}

func (client *HttpClient) Delete() *HttpClient {
	client.Method = "DELETE"
	return client.prepareRequest(nil)
}

// DeleteCtx 带上下文的 DELETE 请求
func (client *HttpClient) DeleteCtx(ctx context.Context) *HttpClient {
	client.Method = "DELETE"
	client.ctx = ctx
	return client.prepareRequest(nil)
}

func (client *HttpClient) Put(body io.Reader) *HttpClient {
	client.Method = "PUT"
	return client.prepareRequest(body)
}

// PutCtx 带上下文的 PUT 请求
func (client *HttpClient) PutCtx(ctx context.Context, body io.Reader) *HttpClient {
	client.Method = "PUT"
	client.ctx = ctx
	return client.prepareRequest(body)
}

// prepareRequest 准备请求，缓存请求体并构建请求对象
func (client *HttpClient) prepareRequest(body io.Reader) *HttpClient {
	if body != nil {
		b, err := io.ReadAll(body)
		if err != nil {
			client.err = err
			return client
		}
		client.body = b
	}

	client.req = client.buildRequest()
	return client
}

// buildRequest 从存储的字段构建新的 HTTP 请求对象
func (client *HttpClient) buildRequest() *http.Request {
	ctx := client.ctx
	if ctx == nil {
		ctx = context.Background()
	}

	var bodyReader io.Reader
	if client.body != nil {
		bodyReader = bytes.NewReader(client.body)
	}

	req, _ := http.NewRequestWithContext(ctx, client.Method, client.Url, bodyReader)

	for k, values := range client.headers {
		for _, v := range values {
			req.Header.Add(k, v)
		}
	}

	for _, c := range client.Cookies {
		req.AddCookie(&http.Cookie{Name: c.Name, Value: c.Value, HttpOnly: c.HttpOnly})
	}

	return req
}

func (client *HttpClient) Bytes() (body []byte, err error) {
	body, _, err = client.Respond()
	return
}

func (client *HttpClient) Respond() (body []byte, cookies []*http.Cookie, err error) {
	if client.err != nil {
		err = client.err
		return
	}

	ctx := client.ctx
	if ctx == nil {
		ctx = context.Background()
	}

	var resp *http.Response
	for attempt := 0; attempt <= client.maxRetries; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				err = ctx.Err()
				return
			case <-time.After(client.retryDelay):
			}
		}

		client.req = client.buildRequest()

		if resp, err = client.defaultClient.Do(client.req); err != nil {
			if attempt < client.maxRetries && client.isRetryableError(err) {
				continue
			}
			return
		}

		cookies = resp.Cookies()
		respBody, readErr := io.ReadAll(resp.Body)
		resp.Body.Close()

		if readErr != nil {
			err = readErr
			return
		}

		if resp.StatusCode >= 400 {
			if attempt < client.maxRetries && slices.Contains(client.retryableCodes, resp.StatusCode) {
				continue
			}
			err = fmt.Errorf("method:%v url:%v code:%d body:%s", client.Method, client.Url, resp.StatusCode, string(respBody))
			return
		}

		body = respBody
		return
	}

	return
}

// isRetryableError 判断错误是否是瞬时网络错误，可进行重试
func (client *HttpClient) isRetryableError(err error) bool {
	var urlErr *url.Error
	if errors.As(err, &urlErr) {
		var netErr net.Error
		if errors.As(urlErr.Err, &netErr) {
			if netErr.Timeout() {
				return true
			}
		}
		if errors.Is(urlErr.Err, io.EOF) {
			return true
		}
		if errors.Is(urlErr.Err, net.ErrClosed) {
			return true
		}
		if opErr, ok := urlErr.Err.(*net.OpError); ok {
			if opErr.Op == "dial" {
				return true
			}
		}
	}
	return false
}
