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
	"log"
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
	var dryRun bool

	uploadCmd := &cobra.Command{
		Use:   "upload",
		Short: "Upload new Strava activities",
		Long: `Upload new Strava activities.

See https://github.com/vangent/stravacli#upload-activities
for detailed instructions.`,
		Args: cobra.NoArgs,
		RunE: func(_ *cobra.Command, _ []string) error {
			return doUpload(accessToken, inFile, dryRun)
		},
	}
	uploadCmd.Flags().StringVarP(&accessToken, "access_token", "t", "", "Strava access token; use the auth command to get one")
	uploadCmd.MarkFlagRequired("access_token")
	uploadCmd.Flags().StringVar(&inFile, "in", "", ".csv with activities to upload")
	uploadCmd.MarkFlagRequired("in")
	uploadCmd.Flags().BoolVar(&dryRun, "dryrun", false, "do a dry run: print out proposed changes")
	rootCmd.AddCommand(uploadCmd)
}

type uploadActivity struct {
	ExternalID   string `csv:"External ID"`
	ActivityType string `csv:"Activity Type"`
	Name         string `csv:"Name"`
	Description  string `csv:"Description"`
	Commute      bool   `csv:"Commute?"`
	Trainer      bool   `csv:"Trainer?"`
	FileType     string `csv:"File Type"`
	Filename     string `csv:"Filename"`
}

func (a *uploadActivity) String() string {
	return fmt.Sprintf("[%s from %s]", a.Name, a.Filename)
}

var validActivityType = map[string]bool{
	"AlpineSki":       true,
	"BackcountrySki":  true,
	"Canoeing":        true,
	"Crossfit":        true,
	"EBikeRide":       true,
	"Elliptical":      true,
	"Golf":            true,
	"Handcycle":       true,
	"Hike":            true,
	"IceSkate":        true,
	"InlineSkate":     true,
	"Kayaking":        true,
	"Kitesurf":        true,
	"NordicSki":       true,
	"Ride":            true,
	"RockClimbing":    true,
	"RollerSki":       true,
	"Rowing":          true,
	"Run":             true,
	"Sail":            true,
	"Skateboard":      true,
	"Snowboard":       true,
	"Snowshoe":        true,
	"Soccer":          true,
	"StairStepper":    true,
	"StandUpPaddling": true,
	"Surfing":         true,
	"Swim":            true,
	"Velomobile":      true,
	"VirtualRide":     true,
	"VirtualRun":      true,
	"Walk":            true,
	"WeightTraining":  true,
	"Wheelchair":      true,
	"Windsurf":        true,
	"Workout":         true,
	"Yoga":            true,
}

var validFileType = map[string]bool{
	"fit":    true,
	"fit.gz": true,
	"tcx":    true,
	"tcx.gz": true,
	"gpx":    true,
	"gpx.gz": true,
}

// Verify checks to see that a looks like it can be uploaded.
func (a *uploadActivity) Verify() error {
	if a.ActivityType == "" {
		return errors.New("missing Activity Type")
	}
	if !validActivityType[a.ActivityType] {
		return fmt.Errorf("invalid Activity Type %q", a.ActivityType)
	}
	if a.Name == "" {
		return errors.New("missing Name")
	}
	if a.FileType == "" {
		return errors.New("missing File Type")
	}
	if !validFileType[a.FileType] {
		return fmt.Errorf("invalid File Type %q", a.FileType)
	}
	if a.Filename == "" {
		return errors.New("missing Filename")
	}
	if _, err := os.Stat(a.Filename); os.IsNotExist(err) {
		return fmt.Errorf("Filename %q not found", a.Filename)
	}
	return nil
}

func doUpload(accessToken, inFile string, dryRun bool) error {
	activities, err := loadActivitiesFromCSV(inFile)
	if err != nil {
		return err
	}

	ctx := context.WithValue(context.Background(), strava.ContextAccessToken, accessToken)
	uploadSvc := strava.NewAPIClient(strava.NewConfiguration()).UploadsApi

	fmt.Printf("Found %d activities in %q to upload.\n", len(activities), inFile)
	nUploads := 0
	for i, a := range activities {
		if err := uploadOne(ctx, uploadSvc, a, dryRun); err != nil {
			return fmt.Errorf("failed to upload activity %v near line %d: %v", a, i+1, err)
		}
		nUploads++
	}
	if dryRun {
		fmt.Printf("Found %d activities to be uploaded.\n", nUploads)
	} else {
		fmt.Printf("Uploaded %d activities.\n", nUploads)
	}
	return nil
}

func loadActivitiesFromCSV(filename string) ([]*uploadActivity, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open %q: %v", filename, err)
	}
	defer f.Close()
	var activities []*uploadActivity
	if err := gocsv.UnmarshalFile(f, &activities); err != nil {
		return nil, fmt.Errorf("failed to parse %q: %v", filename, err)
	}
	return activities, nil
}

func uploadOne(ctx context.Context, uploadSvc *strava.UploadsApiService, a *uploadActivity, dryRun bool) error {
	if err := a.Verify(); err != nil {
		return err
	}
	f, err := os.Open(a.Filename)
	if err != nil {
		return fmt.Errorf("failed to open %q: %v", a.Filename, err)
	}
	if dryRun {
		fmt.Printf("  Would upload %v...\n", a)
		return nil
	}
	fmt.Printf("  Uploading %v...\n", a)

	opts := strava.CreateUploadOpts{
		Name:     optional.NewString(a.Name),
		Type:     optional.NewString(a.ActivityType),
		DataType: optional.NewString(a.FileType),
		File:     optional.NewInterface(f),
	}
	defer f.Close()
	if a.ExternalID != "" {
		opts.ExternalId = optional.NewString(a.ExternalID)
	}
	if a.Description != "" {
		opts.Description = optional.NewString(a.Description)
	}
	if a.Trainer {
		opts.Trainer = optional.NewInt32(1)
	}
	if a.Commute {
		opts.Commute = optional.NewInt32(1)
	}
	upload, resp, err := uploadSvc.CreateUpload(ctx, &opts)
	for {
		if err != nil {
			var msg string
			if resp != nil {
				body, _ := ioutil.ReadAll(resp.Body)
				msg = string(body)
			}
			return fmt.Errorf("%v %v %s", err, resp, msg)
		}
		if upload.Error_ != "" {
			return fmt.Errorf("upload failed: %s", upload.Error_)
		}
		if upload.ActivityId != 0 {
			break
		}
		time.Sleep(1 * time.Second)
		log.Printf("    checking on status...")
		upload, resp, err = uploadSvc.GetUploadById(ctx, upload.Id)
	}
	fmt.Printf("  --> https://www.strava.com/activities/%d\n", upload.ActivityId)
	return nil
}
