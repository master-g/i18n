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

package packer

import (
	"encoding/json"
	"errors"
	"fmt"
	"image"
	"image/png"
	"os"
	"path/filepath"
	"sort"

	"github.com/sirupsen/logrus"
)

// Config of the texture packer
type Config struct {
	OutputWidth      int
	OutputHeight     int
	Padding          int
	OutputImagePath  string
	OutputSchemaPath string
	IgnoreLargeImage bool
}

// ImageJson representation of image info
type ImageJson struct {
	X int `json:"x"`
	Y int `json:"y"`
	W int `json:"width"`
	H int `json:"height"`
}

// MetaJson representation of atlas meta info
type MetaJson struct {
	Filename string `json:"filename"`
	W        int    `json:"width"`
	H        int    `json:"height"`
	Padding  int    `json:"padding"`
}

// AtlasJson representation of all images
type AtlasJson struct {
	Meta  *MetaJson             `json:"meta"`
	Atlas map[string]*ImageJson `json:"atlas"`
}

type Packer struct {
	cfg    *Config
	root   *Node
	canvas *image.RGBA
	atlas  *AtlasJson
}

// NewPacker returns a new packer instance created with config
func NewPacker(cfg *Config) *Packer {
	if cfg == nil {
		return nil
	}

	upLeft := image.Point{X: 0, Y: 0}
	lowRight := image.Point{X: cfg.OutputWidth, Y: cfg.OutputHeight}
	canvas := image.NewRGBA(image.Rectangle{Min: upLeft, Max: lowRight})

	root := &Node{
		rc: Rectangle{
			Left:   0,
			Top:    0,
			Right:  cfg.OutputWidth,
			Bottom: cfg.OutputHeight,
		},
	}

	return &Packer{
		cfg:    cfg,
		root:   root,
		canvas: canvas,
		atlas: &AtlasJson{
			Meta: &MetaJson{
				Filename: filepath.Base(cfg.OutputImagePath),
				W:        cfg.OutputWidth,
				H:        cfg.OutputHeight,
				Padding:  cfg.Padding,
			},
			Atlas: make(map[string]*ImageJson),
		},
	}
}

func (p *Packer) Pack(images map[string]string) (err error) {
	if len(images) == 0 {
		return errors.New("empty images")
	}

	sortedPath := make([]string, 0, len(images))
	for k := range images {
		sortedPath = append(sortedPath, k)
	}
	sort.Strings(sortedPath)

	packed := 0
	for _, absPath := range sortedPath {
		logrus.Debugf("packing %v", absPath)

		// read image file
		imgInfo := NewImageInfoParseFrom(absPath, images[absPath], p.cfg.Padding)
		if imgInfo == nil {
			continue
		}

		err = p.insert(imgInfo)
		if err != nil {
			if !p.cfg.IgnoreLargeImage {
				return
			}
			logrus.Debug("ignore oversize image:", imgInfo.absolutePath)
		} else {
			packed++
		}
	}

	percentage := float32(packed) / float32(len(sortedPath))
	logrus.Infof("%v image packed (%.1f%%)", packed, percentage*100)

	// output image

	var outputFile *os.File
	outputFile, err = os.Create(p.cfg.OutputImagePath)
	defer func() {
		err = outputFile.Close()
		if err != nil {
			logrus.Warnf("cannot close file: %v, err: %v", p.cfg.OutputImagePath, err)
		}
	}()
	if err != nil {
		return
	}
	err = png.Encode(outputFile, p.canvas)

	// output atlas
	var outputAtlasFile *os.File
	outputAtlasFile, err = os.Create(p.cfg.OutputSchemaPath)
	defer func() {
		err = outputAtlasFile.Close()
		if err != nil {
			logrus.Warnf("cannot close file: %v, err: %v", p.cfg.OutputSchemaPath, err)
		}
	}()

	atlasJson, err := json.Marshal(p.atlas)
	if err != nil {
		logrus.Warnf("cannot marshal atlas json, err: %v", err)
		return
	}
	_, err = outputAtlasFile.Write(atlasJson)
	return
}

func (p *Packer) insert(img *ImageInfo) (err error) {
	node := p.root.insert(img)
	if node != nil {
		// copy pixels
		node.image = img
		img.CopyToImage(p.canvas, node.rc)

		p.atlas.Atlas[img.Name] = &ImageJson{
			X: img.Left,
			Y: img.Top,
			W: img.Width,
			H: img.Height,
		}
	} else {
		err = errors.New(fmt.Sprintf("cannot pack %v, image oversize: %vx%v", img.absolutePath, img.Width, img.Height))
	}
	return
}
