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
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/master-g/i18n/internal/i18n"
	"github.com/master-g/texturepacker/pkg/config"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var paramList = []config.Flag{
	{Name: "out", Type: config.String, Shorthand: "o", Value: "", Usage: "output json file path."},
	{Name: "overwrite", Type: config.Bool, Shorthand: "", Value: false, Usage: "overwrite exists output file."},
	{Name: "append", Type: config.Bool, Shorthand: "a", Value: false, Usage: "append output to exists json file."},
	{Name: "metadata", Type: config.Bool, Shorthand: "m", Value: false, Usage: "write metadata info to output json file."},
	{Name: "verbose", Type: config.Bool, Shorthand: "v", Value: false, Usage: "show verbose information during the processing."},
}

// convertCmd represents the serve command
var convertCmd = &cobra.Command{
	Use:   "convert [input csv file]",
	Short: "convert csv to json",
	Long:  `convert csv to json file`,
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		setupLogger()
		runApplication(args)
	},
}

func init() {
	rootCmd.AddCommand(convertCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// convertCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// bind flags with viper
	err := config.BindFlags(convertCmd, paramList)
	if err != nil {
		logrus.Errorln("unable to config i18n converter:", err)
	}
}

func setupLogger() {
	if viper.GetBool("verbose") {
		logrus.SetLevel(logrus.DebugLevel)
	} else {
		logrus.SetLevel(logrus.InfoLevel)
	}

	logrus.SetFormatter(&logrus.TextFormatter{
		DisableColors: true,
		FullTimestamp: true,
	})
}

func filterCSVFile(p string, info os.FileInfo) bool {
	return strings.EqualFold(".csv", filepath.Ext(p))
}

func getDefaultOutputFile(dir, ext string) string {
	prefix := "i18n_"
	ts := time.Now().Format("20060102T150405")
	return filepath.Join(dir, strings.Join([]string{prefix, ts, ext}, ""))
}

func getCurrentDir() string {
	exe, err := os.Executable()
	if err != nil {
		logrus.Fatal("cannot obtain pwd for default output, err", err)
	}

	exePath := filepath.Dir(exe)
	return exePath
}

func runApplication(args []string) {
	if viper.GetBool("overwrite") && viper.GetBool("append") {
		logrus.Fatal("overwrite or append, use only one at a time.")
	}
	// STEP 1. inspect input files and directories

	// absolute path : csv file
	csvFiles := make(map[string]string)

	var fi os.FileInfo
	var absPath string
	var err error
	// iterate all arguments
	for _, arg := range args {
		// check if file/directory is accessible
		if fi, err = os.Stat(arg); err == nil {
			switch mode := fi.Mode(); {
			case mode.IsDir():
				// walk dir
				logrus.Debug("searching ", arg, " ...")
				err = filepath.Walk(arg, func(path string, info os.FileInfo, err2 error) error {
					absPath, err = filepath.Abs(path)
					if info.Mode().IsRegular() && filterCSVFile(path, info) {
						logrus.Debug("found ", absPath)
						csvFiles[absPath] = info.Name()
					}
					return nil
				})
			case mode.IsRegular():
				absPath, err = filepath.Abs(arg)
				if err == nil && filterCSVFile(arg, fi) {
					csvFiles[absPath] = fi.Name()
				} else if err != nil {
					logrus.Warnf("cannot obtain absolute path of %v, err: %v", arg, err)
				}
			}
		} else {
			logrus.Warnf("cannot access %v, err: %v", arg, err)
		}
	}

	if len(csvFiles) == 0 {
		logrus.Fatal("no input csv file found, abort")
	}

	// STEP 2. inspect output path, if output path is an existed directory, generates output filename for user
	outputJSONPath := viper.GetString("out")
	if outputJSONPath == "" {
		outputJSONPath = getDefaultOutputFile(getCurrentDir(), ".json")
	} else {
		fi, err = os.Stat(outputJSONPath)
		if err == nil && fi.IsDir() {
			// output path is a dir, complete the output file name
			outputJSONPath = getDefaultOutputFile(outputJSONPath, ".json")
		} else if err == nil && fi.Mode().IsRegular() {
			if !viper.GetBool("overwrite") && !viper.GetBool("append") {
				// output path is a regular file, and append/overwrite flag is not specified
				logrus.Fatal("output file already exists, use --append or --overwrite flag")
			}
		}
	}

	// STEP 3. run the converter
	converter := i18n.NewConverter(&i18n.Config{
		OutputJSONPath: outputJSONPath,
		Append:         viper.GetBool("append"),
		HasMeta:        viper.GetBool("metadata"),
		Overwrite:      viper.GetBool("overwrite"),
	})

	if converter == nil {
		logrus.Fatal("cannot create converter instance")
	}

	if viper.GetBool("append") {
		err = converter.ReadAppendFile(outputJSONPath)
		if err != nil {
			logrus.Fatal("unable to read append json file, err:%v", err)
		}
	}

	// stat
	logrus.Infof("output: %v", outputJSONPath)
	logrus.Infof("append: %v", viper.GetBool("append"))
	logrus.Infof("metadata: %v", viper.GetBool("metadata"))
	logrus.Infof("overwrite: %v", viper.GetBool("overwrite"))

	err = converter.Convert(csvFiles)
	if err != nil {
		logrus.Fatal("cannot convert i18n csv, err:", err)
	}

	logrus.Info("goodbye, have a nice day ❤")
}
