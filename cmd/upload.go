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
	"log"
	"os"

	"github.com/gocarina/gocsv"
	"github.com/spf13/cobra"
)

func init() {
	var accessToken string
	var origFile string
	var updatedFile string
	var dryRun bool

	uploadCmd := &cobra.Command{
		Use:   "upload",
		Short: "Upload new and/or modified Strava activites",
		Long:  `Upload new and/or modified Strava activites.`,
		Args:  cobra.NoArgs,
		RunE: func(_ *cobra.Command, _ []string) error {
			return upload(accessToken, origFile, updatedFile, dryRun)
		},
	}
	uploadCmd.Flags().StringVarP(&accessToken, "access_token", "t", "", "Strava access token; use the auth command to get one")
	uploadCmd.MarkFlagRequired("access_token")
	uploadCmd.Flags().StringVar(&origFile, "orig", "", "original .csv file from download")
	uploadCmd.MarkFlagRequired("orig")
	uploadCmd.Flags().StringVar(&updatedFile, "updated", "", ".csv with modifications")
	uploadCmd.MarkFlagRequired("updated")
	uploadCmd.Flags().BoolVar(&dryRun, "dry_run", false, "do a dry run: print out proposed changes")
	rootCmd.AddCommand(uploadCmd)
}

func upload(accessToken, origFile, updatedFile string, dryRun bool) error {

	var orig map[int64]*Activity
	activities, err := loadCSV(origFile)
	if err != nil {
		return err
	}
	orig = map[int64]*Activity{}
	for _, a := range activities {
		orig[a.ID] = a
	}

	activities, err = loadCSV(updatedFile)
	if err != nil {
		return err
	}

	fmt.Printf("Loaded %d updated activities and %d original activities.\n", len(activities), len(orig))
	nUpdates, nCreates := 0, 0
	for _, a := range activities {
		if a.ID == 0 {
			// Manual upload.
			// TODO: Improve dryRun printout.
			// TODO: Manual upload.
			// TODO: Check for possible duplicates?
			fmt.Printf("Would create: %v\n", a)
			nCreates++
			continue
		}
		// Possible update.
		prev := orig[a.ID]
		if prev == nil {
			return fmt.Errorf("activity ID %d from %q isn't present in %q", a.ID, updatedFile, origFile)
		}
		if *prev == *a {
			log.Printf("no change for ID %d", a.ID)
			continue
		}
		// TODO: Improve dryRun printout.
		// TODO: UpdateActivity.
		fmt.Printf("Would update ID %d: DELTA\n", a.ID)
		nUpdates++
	}
	fmt.Printf("Found %d updates and %d creates.\n", nUpdates, nCreates)
	return nil
}

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
