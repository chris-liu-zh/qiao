package Http

import (
	"crypto/tls"

	"golang.org/x/crypto/acme/autocert"
)

func DNS01Challenge(dnsProvider DNSProvider, domains []string, cacheDir, mail, checkTime string) (*tls.Config, error) {
	DCM, err := NewDNSCertManager(dnsProvider, mail, domains, cacheDir)
	if err != nil {
		return nil, err
	}
	if checkTime != "" {
		go DCM.StartCertificateExpiryMonitor(checkTime)
	}
	// 配置TLS使用DNS证书管理器
	tlsConfig := &tls.Config{
		GetCertificate: DCM.GetCertificateFunc(),
	}
	return tlsConfig, nil
}

// HTTP01 使用标准的HTTP-01认证
func HTTP01Challenge(domains []string, cacheDir, mail string) *tls.Config {
	manager := &autocert.Manager{
		Prompt:     autocert.AcceptTOS,
		Email:      mail,
		Cache:      autocert.DirCache(cacheDir),
		HostPolicy: autocert.HostWhitelist(domains...),
	}
	return manager.TLSConfig()
}
