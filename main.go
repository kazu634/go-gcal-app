package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/user"
	"path/filepath"
	"time"

	"golang.org/x/net/context"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/calendar/v3"
)

// getClient uses a Context and Config to retrieve a Token
// then generate a Client. It returns the generated Client.
func getClient(ctx context.Context, config *oauth2.Config) *http.Client {
	cacheFile, err := tokenCacheFile()
	if err != nil {
		log.Fatalf("Unable to get path to cached credential file. %v", err)
	}
	tok, err := tokenFromFile(cacheFile)
	if err != nil {
		tok = getTokenFromWeb(config)
		saveToken(cacheFile, tok)
	}
	return config.Client(ctx, tok)
}

// getTokenFromWeb uses Config to request a Token.
// It returns the retrieved Token.
func getTokenFromWeb(config *oauth2.Config) *oauth2.Token {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("Go to the following link in your browser then type the "+
		"authorization code: \n%v\n", authURL)

	var code string
	if _, err := fmt.Scan(&code); err != nil {
		log.Fatalf("Unable to read authorization code %v", err)
	}

	tok, err := config.Exchange(oauth2.NoContext, code)
	if err != nil {
		log.Fatalf("Unable to retrieve token from web %v", err)
	}
	return tok
}

// tokenCacheFile generates credential file path/filename.
// It returns the generated credential path/filename.
func tokenCacheFile() (string, error) {
	conf := os.Getenv("CONF")

	if conf != "" {
		return conf, nil
	} else {
		usr, err := user.Current()
		if err != nil {
			return "", err
		}
		tokenCacheDir := filepath.Join(usr.HomeDir, ".credentials")
		os.MkdirAll(tokenCacheDir, 0700)
		return filepath.Join(tokenCacheDir,
			url.QueryEscape("calendar-go-quickstart.json")), err
	}
}

// tokenFromFile retrieves a Token from a given file path.
// It returns the retrieved Token and any read error encountered.
func tokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	t := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(t)
	defer f.Close()
	return t, err
}

// saveToken uses a file path to create a file and store the
// token in it.
func saveToken(file string, token *oauth2.Token) {
	fmt.Printf("Saving credential file to: %s\n", file)
	f, err := os.OpenFile(file, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Fatalf("Unable to cache oauth token: %v", err)
	}
	defer f.Close()
	json.NewEncoder(f).Encode(token)
}

func timeConv(t string) string {
	from := "2006-01-02T15:04:05-07:00"
	to := "2006/01/02 15:04"

	tmp, _ := time.Parse(from, t)

	return tmp.Format(to)
}

func main() {
	ctx := context.Background()

	b, err := ioutil.ReadFile("client_secret.json")
	if err != nil {
		log.Fatalf("Unable to read client secret file: %v", err)
	}

	// If modifying these scopes, delete your previously saved credentials
	// at ~/.credentials/calendar-go-quickstart.json
	config, err := google.ConfigFromJSON(b, calendar.CalendarReadonlyScope)
	if err != nil {
		log.Fatalf("Unable to parse client secret file to config: %v", err)
	}
	client := getClient(ctx, config)

	srv, err := calendar.New(client)
	if err != nil {
		log.Fatalf("Unable to retrieve calendar Client %v", err)
	}

	calendars, err := srv.CalendarList.List().Fields("items(id,summary)").Do()
	if err != nil {
		log.Fatalf("Unable to retrieve list of calendars: %v", err)
	}

	oneDayBefore := time.Now().Add(-24 * time.Hour).Format(time.RFC3339)
	today := time.Now().Format(time.RFC3339)

	if len(calendars.Items) > 0 {
		for _, c := range calendars.Items {
      if c.Summary != "Work" {
        continue
      }

			events, err := srv.Events.List(c.Id).ShowDeleted(true).SingleEvents(true).MaxResults(100).UpdatedMin(oneDayBefore).TimeMin(today).OrderBy("startTime").Do()
			if err != nil {
				log.Fatalf("Unable to retrieve calendar events list: %v", err)
			}

			var start, end string

			for _, e := range events.Items {
				// If the DateTime is an empty string the Event is an all-day Event.
				// So only Date is available.
				if e.Start.DateTime != "" {
					start = timeConv(e.Start.DateTime)
					end = timeConv(e.End.DateTime)
				} else {
					start = e.Start.Date
					end = e.End.Date
				}
				fmt.Printf("%s [%s - %s]\n", e.Summary, start, end)
			}
		}
	}
}