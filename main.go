package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"gopkg.in/alecthomas/kingpin.v2"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	const withingsAPIBaseURL = "https://wbsapi.withings.net"
	const scopes = "user.info,user.metrics"
	var accessToken, refreshToken string
	var expiryTime time.Time

	clientID := kingpin.Flag("api-client-id", "Withings API OAuth client ID (https://account.withings.com/partner/add_oauth2)").Default("").OverrideDefaultFromEnvar("WITHINGS_API_CLIENT_ID").String()
	clientSecret := kingpin.Flag("api-client-secret", "Withings API OAuth client secret (https://account.withings.com/partner/add_oauth2)").Default("").OverrideDefaultFromEnvar("WITHINGS_API_CLIENT_SECRET").String()
	metricsPort := kingpin.Flag("metrics-port", "The port to bind to for serving metrics").Default("8080").OverrideDefaultFromEnvar("METRICS_PORT").Int()
	metricsScrapeInterval := kingpin.Flag("scrape-interval", "Time in seconds between scrapes").Default("1800").OverrideDefaultFromEnvar("METRICS_SCRAPE_INTERVAL").Int64()
	kingpin.Version("1.0.0")
	kingpin.Parse()

	if *clientID == "" || *clientSecret == "" {
		log.Println("Cannot talk to the Withings API. Pass `--api-client-id` and/or `--api-client-secret` flags with values. Or set `WITHINGS_API_CLIENT_ID` or `WITHINGS_API_CLIENT_SECRET` environment variables.")
		os.Exit(1)
	}

	if accessToken == "" {
		accessToken, refreshToken, expiryTime = oauthFlow(withingsAPIBaseURL, *clientID, *clientSecret, scopes, "", false)
	}

	registerMetrics()

	ticker := time.NewTicker(time.Duration(*metricsScrapeInterval) * time.Second)
	go func() {
		for {
			select {
			case <-ticker.C:
				if time.Now().After(expiryTime) {
					log.Println("Refreshing credentials...")
					accessToken, refreshToken, expiryTime = oauthFlow(withingsAPIBaseURL, *clientID, *clientSecret, scopes, refreshToken, true)
				}

				log.Println("Updating data...")
				updateMetrics(
					currentWeightMetric, getMeasurements(withingsAPIBaseURL, accessToken, "weight"),
					hydrationMetric, getMeasurements(withingsAPIBaseURL, accessToken, "hydration"),
				)
			}
		}
	}()

	log.Println("Getting initial values...")
	updateMetrics(
		currentWeightMetric, getMeasurements(withingsAPIBaseURL, accessToken, "weight"),
		hydrationMetric, getMeasurements(withingsAPIBaseURL, accessToken, "hydration"),
	)

	http.Handle("/metrics", promhttp.Handler())
	log.Printf("Serving metrics on http://localhost:%d/metrics. Configure your Prometheus to scrape accordingly.", *metricsPort)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", *metricsPort), nil))
}

func oauthFlow(withingsAPIBaseURL string, clientID string, clientSecret string, scopes string, refreshToken string, isRefresh bool) (string, string, time.Time) {
	var url string

	if !isRefresh {
		authCode := ""
		fmt.Printf("Go to https://account.withings.com/oauth2_user/authorize2?response_type=code&client_id=%s&scope=%s&state=issyl0-withings&redirect_uri=http://localhost\n", clientID, scopes)
		fmt.Println("Enter the value of `code` from the returned query string:")
		fmt.Scanln(&authCode)

		url = fmt.Sprintf("%s/v2/oauth2?action=requesttoken&grant_type=authorization_code&client_id=%s&client_secret=%s&code=%s&redirect_uri=http://localhost", withingsAPIBaseURL, clientID, clientSecret, authCode)
	} else {
		url = fmt.Sprintf("%s/v2/oauth2?action=requesttoken&grant_type=refresh_token&client_id=%s&client_secret=%s&refresh_token=%s&redirect_uri=http://localhost", withingsAPIBaseURL, clientID, clientSecret, refreshToken)
	}

	client := &http.Client{}
	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		fmt.Println(err)
	}

	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
	}

	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)

	parsedRequestToken := RequestToken{}
	json.Unmarshal(body, &parsedRequestToken)

	expiryTime := tokenExpiryTime(time.Now(), parsedRequestToken.Body.ExpiresIn)

	return parsedRequestToken.Body.AccessToken, parsedRequestToken.Body.RefreshToken, expiryTime
}

func tokenExpiryTime(issuedTime time.Time, expiresIn int64) time.Time {
	return issuedTime.Add(time.Second * time.Duration(expiresIn))
}

func getMeasurements(withingsAPIBaseURL string, accessToken string, measurementType string) float64 {
	var measurementAPIType int
	switch measurementType {
	case "weight":
		measurementAPIType = 1
	case "hydration":
		measurementAPIType = 77
	}

	url := fmt.Sprintf("%s/measure?action=getmeas&meastypes=%d&category=1&lastupdate=integer", withingsAPIBaseURL, measurementAPIType)
	method := "POST"

	client := &http.Client{}
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		fmt.Println(err)
	}
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", accessToken))

	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
	}

	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)

	parsedMeasures := Measures{}
	json.Unmarshal(body, &parsedMeasures)

	if measurementType == "weight" || measurementType == "hydration" {
		return parsedMeasures.Body.MeasureGroups[0].Measures[0].Value / 1000
	}

	return 0.0
}

func updateMetrics(currentWeightMetric prometheus.Gauge, currentWeight float64, hydrationMetric prometheus.Gauge, hydration float64) {
	currentWeight, err := strconv.ParseFloat(fmt.Sprintf("%.1f", currentWeight), 64)
	hydration, err = strconv.ParseFloat(fmt.Sprintf("%.1f", hydration), 64)

	if err != nil {
		fmt.Println(err)
	}

	log.Printf("Setting withings_current_weight metric to %.1f kg.\n", currentWeight)
	currentWeightMetric.Set(currentWeight)
	log.Printf("Setting withings_current_hydration metric to %.1f/kg.\n", hydration)
	hydrationMetric.Set(hydration)
}
