package qiao

import (
	"log"
	"os"
	"strings"
	"testing"

	"github.com/chris-liu-zh/qiao/Http"
	"github.com/chris-liu-zh/qiao/ddns/aliyun"
)

// Config 配置结构体
type Config struct {
	Domains               []string
	Email                 string
	CacheDir              string
	HTTPSPort             string
	HTTPPort              string
	AliyunAccessKeyID     string
	AliyunAccessKeySecret string
	AliyunRegionID        string
	ChallengeType         string // 认证类型："http-01" 或 "dns-01"
}

// loadConfig 从环境变量加载配置
func loadConfig() *Config {
	config := &Config{
		Domains:               strings.Split(getEnv("DOMAINS", "memory.zhjyxh.com"), ","),
		Email:                 getEnv("EMAIL", "zhxsglhq@hotmail.com"),
		CacheDir:              getEnv("CACHE_DIR", "./certs"),
		HTTPSPort:             getEnv("HTTPS_PORT", "8443"),
		HTTPPort:              getEnv("HTTP_PORT", "8080"),
		AliyunAccessKeyID:     getEnv("ALIYUN_ACCESS_KEY_ID", ""),
		AliyunAccessKeySecret: getEnv("ALIYUN_ACCESS_KEY_SECRET", ""),
		AliyunRegionID:        getEnv("ALIYUN_REGION_ID", "cn-hangzhou"),
		ChallengeType:         getEnv("CHALLENGE_TYPE", "dns-01"), // 默认使用HTTP-01认证
	}

	// 验证必要配置
	if len(config.Domains) == 0 || config.Domains[0] == "" {
		log.Fatal("DOMAINS环境变量必须设置")
	}
	if config.Email == "" {
		log.Fatal("EMAIL环境变量必须设置")
	}

	// 验证DNS-01认证所需配置
	if config.ChallengeType == "dns-01" {
		if config.AliyunAccessKeyID == "" || config.AliyunAccessKeySecret == "" {
			log.Fatal("使用DNS-01认证时，ALIYUN_ACCESS_KEY_ID和ALIYUN_ACCESS_KEY_SECRET环境变量必须设置")
		}
	}

	return config
}

// getEnv 获取环境变量，如果不存在则返回默认值
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func Test_Https(t *testing.T) {
	config := loadConfig()
	dnsProvider, err := aliyun.NewDNS01ChallengeProvider(
		config.AliyunAccessKeyID,
		config.AliyunAccessKeySecret,
		config.AliyunRegionID,
	)
	if err != nil {
		return
	}
	r := Http.NewRouter()
	r.Get("/version", GetVersion)
	tlsconfig, err := Http.DNS01Challenge(dnsProvider, config.Domains, config.CacheDir, config.Email, "02:00", 20)
	if err != nil {
		t.Fatalf("创建TLS配置失败: %v", err)
	}
	if err = Http.NewHttpServer(":8083", r).StartTLS(tlsconfig); err == nil {
		log.Println("HTTPS服务器已启动，监听端口8083")
	}
	t.Fatalf("启动HTTPS服务器失败: %v", err)
}
