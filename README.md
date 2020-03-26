[![CircleCI](https://circleci.com/gh/cinemast/covid19-at.svg?style=svg)](https://circleci.com/gh/cinemast/covid19-at)
[![codecov](https://codecov.io/gh/cinemast/covid19-at/branch/master/graph/badge.svg)](https://codecov.io/gh/cinemast/covid19-at)
[![Go Report Card](https://goreportcard.com/badge/github.com/cinemast/covid19-at)](https://goreportcard.com/report/github.com/cinemast/covid19-at)

# covid19-at

This is a small go-application that parses the following URLs for statistics regarding covid19-infections and deaths
provided by the [Austrian ministry for Health](https://www.sozialministerium.at/public.html)

- https://info.gesundheitsministerium.at
- https://www.sozialministerium.at/Informationen-zum-Coronavirus/Neuartiges-Coronavirus-(2019-nCov).html
- https://www.ecdc.europa.eu/en/geographical-distribution-2019-ncov-cases

It then exposes the gathered metrics as [prometheus](https://prometheus.io/) endpoint under `http://localhost:8282/metrics`

## Usage
- ```docker-compose up```
- Open [http://localhost:3000/](http://localhost:3000/) for Grafana (Credentials: admin/admin)
- Open [http://localhost:9090/prometheus](http://localhost:9090/prometheus) for Prometheus
- Open [http://localhost:8282/metrics](http://localhost:8282/metrics) for the metric exporter

## API 
- `GET` [http://localhost:8282/api/bundesland](http://localhost:8282/api/bundesland)
- `GET` [http://localhost:8282/api/bezirk](http://localhost:8282/api/bezirk)
- `GET` [http://localhost:8282/api/total](http://localhost:8282/api/total)

## Docker Image
- https://hub.docker.com/r/cinemast/covid19-at
- `docker pull cinemast/covid19-at`

## Demo Setup

- https://covid19.spiessknafl.at
- https://covid19.spiessknafl.at/prometheus
- https://covid19.spiessknafl.at/covid19/metrics
- https://covid19.spiessknafl.at/covid19/api/total
- https://covid19.spiessknafl.at/covid19/api/bundesland
- https://covid19.spiessknafl.at/covid19/api/bezirk

## Screenshot
![](screenshots/grafana.png)