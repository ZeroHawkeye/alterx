package alterx

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/projectdiscovery/gologger"
	urlutil "github.com/projectdiscovery/utils/url"
	"golang.org/x/net/publicsuffix"
)

// Input 包含URL的解析/评估数据
type Input struct {
	TLD        string   // 顶级域名（子域名的最右部分），例如：`.uk`
	ETLD       string   // 公共后缀，例如：co.uk
	SLD        string   // 二级域名，例如：scanme
	Root       string   // 根域名（eTLD+1）
	Sub        string   // 子域名或子域名的最左前缀
	Suffix     string   // 后缀是除了`Sub`之外的所有内容（注意：如果域名不是多级的，Suffix==Root）
	MultiLevel []string // （可选）存储多级子域名的前缀
}

// GetMap 返回输入的变量映射
func (i *Input) GetMap() map[string]interface{} {
	m := map[string]interface{}{
		"tld":    i.TLD,
		"etld":   i.ETLD,
		"sld":    i.SLD,
		"root":   i.Root,
		"sub":    i.Sub,
		"suffix": i.Suffix,
	}
	// 添加多级子域名到映射
	for k, v := range i.MultiLevel {
		m["sub"+strconv.Itoa(k+1)] = v
	}
	// 清除空值
	for k, v := range m {
		if v == "" {
			// 清除空变量
			delete(m, k)
		}
	}
	return m
}

// NewInput 将URL解析为Input变量
func NewInput(inputURL string) (*Input, error) {
	URL, err := urlutil.Parse(inputURL)
	if err != nil {
		return nil, err
	}
	// 检查主机名是否包含*
	if strings.Contains(URL.Hostname(), "*") {
		if strings.HasPrefix(URL.Hostname(), "*.") {
			tmp := strings.TrimPrefix(URL.Hostname(), "*.")
			URL.Host = strings.Replace(URL.Host, URL.Hostname(), tmp, 1)
		}
		// 如果*出现在中间，例如：prod.*.hackerone.com
		// 跳过它
		if strings.Contains(URL.Hostname(), "*") {
			return nil, fmt.Errorf("input %v is not a valid url , skipping", inputURL)
		}
	}
	ivar := &Input{}
	// 获取公共后缀
	suffix, _ := publicsuffix.PublicSuffix(URL.Hostname())
	if strings.Contains(suffix, ".") {
		ivar.ETLD = suffix
		arr := strings.Split(suffix, ".")
		ivar.TLD = arr[len(arr)-1]
	} else {
		ivar.TLD = suffix
	}
	// 获取有效顶级域名加一（eTLD+1）
	rootDomain, err := publicsuffix.EffectiveTLDPlusOne(URL.Hostname())
	if err != nil {
		// 如果输入域名根本没有eTLD+1，则会发生这种情况，例如：`.com`或`co.uk`
		gologger.Warning().Msgf("input domain %v is eTLD/publicsuffix and not a valid domain name", URL.Hostname())
		return ivar, nil
	}
	ivar.Root = rootDomain
	// 计算二级域名
	if ivar.ETLD != "" {
		ivar.SLD = strings.TrimSuffix(rootDomain, "."+ivar.ETLD)
	} else {
		ivar.SLD = strings.TrimSuffix(rootDomain, "."+ivar.TLD)
	}
	// 根域名之前的任何内容都是子域名
	subdomainPrefix := strings.TrimSuffix(URL.Hostname(), rootDomain)
	subdomainPrefix = strings.TrimSuffix(subdomainPrefix, ".")
	if strings.Contains(subdomainPrefix, ".") {
		// 这是一个多级子域名
		// 例如：something.level.scanme.sh
		// 在这种情况下，变量名从第一个前缀之后开始
		prefixes := strings.Split(subdomainPrefix, ".")
		ivar.Sub = prefixes[0]
		ivar.MultiLevel = prefixes[1:]
	} else {
		ivar.Sub = subdomainPrefix
	}
	ivar.Suffix = strings.TrimPrefix(URL.Hostname(), ivar.Sub+".")
	return ivar, nil
}
