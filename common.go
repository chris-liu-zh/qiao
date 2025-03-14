/*
 * @Author: Chris
 * @Date: 2022-11-27 20:42:44
 * @LastEditors: Chris
 * @LastEditTime: 2025-03-09 00:20:20
 * @Description: 请填写简介
 */
package qiao

import (
	"io"
	"os"
)

// var lock sync.RWMutex

func CheckFileIsExist(filename string) bool {
	var exist = true
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		exist = false
	}
	return exist
}

func ReadFile(filePath string) (chunk []byte, err error) {
	f, err := os.Open(filePath)
	if err != nil {
		return
	}
	defer f.Close()
	buf := make([]byte, 1024)

	for {
		var n int
		if n, err = f.Read(buf); err != nil && err != io.EOF {
			return
		}
		if n == 0 {
			break
		}
		chunk = append(chunk, buf[:n]...)
	}
	return chunk, nil
}

/**
 * @description: 截取文字
 * @param {string} str
 * @param {int} l
 * @return {*}
 */
func Intercept(str string, l int) string {
	if len([]rune(str)) < l {
		return str
	}
	s := string([]rune(str)[:l])
	return s
}

func GetMaxNum(ary []int64) int64 {
	if len(ary) == 0 {
		return 0
	}

	maxVal := ary[0]
	for i := range ary {
		if maxVal < ary[i] {
			maxVal = ary[i]
		}
	}

	return maxVal
}
