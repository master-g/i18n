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

package config

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// Type indicates the value type of a flag
type Type int

const (
	// Bool value type for bool
	Bool Type = 1
	// Int value type for int
	Int Type = 2
	// String value type for string
	String Type = 3
	// StringSlice value type for []string
	StringSlice Type = 4
)

// Flag structure for flags in configuration
type Flag struct {
	Name      string
	Type      Type
	Shorthand string
	Value     interface{}
	Usage     string
}

// BindFlags will bind each flag to cobra's command, so it can extract them
func BindFlags(cmd *cobra.Command, flags []Flag) (err error) {
	cmdFlags := cmd.PersistentFlags()
	for _, f := range flags {
		switch f.Type {
		case Bool:
			cmdFlags.BoolP(f.Name, f.Shorthand, f.Value.(bool), f.Usage)
		case Int:
			cmdFlags.IntP(f.Name, f.Shorthand, f.Value.(int), f.Usage)
		case String:
			cmdFlags.StringP(f.Name, f.Shorthand, f.Value.(string), f.Usage)
		case StringSlice:
			cmdFlags.StringSliceP(f.Name, f.Shorthand, f.Value.([]string), f.Usage)
		}
		err = viper.BindPFlag(f.Name, cmdFlags.Lookup(f.Name))
		if err != nil {
			break
		}
	}

	return
}
