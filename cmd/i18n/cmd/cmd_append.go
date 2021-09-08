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
	"encoding/json"
	"fmt"
	"golang.org/x/sys/windows"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/master-g/i18n/internal/appender"
	"github.com/master-g/i18n/internal/model"
	"github.com/master-g/i18n/internal/parser"
	"github.com/master-g/i18n/pkg/wkfs"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type Message struct {
	Key, Alias string
}

var appendCmd = &cobra.Command{
	Use:   "append",
	Short: "append translate text to android string xml.",
	PreRun: func(cmd *cobra.Command, args []string) {
		bindFlag(cmd, "src")
		bindFlag(cmd, "out")
		bindFlag(cmd, flagsInteract)
		bindFlag(cmd, flagsPreferNew)
		bindFlag(cmd, flagsNoLint)
		bindFlag(cmd, flagsNoEscape)
		bindFlag(cmd, flagsAutoPlaceHolder)
		bindFlag(cmd, flagsKeyConfigure)
	},
	Run: func(cmd *cobra.Command, args []string) {
		srcFiles := make(map[string]string)

		var absPath string
		var err error

		// STEP 1. iterate all source parameters, find all .csv files
		logrus.Info("checking sources...")
		sources := viper.GetStringSlice("src")
		if len(sources) == 0 {
			logrus.Info("source missing")
			//exit(0)
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
		logrus.Infof("%d source file(s) found", len(srcFiles))
		for _, f := range srcFiles {
			logrus.Info(f)
		}

		// STEP 2. check output directory
		logrus.Info("checking output directory...")
		outputDir := viper.GetString("out")
		if !wkfs.IsDir(outputDir) {
			logrus.Errorf("'%v' is not a valid output directory", outputDir)
			exit(1)
		}

		var folders []string
		_, folders, err = wkfs.Scan(outputDir, wkfs.WithFoldersOnly())
		if err != nil {
			logrus.Errorf("cannot walk through directory %v, err:%v", outputDir, err)
			exit(1)
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

		originOutputDir := outputDir
		outputDir = ""
		interact := viper.GetBool(flagsInteract)
		if interact {
			if len(filteredPath) == 0 {
				force := false
				prompt := &survey.Confirm{
					Message: fmt.Sprintf("%v is not an valid output directory, process anyway?", originOutputDir),
				}
				err = survey.AskOne(prompt, &force)
				if err != nil {
					logrus.Error(err)
					exit(1)
				}
				if !force {
					logrus.Info("abort")
					exit(0)
				}
			} else if len(filteredPath) > 1 && len(filteredPath) < 42 {
				prompt := &survey.Select{
					Message: fmt.Sprintf("there are %v available output directories", len(filteredPath)),
					Options: filteredPath,
				}
				err = survey.AskOne(prompt, &outputDir)
				if err != nil {
					logrus.Error(err)
					exit(1)
				}
			} else if len(filteredPath) == 1 {
				outputDir = filteredPath[0]
			} else {
				logrus.Errorf("too many candidate output directories(%v), abort", len(filteredPath))
				exit(1)
			}
		} else {
			if len(filteredPath) == 0 {
				logrus.Errorf("the output might not be an android resource directory")
				logrus.Info("you might want to run the command again with --interact option")
				exit(1)
			} else if len(filteredPath) > 1 {
				logrus.Errorf("there are multiple android resource directories")
				for _, v := range filteredPath {
					logrus.Info(v)
				}
				logrus.Info("you might want to run the command again with --interact option")
				exit(1)
			} else {
				outputDir = filteredPath[0]
			}
		}

		if outputDir == "" {
			logrus.Error("no available output directory, abort")
			exit(1)
		}

		var valueFolders []string
		_, valueFolders, err = wkfs.Scan(outputDir, wkfs.WithFoldersOnly(), wkfs.WithPatterns("values"))
		if err != nil {
			logrus.Errorf("cannot walk through output directory %v, err:%v", outputDir, err)
			exit(1)
		}

		filteredPath = []string{}
		for _, f := range valueFolders {
			if wkfs.IsFile(filepath.Join(f, "strings.xml")) {
				filteredPath = append(filteredPath, f)
			}
		}

		lang2stringFolders := make(map[string]string)
		for _, v := range filteredPath {
			base := filepath.Base(v)
			if strings.EqualFold(base, "values") {
				// en
				lang2stringFolders["en"] = v
			} else {
				i := strings.IndexRune(base, '-')
				if i < 0 {
					continue
				}
				lang := base[i+1:]
				lang2stringFolders[lang] = v
			}
		}

		// STEP 3. load all source files
		logrus.Info("loading source files...")
		allSources := make(map[string]*model.SourceFile)
		var collisionResolver parser.CollisionResolver

		if interact {
			// prepare collision resolver
			collisionResolver = func(path, key, pre, cur string) string {
				var answer string
				prompt := &survey.Select{
					Message: fmt.Sprintf("key %v collision in source %v", key, path),
					Options: []string{pre, cur},
				}
				err = survey.AskOne(prompt, &answer)
				if err != nil {
					logrus.Error(err)
					exit(1)
				}
				return answer
			}
		}

		for _, v := range srcFiles {
			// iterate all source file and load them up
			source, err := parser.LoadCSV(v, collisionResolver)
			if err != nil {
				logrus.Errorf("cannot load source csv file %v, err:%v", v, err)
				exit(1)
			}
			allSources[v] = source
		}

		if viper.GetBool(flagsNoLint) {
			logrus.Info("flag 'nolint' specified, skip linting...")
		} else {
			// lint
			logrus.Info("linting...")
			sanitized := true
			for _, source := range allSources {
				lintResult := source.Lint(model.WithDefaultLinters())
				if len(lintResult) != 0 {
					sanitized = false
					logrus.Warnf("%v found %d issues", source.AbsPath, len(lintResult))
					for _, lint := range lintResult {
						logrus.Warnf("%v", lint.Desc)
					}
				}
			}
			if !sanitized {
				logrus.Warn("fix issues before continue, or add '--nolint' flag")
				return
			}
		}

		var mergeResolver model.MergeCollisionResolver
		if interact {
			mergeResolver = func(collision *model.Collision) string {
				type Entry struct {
					File    string `json:"file"`
					Content string `json:"content"`
				}
				selections := make([]string, 0, len(collision.Values))
				for i := 0; i < len(collision.Values); i++ {
					entry := &Entry{
						File:    collision.Files[i],
						Content: collision.Values[i],
					}
					var raw []byte
					raw, err = json.Marshal(entry)
					if err != nil {
						logrus.Error("cannot marshal collision entry, err: %v", err)
						exit(1)
					}

					selections = append(selections, string(raw))
				}

				var answer string
				prompt := &survey.Select{
					Message: fmt.Sprintf("key '%v' has %d collisions", collision.Key, len(collision.Values)),
					Options: selections,
				}
				err = survey.AskOne(prompt, &answer)
				if err != nil {
					logrus.Error(err)
					exit(1)
				}

				entry := &Entry{}
				err = json.Unmarshal([]byte(answer), entry)
				if err != nil {
					logrus.Error("cannot unmarshal collision entry, err: %v", err)
					exit(1)
				}

				return entry.Content
			}
		}

		srcModelList := make([]*model.SourceFile, 0, len(allSources))
		for _, src := range allSources {
			srcModelList = append(srcModelList, src)
		}
		merged := model.Merge(srcModelList, mergeResolver)

		// unescape
		if viper.GetBool(flagsNoEscape) {
			logrus.Info("flag 'noescape' specified, skip escaping")
		} else {
			logrus.Info("escaping...")
			for _, kvs := range merged {
				for k, v := range kvs {
					kvs[k] = model.EscapeString(v)
				}
			}
		}

		// auto placeholder
		if viper.GetBool(flagsAutoPlaceHolder) {
			logrus.Info("processing auto placeholder...")
			for _, kvs := range merged {
				for k, v := range kvs {
					kvs[k] = model.AutoPlaceholder(v)
				}
			}
		}

		// key-configure
		key := viper.GetString(flagsKeyConfigure)
		var messageArray []Message
		if "" != key {
			logrus.Info("key-configure path: ", key)
			if wkfs.IsFile(key) {
				data, err := os.ReadFile(key)
				if err != nil {
					logrus.Info("read file failed")
					exit(1)
				}

				fileData := windows.ByteSliceToString(data)
				dec := json.NewDecoder(strings.NewReader(fileData))

				for {
					var m Message
					if err := dec.Decode(&m); err == io.EOF {
						break
					} else if err != nil {
						logrus.Fatal(err)
						exit(1)
					}
					messageArray = append(messageArray, m)
				}
			}
		}

		// STEP 4. append to target xml files

		// collision resolve
		var appendCollisionResolver appender.CollisionResolver
		if interact {
			appendCollisionResolver = func(path string, pos int, key, old, newer string) string {
				var answer string
				prompt := &survey.Select{
					Message: fmt.Sprintf("key %v collision detected in file %v, line %d", key, path, pos),
					Options: []string{old, newer},
				}
				err = survey.AskOne(prompt, &answer)
				if err != nil {
					logrus.Error(err)
					exit(1)
				}
				return answer
			}
		} else {
			preferNewer := viper.GetBool(flagsPreferNew)
			if preferNewer {
				logrus.Info("collision resolver will accept newer over older")
			} else {
				logrus.Info("collision resolver will not change previous value")
			}
			appendCollisionResolver = func(file string, pos int, key, old, newer string) string {
				var result string
				var oldMark string
				var newMark string
				if preferNewer {
					result = newer
					oldMark = " "
					newMark = "*"
				} else {
					result = old
					oldMark = "*"
					newMark = " "
				}
				dir := filepath.Base(filepath.Dir(file))
				logrus.Infof("'%v' collision in '%v' line '%d'", key, dir, pos)
				logrus.Debugf("previous %s: %s", oldMark, old)
				logrus.Debugf("newer    %s: %s", newMark, newer)

				return result
			}
		}

		// merge
		for lang, kvs := range merged {
			if xmlFolder, ok := lang2stringFolders[translateLangKey(messageArray, lang)]; ok {
				stringFilePath := filepath.Join(xmlFolder, "strings.xml")
				logrus.Infof("appending to %v ...", stringFilePath)

				var keyCollisions, keyAppended int
				keyCollisions, keyAppended, err = appender.AppendToXML(kvs, stringFilePath, appendCollisionResolver)
				if err != nil {
					logrus.Errorf("cannot append to %v, err:%v", stringFilePath, err)
					exit(1)
				}
				logrus.Infof("%d key collisions, %d key appended", keyCollisions, keyAppended)
			} else {
				logrus.Infof("lang %v missing output resource folder, skipped", lang)
			}
		}
	},
}

func translateLangKey(keyMaps []Message, lang string) string {
	for _, keys := range keyMaps {
		for _, v := range strings.Split(keys.Alias, ",") {
			if lang == v {
				logrus.Info("translate ", lang, " key alias:", keys.Key)
				return keys.Key
			}
		}
	}
	return lang
}

func exit(num int) {
	os.Exit(num)
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
	rootCmd.AddCommand(appendCmd)

	appendCmd.Flags().StringSliceP("src", "s", []string{}, "source csv file/directories")
	appendCmd.Flags().StringP("out", "o", "", "output directory")
	appendCmd.Flags().BoolP(flagsInteract, "", false, "handle collision in an interactive mode")
	appendCmd.Flags().BoolP(flagsPreferNew, "", false, "prefer new value to old value when there are key collisions")
	appendCmd.Flags().BoolP(flagsNoLint, "", false, "ignore common mistakes in translation text (﹪, s%, $n% etc.)")
	appendCmd.Flags().BoolP(flagsNoEscape, "", false, "do not escape special characters in translation")
	appendCmd.Flags().BoolP(flagsAutoPlaceHolder, "", false, "auto check and format placeholder like %s, %AA, %BB")
	appendCmd.Flags().StringP(flagsKeyConfigure, "c", "", "key map. e.g. {\"Key\": \"en\", \"Alias\": \"English\"}")
}
