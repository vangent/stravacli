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

	"github.com/antihax/optional"
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
		Short: "Upload modified Strava activites",
		Long:  `Upload modified Strava activites.`,
		Args:  cobra.NoArgs,
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

func doUpdate(accessToken, origFile, updatedFile string, dryRun bool) error {
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

	ctx := context.WithValue(context.Background(), strava.ContextAccessToken, accessToken)
	apiSvc := strava.NewAPIClient(strava.NewConfiguration()).ActivitiesApi

	fmt.Printf("Found %d activities in %q and %d activities in %q.\n", len(activities), updatedFile, len(orig), origFile)
	nUpdates := 0
	for i, a := range activities {
		prev := orig[a.ID]
		if prev == nil {
			return fmt.Errorf("activity ID %d from %q isn't present in %q", a.ID, updatedFile, origFile)
		}
		if *prev == *a {
			log.Printf("no change for ID %d", a.ID)
			continue
		}
		if err := updateOne(ctx, apiSvc, a, prev, dryRun); err != nil {
			return fmt.Errorf("failed to update activity %v near line %d: %v", a, i+1, err)
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

func updateOne(ctx context.Context, apiSvc *strava.ActivitiesApiService, a, prev *Activity, dryRun bool) error {
	if err := a.VerifyForUpdate(prev); err != nil {
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
