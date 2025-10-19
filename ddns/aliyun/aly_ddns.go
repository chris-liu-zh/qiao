package aliyun

import (
	"errors"

	alidns20150109 "github.com/alibabacloud-go/alidns-20150109/v4/client"
	openapi "github.com/alibabacloud-go/darabonba-openapi/v2/client"
)

type ipResp struct {
	alyIp    string
	RR       *string
	RecordId *string
}

type api struct {
	Client     *alidns20150109.Client
	DomainName string
	ipResp     *ipResp
}

var endpoint = "alidns.cn-hangzhou.aliyuncs.com"
var types = "A"

// NewClient 创建发起请求的client
func NewAlyDdnsClient(accessKeyId, accessKeySecret, DomainName string) (a *api, err error) {
	config := &openapi.Config{
		AccessKeyId:     &accessKeyId,
		AccessKeySecret: &accessKeySecret,
	}

	config.Endpoint = &endpoint

	result, err := alidns20150109.NewClient(config)
	if err != nil {
		return
	}
	a = &api{
		Client:     result,
		DomainName: DomainName,
	}
	return
}

// 封装获取对应解析记录的方法
func (a *api) getRecordIp(rr *string) *ipResp {
	describeDomainRecordsRequest := &alidns20150109.DescribeDomainRecordsRequest{
		DomainName: &a.DomainName,
	}
	result, err := a.Client.DescribeDomainRecords(describeDomainRecordsRequest)
	if err != nil {
		return nil
	}
	records := result.Body.DomainRecords.Record
	ipresp := &ipResp{}
	for _, record := range records {
		if *record.RR == *rr {
			ipresp.alyIp = *record.Value
			ipresp.RR = rr
			ipresp.RecordId = record.RecordId
		}
	}

	return ipresp
}

// 执行比对当前ip和dns值，并更新操作
func (a *api) SetDdns(clinetip string, prefix string) (ip string, err error) {
	RR := prefix
	if clinetip == "" {
		err = errors.New("客户端IP为空")
		return
	}

	if prefix == "" {
		err = errors.New("客户端前缀为空")
		return
	}

	a.ipResp = a.getRecordIp(&RR)
	if a.ipResp.alyIp == "" {
		return a.addDdns(clinetip, prefix)
	} else {
		return a.upDdns(clinetip)
	}

}

// 执行比对当前ip和dns值，并更新操作
func (a *api) upDdns(clinetip string) (ip string, err error) {
	if a.ipResp.alyIp == clinetip {
		return clinetip, nil
	}
	updateDomainRecordRequest := &alidns20150109.UpdateDomainRecordRequest{
		RecordId: a.ipResp.RecordId,
		RR:       a.ipResp.RR,
		Type:     &types,
		Value:    &clinetip,
	}
	if _, err = a.Client.UpdateDomainRecord(updateDomainRecordRequest); err != nil {
		return
	}
	ipresp := a.getRecordIp(a.ipResp.RR)
	return ipresp.alyIp, nil
}

// 添加解析记录
func (a *api) addDdns(clinetip string, prefix string) (ip string, err error) {
	RR := prefix
	addDomainRecordRequest := &alidns20150109.AddDomainRecordRequest{
		DomainName: &a.DomainName,
		RR:         &RR,
		Type:       &types,
		Value:      &clinetip,
	}
	if _, err = a.Client.AddDomainRecord(addDomainRecordRequest); err != nil {
		return
	}
	ipresp := a.getRecordIp(&RR)
	return ipresp.alyIp, nil
}
