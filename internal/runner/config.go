package runner

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/projectdiscovery/alterx"
	"github.com/projectdiscovery/gologger"
	fileutil "github.com/projectdiscovery/utils/file"
	"gopkg.in/yaml.v3"
)

// getUserHomeDir 获取用户的主目录路径
// 此函数在不同操作系统上都能正确获取用户的主目录
func getUserHomeDir() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}
	return homeDir
}

// init 函数在包被导入时自动执行，用于初始化配置
// Go语言的特殊函数，无需显式调用，程序启动时会自动执行
func init() {
	// 构建默认置换配置文件的路径
	defaultPermutationCfg := filepath.Join(getUserHomeDir(), fmt.Sprintf(".config/alterx/permutation_%v.yaml", version))

	// 检查默认配置文件是否存在
	if fileutil.FileExists(defaultPermutationCfg) {
		// 如果配置文件已存在，则读取其内容作为默认配置
		if bin, err := os.ReadFile(defaultPermutationCfg); err == nil {
			var cfg alterx.Config
			if errx := yaml.Unmarshal(bin, &cfg); errx == nil {
				// 将读取的配置设置为全局默认配置
				alterx.DefaultConfig = cfg
				return
			}
		}
	}

	// 如果配置文件不存在或读取失败，则创建配置目录
	if err := validateDir(filepath.Join(getUserHomeDir(), ".config/alterx")); err != nil {
		gologger.Error().Msgf("alterx config dir not found and failed to create got: %v", err)
	}

	// 将内置的默认配置写入到配置文件
	// 0600权限表示文件所有者可读写，其他用户无权限
	if err := os.WriteFile(defaultPermutationCfg, alterx.DefaultPermutationsBin, 0600); err != nil {
		gologger.Error().Msgf("failed to save default config to %v got: %v", defaultPermutationCfg, err)
	}
}

// validateDir 检查目录是否存在，如果不存在则创建它
// 这是一个辅助函数，用于确保配置目录存在
func validateDir(dirPath string) error {
	if fileutil.FolderExists(dirPath) {
		return nil
	}
	return fileutil.CreateFolder(dirPath)
}
