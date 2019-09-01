package main

import (
	"encoding/json"
	"fmt"
	"github.com/dghubble/go-twitter/twitter"
	"github.com/dghubble/oauth1"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

const (
	twitterBotName   = "seattletidedata"
	noaa             = "https://tidesandcurrents.noaa.gov/api/datagetter"
	seattleStationID = "9447130"
	ListenAddrPort   = ":8080"

	// ContentType constant string
	ContentType = "Content-Type"

	// ApplicationJSON is a content type for JSON data
	ApplicationJSON = "application/json"
)

func main() {
	fmt.Println("Seattle Tide Twitter Bot - @seattletidedata")
	creds := credentials{
		AccessToken:       os.Getenv("TWITTER_ACCESS_TOKEN"),
		AccessTokenSecret: os.Getenv("TWITTER_ACCESS_TOKEN_SECRET"),
		ConsumerKey:       os.Getenv("TWITTER_CONSUMER_KEY"),
		ConsumerSecret:    os.Getenv("TWITTER_CONSUMER_SECRET"),
	}

	client, err := getClient(&creds)
	if err != nil {
		log.Println("Error getting Twitter Client")
		log.Println(err)
	}

	http.HandleFunc("/", HandleRoot)
	http.HandleFunc("/health", HandleHealth)
	http.HandleFunc("/tide", HandleGetPrediction(client))

	log.Fatal(http.ListenAndServe(ListenAddrPort, nil))
}

// HandleRoot handles all routes not otherwise explicitly defined
func HandleRoot(w http.ResponseWriter, r *http.Request) {
	rMsg := "Seattle Tide"
	fmt.Fprintf(w, rMsg+" : %s\n", r.URL.Path)
	fmt.Fprint(w, "twitter bot.")
}

// HandleHealth handles health checks
func HandleHealth(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "healthy")
}

// HandlerGetPrediction gets tide predictions from NOAA data source.
func HandleGetPrediction(client *twitter.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Request: %s", r.RequestURI)

		now := time.Now()

		tide, err := getTidePrediction(now)
		if err != nil {
			fmt.Println(err)
		}

		tString, err := json.Marshal(tide)
		if err != nil {
			log.Println(err)
		}
		log.Println(string(tString))

		//Date format string: Mon Jan 2 15:04:05 MST 2006
		header := fmt.Sprintf("Seattle tide predictions for %v\n", now.Format("Monday, January 2, 2006"))

		lines := []string{header}
		for _, p := range tide.Predictions {
			lines = append(lines, fmt.Sprintf("%v    %v feet    (%v)", p.Time, p.Value, p.Type))
		}

		t := strings.Join(lines, "\n")

		// Show tweet text in logs
		fmt.Println(t)

		// Tweet
		tweet(client, t)

		w.Header().Set(ContentType, ApplicationJSON)
		fmt.Fprint(w, "tweeted")
	}
}

type credentials struct {
	ConsumerKey       string
	ConsumerSecret    string
	AccessToken       string
	AccessTokenSecret string
}

func getClient(creds *credentials) (*twitter.Client, error) {
	// Pass in your consumer key (API Key) and your Consumer Secret (API Secret)
	config := oauth1.NewConfig(creds.ConsumerKey, creds.ConsumerSecret)
	// Pass in your Access Token and your Access Token Secret
	token := oauth1.NewToken(creds.AccessToken, creds.AccessTokenSecret)

	httpClient := config.Client(oauth1.NoContext, token)
	client := twitter.NewClient(httpClient)

	// Verify Credentials
	verifyParams := &twitter.AccountVerifyParams{
		SkipStatus:   twitter.Bool(true),
		IncludeEmail: twitter.Bool(true),
	}

	// we can retrieve the user and verify if the credentials
	// we have used successfully allow us to log in!
	user, _, err := client.Accounts.VerifyCredentials(verifyParams)
	if err != nil {
		return nil, err
	}

	_ = user

	//userJSON, err := prettyJSON(user)
	//if err != nil {
	//	log.Print(err)
	//}
	//log.Printf("User's ACCOUNT:\n%v\n", string(userJSON))

	return client, nil
}

// PrettyJSON marshals and formats the JSON byte slice
func prettyJSON(obj interface{}) (jsonResponse []byte, err error) {
	jsonResponse, err = json.MarshalIndent(obj, "", "  ")
	return jsonResponse, err
}

// Tweet stuff
func tweet(client *twitter.Client, t string) {
	_, _, err := client.Statuses.Update(t, nil)
	if err != nil {
		log.Print(err)
	}

	//j, err := prettyJSON(tweet)
	//if err != nil {
	//	log.Print(err)
	//}

	//fmt.Printf("Tweet: %v \nResponse Status Code: %v\n", string(j), resp.StatusCode)
}

// ymd returns the year month day as a string
func ymd(t time.Time) string {
	year, month, day := t.Date()

	mInt := int(month)
	m := fmt.Sprintf("%d", mInt)
	if mInt < 10 {
		m = fmt.Sprintf("0%d", mInt)
	}

	d := fmt.Sprintf("%d", day)
	fmt.Println("d: " + d)
	if day < 10 {
		d = fmt.Sprintf("0%d", day)
	}

	ymd := fmt.Sprintf("%d%s%s", year, m, d)
	log.Printf("Date: %v", ymd)
	return ymd
}

func getTidePrediction(t time.Time) (*tideData, error) {
	// https://tidesandcurrents.noaa.gov/api/
	// URL: https://tidesandcurrents.noaa.gov/api/datagetter?product=predictions&application=TwitterBot&begin_date=20190830&end_date=20190830&datum=MLLW&station=9447130&time_zone=lst_ldt&units=english&interval=hilo&format=json
	// output: { "predictions" : [ {"t":"2019-08-30 04:39", "v":"11.110", "type":"H"},{"t":"2019-08-30 11:23", "v":"-2.062", "type":"L"},{"t":"2019-08-30 18:18", "v":"11.894", "type":"H"} ]}

	d := ymd(t)
	log.Printf("Date: %v", d)

	url, err := url.Parse(noaa)
	query := url.Query()
	query.Set("application", twitterBotName)
	query.Set("product", "predictions")
	query.Set("begin_date", d)
	query.Set("end_date", d)
	query.Set("datum", "MLLW")
	query.Set("station", seattleStationID)
	query.Set("time_zone", "lst_ldt")
	query.Set("units", "english")
	query.Set("interval", "hilo")
	query.Set("format", "json")
	url.RawQuery = query.Encode()

	log.Println(url.String())

	client := http.DefaultClient
	client.Timeout = 5 * time.Second

	req, _ := http.NewRequest(http.MethodGet, url.String(), nil)
	res, err := client.Do(req)
	if err != nil {
		log.Printf("%v", err)
		return nil, err
	}

	defer res.Body.Close()
	bodyBytes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Printf("%v", err)
		return nil, err
	}

	v := tideData{}
	err = json.Unmarshal(bodyBytes, &v)
	if err != nil {
		log.Println(err)
		//bstr := string(bodyBytes)
		//log.Printf("Error unmarshalling: %v \n%v", bstr, err)
		return nil, err
	}
	return &v, nil
}

type tideData struct {
	Predictions []tide `json:"predictions"`
}

type tide struct {
	Time  string `json:"t"`
	Value string `json:"v"`
	Type  string `json:"type"`
}
