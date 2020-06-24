// Copyright © 2019 Master.G
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package cmd

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/master-g/i18n/pkg/wkfs"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var execCmd = &cobra.Command{
	Use:   "exec",
	Short: "exec i18n, append translate text to android string xml.",
	PreRun: func(cmd *cobra.Command, args []string) {
		bindFlag(cmd, "nocheck")
		bindFlag(cmd, "src")
		bindFlag(cmd, "interact")
		bindFlag(cmd, "out")
	},
	Run: func(cmd *cobra.Command, args []string) {
		srcFiles := make(map[string]string)

		var absPath string
		var err error

		// STEP 1. iterate all source parameters, find all .csv files
		sources := viper.GetStringSlice("src")
		if len(sources) == 0 {
			logrus.Info("source missing")
			os.Exit(0)
		}

		var files []string
		for _, s := range sources {
			if wkfs.IsFile(s) && mightBeCSVFile(s) {
				files = append(files, s)
			} else if wkfs.IsDir(s) {
				var csvs []string
				csvs, _, err = wkfs.Scan(s, wkfs.WithFilesOnly(), wkfs.WithTypes("csv"))
				if err != nil {
					logrus.Errorf("cannot walk through directory %v, err:%v", s, err)
					continue
				}
				files = append(files, csvs...)
			}
		}

		for _, f := range files {
			absPath, err = filepath.Abs(f)
			if err != nil {
				logrus.Errorf("cannot get fullpath for %v, err:%v", f, err)
				continue
			}
			sum := md5.Sum([]byte(absPath))
			srcFiles[hex.EncodeToString(sum[:])] = absPath
		}

		// STEP 2. check output directory
		outputDir := viper.GetString("out")
		if !wkfs.IsDir(outputDir) {
			logrus.Infof("'%v' is not a valid output directory", outputDir)
			os.Exit(1)
		}

		var folders []string
		_, folders, err = wkfs.Scan(outputDir, wkfs.WithFoldersOnly())
		if err != nil {
			logrus.Errorf("cannot walk through directory %v, err:%v", outputDir, err)
			os.Exit(1)
		}

		var filteredPath []string
		for _, f := range folders {
			f = filepath.Clean(f)
			if strings.HasSuffix(f, "res") &&
				wkfs.IsDir(filepath.Join(f, "values")) &&
				wkfs.IsFile(filepath.Join(f, "values", "strings.xml")) &&
				wkfs.IsFile(filepath.Join(filepath.Dir(f), "AndroidManifest.xml")) {

				filteredPath = append(filteredPath, f)
			}
		}

		outputDir = ""
		interact := viper.GetBool("interact")
		if interact {
			if len(filteredPath) == 0 {
				force := false
				prompt := &survey.Confirm{
					Message: fmt.Sprintf("%v is not an valid output directory, process anyway?"),
				}
				err = survey.AskOne(prompt, &force)
				if err != nil {
					logrus.Error(err)
					os.Exit(1)
				}
				if !force {
					logrus.Info("abort")
					os.Exit(0)
				}
			} else if len(filteredPath) > 1 && len(filteredPath) < 42 {
				prompt := &survey.Select{
					Message: fmt.Sprintf("there are %v available output directories", len(filteredPath)),
					Options: filteredPath,
				}
				err = survey.AskOne(prompt, &outputDir)
				if err != nil {
					logrus.Error(err)
					os.Exit(1)
				}
			} else if len(filteredPath) == 1 {
				outputDir = filteredPath[0]
			} else {
				logrus.Errorf("too many candidate output directories(%v), abort", len(filteredPath))
				os.Exit(1)
			}
		} else {
			if len(filteredPath) == 0 {
				logrus.Errorf("the output might not be an android resource directory")
				logrus.Info("you might want to run the command again with --interact option")
				os.Exit(1)
			} else if len(filteredPath) > 1 {
				logrus.Errorf("there are multiple android resource directories")
				for _, v := range filteredPath {
					logrus.Info(v)
				}
				logrus.Info("you might want to run the command again with --interact option")
				os.Exit(1)
			} else {
				outputDir = filteredPath[0]
			}
		}

		if outputDir == "" {
			logrus.Error("no available output directory, abort")
			os.Exit(1)
		}

		var valueFolders []string
		_, valueFolders, err = wkfs.Scan(outputDir, wkfs.WithFoldersOnly(), wkfs.WithPatterns("values"))
		if err != nil {
			logrus.Errorf("cannot walk through output directory %v, err:%v", outputDir, err)
			os.Exit(1)
		}

		filteredPath = []string{}
		for _, f := range valueFolders {
			if wkfs.IsFile(filepath.Join(f, "strings.xml")) {
				filteredPath = append(filteredPath, f)
			}
		}

		langMap := make(map[string]string)
		for _, v := range filteredPath {
			base := filepath.Base(v)
			if strings.EqualFold(base, "values") {
				// en
				langMap["en"] = v
			} else {
				i := strings.IndexRune(base, '-')
				if i < 0 {
					continue
				}
				lang := base[i+1:]
				langMap[lang] = v
			}
		}
		for k, v := range langMap {
			logrus.Infof("%v -> %v", k, v)
		}
	},
}

func hasExtension(fp, ext string) bool {
	b := filepath.Ext(fp)
	return strings.Index(strings.ToLower(b), strings.ToLower(ext)) >= 0
}

func mightBeCSVFile(p string) bool {
	return hasExtension(p, "csv")
}

func mightBeXMLFile(p string) bool {
	return hasExtension(p, "xml")
}

func init() {
	rootCmd.AddCommand(execCmd)

	execCmd.Flags().BoolP("nocheck", "", false, "ignore common mistakes in translation text (﹪, s%, $n% etc.)")
	execCmd.Flags().StringSliceP("src", "s", []string{}, "source csv file/directories")
	execCmd.Flags().BoolP("interact", "", false, "handle collision in an interactive mode")
	execCmd.Flags().StringP("out", "o", "", "output directory")
}
