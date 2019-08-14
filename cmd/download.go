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
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/spf13/cobra"
	"github.com/strava/go.strava"
)

func init() {
	var accessToken string
	var outFile string
	var maxActivities int

	downloadCmd := &cobra.Command{
		Use:   "download",
		Short: "Download Strava activites to a .csv file",
		Long:  `Download Strava activites to a .csv file.`,
		Args:  cobra.NoArgs,
		RunE: func(_ *cobra.Command, _ []string) error {
			return download(accessToken, outFile, maxActivities)
		},
	}
	downloadCmd.Flags().StringVarP(&accessToken, "access_token", "t", "", "Strava access token")
	downloadCmd.MarkFlagRequired("access_token")
	downloadCmd.Flags().StringVar(&outFile, "out", "", "output filename, or leave empty to output to stdout")
	downloadCmd.Flags().IntVar(&maxActivities, "max", 0, "maximum # of activities to download (default 0 means no limit)")
	rootCmd.AddCommand(downloadCmd)
}

const pageSize = 30

func download(accessToken, outFile string, maxActivities int) error {
	client := strava.NewClient(accessToken)
	athleteSvc := strava.NewCurrentAthleteService(client)

	var w io.Writer
	if outFile == "" {
		w = os.Stdout
	} else {
		f, err := os.Create(outFile)
		if err != nil {
			return fmt.Errorf("failed to open output file %q: %v", outFile, err)
		}
		defer f.Close()
		w = f
	}
	out := csv.NewWriter(w)
	defer out.Flush()

	header := []string{
		"ID",
		"StartDate",
		"Name",
	}
	out.Write(header)
	out.Flush()
	page := 1
	n := 0
PageLoop:
	for {
		activities, err := athleteSvc.ListActivities().Page(page).PerPage(pageSize).Do()
		if err != nil {
			return fmt.Errorf("failed ListActivities call (page %d, per page %d)", page, pageSize)
		}
		for _, activity := range activities {
			/*
				&strava.ActivitySummary{Id:2616421579, ExternalId:"", UploadId:0, Name:"Lunch Spin Workout", Distance:0, MovingTime:1800, ElapsedTime:1800, TotalElevationGain:0, Type:"Ride", StartDate, Trainer:true, Commute:false, Manual:true, Private:false, Flagged:false, GearId:"", AverageSpeed:0, MaximunSpeed:0, AverageCadence:0, AverageTemperature:0, AveragePower:0, WeightedAveragePower:0, Kilojoules:0, DeviceWatts:false, AverageHeartrate:0, MaximumHeartrate:0, Truncated:0, HasKudoed:false}
			*/
			row := []string{
				fmt.Sprintf("%d", activity.Id),
				activity.StartDate.Format("2006-01-02"),
				activity.Name,
			}
			out.Write(row)
			out.Flush()
			n++
			if maxActivities != -1 && n == maxActivities {
				break PageLoop
			}
		}
		if len(activities) < pageSize {
			break
		}
		log.Printf("Handled %d activities, fetching next page...", n)
		page++
	}
	log.Printf("Downloaded %d activities.", n)
	return nil
}
