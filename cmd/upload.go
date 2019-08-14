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

	"github.com/spf13/cobra"
	"github.com/strava/go.strava"
)

func init() {
	var accessToken string

	uploadCmd := &cobra.Command{
		Use:   "upload",
		Short: "TODO",
		Long:  `TODO`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return upload(accessToken, args[0])
		},
	}
	uploadCmd.Flags().StringVarP(&accessToken, "access_token", "t", "", "Strava access token")
	rootCmd.AddCommand(uploadCmd)
}

func upload(accessToken, inFile string) error {
	client := strava.NewClient(accessToken)

	athleteSvc := strava.NewCurrentAthleteService(client)

	page := 1
	nActivities := 0
	for {
		fmt.Printf("Handled %d activities...", nActivities)
		activities, err := athleteSvc.ListActivities().Page(page).PerPage(30).Do()
		if err != nil {
			return err
		}
		nActivities += len(activities)
		for _, activity := range activities {
			fmt.Printf("  %s %s (%d) has private %v", activity.StartDate.Format("2006-01-02"), activity.Name, activity.Id, activity.Private)
		}
		if len(activities) < 30 {
			break
		}
		page++
	}
	fmt.Printf("Found %d activities.", nActivities)
	return nil
}
