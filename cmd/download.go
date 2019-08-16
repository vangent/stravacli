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
	"time"

	"github.com/antihax/optional"
	"github.com/spf13/cobra"
	"github.com/vangent/strava"
)

const (
	pageSize  = 25           // # of activities to download per page
	dayFormat = "2006-01-02" // format for date flags
)

func init() { //
	var accessToken string
	var outFile string
	var maxActivities int
	var beforeStr, afterStr string

	downloadCmd := &cobra.Command{
		Use:   "download",
		Short: "Download Strava activites",
		Long:  `Download Strava activites.`,
		Args:  cobra.NoArgs,
		RunE: func(_ *cobra.Command, _ []string) error {
			var before, after time.Time
			var err error
			if beforeStr != "" {
				if before, err = time.Parse(dayFormat, beforeStr); err != nil {
					return fmt.Errorf("invalid --before %q (should be YYYY-MM-DD): %v", beforeStr, err)
				}
			}
			if afterStr != "" {
				if after, err = time.Parse(dayFormat, afterStr); err != nil {
					return fmt.Errorf("invalid --after %q (should be YYYY-MM-DD): %v", afterStr, err)
				}
			}
			return doDownload(accessToken, outFile, maxActivities, before, after)
		},
	}
	downloadCmd.Flags().StringVarP(&accessToken, "access_token", "t", "", "Strava access token; use the auth command to get one")
	downloadCmd.MarkFlagRequired("access_token")
	downloadCmd.Flags().StringVar(&outFile, "out", "", "output filename, or leave empty to output to stdout")
	downloadCmd.Flags().IntVar(&maxActivities, "max", 0, "maximum # of activities to download (default 0 means no limit)")
	downloadCmd.Flags().StringVar(&beforeStr, "before", "", "only download activities before this date (YYYY-MM-DD)")
	downloadCmd.Flags().StringVar(&afterStr, "after", "", "only download activities after this date (YYYY-MM-DD)")
	rootCmd.AddCommand(downloadCmd)
}

func doDownload(accessToken, outFile string, maxActivities int, before, after time.Time) error {
	ctx := context.WithValue(context.Background(), strava.ContextAccessToken, accessToken)
	cfg := strava.NewConfiguration()
	client := strava.NewAPIClient(cfg)

	page := int32(1)
	var activities []*Activity

PageLoop:
	for {
		req := &strava.GetLoggedInAthleteActivitiesOpts{
			Page:    optional.NewInt32(page),
			PerPage: optional.NewInt32(pageSize),
		}
		if !before.IsZero() {
			req.Before = optional.NewInt32(int32(before.Unix()))
		}
		if !after.IsZero() {
			req.After = optional.NewInt32(int32(after.Unix()))
		}
		summaries, _, err := client.ActivitiesApi.GetLoggedInAthleteActivities(ctx, req)
		if err != nil {
			return fmt.Errorf("failed ListActivities call (page %d, per page %d)", page, pageSize)
		}
		for _, a := range summaries {
			activity := &Activity{
				ID:          a.Id,
				Start:       a.StartDate,
				Type:        string(*a.Type_),
				Name:        a.Name,
				Description: "", // not included in ActivitySummary
				Duration:    a.ElapsedTime,
				Distance:    a.Distance,
				Private:     a.Private,
				Commute:     a.Commute,
				Trainer:     a.Trainer,
			}
			activities = append(activities, activity)
			if maxActivities != -1 && len(activities) == maxActivities {
				break PageLoop
			}
		}
		if len(activities) < pageSize {
			break
		}
		fmt.Printf("%d activities so far, fetching next %d...\n", len(activities), pageSize)
		page++
	}
	fmt.Printf("Downloaded %d activities.\n", len(activities))
	return writeCSV(outFile, activities)
}
