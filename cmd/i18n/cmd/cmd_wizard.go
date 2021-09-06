package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func askBoolean(msg string) bool {
	var answer string
	prompt := &survey.Select{
		Message: msg,
		Options: []string{"Yes", "No"},
	}
	err := survey.AskOne(prompt, &answer)
	if err != nil {
		logrus.Fatal(err)
		os.Exit(1)
	}

	return strings.EqualFold(answer, "yes")
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
			}
			err = survey.AskOne(prompt, &output, survey.WithValidator(survey.Required))
			if err != nil {
				logrus.Fatal(err)
				os.Exit(1)
			}
		}

		// interactive
		interactive := askBoolean("需要在出现文案冲突时进行人工干预吗")
		var preferOld bool
		if !interactive {
			preferOld = askBoolean("在遇到键值冲突时保留 xml 中的旧值吗")
		}
		lint := askBoolean("需要检查一些常见的文案问题吗")
		escape := askBoolean("需要自动转换特殊字符吗")
		placeholder := askBoolean("需要自动转换占位符吗")
		verbose := askBoolean("是否输出额外的日志信息")

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
