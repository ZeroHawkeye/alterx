package alterx

import (
	"fmt"
	"regexp"
	"strings"
	"unsafe"
)

// 用于匹配变量的正则表达式模式，如 {{variable}}
var varRegex = regexp.MustCompile(`\{\{([a-zA-Z0-9]+)\}\}`)

// getVarCount 返回语句中存在的变量数量
func getVarCount(data string) int {
	return len(varRegex.FindAllStringSubmatch(data, -1))
}

// getAllVars 返回所有变量的名称
func getAllVars(data string) []string {
	var values []string
	for _, v := range varRegex.FindAllStringSubmatch(data, -1) {
		if len(v) >= 2 {
			values = append(values, v[1])
		}
	}
	return values
}

// getSampleMap 返回一个包含输入变量和有效载荷变量的示例映射
func getSampleMap(inputVars map[string]interface{}, payloadVars map[string][]string) map[string]interface{} {
	sMap := map[string]interface{}{}
	for k, v := range inputVars {
		sMap[k] = v
	}
	for k, v := range payloadVars {
		if k != "" && len(v) > 0 {
			sMap[k] = "temp"
		}
	}
	return sMap
}

// checkMissing 检查是否所有变量/占位符都已成功替换
// 如果没有，则会抛出带有描述的错误
func checkMissing(template string, data map[string]interface{}) error {
	got := Replace(template, data)
	if res := varRegex.FindAllString(got, -1); len(res) > 0 {
		return fmt.Errorf("values of `%v` variables not found", strings.Join(res, ","))
	}
	return nil
}

// unsafeToBytes 将字符串转换为字节切片，且零分配
//
// 参考 - https://stackoverflow.com/questions/59209493/how-to-use-unsafe-get-a-byte-slice-from-a-string-without-memory-copy
func unsafeToBytes(data string) []byte {
	return unsafe.Slice(unsafe.StringData(data), len(data))
}
