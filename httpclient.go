/*
 * @Author: Chris
 * @Date: 2022-12-07 15:28:22
 * @LastEditors: Chris
 * @LastEditTime: 2025-03-09 00:17:36
 * @Description: 请填写简介
 */
package qiao

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
)

type httpClient struct {
	req     *http.Request
	Url     string
	Method  string
	Cookies []*http.Cookie
	err     error
}

func NewHttpClient(url string) *httpClient {
	http := &httpClient{}
	http.Url = url
	return http
}

func (client *httpClient) Cookie(cookies []*http.Cookie) *httpClient {
	client.Cookies = cookies
	return client
}

func (client *httpClient) Post() *httpClient {
	client.Method = "POST"
	return client
}

func (client *httpClient) Get() *httpClient {
	if client.req, client.err = http.NewRequest("GET", client.Url, nil); client.err != nil {
		return client
	}
	return client
}

func (client *httpClient) Delete() *httpClient {
	client.Method = "DELETE"
	return client
}

func (client *httpClient) Put() *httpClient {
	client.Method = "PUT"
	return client
}

func (client *httpClient) DoJson(bodyJson []byte) *httpClient {
	reqBody := io.NopCloser(bytes.NewReader(bodyJson))
	if client.req, client.err = http.NewRequest(client.Method, client.Url, reqBody); client.err != nil {
		return client
	}
	for _, c := range client.Cookies {
		client.req.AddCookie(&http.Cookie{Name: c.Name, Value: c.Value, HttpOnly: c.HttpOnly})
	}
	return client
}

func (client *httpClient) HeaderAdd(header map[string]string) *httpClient {
	if client.err != nil {
		return client
	}
	for k, v := range header {
		client.req.Header.Add(k, v)
	}
	return client
}

// func (client *httpClient) DoBody(bodys map[string]string) *httpClient {
// 	payload := &bytes.Buffer{}
// 	writer := multipart.NewWriter(payload)
// 	for k, v := range bodys {
// 		_ = writer.WriteField(k, v)
// 	}
// 	if client.err = writer.Close(); client.err != nil {
// 		return client
// 	}
// 	if client.req, client.err = http.NewRequest(client.Method, client.Url, payload); client.err != nil {
// 		return client
// 	}
// 	for _, c := range client.Cookies {
// 		client.req.AddCookie(&http.Cookie{Name: c.Name, Value: c.Value, HttpOnly: c.HttpOnly})
// 	}
// 	client.req.Header.Add("X-Requested-With", "XMLHttpRequest")
// 	client.req.Header.Set("Content-Type", writer.FormDataContentType())
// 	return client
// }

func (client *httpClient) Respond() (body []byte, cookies []*http.Cookie, err error) {
	if client.err != nil {
		err = client.err
		return
	}
	tr := &http.Transport{
		// TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	c := &http.Client{Transport: tr}
	var resp *http.Response
	if resp, err = c.Do(client.req); err != nil {
		return
	}
	cookies = resp.Cookies()
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		err = fmt.Errorf("method:%v url: %v code:%d", client.Method, client.Url, resp.StatusCode)
		return
	}
	if body, err = io.ReadAll(resp.Body); err != nil {
		return
	}
	return
}
