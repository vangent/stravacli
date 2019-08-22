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
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/antihax/optional"
	"github.com/gocarina/gocsv"
	"github.com/spf13/cobra"
	"github.com/vangent/strava"
)

func init() {
	var accessToken string
	var origFile string
	var updatedFile string
	var dryRun bool

	updateCmd := &cobra.Command{
		Use:   "update",
		Short: "Upload modified Strava activities",
		Long: `Upload modified Strava activities.

See https://github.com/vangent/stravacli/#bulk-update-existing-activities
for detailed instructions.`,
		Args: cobra.NoArgs,
		RunE: func(_ *cobra.Command, _ []string) error {
			return doUpdate(accessToken, origFile, updatedFile, dryRun)
		},
	}
	updateCmd.Flags().StringVarP(&accessToken, "access_token", "t", "", "Strava access token; use the auth command to get one")
	updateCmd.MarkFlagRequired("access_token")
	updateCmd.Flags().StringVar(&origFile, "orig", "", "original .csv file from download")
	updateCmd.MarkFlagRequired("orig")
	updateCmd.Flags().StringVar(&updatedFile, "updated", "", ".csv with modifications")
	updateCmd.MarkFlagRequired("updated")
	updateCmd.Flags().BoolVar(&dryRun, "dryrun", false, "do a dry run: print out proposed changes")
	rootCmd.AddCommand(updateCmd)
}

func loadUpdatableActivitiesFromCSV(filename string) ([]*updatableActivity, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open %q: %v", filename, err)
	}
	defer f.Close()
	var activities []*updatableActivity
	if err := gocsv.UnmarshalFile(f, &activities); err != nil {
		return nil, fmt.Errorf("failed to parse %q: %v", filename, err)
	}
	return activities, nil
}

func doUpdate(accessToken, origFile, updatedFile string, dryRun bool) error {
	activities, err := loadUpdatableActivitiesFromCSV(origFile)
	if err != nil {
		return err
	}
	orig := map[int64]*updatableActivity{}
	for _, a := range activities {
		orig[a.ID] = a
	}

	activities, err = loadUpdatableActivitiesFromCSV(updatedFile)
	if err != nil {
		return err
	}

	if len(activities) != len(orig) {
		return fmt.Errorf("%q has %d activities, but %q has %d; for update, they should be the same", origFile, len(orig), updatedFile, len(activities))
	}
	ctx := context.WithValue(context.Background(), strava.ContextAccessToken, accessToken)
	apiSvc := strava.NewAPIClient(strava.NewConfiguration()).ActivitiesApi

	fmt.Printf("Found %d activities....\n", len(activities))
	nUpdates := 0
	for i, a := range activities {
		prev := orig[a.ID]
		if prev == nil {
			return fmt.Errorf("activity ID %d from %q not found in %q", a.ID, updatedFile, origFile)
		}
		if *prev == *a {
			log.Printf("no change for ID %d", a.ID)
			continue
		}
		if err := updateOne(ctx, apiSvc, a, prev, dryRun); err != nil {
			return fmt.Errorf("failed to update activity %v on line %d: %v", a, i+1, err)
		}
		nUpdates++
	}
	if dryRun {
		fmt.Printf("Found %d activities to be updated.\n", nUpdates)
	} else {
		fmt.Printf("Updated %d activities.\n", nUpdates)
	}
	return nil
}

func updateOne(ctx context.Context, apiSvc *strava.ActivitiesApiService, a, prev *updatableActivity, dryRun bool) error {
	if err := a.Verify(prev); err != nil {
		return err
	}
	if dryRun {
		fmt.Printf("  Would update %v...\n", a)
		return nil
	}
	fmt.Printf("  Updating %v...\n", a)
	activityType := strava.ActivityType(a.Type)
	update := strava.UpdatableActivity{
		Commute: a.Commute,
		Trainer: a.Trainer,
		Name:    a.Name,
		Type_:   &activityType,
	}
	detailedActivity, resp, err := apiSvc.UpdateActivityById(ctx, a.ID, &strava.UpdateActivityByIdOpts{Body: optional.NewInterface(update)})
	if err != nil {
		var msg string
		if resp != nil {
			body, _ := ioutil.ReadAll(resp.Body)
			msg = string(body)
		}
		return fmt.Errorf("%v %s", err, msg)
	}
	fmt.Printf("  --> https://www.strava.com/activities/%d\n", detailedActivity.Id)
	return nil
}
