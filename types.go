package main

// RequestToken response from Withings API
// https://developer.withings.com/oauth2/#operation/oauth2-getaccesstoken
type RequestToken struct {
	Status int    `json:"status"`
	Error  string `json:"error"`
	Body   struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
		Scope        string `json:"scope"`
		ExpiresIn    int64  `json:"expires_in"`
		TokenType    string `json:"token_type"`
	} `json:"body"`
}

// Measures response from Withings API
// https://developer.withings.com/oauth2/#operation/measure-getmeas
type Measures struct {
	Status int `json:"status"`
	Body   struct {
		MeasureGroups []struct {
			Date     int64 `json:"date"`
			Created  int64 `json:"created"`
			Measures []struct {
				Value float64 `json:"value"`
				Type  int     `json:"type"`
			}
		} `json:"measuregrps"`
	} `json:"body"`
}
