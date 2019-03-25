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

	"github.com/master-g/texturepacker/internal/packer"
	"github.com/master-g/texturepacker/pkg/config"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var paramList = []config.Flag{
	{Name: "width", Type: config.Int, Shorthand: "", Value: 1024, Usage: "output image width."},
	{Name: "height", Type: config.Int, Shorthand: "", Value: 1024, Usage: "output image height."},
	{Name: "padding", Type: config.Int, Shorthand: "", Value: 1, Usage: "atlas padding."},
	{Name: "out", Type: config.String, Shorthand: "o", Value: "", Usage: "output image file path."},
	{Name: "format", Type: config.String, Shorthand: "f", Value: "png", Usage: "output image format."},
	{Name: "schema", Type: config.String, Shorthand: "s", Value: "json", Usage: "output schema format."},
	{Name: "ignore-large-image", Type: config.Bool, Shorthand: "i", Value: false, Usage: "ignore image too large to fit in the atlas."},
	{Name: "verbose", Type: config.Bool, Shorthand: "v", Value: false, Usage: "show verbose information during the packing."},
}

// packCmd represents the serve command
var packCmd = &cobra.Command{
	Use:   "pack [input images and directories]",
	Short: "pack textures",
	Long:  `pack multiple image files into a single texture image`,
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		setupLogger()
		runApplication(args)
	},
}

func init() {
	rootCmd.AddCommand(packCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// packCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// bind flags with viper
	err := config.BindFlags(packCmd, paramList)
	if err != nil {
		logrus.Errorln("unable to config texture packer:", err)
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

func filterImageFile(p string, info os.FileInfo) bool {
	lowerExt := strings.ToLower(filepath.Ext(p))
	if lowerExt == ".png" || lowerExt == ".jpg" || lowerExt == ".jpeg" || lowerExt == ".bmp" {
		return true
	}
	return false
}

func getCurrentDir() string {
	exe, err := os.Executable()
	if err != nil {
		logrus.Fatal("cannot obtain pwd for default output, err", err)
	}

	exePath := filepath.Dir(exe)
	return exePath
}

func getDefaultOutputFile(ext string) string {
	prefix := "packed_"
	pwd := getCurrentDir()
	ts := time.Now().Format("20060102T150405")
	return filepath.Join(pwd, strings.Join([]string{prefix, ts, ext}, ""))
}

func isPowerOfTwo(v int) bool {
	return v > 0 && (v&(v-1)) == 0
}

func runApplication(args []string) {
	// STEP 1 - basic validation

	width := viper.GetInt("width")
	height := viper.GetInt("height")
	if width <= 0 || height <= 0 {
		logrus.Fatalf("invalid output image size %vx%v", width, height)
	}
	if width > 4096 || height > 4096 {
		logrus.Fatal("invalid output image size, should not larger than 4096x4096")
	}
	if !isPowerOfTwo(width) || !isPowerOfTwo(height) {
		logrus.Warnf("output size %vx%v is not power of 2", width, height)
	}

	padding := viper.GetInt("padding")
	if padding < 0 {
		logrus.Fatal("invalid padding ", padding)
	}
	if padding >= width || padding >= height {
		logrus.Fatalf("invalid padding %v larger than output image %vx%v", padding, width, height)
	}

	if viper.GetString("schema") != "json" {
		logrus.Fatal("sorry, current version only supports json schema.")
	}

	outputFormat := strings.ToLower(viper.GetString("format"))
	if outputFormat != "png" {
		logrus.Fatal("sorry, current version only supports png output.")
	}

	// STEP 2 - inspect input files and input dirs

	// absolute path : filename
	imageMaps := make(map[string]string)

	var err error
	var absPath string
	// iterate all arguments
	for _, arg := range args {
		var fi os.FileInfo
		// check if file/directory is accessible
		if fi, err = os.Stat(arg); err == nil {
			switch mode := fi.Mode(); {
			case mode.IsDir():
				// walk dir
				logrus.Debug("searching ", arg, " ...")
				err = filepath.Walk(arg, func(path string, info os.FileInfo, err2 error) error {
					absPath, err = filepath.Abs(path)
					if info.Mode().IsRegular() && filterImageFile(path, info) {
						logrus.Debug("found ", absPath)
						imageMaps[absPath] = info.Name()
					}
					return nil
				})
			case mode.IsRegular():
				absPath, err = filepath.Abs(arg)
				if err == nil && filterImageFile(arg, fi) {
					imageMaps[absPath] = fi.Name()
				} else if err != nil {
					logrus.Warn("cannot obtain absolute path of ", arg)
				}
			}
		} else {
			logrus.Warn("cannot access ", arg, " err:", err)
		}
	}

	if len(imageMaps) == 0 {
		logrus.Fatal("no available input image, abort")
	}

	// STEP 3 - inspect output path, if output path is a existed directory, generates output filename for the user
	outputImagePath := viper.GetString("out")
	var outputAtlasPath string
	if outputImagePath == "" {
		outputImagePath = getDefaultOutputFile(".png")
	}
	outputExt := filepath.Ext(outputImagePath)
	withoutExt := strings.TrimSuffix(outputImagePath, outputExt)
	outputAtlasPath = strings.Join([]string{withoutExt, ".json"}, "")

	// 4. run the algorithm
	p := packer.NewPacker(&packer.Config{
		OutputWidth:      width,
		OutputHeight:     height,
		Padding:          padding,
		OutputImagePath:  outputImagePath,
		OutputSchemaPath: outputAtlasPath,
		IgnoreLargeImage: viper.GetBool("ignore-large-image"),
	})

	if p == nil {
		logrus.Fatal("cannot create packer instance")
	}

	// start
	logrus.Infof("output image size: %vx%v, with %v pixel padding", width, height, padding)
	logrus.Infof("output image format: %v", viper.GetString("format"))
	logrus.Infof("ignore oversize image: %v", viper.GetBool("ignore-large-image"))
	logrus.Infof("output atlas schema: %v", viper.GetString("schema"))
	logrus.Infof("start packing %v images", len(imageMaps))

	err = p.Pack(imageMaps)
	if err != nil {
		logrus.Fatal("cannot pack texture up, err:", err)
	}

	logrus.Infof("output image: %v", outputImagePath)
	logrus.Infof("output atlas: %v", outputAtlasPath)

	logrus.Info("goodbye, have a nice day ❤")
}
