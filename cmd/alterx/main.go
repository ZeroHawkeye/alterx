package main

import (
	"io"
	"os"

	"github.com/projectdiscovery/alterx"
	"github.com/projectdiscovery/alterx/internal/runner"
	"github.com/projectdiscovery/gologger"
)

func main() {
	// 解析命令行参数
	cliOpts := runner.ParseFlags()

	// 创建alterx配置选项
	alterOpts := alterx.Options{
		Domains:  cliOpts.Domains,  // 要处理的域名列表
		Patterns: cliOpts.Patterns, // 自定义置换模式
		Payloads: cliOpts.Payloads, // 自定义有效载荷/词表
		Limit:    cliOpts.Limit,    // 限制结果数量
		Enrich:   cliOpts.Enrich,   // 是否从输入中提取词汇丰富词表
		MaxSize:  cliOpts.MaxSize,  // 输出文件的最大大小
	}

	if cliOpts.PermutationConfig != "" {
		// 读取配置文件
		config, err := alterx.NewConfig(cliOpts.PermutationConfig)
		if err != nil {
			gologger.Fatal().Msgf("failed to read %v file got: %v", cliOpts.PermutationConfig, err)
		}
		// 如果配置文件中有模式，则使用配置文件中的模式
		if len(config.Patterns) > 0 {
			alterOpts.Patterns = config.Patterns
		}
		// 如果配置文件中有有效载荷，则使用配置文件中的有效载荷
		if len(config.Payloads) > 0 {
			alterOpts.Payloads = config.Payloads
		}
	}

	// 配置输出写入器
	var output io.Writer
	if cliOpts.Output != "" {
		// 如果指定了输出文件，则打开文件
		fs, err := os.OpenFile(cliOpts.Output, os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			gologger.Fatal().Msgf("failed to open output file %v got %v", cliOpts.Output, err)
		}
		output = fs
		defer fs.Close()
	} else {
		// 否则使用标准输出
		output = os.Stdout
	}

	// 使用配置选项创建新的alterx实例
	m, err := alterx.New(&alterOpts)
	if err != nil {
		gologger.Fatal().Msgf("failed to parse alterx config got %v", err)
	}

	// 如果只是估算结果数量而不生成结果
	if cliOpts.Estimate {
		gologger.Info().Msgf("Estimated Payloads (including duplicates) : %v", m.EstimateCount())
		return
	}

	// 执行alterx并将结果写入到输出
	if err = m.ExecuteWithWriter(output); err != nil {
		gologger.Error().Msgf("failed to write output to file got %v", err)
	}
}
