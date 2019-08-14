// A utility for uploading activities exported from Runkeeper to Strava.
package main

import (
	"flag"
	"log"

	"github.com/strava/go.strava"
)

func main() {
	var accessToken string
	var dryrun bool
	var max int

	// Provide an access token, with write permissions.
	// You'll need to complete the oauth flow to get one.
	flag.StringVar(&accessToken, "token", "0ff1fe6e4529396cec77a001315adebe408b16b5", "Access Token")
	flag.BoolVar(&dryrun, "dryrun", false, "Dryrun mode")
	flag.IntVar(&max, "max", -1, "Max # to upload, or -1 for no max")
	flag.Parse()

	client := strava.NewClient(accessToken)
	athleteSvc := strava.NewCurrentAthleteService(client)
	activitiesService := strava.NewActivitiesService(client)
	page := 1
	nActivities, nUpdates := 0, 0
	for max == -1 || nUpdates <= max {
		log.Printf("Handled %d activities...", nActivities)
		activities, err := athleteSvc.ListActivities().Page(page).PerPage(30).Do()
		if err != nil {
			log.Fatal(err)
		}
		nActivities += len(activities)
		for _, activity := range activities {
			log.Printf("  %s %s (%d) has private %v", activity.StartDate.Format("2006-01-02"), activity.Name, activity.Id, activity.Private)
			if !activity.Private {
				continue
			}
			log.Printf("  updating %s %s (%d)", activity.StartDate.Format("2006-01-02"), activity.Name, activity.Id)
			nUpdates++
			if dryrun {
				continue
			}
			_, err := activitiesService.Update(activity.Id).Private(false).Do()
			if err != nil {
				log.Fatal(err)
			}
		}
		if len(activities) < 30 {
			break
		}
		page++
	}
	log.Printf("Found %d activities, updated %d of them.", nActivities, nUpdates)
}

/*
// RKActivity represents a RunKeeper activity.
type RKActivity struct {
	ActivityID       string
	Date             string
	Type             string
	RouteName        string
	Distance         string
	Duration         string
	AveragePace      string
	AverageSpeed     string
	CaloriesBurned   string
	Climb            string
	AverageHeartRate string
	Friends          string
	Notes            string
	GPXFile          string
}

func main() {
	var accessToken string
	var inputDir string
	var dryrun bool
	var skipgpx bool
	var skipmanual bool
	var max int
	var noexisting bool

	// Provide an access token, with write permissions.
	// You'll need to complete the oauth flow to get one.
	flag.StringVar(&accessToken, "token", "0ff1fe6e4529396cec77a001315adebe408b16b5", "Access Token")
	flag.StringVar(&inputDir, "input", "testinput", "Input directory")
	flag.BoolVar(&dryrun, "dryrun", false, "Dryrun mode")
	flag.BoolVar(&skipgpx, "skipgpx", false, "Skip GPX activities")
	flag.BoolVar(&skipmanual, "skipmanual", false, "Skip manual activities")
	flag.IntVar(&max, "max", -1, "Max # to upload, or -1 for no max")
	flag.BoolVar(&noexisting, "noexisting", false, "Don't get any existing activities")
	flag.Parse()

	client := strava.NewClient(accessToken)

	athleteSvc := strava.NewCurrentAthleteService(client)
	page := 1
	nActivities := 0
	existingGPX, existingManual := map[string]bool{}, map[string]bool{}
	for !noexisting {
		activities, err := athleteSvc.ListActivities().Page(page).PerPage(30).Do()
		if err != nil {
			log.Fatal(err)
		}
		for _, activity := range activities {
			if activity.ExternalId != "" {
				id := strings.TrimSuffix(activity.ExternalId, ".gpx")
				if existingGPX[id] {
					log.Printf("duplicate gpx ID: %s", id)
				}
				existingGPX[id] = true
			} else {
				start := activity.StartDate.In(time.Local)
				id := uniqueIDForManual(start, activity.Name, activity.Type)
				if existingManual[id] {
					log.Printf("duplicate manual ID: %s", id)
				}
				existingManual[id] = true
			}
		}
		nActivities += len(activities)
		if len(activities) < 30 {
			break
		}
		page++
	}
	log.Printf("Found %d existing activities (%d GPX + %d manual = %d)...", nActivities, len(existingGPX), len(existingManual), len(existingGPX)+len(existingManual))

	// Open CSV file
	f, err := os.Open(filepath.Join(inputDir, "cardioActivities.csv"))
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	lines, err := csv.NewReader(f).ReadAll()
	if err != nil {
		log.Fatal(err)
	}
	activitiesService := strava.NewActivitiesService(client)
	uploadService := strava.NewUploadsService(client)
	count := 0
	for i, line := range lines {
		if i == 0 {
			continue
		}
		rkActivity := RKActivity{line[0], line[1], line[2], line[3], line[4], line[5], line[6], line[7], line[8], line[9], line[10], line[11], line[12], line[13]}

		var did bool
		if rkActivity.GPXFile == "" {
			if skipmanual {
				continue
			}
			did, err = doManualCreate(activitiesService, &rkActivity, existingManual, dryrun)
		} else {
			if skipgpx {
				continue
			}
			did, err = doUpload(uploadService, inputDir, &rkActivity, existingGPX, dryrun)
		}
		if err != nil {
			log.Fatal(err)
		}
		if did || dryrun {
			count++
		}
		if max != -1 && count >= max {
			break
		}
	}
	log.Printf("Completed %d uploads.", count)
}

func uniqueIDForManual(start time.Time, name string, typ strava.ActivityType) string {
	return fmt.Sprintf("%s:%v:%s", start.Format("2006-01-02T15:04:05Z"), typ, name)
}

// translateActivity translates RunKeeper's activity codes to Strava's.
func translateActivity(activity string) (strava.ActivityType, string, bool) {
	switch activity {
	case "Running":
		return strava.ActivityTypes.Run, "Run", false
	case "Cycling":
		return strava.ActivityTypes.Ride, "Ride", false
	case "Mountain Biking":
		return strava.ActivityTypes.Ride, "Mountain Bike Ride", false
	case "Swimming":
		return strava.ActivityTypes.Swim, "Swim", false
	case "Elliptical":
		return strava.ActivityTypes.Elliptical, "Elliptical Workout", true
	case "Spinning":
		return strava.ActivityTypes.Ride, "Spin Workout", true
	case "Stairmaster / Stepwell":
		return strava.ActivityTypes.StairStepper, "Hill/Stairs Workout", false
	case "Yoga":
		return strava.ActivityTypes.Yoga, "Yoga Class", false
	case "Sports":
		return strava.ActivityTypes.Workout, "Ultimate/Soccer", false
	case "Circuit Training":
		return strava.ActivityTypes.WeightTraining, "Weight Training", false
	}
	log.Fatalf("Unsupported activity type: %s", activity)
	return "", "", false
}

func translateDate(date string) (time.Time, string, error) {
	t, err := time.Parse("2006-01-02 15:04:05", date)
	if err != nil {
		return time.Time{}, "", err
	}
	var tod string
	h := t.Hour()
	switch {
	case h < 5:
		tod = "Night"
	case h < 11:
		tod = "Morning"
	case h < 13:
		tod = "Lunch"
	case h < 17:
		tod = "Afternoon"
	case h < 20:
		tod = "Evening"
	default:
		tod = "Night"
	}
	return t, tod, nil
}

func translateDuration(dur string) (int, error) {
	parts := strings.Split(dur, ":")
	var intParts []int
	for _, p := range parts {
		i, err := strconv.Atoi(p)
		if err != nil {
			return 0, err
		}
		intParts = append(intParts, i)
	}
	if len(intParts) == 2 {
		return (intParts[0] * 60) + intParts[1], nil
	}
	if len(intParts) == 3 {
		return (intParts[0] * 60 * 60) + (intParts[1] * 60) + intParts[2], nil
	}
	return 0, fmt.Errorf("bad duration %q", dur)
}

func doManualCreate(svc *strava.ActivitiesService, rkActivity *RKActivity, existing map[string]bool, dryrun bool) (bool, error) {
	activityType, activityDesc, isTrainer := translateActivity(rkActivity.Type)

	distanceInMiles, err := strconv.ParseFloat(rkActivity.Distance, 10)
	if err != nil {
		return false, fmt.Errorf("invalid distance %q: %v", rkActivity.Distance, err)
	}
	distanceInMeters := distanceInMiles * 1609.344

	start, tod, err := translateDate(rkActivity.Date)
	if err != nil {
		return false, fmt.Errorf("invalid date %q: %v", rkActivity.Date, err)
	}

	elapsed, err := translateDuration(rkActivity.Duration)
	if err != nil {
		return false, err
	}
	name := fmt.Sprintf("%s %v", tod, activityDesc)
	if rkActivity.RouteName != "" {
		name += " (" + rkActivity.RouteName + ")"
	}

	uniqueID := uniqueIDForManual(start, name, activityType)
	if existing[uniqueID] {
		// log.Printf("  manual %s exists, skipping", uniqueID)
		return false, nil
	}
	existing[uniqueID] = true
	if dryrun {
		return false, nil
	}

	req := svc.Create(name, activityType, start, elapsed).
		Distance(distanceInMeters).
		ExternalId(rkActivity.ActivityID).
		Description(rkActivity.Notes)
	if isTrainer {
		req.Trainer()
	}
	if strings.Contains(name, "Commute") {
		req.Commute()
	}
	log.Printf("creating manual %q...", uniqueID)
	resp, err := req.Do()
	if err != nil {
		return false, err
	}
	if resp.Id == 0 {
		return false, fmt.Errorf("upload manual didn't return an ID: %#v", resp)
	}
	log.Printf("  -> ID %d", resp.Id)
	return true, nil
}

func doUpload(svc *strava.UploadsService, inputDir string, rkActivity *RKActivity, existing map[string]bool, dryrun bool) (bool, error) {
	if existing[rkActivity.ActivityID] {
		// log.Printf("  gpx %s exists, skipping", rkActivity.ActivityID)
		return false, nil
	}
	existing[rkActivity.ActivityID] = true
	activityType, activityDesc, isTrainer := translateActivity(rkActivity.Type)

	f, err := os.Open(filepath.Join(inputDir, rkActivity.GPXFile))
	if err != nil {
		return false, err
	}
	defer f.Close()

	_, tod, err := translateDate(rkActivity.Date)
	if err != nil {
		return false, fmt.Errorf("invalid date %q: %v", rkActivity.Date, err)
	}
	name := fmt.Sprintf("%s %v", tod, activityDesc)
	if rkActivity.RouteName != "" {
		name += " (" + rkActivity.RouteName + ")"
	}
	if dryrun {
		return false, nil
	}
	req := svc.Create(strava.FileDataTypes.GPX, rkActivity.GPXFile, f).
		ActivityType(activityType).
		ExternalId(rkActivity.ActivityID).
		Name(name).
		Description(rkActivity.Notes)
	if isTrainer {
		req.Trainer()
	}
	if strings.Contains(name, "Commute") {
		req.Commute()
	}
	log.Printf("uploading GPX %q (%s, %s)...", name, rkActivity.GPXFile, rkActivity.ActivityID)
	resp, err := req.Do()
	if err != nil {
		return false, err
	}
	if resp.Error != "" {
		return false, fmt.Errorf("upload request failed: %v", err)
	}

	status := &strava.UploadDetailed{}
	for status.ActivityId == 0 {
		time.Sleep(1 * time.Second)
		// log.Printf("  checking status...")
		status, err = svc.Get(resp.Id).Do()
		if err != nil {
			return true, fmt.Errorf("upload status check failed: %v", err)
		}
		if status.Error != "" {
			log.Printf("problem with %s: %s", rkActivity.GPXFile, status.Error)
			return false, nil
			// return true, fmt.Errorf("upload status check reported problem: %s", status.Error)
		}
	}
	log.Printf("  -> ID %d", status.ActivityId)
	return true, nil
}*/
