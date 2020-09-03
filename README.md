# Prometheus Exporter for the Withings Body Smart Scales

I have a Withings Body+ smart scale. And I wanted to build my first Prometheus exporter.

## What this does

- Outputs a gauge metric for `withings_current_weight`, taking the most recent recorded weight and naively assuming it's in kilograms.
- Outputs all of the usual Go Prometheus client metrics.

## Future plans

- Better authentication - refresh tokens once the original access token expires after 3 hours?
- Some other metrics that the [Withings Measures API](https://developer.withings.com/oauth2/#section/Models/Measures) has that the Body+ scales can provide.
- Support stones and/or pounds for those who are more used to that than kg, or to correctly parse non-kg API responses.

## Running the exporter

```sh
cd /path/to/repo
go run main.go
```

or

```sh
go build .
./withings-exporter
```

## Authentication

- Create a [Withings account](https://account.withings.com/connectionuser/account_create). (You should already have one if you have a Withings product and use the HealthMate app!)
- Make a [Withings API Application](https://account.withings.com/connectionuser/account_create).
- Set `WITHINGS_APP_CLIENT_ID` and `WITHINGS_APP_CLIENT_SECRET` based off that application you created.
- Follow the instructions when you run the exporter to authorize your account to connect with the application, and optionally set `WITHINGS_API_ACCESS_TOKEN`. Access tokens are valid for 3 hours.
