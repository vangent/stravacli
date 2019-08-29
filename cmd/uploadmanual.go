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
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"time"

	"github.com/antihax/optional"
	"github.com/gocarina/gocsv"
	"github.com/spf13/cobra"
	"github.com/vangent/strava"
)

func init() {
	var accessToken string
	var inFile string
	var startRow int
	var dryRun bool

	uploadManualCmd := &cobra.Command{
		Use:   "uploadmanual",
		Short: "Upload new manual Strava activities",
		Long: `Upload new manual Strava activities.

See https://github.com/vangent/stravacli#upload-manual-activities
for detailed instructions.`,
		Args: cobra.NoArgs,
		RunE: func(_ *cobra.Command, _ []string) error {
			return checkPartialSuccess(doUploadManual(accessToken, inFile, startRow, dryRun))
		},
	}
	uploadManualCmd.Flags().StringVarP(&accessToken, "access_token", "t", "", "Strava access token; use the auth command to get one")
	uploadManualCmd.MarkFlagRequired("access_token")
	uploadManualCmd.Flags().StringVar(&inFile, "in", "", ".csv with activities to upload")
	uploadManualCmd.MarkFlagRequired("in")
	uploadManualCmd.Flags().IntVar(&startRow, "start_row", 1, "skip rows in the input up to this row (row 0 is the header row)")
	uploadManualCmd.Flags().BoolVar(&dryRun, "dryrun", false, "do a dry run: print out proposed changes")
	rootCmd.AddCommand(uploadManualCmd)
}

type manualActivity struct {
	Start        time.Time `csv:"Start"`
	ActivityType string    `csv:"Activity Type"`
	Name         string    `csv:"Name"`
	Description  string    `csv:"Description"`
	WorkoutType  int       `csv:"Workout Type"`
	Duration     int32     `csv:"Duration"`
	Distance     float32   `csv:"Distance"`
	Commute      bool      `csv:"Commute?"`
	Trainer      bool      `csv:"Trainer?"`
}

func (a *manualActivity) String() string {
	return fmt.Sprintf("[%s on %s]", a.Name, a.Start.Format(dayFormat))
}

// Verify checks to see that a looks like it can be uploaded.
func (a *manualActivity) Verify() error {
	if a.Start.IsZero() {
		return errors.New("missing Start")
	}
	if a.ActivityType == "" {
		return errors.New("missing Activity Type")
	}
	if !validActivityType[a.ActivityType] {
		return fmt.Errorf("invalid Activity Type %q", a.ActivityType)
	}
	if a.Name == "" {
		return errors.New("missing Name")
	}
	return nil
}

func doUploadManual(accessToken, inFile string, startRow int, dryRun bool) (int, error) {
	activities, err := loadManualActivitiesFromCSV(inFile)
	if err != nil {
		return 0, err
	}

	ctx := context.WithValue(context.Background(), strava.ContextAccessToken, accessToken)
	apiSvc := strava.NewAPIClient(strava.NewConfiguration()).ActivitiesApi

	fmt.Printf("Found %d manual activities in %q to upload%s....\n", len(activities), inFile, startRowMessage(len(activities), startRow))
	n := 0
	for i, a := range activities {
		row := i + 1 // row 0 is the header row
		if row < startRow {
			continue
		}
		if err := uploadManualOne(ctx, apiSvc, a, dryRun); err != nil {
			return row, fmt.Errorf("failed to upload manual activity %v: %v", a, err)
		}
		n++
	}
	if !dryRun {
		fmt.Printf("Uploaded %d manual activities.\n", n)
	}
	return 0, nil
}

func loadManualActivitiesFromCSV(filename string) ([]*manualActivity, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open %q: %v", filename, err)
	}
	defer f.Close()
	var activities []*manualActivity
	if err := gocsv.UnmarshalFile(f, &activities); err != nil {
		return nil, fmt.Errorf("failed to parse %q: %v", filename, err)
	}
	return activities, nil
}

func uploadManualOne(ctx context.Context, apiSvc *strava.ActivitiesApiService, a *manualActivity, dryRun bool) error {
	if err := a.Verify(); err != nil {
		return err
	}
	if dryRun {
		fmt.Printf("  Would upload %v...\n", a)
		return nil
	}
	fmt.Printf("  Uploading %v...\n", a)
	opts := strava.CreateActivityOpts{}
	if a.Description != "" {
		opts.Description = optional.NewString(a.Description)
	}
	if a.Distance != 0 {
		opts.Distance = optional.NewFloat32(a.Distance)
	}
	if a.Trainer {
		opts.Trainer = optional.NewInt32(1)
	}
	if a.Commute {
		opts.Commute = optional.NewInt32(1)
	}
	if a.WorkoutType != 0 {
		opts.WorkoutType = optional.NewInt32(int32(a.WorkoutType))
	}
	detailedActivity, resp, err := apiSvc.CreateActivity(ctx, a.Name, a.ActivityType, a.Start, a.Duration, &opts)
	if err != nil {
		var msg string
		if resp != nil {
			body, _ := ioutil.ReadAll(resp.Body)
			msg = string(body)
		}
		return fmt.Errorf("%v %s", err, msg)
	}
	if detailedActivity.Id != 0 {
		fmt.Printf("  --> https://www.strava.com/activities/%d\n", detailedActivity.Id)
	}
	return nil
}
