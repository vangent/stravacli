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

	"github.com/antihax/optional"
	"github.com/spf13/cobra"
	"github.com/vangent/strava"
)

func init() {
	var accessToken string
	var inFile string
	var outFile string
	var dryRun bool

	createCmd := &cobra.Command{
		Use:   "create",
		Short: "Upload new Strava activites",
		Long:  `Upload new Strava activites.`,
		Args:  cobra.NoArgs,
		RunE: func(_ *cobra.Command, _ []string) error {
			return doCreate(accessToken, inFile, outFile, dryRun)
		},
	}
	createCmd.Flags().StringVarP(&accessToken, "access_token", "t", "", "Strava access token; use the auth command to get one")
	createCmd.MarkFlagRequired("access_token")
	createCmd.Flags().StringVar(&inFile, "in", "", ".csv with activities to upload")
	createCmd.MarkFlagRequired("in")
	createCmd.Flags().StringVar(&outFile, "out", "", ".csv with IDs filled in for succesful uploads")
	createCmd.Flags().BoolVar(&dryRun, "dryrun", false, "do a dry run: print out proposed changes")
	rootCmd.AddCommand(createCmd)
}

func doCreate(accessToken, inFile, outFile string, dryRun bool) error {
	activities, err := loadCSV(inFile)
	if err != nil {
		return err
	}
	defer func() {
		writeCSV(outFile, activities)
	}()

	ctx := context.WithValue(context.Background(), strava.ContextAccessToken, accessToken)
	apiSvc := strava.NewAPIClient(strava.NewConfiguration()).ActivitiesApi

	fmt.Printf("Found %d activities in %q to upload.\n", len(activities), inFile)
	nCreates := 0
	for i, a := range activities {
		if a.ID != 0 {
			fmt.Printf("activity %d near line %d has already been uploaded, skipping\n", a.ID, i+1)
			continue
		}
		if err := createOne(ctx, apiSvc, a, dryRun); err != nil {
			return fmt.Errorf("failed to create activity %v near line %d: %v", a, i+1, err)
		}
		nCreates++
	}
	if dryRun {
		fmt.Printf("Found %d activities to be created.\n", nCreates)
	} else {
		fmt.Printf("Created %d activities.\n", nCreates)
	}
	return nil
}

func createOne(ctx context.Context, apiSvc *strava.ActivitiesApiService, a *Activity, dryRun bool) error {
	if err := a.VerifyForCreate(); err != nil {
		return err
	}
	if dryRun {
		fmt.Printf("  Would create %v...\n", a)
		return nil
	}
	fmt.Printf("  Creating %v...\n", a)
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
	detailedActivity, resp, err := apiSvc.CreateActivity(ctx, a.Name, a.Type, a.Start, a.Duration, &opts)
	if err != nil {
		var msg string
		if resp != nil {
			body, _ := ioutil.ReadAll(resp.Body)
			msg = string(body)
		}
		return fmt.Errorf("%v %s", err, msg)
	}
	fmt.Printf("  --> https://www.strava.com/activities/%d\n", detailedActivity.Id)
	a.ID = detailedActivity.Id
	return nil
}
