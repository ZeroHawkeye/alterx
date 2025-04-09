package alterx

import (
	"os"
	"strings"

	_ "embed"

	"github.com/projectdiscovery/gologger"
	fileutil "github.com/projectdiscovery/utils/file"
	"gopkg.in/yaml.v3"
)

//go:embed permutations.yaml
var DefaultPermutationsBin []byte

// DefaultConfig 包含默认的模式和有效载荷
var DefaultConfig Config

// Config 定义了配置结构
type Config struct {
	Patterns []string            `yaml:"patterns"` // 置换模式列表
	Payloads map[string][]string `yaml:"payloads"` // 各种类型的有效载荷映射
}

// NewConfig 从文件中读取配置
func NewConfig(filePath string) (*Config, error) {
	bin, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	var cfg Config
	if err = yaml.Unmarshal(bin, &cfg); err != nil {
		return nil, err
	}

	var words []string
	for _, p := range cfg.Payloads["word"] {
		if !fileutil.FileExists(p) {
			words = append(words, p)
		} else {
			// 如果是文件路径，则读取文件中的所有单词
			wordBytes, err := os.ReadFile(p)
			if err != nil {
				gologger.Error().Msgf("failed to read wordlist from %v got %v", p, err)
				continue
			}
			words = append(words, strings.Fields(string(wordBytes))...)
		}
	}
	cfg.Payloads["word"] = words
	return &cfg, nil
}

// 初始化默认配置
func init() {
	if err := yaml.Unmarshal(DefaultPermutationsBin, &DefaultConfig); err != nil {
		gologger.Error().Msgf("default wordlist not found: got %v", err)
	}
}
