package Http

import (
	"crypto/tls"
	"golang.org/x/crypto/acme/autocert"
)

func GetTlsCert(domain, cacheDir, mail string) *tls.Config {
	m := autocert.Manager{
		Prompt:     autocert.AcceptTOS,             // 接受服务条款
		HostPolicy: autocert.HostWhitelist(domain), // 只允许指定域名
		Cache:      autocert.DirCache(cacheDir),    // 证书缓存目录
		Email:      mail,
	}
	return m.TLSConfig()
}
