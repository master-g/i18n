# i18n

[TOC]

For english README, checkout [here](./README.md)

此工具可以将处理后的多语言文件(目前仅支持 csv) 批量添加到 Android 工程路径下的 res/value(-lang)/strings.xml 中

## 流程

### 1. 处理产品交付的多语言源文档

产品交付给开发的多语言资源一般为 excel 文档, 例如:

|序号|英文|繁体|西语||
|:---|:---|:---|:---|:---|
|1|My Gift|我的禮物|mi regalo||
|2|Income Record|收入記錄|registro de ingresos||
||||

处理过程如下:

1. 去掉多余的空白行, 列
2. 第一行语言名称改为 res/values 目录对应的后缀, 英语默认后缀用 en 代替
3. 第一列改为多语言文本的 name 值, 即 <string name="string_hello"> 中的 name 值
4. 导出为 csv 格式

结果:

|keys|en|zh-rTW|es|
|:---|:---|:---|:---|
|string_my_gift|My Gift|我的禮物|mi regalo|
|string_income_record|Income Record|收入記錄|registro de ingresos|

### 2. 添加至 Android 工程

根据你的操作系统下载对应的 i18n 可执行文件, 并在命令行环境中运行

用法:

1. 查看帮助

`i18n help [subcommand]`

2. 使用 wizard 子命令来创建完整的命令

`i18n wizard`

3. 直接使用 append 子命令

`i18n append --src [多语言 csv 文件/目录] --out [android 工程 res 目录] [flags]`

可选的 flags 有:

* `--verbose` 运行时打印额外的调试信息
* `--interact` 以交互式运行命令, 在遇到文案冲突等异常情况时询问下一步操作
* `--nolint` 不检查常见的文案错误, 例如全角百分号, 错误的格式化标识符等
* `--noescape` 不自动处理文案中的特殊字符
* `--prefer-new` 在未指定 `--interact` 时, 如遇到键值冲突, 则使用新(`csv` 中的)值替换旧(`xml` 中的)值
* `--auto-placeholder` 自动将 `%AA`, `%BB` 转换为 `%1$s`, `%2$s` (要求 `%` 是半角, `AA`, `BB` 必须大写, 暂时仅支持输出 `%n$s` 格式)
* `--key-mapping-config` 指定语言名称转换配置文件
* `--key` 在命令行参数中指定语言名称转换的源语言名称
* `--alias` 在命令行参数中指定语言名称转换的目标语言名称
* `--dry` 以 dry 模式运行命令，用于检查和调试，不会修改任何文件

**关于语言名称转换**

为了节省您宝贵的时间，现在 `i18n` 支持指定语言名称转换配置，此功能会在运行时自动对 `csv` 中的语言名称进行转换

比如我们有一份如下的 `csv` 文件:

|keys|英语|繁体中文|西语|
|:---|:---|:---|:---|
|string_my_gift|My Gift|我的禮物|mi regalo|
|string_income_record|Income Record|收入記錄|registro de ingresos|

你可以通过指定一份配置文件来自动进行语言名称转换，配置文件如下:

```json
{
  "mapping": [
    {
      "key": "英语",
      "alias": "en"
    }, {
      "key": "繁体中文",
      "alias": "zh-rTW"
    }, {
      "key": "西语",
      "alias": "es"
    }
  ]
}
```

然后像这样执行命令

`i18n --src path-to-csv --out path-to-android-res --key-mapping-config path-to-config-file`

或者你也可以通过在命令行中输入转换配置

`i18n --src path-to-csv --out path-to-android-res --key "英语" --alias "en" --key "繁体中文" --alias "zh-rTW" --key "西语" --alias "es"`

注意，每个 `--key` 必须对应一个 `--alias`

### 3. 检查 `res` 目录下的输出

命令执行无异常后, 请人工核对文案的添加结果并处理可能存在的错误

## 配置文件

`i18n` 支持配置文件

```shell
$ i18n append --config path-to-your-config.yaml [...options]
```

配置文件示例:

```yaml
src: ./source_example.csv
out: /path/to/your/anroid/project/res/
key:
  - English
  - CN
  - TW
alias:
  - en
  - zh-rCN
  - zh-rTW
```

你可以指定多个输入源

```yaml
src:
  - ./source_example.csv
  - /path/to/source.csv
  - /path/to/sources/dir # 也支持目录
out: /path/to/your/android/project/res/
key-mapping-config: /path/to/key-mapping-config.json
```

请注意，配置文件不支持以下 flags:

* `--verbose`
* `--interact`
* `--nolint`
* `--noescape`
* `--prefer-new`
* `--auto-placeholder`
* `--dry`

你需要在运行命令时指定这些 flag
