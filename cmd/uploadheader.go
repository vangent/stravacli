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

	"github.com/gocarina/gocsv"
	"github.com/spf13/cobra"
)

func init() {
	uploadHeaderCmd := &cobra.Command{
		Use:   "uploadheader",
		Short: "Print out the required header for the .csv file for upload",
		Long: `Print out the required header for the .csv file for upload.

Data Columns:
External ID: An external ID for the activity; OK to leave blank.
Activity Type: The activity type. Required. See the available list here: https://developers.strava.com/docs/reference/#api-models-ActivityType.
Name: The activity name. Required. If you leave it blank, Strava will pick one for you, like "Lunch Ride".
Description: Description of the activity.
Commute?: "false" or "true", depending on whether this activity was for a commute. Defaults to "false".
Trainer?: "false" or "true", depending on whether this activity used a trainer. Defaults to "false".
File Type: The type of data file being uploaded; one of "fit", "tcx", or "gpx"; may be suffixed with ".gz" (e.g., "gpx.gz") if the file is gzipped.
Filename: Relative path to the data file.
`,
		Args: cobra.NoArgs,
		RunE: func(_ *cobra.Command, _ []string) error {
			return doUploadHeader()
		},
	}
	rootCmd.AddCommand(uploadHeaderCmd)
}

func doUploadHeader() error {
	var noActivities []*uploadActivity
	csv, err := gocsv.MarshalString(noActivities)
	if err != nil {
		return fmt.Errorf("failed to generate .csv: %v", err)
	}
	fmt.Printf(csv)
	return nil
}
