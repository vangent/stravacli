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

	"github.com/dwj300/strava"
	"github.com/spf13/cobra"
)

func init() {
	var accessToken string
	var inFile string
	var dryRun bool

	uploadCmd := &cobra.Command{
		Use:   "upload",
		Short: "TODO",
		Long:  `TODO`,
		Args:  cobra.NoArgs,
		RunE: func(_ *cobra.Command, _ []string) error {
			return upload(accessToken, inFile, dryRun)
		},
	}
	uploadCmd.Flags().StringVarP(&accessToken, "access_token", "t", "", "Strava access token")
	uploadCmd.MarkFlagRequired("access_token")
	uploadCmd.Flags().BoolVar(&dryRun, "dry_run", false, "dry run")
	rootCmd.AddCommand(uploadCmd)
}

func upload(accessToken, inFile string, dryrun bool) error {
	ctx := context.WithValue(context.Background(), strava.ContextAccessToken, accessToken)
	cfg := strava.NewConfiguration()
	client := strava.NewAPIClient(cfg)
	_ = client
	_ = ctx
	return nil
}
