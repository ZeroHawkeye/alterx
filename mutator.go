package alterx

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"regexp"
	"strings"
	"time"

	"github.com/projectdiscovery/fasttemplate"
	"github.com/projectdiscovery/gologger"
	"github.com/projectdiscovery/utils/dedupe"
	errorutil "github.com/projectdiscovery/utils/errors"
	sliceutil "github.com/projectdiscovery/utils/slice"
)

var (
	extractNumbers   = regexp.MustCompile(`[0-9]+`)
	extractWords     = regexp.MustCompile(`[a-zA-Z0-9]+`)
	extractWordsOnly = regexp.MustCompile(`[a-zA-Z]{3,}`)
	DedupeResults    = true // Dedupe all results (default: true)
)

// Mutator Options
type Options struct {
	// list of Domains to use as base
	Domains []string
	// list of words to use while creating permutations
	// if empty DefaultWordList is used
	Payloads map[string][]string
	// list of pattersn to use while creating permutations
	// if empty DefaultPatterns are used
	Patterns []string
	// Limits output results (0 = no limit)
	Limit int
	// Enrich when true alterx extra possible words from input
	// and adds them to default payloads word,number
	Enrich bool
	// MaxSize limits output data size
	MaxSize int
}

// Mutator 是生成置换的核心结构体
type Mutator struct {
	Options      *Options
	payloadCount int
	Inputs       []*Input // all processed inputs
	timeTaken    time.Duration
	// internal or unexported variables
	maxkeyLenInBytes int
}

// New 创建并返回一个新的置换生成器实例
func New(opts *Options) (*Mutator, error) {
	if len(opts.Domains) == 0 {
		return nil, fmt.Errorf("no input provided to calculate permutations")
	}
	if len(opts.Payloads) == 0 {
		opts.Payloads = map[string][]string{}
		if len(DefaultConfig.Payloads) == 0 {
			return nil, fmt.Errorf("something went wrong, `DefaultWordList` and input wordlist are empty")
		}
		opts.Payloads = DefaultConfig.Payloads
	}
	if len(opts.Patterns) == 0 {
		if len(DefaultConfig.Patterns) == 0 {
			return nil, fmt.Errorf("something went wrong,`DefaultPatters` and input patterns are empty")
		}
		opts.Patterns = DefaultConfig.Patterns
	}
	// 清除重复项
	for k, v := range opts.Payloads {
		dedupe := sliceutil.Dedupe(v)
		if len(v) != len(dedupe) {
			gologger.Warning().Msgf("%v duplicate payloads found in %v. purging them..", len(v)-len(dedupe), k)
			opts.Payloads[k] = dedupe
		}
	}
	m := &Mutator{
		Options: opts,
	}
	if err := m.validatePatterns(); err != nil {
		return nil, err
	}
	if err := m.prepareInputs(); err != nil {
		return nil, err
	}
	if opts.Enrich {
		m.enrichPayloads()
	}
	return m, nil
}

// Execute 使用输入的词表和模式计算所有的置换，并将结果写入字符串通道
func (m *Mutator) Execute(ctx context.Context) <-chan string {
	var maxBytes int
	if DedupeResults {
		count := m.EstimateCount()
		maxBytes = count * m.maxkeyLenInBytes
	}

	results := make(chan string, len(m.Options.Patterns))
	go func() {
		now := time.Now()
		// 遍历所有输入
		for _, v := range m.Inputs {
			varMap := getSampleMap(v.GetMap(), m.Options.Payloads)
			// 遍历所有模式
			for _, pattern := range m.Options.Patterns {
				if err := checkMissing(pattern, varMap); err == nil {
					statement := Replace(pattern, v.GetMap())
					select {
					case <-ctx.Done():
						return
					default:
						// 对每个模式进行组合爆破
						m.clusterBomb(statement, results)
					}
				} else {
					gologger.Warning().Msgf("%v : failed to evaluate pattern %v. skipping", err.Error(), pattern)
				}
			}
		}
		m.timeTaken = time.Since(now)
		close(results)
	}()

	if DedupeResults {
		// 去除结果中的重复项
		d := dedupe.NewDedupe(results, maxBytes)
		d.Drain()
		return d.GetResults()
	}
	return results
}

// ExecuteWithWriter 执行置换生成并将结果直接写入实现io.Writer接口的对象
func (m *Mutator) ExecuteWithWriter(Writer io.Writer) error {
	if Writer == nil {
		return errorutil.NewWithTag("alterx", "writer destination cannot be nil")
	}
	resChan := m.Execute(context.TODO())
	m.payloadCount = 0
	maxFileSize := m.Options.MaxSize
	for {
		value, ok := <-resChan
		if !ok {
			gologger.Info().Msgf("Generated %v permutations in %v", m.payloadCount, m.Time())
			return nil
		}
		if m.Options.Limit > 0 && m.payloadCount == m.Options.Limit {
			// 如果达到了限制，我们不能提前退出，由于抽象，我们必须完成处理以耗尽所有去重器
			continue
		}
		if maxFileSize <= 0 {
			// 当达到最大文件大小时，耗尽所有去重器
			continue
		}

		if strings.HasPrefix(value, "-") {
			continue
		}

		outputData := []byte(value + "\n")
		if len(outputData) > maxFileSize {
			maxFileSize = 0
			continue
		}

		n, err := Writer.Write(outputData)
		if err != nil {
			return err
		}
		// 每次写入后更新最大文件大小限制
		maxFileSize -= n
		m.payloadCount++
	}
}

// EstimateCount 估计将创建的置换数量，而不实际执行/创建置换
func (m *Mutator) EstimateCount() int {
	counter := 0
	for _, v := range m.Inputs {
		varMap := getSampleMap(v.GetMap(), m.Options.Payloads)
		for _, pattern := range m.Options.Patterns {
			if err := checkMissing(pattern, varMap); err == nil {
				// 如果模式是 {{sub}}.{{sub1}}-{{word}}.{{root}}
				// 且输入域名是 api.scanme.sh，显然这里的 {{sub1}} 将是空的/缺失的
				// 在这种情况下，`alterx` 会静默跳过该特定输入的该模式
				// 这样，用户可以有一个长的模式列表，但它们只在所有必需的数据都给出时使用（类似于自包含模板）
				statement := Replace(pattern, v.GetMap())
				bin := unsafeToBytes(statement)
				if m.maxkeyLenInBytes < len(bin) {
					m.maxkeyLenInBytes = len(bin)
				}
				varsUsed := getAllVars(statement)
				if len(varsUsed) == 0 {
					counter += 1
				} else {
					tmpCounter := 1
					for _, word := range varsUsed {
						tmpCounter *= len(m.Options.Payloads[word])
					}
					counter += tmpCounter
				}
			}
		}
	}
	return counter
}

// DryRun 执行但不存储置换，并返回创建的置换数量
// 这个值也存储在变量中，可以通过getter "PayloadCount"访问
func (m *Mutator) DryRun() int {
	m.payloadCount = 0
	err := m.ExecuteWithWriter(io.Discard)
	if err != nil {
		gologger.Error().Msgf("alterx: got %v", err)
	}
	return m.payloadCount
}

// clusterBomb 计算组合爆破攻击的所有有效载荷，并将它们发送到结果通道
func (m *Mutator) clusterBomb(template string, results chan string) {
	// 提前退出：这是使clusterbomb免于堆栈溢出并减少
	// n*len(n)次迭代和n次递归的原因
	varsUsed := getAllVars(template)
	if len(varsUsed) == 0 {
		// 不需要组合爆破
		// 只需将现有模板作为结果发送并退出
		results <- template
		return
	}
	payloadSet := map[string][]string{}
	// 不是发送所有有效载荷，只发送在模板/声明中使用的有效载荷
	leftmostSub := strings.Split(template, ".")[0]
	for _, v := range varsUsed {
		payloadSet[v] = []string{}
		for _, word := range m.Options.Payloads[v] {
			if !strings.HasPrefix(leftmostSub, word) && !strings.HasSuffix(leftmostSub, word) {
				// 跳过已经出现在最左侧子域名中的所有单词
				// 我们很可能永远不会找到 api-api.example.com
				payloadSet[v] = append(payloadSet[v], word)
			}
		}
	}
	payloads := NewIndexMap(payloadSet)
	// 在组合爆破攻击中生成的有效载荷数量...
	callbackFunc := func(varMap map[string]interface{}) {
		results <- Replace(template, varMap)
	}
	ClusterBomb(payloads, callbackFunc, []string{})
}

// prepares input and patterns and calculates estimations
func (m *Mutator) prepareInputs() error {
	var errors []string
	// prepare input
	var allInputs []*Input
	for _, v := range m.Options.Domains {
		i, err := NewInput(v)
		if err != nil {
			errors = append(errors, err.Error())
			continue
		}
		allInputs = append(allInputs, i)
	}
	m.Inputs = allInputs
	if len(errors) > 0 {
		gologger.Warning().Msgf("errors found when preparing inputs got: %v : skipping errored inputs", strings.Join(errors, " : "))
	}
	return nil
}

// validates all patterns by compiling them
func (m *Mutator) validatePatterns() error {
	for _, v := range m.Options.Patterns {
		// check if all placeholders are correctly used and are valid
		if _, err := fasttemplate.NewTemplate(v, ParenthesisOpen, ParenthesisClose); err != nil {
			return err
		}
	}
	return nil
}

// enrichPayloads extract possible words and adds them to default wordlist
func (m *Mutator) enrichPayloads() {
	var temp bytes.Buffer
	for _, v := range m.Inputs {
		temp.WriteString(v.Sub + " ")
		if len(v.MultiLevel) > 0 {
			temp.WriteString(strings.Join(v.MultiLevel, " "))
		}
	}
	numbers := extractNumbers.FindAllString(temp.String(), -1)
	extraWords := extractWords.FindAllString(temp.String(), -1)
	extraWordsOnly := extractWordsOnly.FindAllString(temp.String(), -1)
	if len(extraWordsOnly) > 0 {
		extraWords = append(extraWords, extraWordsOnly...)
		extraWords = sliceutil.Dedupe(extraWords)
	}

	if len(m.Options.Payloads["word"]) > 0 {
		extraWords = append(extraWords, m.Options.Payloads["word"]...)
		m.Options.Payloads["word"] = sliceutil.Dedupe(extraWords)
	}
	if len(m.Options.Payloads["number"]) > 0 {
		numbers = append(numbers, m.Options.Payloads["number"]...)
		m.Options.Payloads["number"] = sliceutil.Dedupe(numbers)
	}
}

// PayloadCount returns total estimated payloads count
func (m *Mutator) PayloadCount() int {
	if m.payloadCount == 0 {
		return m.EstimateCount()
	}
	return m.payloadCount
}

// Time returns time taken to create permutations in seconds
func (m *Mutator) Time() string {
	return fmt.Sprintf("%.4fs", m.timeTaken.Seconds())
}
