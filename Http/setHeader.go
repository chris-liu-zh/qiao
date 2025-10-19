/*
 * @Author: Chris
 * @Date: 2024-07-12 15:38:38
 * @LastEditors: Chris
 * @LastEditTime: 2025-03-13 01:03:01
 * @Description: 请填写简介
 */
package Http

import (
	"net/http"
)

func DefaultHeader(w http.ResponseWriter) {
	// 允许访问所有域，可以换成具体url，注意仅具体url才能带cookie信息
	w.Header().Set("Access-Control-Allow-Origin", "*")
	//设置为true，允许ajax异步请求带cookie信息
	w.Header().Set("Access-Control-Allow-Credentials", "true")
	// //header的类型
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type,Authorization,Timestamp,Sign,Appkey")
	//允许请求方法
	w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
}
