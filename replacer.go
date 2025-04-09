package alterx

import (
	"fmt"

	"github.com/projectdiscovery/fasttemplate"
)

const (
	// General 通用标记（开/关）
	General = "§"
	// ParenthesisOpen 开始占位符的标记
	ParenthesisOpen = "{{"
	// ParenthesisClose 结束占位符的标记
	ParenthesisClose = "}}"
)

// Replace 实时替换模板中的占位符。
// 该函数接收一个模板字符串和一个值映射，将模板中的变量替换为对应的值
func Replace(template string, values map[string]interface{}) string {
	valuesMap := make(map[string]interface{}, len(values))
	for k, v := range values {
		valuesMap[k] = fmt.Sprint(v)
	}
	// 首先替换 {{变量}} 格式的占位符
	replaced := fasttemplate.ExecuteStringStd(template, ParenthesisOpen, ParenthesisClose, valuesMap)
	// 然后替换 §变量§ 格式的占位符
	final := fasttemplate.ExecuteStringStd(replaced, General, General, valuesMap)
	return final
}
