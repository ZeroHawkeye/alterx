package runner

import (
	"github.com/projectdiscovery/gologger"
	updateutils "github.com/projectdiscovery/utils/update"
)

// 程序的ASCII艺术横幅
var banner = `
   ___   ____          _  __
  / _ | / / /____ ____| |/_/
 / __ |/ / __/ -_) __/>  <  
/_/ |_/_/\__/\__/_/ /_/|_|  				 
`

// 当前版本号
var version = "v0.0.6"

// showBanner 用于向用户显示程序的横幅
// 这个函数在程序启动时调用，用于展示程序的标志和版本信息
func showBanner() {
	gologger.Print().Msgf("%s\n", banner)
	gologger.Print().Msgf("\t\tprojectdiscovery.io\n\n")
}

// GetUpdateCallback 返回一个回调函数，用于更新alterx程序
// 当用户使用-update标志时，此函数会被调用
func GetUpdateCallback() func() {
	return func() {
		showBanner()
		updateutils.GetUpdateToolCallback("alterx", version)()
	}
}
