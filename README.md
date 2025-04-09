<h1 align="center">
 AlterX
<br>
</h1>


<p align="center">
<a href="https://opensource.org/licenses/MIT"><img src="https://img.shields.io/badge/license-MIT-_red.svg"></a>
<a href="https://goreportcard.com/badge/github.com/projectdiscovery/alterx"><img src="https://goreportcard.com/badge/github.com/projectdiscovery/alterx"></a>
<a href="https://pkg.go.dev/github.com/projectdiscovery/alterx/pkg/alterx"><img src="https://img.shields.io/badge/go-reference-blue"></a>
<a href="https://github.com/projectdiscovery/alterx/releases"><img src="https://img.shields.io/github/release/projectdiscovery/alterx"></a>
<a href="https://twitter.com/pdiscoveryio"><img src="https://img.shields.io/twitter/follow/pdiscoveryio.svg?logo=twitter"></a>
<a href="https://discord.gg/projectdiscovery"><img src="https://img.shields.io/discord/695645237418131507.svg?logo=discord"></a>
</p>

<p align="center">
  <a href="#特性">特性</a> •
  <a href="#安装">安装</a> •
  <a href="#帮助菜单">使用</a> •
  <a href="#示例">运行AlterX</a> •
  <a href="https://discord.gg/projectdiscovery">加入Discord</a>

</p>

<pre align="center">
<b>
   基于DSL的快速可定制子域名生成器
</b>
</pre>

## 特性
- 快速且可定制化
- **自动单词提取和丰富**
- 预定义变量
- **可配置的模式**
- 支持标准输入/列表输入

## 安装
要安装alterx，你需要在系统上安装Golang 1.19。你可以从[这里](https://go.dev/doc/install)下载Golang。安装Golang后，你可以使用以下命令安装alterx：

```bash
go install github.com/projectdiscovery/alterx/cmd/alterx@latest
```

## 帮助菜单
你可以使用以下命令查看可用的标志和选项：

```console
基于DSL的快速可定制子域名生成器

使用方法:
  ./alterx [flags]

参数:
输入:
   -l, -list string[]     创建置换时使用的子域名 (stdin、逗号分隔、文件)
   -p, -pattern string[]  要生成的自定义置换模式输入 (逗号分隔、文件)
   -pp, -payload value    自定义有效载荷模式输入，使用key=value格式 (-pp 'word=words.txt')

输出:
   -es, -estimate      估算置换数量而不生成有效载荷
   -o, -output string  输出文件路径，用于写入生成的子域名列表
   -ms, -max-size int  最大导出数据大小 (kb, mb, gb, tb) (默认为mb)
   -v, -verbose        显示详细输出
   -silent             仅显示结果
   -version            显示alterx版本

配置:
   -config string  alterx CLI配置文件 (默认为 '$HOME/.config/alterx/config.yaml')
   -en, -enrich    通过从输入中提取单词来丰富词表
   -ac string      alterx置换配置文件 (默认为 '$HOME/.config/alterx/permutation_v0.0.1.yaml')
   -limit int      限制返回结果的数量 (默认为0)

更新:
   -up, -update                 更新alterx到最新版本
   -duc, -disable-update-check  禁用自动alterx更新检查
```

## 项目学习路线图

本项目包含一个SVG格式的学习路线图`project.svg`，展示了AlterX的整体架构和学习路径。该路线图包含：

1. **核心组件** - 展示了各个主要源代码文件的作用和关系
2. **主要特性** - 列出了AlterX的关键功能点
3. **处理流程** - 展示了从输入到输出的整个处理流程
4. **学习路径建议** - 提供了学习本项目代码的推荐顺序

你可以使用任何支持SVG的浏览器或图像查看器打开此文件。对于初学者，建议按照学习路径建议的顺序逐步了解项目代码。

## 为什么选择 `alterx`？

`alterx`与其他子域名置换工具（如`goaltdns`）的不同之处在于它的`脚本化`特性。alterx接受模式作为输入，并基于这些模式生成子域名置换词表，类似于[nuclei](https://github.com/projectdiscovery/nuclei)与[fuzzing-templates](https://github.com/projectdiscovery/fuzzing-templates)的工作方式。

`主动子域名枚举`的难点在于找到实际存在的域名的概率。如果用一个比例尺来表示可能的子域名查找方法，它应该是这样的：

```console
   使用词表 < 使用子域名生成置换(goaltdns) < alterx
```

几乎所有流行的子域名置换工具都有硬编码的模式，当这些工具运行时，它们会创建包含数百万个子域名的词表，这降低了使用dnsx等工具暴力破解它们的可行性。命名子域名没有实际的约定，通常取决于注册子域名的人。使用`alterx`，可以基于`被动子域名枚举`结果创建模式，这增加了找到子域名的概率并提高了暴力破解它们的可行性。

## 变量

`alterx`使用类似于nuclei模板的变量语法。用户可以使用这些变量编写自己的模式。当域名作为输入传递时，`alterx`会评估输入并从中提取变量。

### 基本/常用变量

```yaml
{{sub}}     :  子域名前缀或子域名的最左部分
{{suffix}}  :  子域名名称中除{{sub}}之外的所有内容是后缀
{{tld}}     :  顶级域名(例如 com,uk,in等)
{{etld}}    :  也称为公共后缀(例如 co.uk, gov.in等)
```

| 变量         | api.scanme.sh | admin.dev.scanme.sh | cloud.scanme.co.uk |
| ---------- | ------------- | ------------------- | ------------------ |
| `{{sub}}`    | `api`         | `admin`             | `cloud`          |
| `{{suffix}}` | `scanme.sh`   | `dev.scanme.sh`     | `scanme.co.uk`   |
| `{{tld}}`    | `sh`          | `sh`                | `uk`             |
| `{{etld}}`   | `-`           | `-`                 | `co.uk`          |

### 高级变量

```yaml
{{sld}}   :  二级域名(例如对于 api.scanme.sh => {{sld}} 是 scanme)
{{root}}  :  也称为 eTLD+1，即只有根域名(例如对于 api.scanme.sh => {{root}} 是 scanme.sh)
{{subN}}  :  这里N是一个整数(例如 {{sub1}}, {{sub2}} 等)。

// {{subN}} 是高级变量，取决于输入而存在
// 假设有一个多级域名 cloud.nuclei.scanme.sh
// 在这种情况下 {{sub}} = cloud 且 {{sub1}} = nuclei
```

| 变量       | api.scanme.sh | admin.dev.scanme.sh | cloud.scanme.co.uk |
| -------- | ------------- | ------------------- | ------------------ |
| `{{sld}}` | `scanme`      | `scanme`            | `scanme`         |
| `{{root}}` | `scanme.sh`   | `scanme.sh`         | `scanme.co.uk`   |
| `{{sub1}}` | `-`           | `dev`               | `-`              |
| `{{sub2}}` | `-`           | `-`                 | `-`              |


## 模式

模式简单来说可以被视为`模板`，它描述了alterx应该生成什么类型的模式。

```console
// 以下是一些可用于生成置换的示例模式
// 假设api.scanme.sh被给定为输入，变量{{word}}被给定为输入，只有一个值prod
// alterx为以下模式生成子域名

"{{sub}}-{{word}}.{{suffix}}" // 例如: api-prod.scanme.sh
"{{word}}-{{sub}}.{{suffix}}" // 例如: prod-api.scanme.sh
"{{word}}.{{sub}}.{{suffix}}" // 例如: prod.api.scanme.sh
"{{sub}}.{{word}}.{{suffix}}" // 例如: api.prod.scanme.sh
```

这里有一个示例模式配置文件 - https://github.com/projectdiscovery/alterx/blob/main/permutations.yaml，可以根据需要轻松定制。

这个配置文件使用可定制的模式和动态有效载荷为安全评估或渗透测试生成子域名置换。模式包括基于破折号、点和其他方式的组合。用户可以创建自定义有效载荷部分，如单词、区域标识符或数字，以满足他们的特定需求。

例如，用户可以定义一个新的有效载荷部分`env`，其中包含`prod`和`dev`等值，然后在模式中使用它，如`{{env}}-{{word}}.{{suffix}}`，以生成如`prod-app.example.com`和`dev-api.example.com`的子域名。这种灵活性允许为独特的测试场景和目标环境定制子域名列表。

用于生成的默认模式配置文件存储在`$HOME/.config/alterx/`目录下，也可以使用`-ac`选项使用自定义配置文件。

## 示例

在`tesla.com`的现有被动子域名列表上运行alterx，使用[dnsx](https://github.com/projectdiscovery/dnsx)解析后，我们得到了**10个额外的新的**且**有效的子域名**。

```console
$ chaos -d tesla.com | alterx | dnsx



   ___   ____          _  __
  / _ | / / /____ ____| |/_/
 / __ |/ / __/ -_) __/>  <
/_/ |_/_/\__/\__/_/ /_/|_|

      projectdiscovery.io

[INF] Generated 8312 permutations in 0.0740s
auth-global-stage.tesla.com
auth-stage.tesla.com
digitalassets-stage.tesla.com
errlog-stage.tesla.com
kronos-dev.tesla.com
mfa-stage.tesla.com
paymentrecon-stage.tesla.com
sso-dev.tesla.com
shop-stage.tesla.com
www-uat-dev.tesla.com
```

类似地，`-enrich`选项可以用来填充已知子域名作为单词输入，以生成**目标感知置换**。

```console
$ chaos -d tesla.com | alterx -enrich

   ___   ____          _  __
  / _ | / / /____ ____| |/_/
 / __ |/ / __/ -_) __/>  <
/_/ |_/_/\__/\__/_/ /_/|_|

      projectdiscovery.io

[INF] Generated 662010 permutations in 3.9989s
```

你可以使用`-pattern` CLI选项在运行时更改默认模式。

```console
$ chaos -d tesla.com | alterx -enrich -p '{{word}}-{{suffix}}'

   ___   ____          _  __
  / _ | / / /____ ____| |/_/
 / __ |/ / __/ -_) __/>  <
/_/ |_/_/\__/\__/_/ /_/|_|

      projectdiscovery.io

[INF] Generated 21523 permutations in 0.7984s
```

也可以使用`-payload` CLI选项覆盖现有变量值。

```console
$ alterx -list tesla.txt -enrich -p '{{word}}-{{year}}.{{suffix}}' -pp word=keywords.txt -pp year=2023

   ___   ____          _  __
  / _ | / / /____ ____| |/_/
 / __ |/ / __/ -_) __/>  <
/_/ |_/_/\__/\__/_/ /_/|_|

      projectdiscovery.io

[INF] Generated 21419 permutations in 1.1699s
```

**欲了解更多信息，请查看发布博客** - https://blog.projectdiscovery.io/introducing-alterx-simplifying-active-subdomain-enumeration-with-patterns/


也请查看以下类似的开源项目，它们可能适合你的工作流程：

[altdns](https://github.com/infosec-au/altdns), [goaltdns](https://github.com/subfinder/goaltdns), [gotator](https://github.com/Josue87/gotator), [ripgen](https://github.com/resyncgg/ripgen/), [dnsgen](https://github.com/ProjectAnte/dnsgen), [dmut](https://github.com/bp0lr/dmut), [permdns](https://github.com/hpy/permDNS), [str-replace](https://github.com/j3ssie/str-replace), [dnscewl](https://github.com/codingo/DNSCewl), [regulator](https://github.com/cramppet/regulator)


--------

<div align="center">

**alterx** 由 [projectdiscovery](https://projectdiscovery.io) 团队用❤️制作，并在 [MIT许可证](LICENSE.md) 下分发。


<a href="https://discord.gg/projectdiscovery"><img src="https://raw.githubusercontent.com/projectdiscovery/nuclei-burp-plugin/main/static/join-discord.png" width="300" alt="加入Discord"></a>

</div>
