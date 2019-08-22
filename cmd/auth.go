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
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"

	"github.com/skratchdot/open-golang/open"
	"github.com/spf13/cobra"
)

func init() {
	var clientID string
	var clientSecret string
	var port int
	var readOnly bool

	authCmd := &cobra.Command{
		Use:   "auth",
		Short: "Get a Strava access token",
		Long: `Get a Strava access token. See https://github.com/vangent/stravacli
for detailed instructions.`,
		Args: cobra.NoArgs,
		RunE: func(_ *cobra.Command, _ []string) error {
			return doAuth(clientID, clientSecret, port, readOnly)
		},
	}
	authCmd.Flags().StringVar(&clientID, "client_id", "", "Strava client ID from https://www.strava.com/settings/api")
	authCmd.MarkFlagRequired("client_id")
	authCmd.Flags().StringVar(&clientSecret, "client_secret", "", "Strava client secret from https://www.strava.com/settings/api")
	authCmd.MarkFlagRequired("client_secret")
	authCmd.Flags().IntVar(&port, "port", 8080, "port to run local server on")
	authCmd.Flags().BoolVar(&readOnly, "read_only", false, "get a read-only token")
	rootCmd.AddCommand(authCmd)
}

// doAuth performs the oauth authentication workflow.
// https://developers.strava.com/docs/authentication/
func doAuth(clientID, clientSecret string, port int, readOnly bool) error {
	type authResult struct {
		err  error
		code string
	}

	ch := make(chan *authResult)
	go func() {
		http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintf(w, "You can now close this window.")

			res := &authResult{}
			defer func() { ch <- res }()

			q := r.URL.Query()
			if err := q.Get("error"); err != "" {
				res.err = fmt.Errorf("authorization failed: %s", err)
				return
			}
			res.code = q.Get("code")
			if res.code == "" {
				res.err = fmt.Errorf("authorization didn't include a code: %q", r.URL.String())
			}
		})
		http.ListenAndServe(fmt.Sprintf("127.0.0.1:%d", port), nil)
	}()

	u, _ := url.Parse("https://www.strava.com/oauth/authorize")
	q := u.Query()
	q.Add("client_id", clientID)
	q.Add("redirect_uri", fmt.Sprintf("http://127.0.0.1:%d", port))
	q.Add("response_type", "code")
	if readOnly {
		q.Add("scope", "activity:read_all")
	} else {
		q.Add("scope", "activity:read_all,activity:write")
	}
	u.RawQuery = q.Encode()
	urlstr := u.String()

	fmt.Printf("Pointing your browser to %s. If it doesn't work, please copy the URL and paste it into your browser.\n", urlstr)
	if err := open.Start(urlstr); err != nil {
		return err
	}

	// Wait for the redirect.
	res := <-ch
	if res.err != nil {
		return res.err
	}
	log.Printf("got code %s", res.code)

	form := url.Values{}
	form.Set("client_id", clientID)
	form.Set("client_secret", clientSecret)
	form.Set("code", res.code)
	resp, err := http.PostForm("https://www.strava.com/oauth/token", form)
	if err != nil {
		return fmt.Errorf("authentication failed at POST: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("authentication failed at POST, status code %d", resp.StatusCode)
	}
	body, err := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	log.Printf("POST body: %s", string(body))
	m := map[string]interface{}{}
	if err := json.Unmarshal(body, &m); err != nil {
		return fmt.Errorf("authentication failed, POST response was not JSON: %v", err)
	}
	accessToken := m["access_token"]
	if accessToken == "" {
		return fmt.Errorf("authentication failed, no access token received: %v", m)
	}
	if athlete, ok := m["athlete"].(map[string]interface{}); ok {
		fmt.Printf("Hello, %s %s!\n", athlete["firstname"], athlete["lastname"])
	}
	fmt.Printf("Your Strava access token is: %s\n", m["access_token"])
	return nil
}
