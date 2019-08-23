# stravacli

A command-line tool for working with Strava activities.

*   Bulk upload activities, including manual activities, from a `.csv` file.
*   Bulk edit existing your activities by downloading a `.csv` file and editing
    it in an editor or spreadsheet application, then uploading the changes.

## Instructions

### Build

1.  [Install Go](https://golang.org/dl).
2.  Run

```bash
go install github.com/vangent/stravacli
```

Alternatively, you can clone this Github repository and build locally.

### Create a Strava API application

To use the CLI, you need to create your own API application. This is kind of a
pain, but it's a one-time thing, and it means that you're not giving any third
parties access to your Strava data.

Follow the instructions at
https://developers.strava.com/docs/getting-started/#b-how-to-create-an-account
to create an API application.

*   You can put anything you want for `Application Name`, `Category`, `Club`,
    and `Website`; they don't matter for this application. `http://unused.com`
    will work fine for `Website`.
*   Set the "Authorization Callback Domain" to "localhost".
*   You can use any logo you'd like;
    [here's](https://www.google.com/search?q=free+logo+download+png&tbm=isch) a
    link to download a free one.
*   Take note of the `Client ID` and `Client Secret` fields, you'll need them in
    the next step.

### Authenticate

Run

```bash
stravacli auth --client_id=<YOUR_CLIENT_ID> --client_secret=<YOUR_CLIENT_SECRET>
```

with the `Client ID` and `Client Secret` from your Strava API application. Your
browser will be redirected to Strava, where you'll need to log in (if you're not
already logged in) and authorize your application to connect to Strava. Note
that this is your own personal application, so you're not really giving anyone
besides yourself access. Once you've clicked `Authorize`, go back to your
terminal and `stravacli` will have printed out an access token for you to use
with other `stravacli` commands.

You may have to repeat this step periodically if your access token expires.

### Update Existing Activities

To bulk update existing Strava activities, first download them:

```bash
stravacli download --access_token <YOUR_ACCESS_TOKEN> --out orig.csv
```

This will download your existing activities into a file called `orig.csv`, in
`csv` format. See `stravacli download help` for more detailed help on available
flags, and what the columns mean. You can now open or import the `csv` file in a
spreadsheet application of your choice; see the bottom of the page for tips.

Edit away; all of the columns are editable except for `ID` and `Start`. Sadly,
there are a lot of fields for activities that are not editable via the Strava
API.

When you are done editing, export the data as a `.csv` file again; again, see
the bottom of the page for tips. Make sure not to clobber the original `.csv`;
the instructions below assume you name the file `updated.csv`.

Finally, use `stravacli` to apply the changes. You can use `--dryrun` to see
what changes would be made without actually making them.

```bash
stravacli update  --access_token <YOUR_ACCESS_TOKEN> --orig=orig.csv --updated updated.csv
```

See `stravacli update help` for more detailed help on available flags.

### Upload Activities

See the next section for Manual Activities; this section is for activities with
an associated `.gpx` or similar file. To bulk upload activities, first get the
required header:

```bash
stravacli uploadheader
```

See `stravacli uploadheader help` for detailed descriptions of the data columns.

Copy/paste the header data into a spreadsheet application of your choice; see
the bottom of the page for tips. Add rows for the activities you'd like to
create.

When you're done, export the data as a `.csv` file; again, see the bottom of the
page for tips. The instructions below assume you name the file `activities.md`.

Finally, use `stravacli` to upload. You can use `--dryrun` to see what changes
would be made without actually making them.

```bash
stravacli upload  --access_token <YOUR_ACCESS_TOKEN> --in=activities.csv
```

See `stravacli upload help` for more detailed help on available flags.

### Upload Manual Activities

To bulk upload manual activities, first get the required header:

```bash
stravacli uploadmanualheader
```

See `stravacli uploadmanualheader help` for detailed descriptions of the data
columns.

Copy/paste the header data into a spreadsheet application of your choice; see
the bottom of the page for tips. Add rows for the activities you'd like to
create. Note that `Duration` is in seconds, and `Distance` is in meters!

When you're done, export the data as a `.csv` file; again, see the bottom of the
page for tips. The instructions below assume you name the file `activities.md`.

Finally, use `stravacli` to upload. You can use `--dryrun` to see what changes
would be made without actually making them.

```bash
stravacli uploadmanual  --access_token <YOUR_ACCESS_TOKEN> --in=activities.csv
```

See `stravacli uploadmanual help` for more detailed help on available flags.

### Cleanup

If you are done using `stravacli`, you can revoke its API access
[here](https://www.strava.com/settings/apps).

### Editing CSV Files

There are lots of ways to edit
[CSV](https://en.wikipedia.org/wiki/Comma-separated_values) data, including:

*   Using an editor like `vi` or `emacs`.
*   In a spreadsheet application like `Google Sheets` or `Microsoft Excel`.

Since `Google Sheets` is free...
[Here](https://support.google.com/docs/answer/40608) is help on how to import a
`.csv` into Google Sheets. To export back to `.csv`, choose `File -> Download ->
Comma-separated values`.
