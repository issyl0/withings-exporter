package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	const withingsAPIBaseURL = "https://wbsapi.withings.net"
	const scopes = "user.info,user.metrics"
	var accessToken, refreshToken string
	var expiryTime time.Time

	clientID, clientSecret := checkForAPIClientCredentials()

	if accessToken == "" {
		accessToken, refreshToken, expiryTime = oauthFlow(withingsAPIBaseURL, clientID, clientSecret, scopes, "", false)
	}

	registerMetrics()

	ticker := time.NewTicker(1800 * time.Second)
	go func() {
		for {
			select {
			case <-ticker.C:
				if time.Now().After(expiryTime) {
					log.Println("Refreshing credentials...")
					accessToken, refreshToken, expiryTime = oauthFlow(withingsAPIBaseURL, clientID, clientSecret, scopes, refreshToken, true)
				}

				log.Println("Updating data...")
				updateMetrics(currentWeightMetric, getWeightMeasurements(withingsAPIBaseURL, accessToken))
			}
		}
	}()

	log.Println("Getting initial values...")
	updateMetrics(currentWeightMetric, getWeightMeasurements(withingsAPIBaseURL, accessToken))

	http.Handle("/metrics", promhttp.Handler())
	log.Println("Serving metrics on http://localhost:8080/metrics. Configure your Prometheus to scrape accordingly.")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func checkForAPIClientCredentials() (string, string) {
	clientID := os.Getenv("WITHINGS_APP_CLIENT_ID")
	clientSecret := os.Getenv("WITHINGS_APP_CLIENT_SECRET")

	if clientID == "" || clientSecret == "" {
		fmt.Println("Set your Withings API application up with `WITHINGS_APP_CLIENT_ID` and `WITHINGS_APP_CLIENT_SECRET` envvars.")
		os.Exit(1)
	}

	return clientID, clientSecret
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

func getWeightMeasurements(withingsAPIBaseURL string, accessToken string) float64 {
	var weightMeasurementAPITypes = 1
	url := fmt.Sprintf("%s/measure?action=getmeas&meastypes=%d&category=1&lastupdate=integer", withingsAPIBaseURL, weightMeasurementAPITypes)
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

	return parsedMeasures.Body.MeasureGroups[0].Measures[0].Value / 1000
}

func updateMetrics(currentWeightMetric prometheus.Gauge, currentWeight float64) {
	currentWeightMetric.Set(currentWeight)
	log.Printf("Setting withings_current_weight metric to %f kg.\n", currentWeight)
}
