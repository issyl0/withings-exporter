# Prometheus Exporter for the Withings Body Smart Scales

I have a Withings Body+ smart scale. And I wanted to build my first Prometheus exporter.

## What this does

- Outputs a gauge metric for `withings_current_weight`, taking the most recent recorded weight. The API returns the weight in kilograms.
- OAuth token refresh.
- Metrics refresh after 30 minutes.
- Outputs all of the usual Go Prometheus client metrics.

## Future plans

- Some other metrics that the [Withings Measures API](https://developer.withings.com/oauth2/#operation/measure-getmeas) has that the Body+ scales can provide.
- The ability to pass flags for metric refresh rate and port to run on.

## Running the exporter

```sh
cd /path/to/repo
go build .
./withings-exporter
```

or

```sh
cd /path/to/repo
go run main.go metrics.go types.go
```

## Authentication

- Create a [Withings account](https://account.withings.com/connectionuser/account_create). (You should already have one if you have a Withings product and use the HealthMate app!)
- Make a [Withings API Application](https://account.withings.com/connectionuser/account_create).
- Set `WITHINGS_APP_CLIENT_ID` and `WITHINGS_APP_CLIENT_SECRET` based off that application you created.
- Follow the instructions when you run the exporter to authorize your account to connect with the application. Access tokens are valid for three hours, then this auto-refreshes.
