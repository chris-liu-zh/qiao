package aliyun

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/aliyun/alibaba-cloud-sdk-go/services/alidns"
)

// AliyunDNSProvider 实现DNS-01验证的阿里云DNS提供者
type AliyunDNSProvider struct {
	client *alidns.Client
}

// NewAliyunDNSProvider 创建新的阿里云DNS提供者实例
func NewAliyunDNSProvider(accessKeyID, accessKeySecret, regionID string) (*AliyunDNSProvider, error) {
	client, err := alidns.NewClientWithAccessKey(regionID, accessKeyID, accessKeySecret)
	if err != nil {
		return nil, fmt.Errorf("创建阿里云客户端失败: %v", err)
	}

	return &AliyunDNSProvider{
		client: client,
	}, nil
}

// Present 实现DNS-01验证的Present方法，添加TXT记录
func (p *AliyunDNSProvider) Present(domain, token, keyAuth string) error {
	// 手动构建DNS-01挑战的FQDN和值
	fqdn := "_acme-challenge." + domain
	// 对于DNS-01挑战，keyAuth需要编码为TXT记录的值
	// 通常keyAuth已经是正确的格式，但我们需要确保它被正确使用
	value := keyAuth

	fmt.Printf("开始DNS-01挑战: domain=%s, fqdn=%s, token=%s\n", domain, fqdn, token)

	// 获取域名
	domainName := extractDomain(fqdn)
	fmt.Printf("提取的域名: %s\n", domainName)

	// 检查是否已存在TXT记录
	records, err := p.getTXTRecords(domainName, fqdn)
	if err != nil {
		fmt.Printf("获取TXT记录失败: %v\n", err)
		return fmt.Errorf("获取TXT记录失败: %v", err)
	}

	fmt.Printf("找到 %d 条现有TXT记录\n", len(records))

	// 如果记录已存在，先删除
	if len(records) > 0 {
		fmt.Printf("删除现有记录...\n")
		for _, record := range records {
			fmt.Printf("删除记录ID: %s, RR: %s, Value: %s\n", record.RecordId, record.RR, record.Value)
			if err = p.deleteRecord(record.RecordId); err != nil {
				fmt.Printf("删除记录失败: %v\n", err)
				return fmt.Errorf("删除现有记录失败: %v", err)
			}
			fmt.Printf("成功删除记录ID: %s\n", record.RecordId)
		}
		// 等待一小段时间让删除操作生效
		fmt.Printf("等待删除操作生效...\n")
		time.Sleep(5 * time.Second)
	}

	// 添加新的TXT记录
	request := alidns.CreateAddDomainRecordRequest()
	request.Scheme = "https"
	request.DomainName = domainName
	request.RR = extractSubdomain(fqdn, domainName)
	request.Type = "TXT"
	request.Value = value
	request.TTL = "600" // 10分钟TTL

	fmt.Printf("添加TXT记录: RR=%s, Type=%s, Value=%s\n", request.RR, request.Type, request.Value)

	response, err := p.client.AddDomainRecord(request)
	if err != nil {
		fmt.Printf("添加TXT记录API调用失败: %v\n", err)
		return fmt.Errorf("添加TXT记录失败: %v", err)
	}

	fmt.Printf("成功添加TXT记录: %s -> %s (记录ID: %s)\n", fqdn, value, response.RecordId)
	return nil
}

// CleanUp 实现DNS-01验证的CleanUp方法，清理TXT记录
func (p *AliyunDNSProvider) CleanUp(domain, token, keyAuth string) error {
	// 手动构建DNS-01挑战的FQDN
	fqdn := "_acme-challenge." + domain

	fmt.Printf("清理DNS-01挑战记录: domain=%s, fqdn=%s\n", domain, fqdn)

	domainName := extractDomain(fqdn)
	records, err := p.getTXTRecords(domainName, fqdn)
	if err != nil {
		fmt.Printf("获取TXT记录失败: %v\n", err)
		return fmt.Errorf("获取TXT记录失败: %v", err)
	}

	fmt.Printf("找到 %d 条需要清理的TXT记录\n", len(records))

	for _, record := range records {
		fmt.Printf("删除记录ID: %s\n", record.RecordId)
		if err := p.deleteRecord(record.RecordId); err != nil {
			fmt.Printf("删除记录失败: %v\n", err)
			return fmt.Errorf("删除记录失败: %v", err)
		}
		fmt.Printf("成功删除TXT记录: %s (记录ID: %s)\n", fqdn, record.RecordId)
	}

	return nil
}

// getTXTRecords 获取指定域名的TXT记录
func (p *AliyunDNSProvider) getTXTRecords(domainName, fqdn string) ([]alidns.Record, error) {
	request := alidns.CreateDescribeDomainRecordsRequest()
	request.Scheme = "https"
	request.DomainName = domainName
	request.TypeKeyWord = "TXT"
	request.SearchMode = "EXACT"

	response, err := p.client.DescribeDomainRecords(request)
	if err != nil {
		return nil, fmt.Errorf("查询域名记录失败: %v", err)
	}

	var matchingRecords []alidns.Record
	expectedRR := extractSubdomain(fqdn, domainName)

	for _, record := range response.DomainRecords.Record {
		if record.RR == expectedRR && record.Type == "TXT" {
			matchingRecords = append(matchingRecords, record)
		}
	}

	return matchingRecords, nil
}

// deleteRecord 删除指定的DNS记录
func (p *AliyunDNSProvider) deleteRecord(recordID string) error {
	request := alidns.CreateDeleteDomainRecordRequest()
	request.Scheme = "https"
	request.RecordId = recordID

	_, err := p.client.DeleteDomainRecord(request)
	if err != nil {
		return fmt.Errorf("删除记录失败: %v", err)
	}

	return nil
}

// extractDomain 从FQDN中提取域名
func extractDomain(fqdn string) string {
	parts := strings.Split(fqdn, ".")
	if len(parts) >= 2 {
		return strings.Join(parts[len(parts)-2:], ".")
	}
	return fqdn
}

// extractSubdomain 从FQDN中提取子域名部分
func extractSubdomain(fqdn, domain string) string {
	if strings.HasSuffix(fqdn, "."+domain) {
		subdomain := strings.TrimSuffix(fqdn, "."+domain)
		// 如果子域名以点结尾，去掉点
		return strings.TrimSuffix(subdomain, ".")
	}
	// 如果无法提取，返回默认值
	return "_acme-challenge"
}

// DNS01ChallengeProvider 实现acme.ChallengeProvider接口
type DNS01ChallengeProvider struct {
	provider *AliyunDNSProvider
}

// NewDNS01ChallengeProvider 创建新的DNS-01挑战提供者
func NewDNS01ChallengeProvider(accessKeyID, accessKeySecret, regionID string) (*DNS01ChallengeProvider, error) {
	provider, err := NewAliyunDNSProvider(accessKeyID, accessKeySecret, regionID)
	if err != nil {
		return nil, err
	}

	return &DNS01ChallengeProvider{
		provider: provider,
	}, nil
}

// Present 实现acme.ChallengeProvider接口
func (p *DNS01ChallengeProvider) Present(ctx context.Context, domain, token, keyAuth string) error {
	return p.provider.Present(domain, token, keyAuth)
}

// CleanUp 实现acme.ChallengeProvider接口
func (p *DNS01ChallengeProvider) CleanUp(ctx context.Context, domain, token, keyAuth string) error {
	return p.provider.CleanUp(domain, token, keyAuth)
}
