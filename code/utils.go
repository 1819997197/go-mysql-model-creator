package code

import (
	"fmt"
	"github.com/laixyz/utils"
)

// StringArrayAppend 往一个字符数组添加一个字符串，并保持唯一性
func StringArrayAppend(arr []string, str string) []string {
	for _, v := range arr {
		if v == str {
			return arr
		}
	}
	arr = append(arr, str)
	return arr
}

// StringArrayExists 判断一个字符串是否存在数组里
func StringArrayExists(arr []string, str string) bool {
	for _, v := range arr {
		if v == str {
			return true
		}
	}
	return false
}

// DebugInfo 调式输出函数
func DebugInfo(a ...interface{}) {
	if DebugMode {
		fmt.Println(a...)
	}
}

// Debug2Json 输出JSON
func Debug2Json(a interface{}) {
	if DebugMode {
		fmt.Println(utils.JSONEncode(a))
	}
}

// DebugPrintf 调式输出函数,带格式
func DebugPrintf(format string, a ...interface{}) {
	if DebugMode {
		fmt.Println(fmt.Sprintf(format, a...))
	}
}

// Panic 异常方法
func Panic(a ...interface{}) {
	panic(fmt.Sprint(a...))
}
