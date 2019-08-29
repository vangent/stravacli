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
	uploadManualHeaderCmd := &cobra.Command{
		Use:   "uploadmanualheader",
		Short: "Print out the required header for the .csv file for uploadmanual",
		Long: `Print out the required header for the .csv file for uploadmanual.

Data Columns:
Start: The start time. Required. The time format looks like YYYY-MM-DDTHH:mm:ssZ; for example, 2019-02-22T18:53:46Z".
Activity Type: The activity type. Required. See the available list here: https://developers.strava.com/docs/reference/#api-models-ActivityType.
Name: The activity name. Required. If you leave it blank, Strava will pick one for you, like "Lunch Ride".
Description: Description of the activity.
Workout Type: The type of workout. 0=default/none. For Ride: 11=Race, 12=Workout; for Run: 1=Race, 2=Long Run, 3=Workout. You can figure out other values by setting the field to what you want in Strava, then using "download" to view it.
Duration: The elapsed time, in seconds.
Distance: The distance, in meters.
Commute?: "false" or "true", depending on whether this activity was for a commute. Defaults to "false".
Trainer?: "false" or "true", depending on whether this activity used a trainer. Defaults to "false".
`,
		Args: cobra.NoArgs,
		RunE: func(_ *cobra.Command, _ []string) error {
			return doUploadManualHeader()
		},
	}
	rootCmd.AddCommand(uploadManualHeaderCmd)
}

func doUploadManualHeader() error {
	var noActivities []*manualActivity
	csv, err := gocsv.MarshalString(noActivities)
	if err != nil {
		return fmt.Errorf("failed to generate .csv: %v", err)
	}
	fmt.Printf(csv)
	return nil
}
