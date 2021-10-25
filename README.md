# i18n

中文文档请移步[这里](./README_CH.md)

A little tool that converts csv translations and append to android project.

## Workflow

### 1. Preprocess translation documentation

What developers got are usually excel documentations, for example:

|序号|英文|繁体|西语||
|:---|:---|:---|:---|:---|
|1|My Gift|我的禮物|mi regalo||
|2|Income Record|收入記錄|registro de ingresos||
||||

the process flow is:

1. remove blank columns and rows
2. change first row to the keys of `res/values`, english is default to `en`
3. change first column to the keys in `strings.xml`, that is `<string name="key">`
4. export to `csv` format

results be like:

|keys|en|zh-rTW|es|
|:---|:---|:---|:---|
|string_my_gift|My Gift|我的禮物|mi regalo|
|string_income_record|Income Record|收入記錄|registro de ingresos|

### 2. Append to android project

Download corresponding executable `i18n` for your OS, run in cli env

usage:

1. check help

`i18n help [subcommand]`

2. use `wizard` subcommand to create full command

3. use `append` subcommand directly

`i18n append --src [path to csv file/directory] --out [path to android project res directory] [flags]`

available flags are:

* `--verbose` print extra debug info at runtime
* `--interact` run command in interactive mode, will ask options when text conflicts occurs
* `--nolint` DO NOT check common text mistakes, e.g. full-width space, wrong format placeholders
* `--noescape` DO NOT convert [special characters](https://developer.android.com/guide/topics/resources/string-resource#FormattingAndStyling) in text
* `--prefer-new` when `--interact` is not specified, use new value (in `csv`) if there are any conflicts in text (existed `xml`)
* `--auto-placehoder` automatically convert from `%AA`, `%BB` to `%1$s`, `%2$s` (`%` must be half width, `AA`, `BB` must be uppercase, only supports output `%n$s` format)
* `--key-mapping-config` language key mapping config file, will convert language key in `csv`
* `--key` language key
* `--alias` language key mapping value
* `--dry` run the command in dry mode, will not modify any files

**about language key mapping**

to save time, you can specify language key mapping now

for example, a `csv` source file:

|keys|英语|繁体中文|西语|
|:---|:---|:---|:---|
|string_my_gift|My Gift|我的禮物|mi regalo|
|string_income_record|Income Record|收入記錄|registro de ingresos|

you can specify a language key mapping configuration file to let `i18n` do the key convention for you

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

and run the command like:

`i18n --src path-to-csv --out path-to-android-res --key-mapping-config path-to-config-file`

or you can specify the key mapping in arguments via `--key` and `--alias`

`i18n --src path-to-csv --out path-to-android-res --key "英语" --alias "en" --key "繁体中文" --alias "zh-rTW" --key "西语" --alias "es"`

### 3. check output in `res` directory

after execution of `i18n`, check the result in `res` folder of your Android Project, and fix any potential bugs
