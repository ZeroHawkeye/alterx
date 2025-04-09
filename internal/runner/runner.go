package runner

import (
	"fmt"
	"io"
	"math"
	"os"
	"strings"

	"github.com/projectdiscovery/goflags"
	"github.com/projectdiscovery/gologger"
	"github.com/projectdiscovery/gologger/levels"
	fileutil "github.com/projectdiscovery/utils/file"
	updateutils "github.com/projectdiscovery/utils/update"
)

// Options 定义了程序的命令行选项
type Options struct {
	Domains            goflags.StringSlice // 用作基础的子域名列表
	Patterns           goflags.StringSlice // 输入的自定义模式
	Payloads           map[string][]string // 输入的有效载荷/词表
	Output             string              // 输出文件路径
	Config             string              // 配置文件路径
	PermutationConfig  string              // 置换配置文件路径
	Estimate           bool                // 是否仅估算而不生成置换
	DisableUpdateCheck bool                // 是否禁用自动更新检查
	Verbose            bool                // 是否显示详细输出
	Silent             bool                // 是否仅显示结果
	Enrich             bool                // 是否从输入中提取词汇丰富词表
	Limit              int                 // 限制返回结果的数量
	MaxSize            int                 // 最大导出数据大小
	// 内部/未导出字段
	wordlists goflags.RuntimeMap // 运行时词表映射，用于存储-pp参数
}

// ParseFlags 解析命令行参数并返回配置选项
// 这是程序启动时的主要配置点，处理所有命令行参数并设置默认值
func ParseFlags() *Options {
	var maxFileSize goflags.Size
	opts := &Options{}
	flagSet := goflags.NewFlagSet()
	flagSet.SetDescription(`Fast and customizable subdomain wordlist generator using DSL.`)

	// 输入参数组
	// 这些选项用于指定程序的输入源和处理方式
	flagSet.CreateGroup("input", "Input",
		// -l/-list 指定要用于创建置换的子域名列表(标准输入、逗号分隔的列表、文件)
		flagSet.StringSliceVarP(&opts.Domains, "list", "l", nil, "subdomains to use when creating permutations (stdin, comma-separated, file)", goflags.FileCommaSeparatedStringSliceOptions),
		// -p/-pattern 指定自定义置换模式(逗号分隔的列表、文件)
		flagSet.StringSliceVarP(&opts.Patterns, "pattern", "p", nil, "custom permutation patterns input to generate (comma-seperated, file)", goflags.FileCommaSeparatedStringSliceOptions),
		// -pp/-payload 指定自定义有效载荷模式，格式为key=value (-pp 'word=words.txt')
		flagSet.RuntimeMapVarP(&opts.wordlists, "payload", "pp", nil, "custom payload pattern input to replace/use in key=value format (-pp 'word=words.txt')"),
	)

	// 输出参数组
	// 这些选项控制程序的输出方式和详细程度
	flagSet.CreateGroup("output", "Output",
		// -es/-estimate 估算置换数量而不实际生成
		flagSet.BoolVarP(&opts.Estimate, "estimate", "es", false, "estimate permutation count without generating payloads"),
		// -o/-output 指定输出文件路径
		flagSet.StringVarP(&opts.Output, "output", "o", "", "output file to write altered subdomain list"),
		// -ms/-max-size 指定最大导出数据大小
		flagSet.SizeVarP(&maxFileSize, "max-size", "ms", "", "Max export data size (kb, mb, gb, tb) (default mb)"),
		// -v/-verbose 显示详细输出
		flagSet.BoolVarP(&opts.Verbose, "verbose", "v", false, "display verbose output"),
		// -silent 仅显示结果
		flagSet.BoolVar(&opts.Silent, "silent", false, "display results only"),
		// -version 显示版本信息
		flagSet.CallbackVar(printVersion, "version", "display alterx version"),
	)

	// 配置参数组
	// 这些选项用于指定配置文件和程序行为
	flagSet.CreateGroup("config", "Config",
		// -config 指定命令行配置文件
		flagSet.StringVar(&opts.Config, "config", "", `alterx cli config file (default '$HOME/.config/alterx/config.yaml')`),
		// -en/-enrich 从输入中提取词汇丰富词表
		flagSet.BoolVarP(&opts.Enrich, "enrich", "en", false, "enrich wordlist by extracting words from input"),
		// -ac 指定置换配置文件
		flagSet.StringVar(&opts.PermutationConfig, "ac", "", fmt.Sprintf(`alterx permutation config file (default '$HOME/.config/alterx/permutation_%v.yaml')`, version)),
		// -limit 限制返回结果的数量
		flagSet.IntVar(&opts.Limit, "limit", 0, "limit the number of results to return (default 0)"),
	)

	// 更新参数组
	// 这些选项控制程序的更新行为
	flagSet.CreateGroup("update", "Update",
		// -up/-update 更新alterx到最新版本
		flagSet.CallbackVarP(GetUpdateCallback(), "update", "up", "update alterx to latest version"),
		// -duc/-disable-update-check 禁用自动更新检查
		flagSet.BoolVarP(&opts.DisableUpdateCheck, "disable-update-check", "duc", false, "disable automatic alterx update check"),
	)

	// 解析命令行参数
	if err := flagSet.Parse(); err != nil {
		gologger.Fatal().Msgf("Could not read flags: %s\n", err)
	}

	// 如果指定了配置文件，则合并配置文件中的设置
	if opts.Config != "" {
		if err := flagSet.MergeConfigFile(opts.Config); err != nil {
			gologger.Error().Msgf("failed to read config file got %v", err)
		}
	}

	// 根据silent和verbose设置日志级别
	if opts.Silent {
		gologger.DefaultLogger.SetMaxLevel(levels.LevelSilent)
	} else if opts.Verbose {
		gologger.DefaultLogger.SetMaxLevel(levels.LevelVerbose)
	}
	showBanner()

	// 检查更新（除非禁用）
	if !opts.DisableUpdateCheck {
		latestVersion, err := updateutils.GetToolVersionCallback("alterx", version)()
		if err != nil {
			if opts.Verbose {
				gologger.Error().Msgf("alterx version check failed: %v", err.Error())
			}
		} else {
			gologger.Info().Msgf("Current alterx version %v %v", version, updateutils.GetVersionDescription(version, latestVersion))
		}
	}

	// 设置最大文件大小
	opts.MaxSize = math.MaxInt
	if maxFileSize > 0 {
		opts.MaxSize = int(maxFileSize)
	}

	// 处理词表参数
	// 将-pp参数传入的词表转换为内部使用的格式
	opts.Payloads = map[string][]string{}
	for k, v := range opts.wordlists.AsMap() {
		value, ok := v.(string)
		if !ok {
			continue
		}
		if fileutil.FileExists(value) {
			// 如果值是文件路径，则读取文件内容作为词表
			bin, err := os.ReadFile(value)
			if err != nil {
				gologger.Error().Msgf("failed to read wordlist %v got %v", value, err)
				continue
			}
			wordlist := strings.Fields(string(bin))
			opts.Payloads[k] = wordlist
		} else {
			// 否则将值本身作为单个词条
			opts.Payloads[k] = []string{value}
		}
	}

	// 从标准输入读取数据（如果有）
	if fileutil.HasStdin() {
		bin, err := io.ReadAll(os.Stdin)
		if err != nil {
			gologger.Error().Msgf("failed to read input from stdin got %v", err)
		}
		opts.Domains = strings.Fields(string(bin))
	}

	// 检查是否有输入域名
	// TODO: 将Options.Domains替换为Input String Channel
	if len(opts.Domains) == 0 {
		gologger.Fatal().Msgf("alterx: no input found")
	}

	return opts
}

// printVersion 打印当前版本信息并退出程序
func printVersion() {
	gologger.Info().Msgf("Current version: %s", version)
	os.Exit(0)
}
