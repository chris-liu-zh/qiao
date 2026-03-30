package Http

import (
	"context"
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"time"

	"golang.org/x/crypto/acme"
)

// DNSCertManager 实现基于DNS-01挑战的证书管理器
type DNSCertManager struct {
	client      *acme.Client
	dnsProvider DNSProvider
	email       string
	domains     []string
	cacheDir    string
	privateKey  crypto.PrivateKey
	renew       bool
}

type DNSProvider interface {
	Present(ctx context.Context, domain, token, keyAuth string) error
	CleanUp(ctx context.Context, domain, token, keyAuth string) error
}

// NewDNSCertManager 创建新的DNS证书管理器
func NewDNSCertManager(dnsProvider DNSProvider, email string, domains []string, cacheDir string) (*DNSCertManager, error) {
	// 生成私钥
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("生成私钥失败: %v", err)
	}

	// 创建ACME客户端
	client := &acme.Client{
		Key:          privateKey,
		DirectoryURL: acme.LetsEncryptURL,
		HTTPClient: &http.Client{
			Transport: &http.Transport{
				Proxy: http.ProxyFromEnvironment,
			},
			Timeout: 5 * time.Second,
		},
	}
	// 注册账户
	account := &acme.Account{
		Contact: []string{"mailto:" + email},
	}
	if _, err := client.Register(context.Background(), account, acme.AcceptTOS); err != nil {
		return nil, fmt.Errorf("账户注册失败: %v", err)
	}
	return &DNSCertManager{
		client:      client,
		dnsProvider: dnsProvider,
		email:       email,
		domains:     domains,
		cacheDir:    cacheDir,
		privateKey:  privateKey,
	}, nil
}

// GetCertificate 实现tls.Config的GetCertificate方法
func (m *DNSCertManager) GetCertificate(hello *tls.ClientHelloInfo) (*tls.Certificate, error) {
	// 检查域名是否在允许列表中
	allowed := slices.Contains(m.domains, hello.ServerName)
	if !allowed {
		return nil, fmt.Errorf("域名 %s 不在允许列表中", hello.ServerName)
	}

	// 首先尝试从缓存加载证书
	if !m.renew {
		if cachedCert, err := m.loadCertificate(hello.ServerName); err == nil {
			return cachedCert, nil
		}
	}

	// 为证书生成独立的密钥对（不能使用账户密钥）
	certPrivateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("生成证书密钥对失败: %v", err)
	}

	// 创建证书请求
	template := &x509.CertificateRequest{
		Subject:  pkix.Name{CommonName: hello.ServerName},
		DNSNames: []string{hello.ServerName},
	}

	// 生成证书签名请求（使用证书的独立密钥）
	csrDER, err := x509.CreateCertificateRequest(rand.Reader, template, certPrivateKey)
	if err != nil {
		return nil, fmt.Errorf("创建CSR失败: %v", err)
	}

	_, err = x509.ParseCertificateRequest(csrDER)
	if err != nil {
		return nil, fmt.Errorf("解析CSR失败: %v", err)
	}

	// 使用新版本的ACME库API直接获取证书
	// 生成DNS-01挑战的keyAuth
	// 首先获取授权信息来获取正确的挑战token
	order, err := m.client.AuthorizeOrder(context.Background(), []acme.AuthzID{
		{Type: "dns", Value: hello.ServerName},
	})
	if err != nil {
		return nil, fmt.Errorf("创建订单失败: %v", err)
	}

	// 获取授权信息
	authz, err := m.client.GetAuthorization(context.Background(), order.AuthzURLs[0])
	if err != nil {
		return nil, fmt.Errorf("获取授权失败: %v", err)
	}

	// 找到DNS-01挑战
	var challenge *acme.Challenge
	for _, c := range authz.Challenges {
		if c.Type == "dns-01" {
			challenge = c
			break
		}
	}

	if challenge == nil {
		return nil, fmt.Errorf("未找到DNS-01挑战")
	}

	// 使用正确的token生成keyAuth
	keyAuth, err := m.client.DNS01ChallengeRecord(challenge.Token)
	if err != nil {
		return nil, fmt.Errorf("生成挑战记录失败: %v", err)
	}

	// 呈现DNS记录
	if err = m.dnsProvider.Present(context.Background(), hello.ServerName, "", keyAuth); err != nil {
		return nil, fmt.Errorf("呈现DNS记录失败: %v", err)
	}

	// 等待DNS传播
	log.Printf("等待DNS记录传播,等待10秒...")
	time.Sleep(10 * time.Second) // 增加等待时间，确保DNS完全传播

	// 接受DNS-01挑战...
	if _, err = m.client.Accept(context.Background(), challenge); err != nil {
		m.dnsProvider.CleanUp(context.Background(), hello.ServerName, "", keyAuth)
		return nil, fmt.Errorf("接受挑战失败: %v", err)
	}
	// 等待挑战完成
	for range 30 { // 最多等待5分钟
		authz, err = m.client.GetAuthorization(context.Background(), order.AuthzURLs[0])
		if err != nil {
			log.Printf("获取授权状态失败: %v", err)
			time.Sleep(10 * time.Second)
			continue
		}
		if authz.Status == "valid" {
			log.Printf("DNS-01挑战验证成功!")
			break
		} else if authz.Status == "invalid" {
			m.dnsProvider.CleanUp(context.Background(), hello.ServerName, "", keyAuth)
			return nil, fmt.Errorf("DNS-01挑战验证失败")
		}
		log.Printf("挑战状态: %s，等待10秒后重试...", authz.Status)
		time.Sleep(10 * time.Second)
	}

	if authz.Status != "valid" {
		m.dnsProvider.CleanUp(context.Background(), hello.ServerName, "", keyAuth)
		return nil, fmt.Errorf("DNS-01挑战验证超时，最终状态: %s", authz.Status)
	}

	// 等待订单状态变为ready
	for range 12 { // 最多等待2分钟
		order, err = m.client.GetOrder(context.Background(), order.URI)
		if err != nil {
			log.Printf("获取订单状态失败: %v", err)
			time.Sleep(10 * time.Second)
			continue
		}

		if order.Status == "ready" {
			log.Printf("订单状态已变为ready，可以获取证书")
			break
		} else if order.Status == "invalid" {
			m.dnsProvider.CleanUp(context.Background(), hello.ServerName, "", keyAuth)
			return nil, fmt.Errorf("订单无效")
		}

		log.Printf("订单状态: %s，等待10秒后重试...", order.Status)
		time.Sleep(10 * time.Second)
	}

	if order.Status != "ready" {
		m.dnsProvider.CleanUp(context.Background(), hello.ServerName, "", keyAuth)
		return nil, fmt.Errorf("订单状态未变为ready，最终状态: %s", order.Status)
	}

	// 获取证书
	certs, _, err := m.client.CreateOrderCert(context.Background(), order.FinalizeURL, csrDER, true)
	if err != nil {
		m.dnsProvider.CleanUp(context.Background(), hello.ServerName, "", keyAuth)
		return nil, fmt.Errorf("创建证书失败: %v", err)
	}

	// 清理DNS记录
	if err := m.dnsProvider.CleanUp(context.Background(), hello.ServerName, "", keyAuth); err != nil {
		log.Printf("清理DNS记录警告: %v", err)
	}

	// 构建证书（使用证书的独立密钥）
	cert := &tls.Certificate{
		Certificate: certs,
		PrivateKey:  certPrivateKey,
		Leaf:        &x509.Certificate{}, // 简化实现
	}

	// 保存证书到缓存
	if err := m.saveCertificate(hello.ServerName, cert); err != nil {
		return nil, fmt.Errorf("保存证书到缓存失败: %v", err)
	}
	return cert, nil
}

// GetCertificateFunc 返回GetCertificate函数，用于tls.Config
func (m *DNSCertManager) GetCertificateFunc() func(*tls.ClientHelloInfo) (*tls.Certificate, error) {
	return m.GetCertificate
}

// 保存证书到缓存
func (m *DNSCertManager) saveCertificate(domain string, cert *tls.Certificate) error {
	// 创建证书目录
	if err := os.MkdirAll(m.cacheDir, 0700); err != nil {
		return fmt.Errorf("创建缓存目录失败: %v", err)
	}

	// 保存证书文件
	certFile := filepath.Join(m.cacheDir, domain+".crt")
	keyFile := filepath.Join(m.cacheDir, domain+".key")

	// 保存证书链
	certData := []byte{}
	for _, certBytes := range cert.Certificate {
		block := &pem.Block{
			Type:  "CERTIFICATE",
			Bytes: certBytes,
		}
		certData = append(certData, pem.EncodeToMemory(block)...)
	}

	if err := os.WriteFile(certFile, certData, 0600); err != nil {
		return fmt.Errorf("保存证书文件失败: %v", err)
	}

	// 保存私钥（使用证书的独立密钥）
	keyData, err := x509.MarshalECPrivateKey(cert.PrivateKey.(*ecdsa.PrivateKey))
	if err != nil {
		return fmt.Errorf("序列化私钥失败: %v", err)
	}

	keyBlock := &pem.Block{
		Type:  "EC PRIVATE KEY",
		Bytes: keyData,
	}

	return os.WriteFile(keyFile, pem.EncodeToMemory(keyBlock), 0600)
}

func (m *DNSCertManager) CheckCertExists(domain string) error {
	certFile := filepath.Join(m.cacheDir, domain+".crt")
	keyFile := filepath.Join(m.cacheDir, domain+".key")

	// 检查证书文件是否存在
	if _, err := os.Stat(certFile); os.IsNotExist(err) {
		return fmt.Errorf("证书文件不存在")
	}
	if _, err := os.Stat(keyFile); os.IsNotExist(err) {
		return fmt.Errorf("私钥文件不存在")
	}
	return nil
}

// 从缓存加载证书
func (m *DNSCertManager) loadCertificate(domain string) (*tls.Certificate, error) {
	certFile := filepath.Join(m.cacheDir, domain+".crt")
	keyFile := filepath.Join(m.cacheDir, domain+".key")

	if _, err := os.Stat(certFile); os.IsNotExist(err) {
		return nil, fmt.Errorf("证书文件不存在")
	}
	if _, err := os.Stat(keyFile); os.IsNotExist(err) {
		return nil, fmt.Errorf("私钥文件不存在")
	}

	// 加载证书和私钥
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return nil, fmt.Errorf("加载证书失败: %v", err)
	}

	// 解析证书以获取叶子证书
	if len(cert.Certificate) > 0 {
		leaf, err := x509.ParseCertificate(cert.Certificate[0])
		if err == nil {
			cert.Leaf = leaf
		}
	}
	return &cert, nil
}

// CheckCertificateExpiry 检查证书到期时间，返回剩余天数
func (m *DNSCertManager) CheckCertificateExpiry(domain string) (int, error) {
	cert, err := m.loadCertificate(domain)
	if err != nil {
		return -1, fmt.Errorf("加载证书失败: %v", err)
	}

	if cert.Leaf == nil {
		return -1, fmt.Errorf("无法获取证书信息")
	}

	// 计算剩余天数
	remaining := time.Until(cert.Leaf.NotAfter).Hours() / 24
	return int(remaining), nil
}

// IsCertificateExpiringSoon 检查证书是否即将到期（15天内）
func (m *DNSCertManager) IsCertificateExpiringSoon(domain string) (bool, error) {
	daysLeft, err := m.CheckCertificateExpiry(domain)
	if err != nil {
		return false, err
	}

	// 如果证书不存在或已过期，返回true
	if daysLeft <= 0 {
		return true, nil
	}

	// 如果剩余天数小于等于15天，返回true
	return daysLeft <= 15, nil
}

// RenewCertificate 续期证书
func (m *DNSCertManager) RenewCertificate(domain string) error {
	log.Printf("开始续期证书: %s", domain)

	// 模拟一个假的ClientHelloInfo来触发证书获取
	hello := &tls.ClientHelloInfo{
		ServerName: domain,
	}

	// 调用GetCertificate方法获取新证书
	m.renew = true
	_, err := m.GetCertificate(hello)
	if err != nil {
		return fmt.Errorf("续期证书失败: %v", err)
	}
	m.renew = false
	log.Printf("证书续期成功: %s", domain)
	return nil
}

// CheckAndRenewCertificates 检查所有域名的证书并在需要时续期
func (m *DNSCertManager) CheckAndRenewCertificates(lessDayRenew int) error {
	log.Println("开始检查证书到期状态...")

	for _, domain := range m.domains {
		// 如果证书文件不存在，尝试续期
		if err := m.CheckCertExists(domain); err != nil {
			return m.RenewCertificate(domain)
		}
		expiring, err := m.IsCertificateExpiringSoon(domain)
		if err != nil {
			log.Printf("检查证书 %s 到期状态失败: %v", domain, err)
			return err
		}

		daysLeft, err := m.CheckCertificateExpiry(domain)
		if err != nil {
			log.Printf("检查证书 %s 剩余天数失败: %v", domain, err)
			return err
		} else {
			log.Printf("证书 %s 状态正常 (剩余 %d 天)", domain, daysLeft)
		}

		if expiring || daysLeft <= lessDayRenew {
			if err := m.RenewCertificate(domain); err != nil {
				log.Printf("续期证书 %s 失败: %v", domain, err)
				return err
			}
			log.Printf("证书 %s 续期成功", domain)
		}
	}
	log.Println("证书检查完成")
	return nil
}

// startCertificateExpiryMonitor 启动证书到期检测监控协程
func (m *DNSCertManager) StartCertificateExpiryMonitor(checkTime string, lessDayRenew int) {
	log.Printf("📅 启动证书到期检测监控 (检查时间: %s)...", checkTime)

	// 解析检查时间
	hourStr := checkTime[0:2]
	minuteStr := checkTime[3:5]
	hour, _ := strconv.Atoi(hourStr)
	minute, _ := strconv.Atoi(minuteStr)

	// 立即执行一次证书检查
	log.Println("执行首次证书检查...")
	if err := m.CheckAndRenewCertificates(lessDayRenew); err != nil {
		log.Printf("首次证书检查失败: %v", err)
	}

	// 计算到下一个指定检查时间的时间
	now := time.Now()
	nextCheckTime := time.Date(now.Year(), now.Month(), now.Day(), hour, minute, 0, 0, now.Location())

	// 如果今天的时间已经过了，就设置到明天
	if now.After(nextCheckTime) {
		nextCheckTime = nextCheckTime.Add(24 * time.Hour)
	}

	durationUntilNextCheck := nextCheckTime.Sub(now)

	log.Printf("下次证书检查将在 %s 后执行 (%s)", durationUntilNextCheck, nextCheckTime.Format("2006-01-02 15:04:05"))

	// 等待到指定检查时间
	time.Sleep(durationUntilNextCheck)

	// 启动定时器，每天指定时间执行
	ticker := time.NewTicker(24 * time.Hour)
	defer ticker.Stop()

	for range ticker.C {
		log.Println("🔄 执行每日证书到期检查...")
		if err := m.CheckAndRenewCertificates(lessDayRenew); err != nil {
			log.Printf("证书检查失败: %v", err)
		}
	}
}
