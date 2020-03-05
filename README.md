# covid19-at

This is a small go-application that parses the following URLs for statistics regarding covid19-infections and deaths
provided by the [Austrian ministry for Health](https://www.sozialministerium.at/public.html)

- https://www.sozialministerium.at/Themen/Gesundheit/Uebertragbare-Krankheiten/Infektionskrankheiten-A-Z/Neuartiges-Coronavirus.html
- https://www.sozialministerium.at/Informationen-zum-Coronavirus/Neuartiges-Coronavirus-(2019-nCov).html

It then exposes the gathered metrics as [prometheus](https://prometheus.io/) endpoint under `http://localhost:8282/metrics`

## Example Grafana Dashboard
![](screenshots/grafana.png)