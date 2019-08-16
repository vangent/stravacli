/*
Copyright © 2019 Robert van Gent (vangent@gmail.com)

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/

package cmd

import (
	"fmt"
	"io"
	"os"

	"github.com/gocarina/gocsv"
)

func loadCSV(filename string) ([]*Activity, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open %q: %v", filename, err)
	}
	defer f.Close()
	var activities []*Activity
	if err := gocsv.UnmarshalFile(f, &activities); err != nil {
		return nil, fmt.Errorf("failed to parse %q: %v", filename, err)
	}
	return activities, nil
}

func writeCSV(filename string, activities []*Activity) error {
	var w io.Writer
	if filename == "" {
		w = os.Stdout
	} else {
		f, err := os.Create(filename)
		if err != nil {
			return fmt.Errorf("failed to open output file %q: %v", filename, err)
		}
		defer f.Close()
		w = f
	}
	csv, err := gocsv.MarshalString(activities)
	if err != nil {
		return fmt.Errorf("failed to generate .csv: %v", err)
	}
	fmt.Fprintf(w, csv)
	return nil
}
