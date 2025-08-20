package qiao

import (
	"crypto/rand"
	"encoding/binary"
	"errors"
	"fmt"
	"time"
)

type UUIDVariant byte

// UUID变体常量定义
const (
	VariantNCS       UUIDVariant = iota // 保留用于向后兼容
	VariantRFC9562   UUIDVariant = iota // RFC 4122标准定义的UUID布局
	VariantMicrosoft UUIDVariant = iota // 保留用于Microsoft兼容
	VariantFuture    UUIDVariant = iota // 保留供未来使用
)

const VariantRFC4122 = VariantRFC9562

// UUID 表示一个UUID v7，遵循draft-peabody-dispatch-new-uuid-format-04草案规范
type UUID [16]byte

var (
	// ErrInvalidUUID 当解析无效UUID字符串时返回此错误
	ErrInvalidUUID = errors.New("invalid uuid format")
)

// UUIDV7 生成一个新的UUID v7
func UUIDV7() *UUID {
	var u UUID
	// 获取当前时间戳（Unix纪元以来的毫秒数）
	now := time.Now()
	unixMilli := now.UnixNano() / int64(time.Millisecond)
	// 设置版本号(7)和变体(RFC4122)
	// rand_a(12位) + 版本号(4位，值为7)
	versionAndRandA := 0x7000 | (uint16(randUint32()>>20) & 0x0FFF)
	// rand_b(16位)
	randB := uint16(randUint32() >> 16)
	// rand_c(62位)
	randC := [8]byte{}
	_, _ = rand.Read(randC[:])
	// 设置时间字段(48位)
	binary.BigEndian.PutUint32(u[0:4], uint32(unixMilli>>16)) // 时间戳高32位
	binary.BigEndian.PutUint16(u[4:6], uint16(unixMilli))     // 时间戳低16位
	// 设置rand_a(12位) + 版本号(4位)
	binary.BigEndian.PutUint16(u[6:8], versionAndRandA)
	// 设置rand_b(16位) + 变体(2位，值为10)
	binary.BigEndian.PutUint16(u[8:10], randB&0x3FFF|0x8000)
	// 设置rand_c(62位)
	copy(u[10:], randC[:])
	return &u
}

// Parse 将UUID字符串解析为UUID对象
func Parse(s string) (UUID, error) {
	var u UUID
	// 检查长度和分隔符
	if len(s) != 36 || s[8] != '-' || s[13] != '-' || s[18] != '-' || s[23] != '-' {
		return u, ErrInvalidUUID
	}
	// 解析时间字段
	timeLow, err := parseHex(s[0:8]) // 时间戳高32位
	if err != nil {
		return u, err
	}
	timeMid, err := parseHex(s[9:13]) // 时间戳低16位
	if err != nil {
		return u, err
	}
	binary.BigEndian.PutUint32(u[0:4], timeLow)
	binary.BigEndian.PutUint16(u[4:6], uint16(timeMid))
	// 解析版本号和rand_a
	versionAndRandA, err := parseHex(s[14:18])
	if err != nil {
		return u, err
	}
	binary.BigEndian.PutUint16(u[6:8], uint16(versionAndRandA))
	// 解析变体和rand_b
	variantAndRandB, err := parseHex(s[19:23])
	if err != nil {
		return u, err
	}
	binary.BigEndian.PutUint16(u[8:10], uint16(variantAndRandB))
	// 解析rand_c
	for i := 0; i < 6; i++ {
		byteVal, err := parseHex(s[24+2*i : 26+2*i])
		if err != nil {
			return u, err
		}
		u[10+i] = byte(byteVal)
	}
	return u, nil
}

// Time 返回UUID中嵌入的时间戳
func (u *UUID) Time() time.Time {
	timeLow := binary.BigEndian.Uint32(u[0:4])
	timeMid := binary.BigEndian.Uint16(u[4:6])
	unixMilli := int64(timeLow)<<16 | int64(timeMid)
	return time.Unix(unixMilli/1000, (unixMilli%1000)*int64(time.Millisecond))
}

// Version 返回UUID的版本号
func (u *UUID) Version() uint8 {
	return u[6] >> 4 // 版本号存储在字节6的高4位
}

// String 返回UUID的字符串表示形式
func (u *UUID) String() string {
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x",
		binary.BigEndian.Uint32(u[0:4]),  // 时间戳高32位
		binary.BigEndian.Uint16(u[4:6]),  // 时间戳低16位
		binary.BigEndian.Uint16(u[6:8]),  // rand_a + 版本号
		binary.BigEndian.Uint16(u[8:10]), // rand_b + 变体
		u[10:16])                         // rand_c
}

// Variant 返回UUID的变体类型
func (u *UUID) Variant() UUIDVariant {
	switch {
	case (u[8] >> 7) == 0x00:
		return VariantNCS
	case (u[8] >> 6) == 0x02:
		return VariantRFC9562
	case (u[8] >> 5) == 0x06:
		return VariantMicrosoft
	case (u[8] >> 5) == 0x07:
		fallthrough
	default:
		return VariantFuture
	}
}

func (u *UUID) SetVariant(v UUIDVariant) {
	switch v {
	case VariantNCS:
		u[8] = u[8]&(0xff>>1) | (0x00 << 7)
	case VariantRFC9562:
		u[8] = u[8]&(0xff>>2) | (0x02 << 6)
	case VariantMicrosoft:
		u[8] = u[8]&(0xff>>3) | (0x06 << 5)
	case VariantFuture:
		fallthrough
	default:
		u[8] = u[8]&(0xff>>3) | (0x07 << 5)
	}
}

// randUint32 生成一个随机的uint32
func randUint32() uint32 {
	b := [4]byte{}
	_, _ = rand.Read(b[:])
	return binary.BigEndian.Uint32(b[:])
}

// parseHex 将2字节或4字节的十六进制字符串解析为uint32
func parseHex(s string) (uint32, error) {
	var val uint32
	for i := 0; i < len(s); i++ {
		c := s[i]
		var nibble uint32
		switch {
		case '0' <= c && c <= '9':
			nibble = uint32(c - '0')
		case 'a' <= c && c <= 'f':
			nibble = uint32(c-'a') + 10
		case 'A' <= c && c <= 'F':
			nibble = uint32(c-'A') + 10
		default:
			return 0, ErrInvalidUUID
		}
		val = val<<4 | nibble
	}
	return val, nil
}
