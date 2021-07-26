FROM debian:sid-slim
ADD . /code
RUN apt update && apt install -y ca-certificates golang-go golang-github-prometheus-client-golang-dev && cd /code && go build .
CMD /code/withings-exporter
