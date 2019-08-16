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
	"errors"
	"fmt"
	"time"
)

// Activity represents a single Strava activity.
type Activity struct {
	ID          int64     `csv:"ID (C)"`
	Start       time.Time `csv:"Start (C)"`
	Type        string    `csv:"Type (CE)"`
	Name        string    `csv:"Name (CE)"`
	Description string    `csv:"Description (C)"`
	Duration    int32     `csv:"Duration (seconds) (C)"`
	Distance    float32   `csv:"Distance (C)"`
	Private     bool      `csv:"Private? ()"`
	Commute     bool      `csv:"Commute? (CE)"`
	Trainer     bool      `csv:"Trainer? (CE)"`
}

func (a *Activity) String() string {
	s := fmt.Sprintf("%s on %s", a.Name, a.Start.Format(dayFormat))
	if a.ID != 0 {
		s += fmt.Sprintf(" (ID=%d)", a.ID)
	}
	return "[" + s + "]"
}

// VerifyForCreate checks to see that a looks like it can be uploaded
// as a manually created activity.
func (a *Activity) VerifyForCreate() error {
	if a.Start.IsZero() {
		return errors.New("missing Start")
	}
	if a.Name == "" {
		return errors.New("missing Name")
	}
	if a.Private {
		return errors.New("sorry, can't set Private")
	}
	return nil
}

// VerifyForUpdate checks to see that a looks like it can be uploaded
// as an update to prev.
func (a *Activity) VerifyForUpdate(prev *Activity) error {
	if a.ID == 0 {
		return errors.New("ID must be set for update")
	}
	if !a.Start.Equal(prev.Start) {
		return errors.New("sorry, can't modify Start")
	}
	if a.Description != "" || a.Description != prev.Description {
		return errors.New("sorry, can't modify Description")
	}
	if a.Duration != prev.Duration {
		return errors.New("sorry, can't modify Duration")
	}
	if a.Distance != prev.Distance {
		return errors.New("sorry, can't modify Distance")
	}
	if a.Private != prev.Private {
		return errors.New("sorry, can't modify Private")
	}
	return nil
}
