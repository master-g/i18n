package cmd

import (
	"fmt"
	"github.com/master-g/i18n/pkg/wkfs"
	"os"
	"path/filepath"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func askBoolean(msg string, defaultValue bool) bool {
	var answer string
	var options []string
	if defaultValue {
		options = []string{"Yes", "No"}
	} else {
		options = []string{"No", "Yes"}
	}

	prompt := &survey.Select{
		Message: msg,
		Options: options,
	}
	err := survey.AskOne(prompt, &answer)
	if err != nil {
		logrus.Fatal(err)
		os.Exit(1)
	}

	return strings.EqualFold(answer, "yes")
}

func askOption(msg string, options []string) string {
	var answer string
	prompt := &survey.Select{
		Message: msg,
		Options: options,
	}
	err := survey.AskOne(prompt, &answer)
	if err != nil {
		logrus.Fatal(err)
		os.Exit(1)
	}

	return answer
}

var wizardCmd = &cobra.Command{
	Use:   "wizard",
	Short: "step by step i18n command setup wizard",
	Run: func(cmd *cobra.Command, args []string) {
		exePath, err := os.Executable()
		if err != nil {
			logrus.Fatal(err)
			os.Exit(1)
		}

		// sub command select
		fmt.Println(exePath)

		// source input
		var srcInput string
		{
			prompt := &survey.Input{
				Message: "多语言文案 CSV 文件/目录路径",
				Suggest: func(toComplete string) []string {
					files, _ := filepath.Glob(toComplete + "*")
					return files
				},
			}
			err = survey.AskOne(prompt, &srcInput, survey.WithValidator(survey.Required))
			if err != nil {
				logrus.Fatal(err)
				os.Exit(1)
			}
		}

		// output
		var output string
		{
			prompt := &survey.Input{
				Message: "需要添加多语言的 Android 工程 res 目录路径",
				Suggest: func(toComplete string) []string {
					files, _ := filepath.Glob(toComplete + "*")
					var dirs []string
					set := make(map[string]bool)
					for _, v := range files {
						if wkfs.IsDir(v) && set[v] != true {
							set[v] = true
							dirs = append(dirs, v)
						} else if wkfs.IsFile(v) {
							dir := filepath.Dir(v)
							if set[dir] != true {
								set[dir] = true
								dirs = append(dirs, dir)
							}
						}
					}
					return dirs
				},
			}
			err = survey.AskOne(prompt, &output, survey.WithValidator(survey.Required))
			if err != nil {
				logrus.Fatal(err)
				os.Exit(1)
			}
		}

		// interactive
		interactive := askBoolean("需要在出现文案冲突时进行人工干预吗", true)
		var preferOld bool
		if !interactive {
			preferOld = askBoolean("在遇到键值冲突时保留 xml 中的旧值吗", true)
		}
		lint := askBoolean("需要检查一些常见的文案问题吗", true)
		escape := askBoolean("需要自动转换特殊字符吗", true)
		placeholder := askBoolean("需要自动转换占位符吗", true)
		verbose := askBoolean("是否输出额外的日志信息", true)

		// key mapping
		keyMapResult := ""
		{
			if askBoolean("需要进行语种名称转换吗", false) {
				optReadFromFile := "从文件读取"
				optReadFromInput := "手动输入"
				optCancel := "还是算了"
				switch askOption("是文件还是手动输入", []string{optReadFromFile, optReadFromInput, optCancel}) {
				case optReadFromFile:
					// 从文件加载
					var keyMappingCfgFilePath string
					keyMapCfgPrompt := &survey.Input{
						Message: "请输入语种名称键值配置文件路径",
						Suggest: func(toComplete string) []string {
							files, _ := filepath.Glob(toComplete + "*")
							return files
						},
					}
					err = survey.AskOne(keyMapCfgPrompt, &keyMappingCfgFilePath, survey.WithValidator(survey.Required))
					if err != nil {
						logrus.Fatal(err)
						os.Exit(1)
					}
					keyMapResult = fmt.Sprintf("--%v %v", flagsKeyMappingConfig, keyMappingCfgFilePath)
				case optReadFromInput:
					// 从参数输入
					var keyList []string
					var aliasList []string
					for {
						kp := &survey.Input{
							Message: "请输入要映射的语种源名称, 直接回车结束输入阶段",
						}
						var k string
						err = survey.AskOne(kp, &k, nil)
						if err != nil {
							logrus.Fatal(err)
							os.Exit(1)
						}
						if k == "" {
							break
						}
						ap := &survey.Input{
							Message: "请输入要映射的语种目标名称",
						}
						var a string
						err = survey.AskOne(ap, &a, survey.WithValidator(survey.Required))
						if err != nil {
							logrus.Fatal(err)
							os.Exit(1)
						}
						keyList = append(keyList, k)
						aliasList = append(aliasList, a)
					}
					kvs := make(map[string]string)
					for i := 0; i < len(keyList); i++ {
						if i > len(aliasList) {
							continue
						}
						kvs[keyList[i]] = aliasList[i]
					}
					sb := strings.Builder{}
					sb.WriteString("")
					for k, v := range kvs {
						sb.WriteString(fmt.Sprintf("--key \"%v\" --alias \"%v\" ", k, v))
					}
					keyMapResult = sb.String()
				default:
				}
			}
		}

		generated := []string{"append"}

		generated = append(generated, fmt.Sprintf("--src %v", srcInput))
		generated = append(generated, fmt.Sprintf("--out %v", output))

		appendFlags := func(arr []string, flags string) []string {
			return append(arr, fmt.Sprintf("--%v", flags))
		}

		if interactive {
			generated = appendFlags(generated, flagsInteract)
		} else if !preferOld {
			generated = appendFlags(generated, flagsPreferNew)
		}
		if !lint {
			generated = appendFlags(generated, flagsNoLint)
		}
		if !escape {
			generated = appendFlags(generated, flagsNoEscape)
		}
		if placeholder {
			generated = appendFlags(generated, flagsAutoPlaceHolder)
		}
		generated = append(generated, keyMapResult)
		if verbose {
			generated = appendFlags(generated, flagsVerbose)
		}

		fmt.Println("run following command")
		fmt.Printf("%v %v\n", filepath.Base(exePath), strings.Join(generated, " "))
	},
}

func init() {
	rootCmd.AddCommand(wizardCmd)
}
