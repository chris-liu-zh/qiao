/*
 * @Author: Chris.Liu
 * @Date: 2024-05-24 17:50:36
 * @LastEditors: Chris.Liu
 * @LastEditTime: 2024-09-10 15:49:21
 * @Description: 请填写简介
 */
package DB

import (
	"unicode"
)

/*
获取字段
*/
func getColumn(items []string) string {
	return items[0]
	// for _, v := range items {
	// 	if strings.Contains(v, "column:") {
	// 		c := strings.Split(v, ":")
	// 		return c[1]
	// 	}
	// }
	// return ""
}

/*
获取可写字段
*/
func WritableField(f []string) bool {
	for _, v := range f {
		if v == "Autoincrement" || v == "~>" || v == "~" {
			return false
		}
	}
	return true
}

/*
获取只读字段
*/
func ReadOnlyField(f []string) bool {
	for _, v := range f {
		if v == "<~" || v == "~" {
			return false
		}
	}
	return true
}

// 驼峰单词转下划线单词  "StudentName to student_name"
func CamelCaseToUdnderscore(s string) string {
	var output []rune
	for i, r := range s {
		if i == 0 {
			output = append(output, unicode.ToLower(r))
		} else {
			if unicode.IsUpper(r) {
				output = append(output, '_')
			}

			output = append(output, unicode.ToLower(r))
		}
	}
	return string(output)
}
