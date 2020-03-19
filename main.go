package main

import (
	"fmt"
	"github.com/cinemast/covid19-at/exporter"
	"net/http"
)

var metadataProvider = exporter.NewMetadataProvider()
var ministryExporter = exporter.NewMinistryExporter(metadataProvider)
var ecdcExporter = exporter.NewEcdcExporter(metadataProvider)
var grafanaExporter = exporter.NewGrafanaExporter()

func handleMetrics(w http.ResponseWriter, _ *http.Request) {
	austriaStats, err := ministryExporter.GetMetrics()
	if err == nil {
		exporter.WriteMetrics(austriaStats, w)
	}
	worldStats, err := ecdcExporter.GetMetrics()
	if err == nil {
		exporter.WriteMetrics(worldStats, w)
	}

	bezirkStats, err := grafanaExporter.GetMetrics()
	if err == nil {
		exporter.WriteMetrics(bezirkStats, w)
	}
}

func getErrors() []error {
	errors := make([]error, 0)

	worldStats, _ := ecdcExporter.GetMetrics()
	if len(worldStats) < 200 {
		errors = append(errors, fmt.Errorf("World stats are failing"))
	}

	for _, m := range worldStats {
		country := (*m.Tags)["country"]
		if metadataProvider.GetLocation(country) == nil {
			errors = append(errors, fmt.Errorf("Could not find location for country: %s", country))
		}
	}

	ministryStats, err := ministryExporter.GetMetrics()
	if err != nil {
		errors = append(errors, err)
	}

	if len(ministryStats) < 10 {
		errors = append(errors, fmt.Errorf("Missing ministry stats"))
	}

	err = ministryStats.CheckMetric("cov19_confirmed", "", func(x float64) bool { return x > 1000 })
	if err != nil {
		errors = append(errors, err)
	}

	err = ministryStats.CheckMetric("cov19_tests", "", func(x float64) bool { return x > 10000 })
	if err != nil {
		errors = append(errors, err)
	}

	err = ministryStats.CheckMetric("cov19_healed", "", func(x float64) bool { return x > 5 })
	if err != nil {
		errors = append(errors, err)
	}
	return errors
}

func handleHealth(w http.ResponseWriter, _ *http.Request) {
	errors := getErrors()
	if len(errors) > 0 {
		w.WriteHeader(http.StatusInternalServerError)
		errorResponse := ""
		for _, e := range errors {
			errorResponse += e.Error() + "\n"
		}
		fmt.Fprintf(w, `<html><body><img width="500" src="https://spiessknafl.at/fine.jpg"/><pre>%s</pre></body></html>`, errorResponse)
	} else {
		fmt.Fprintf(w, `<html><body><img width="500" src="https://spiessknafl.at/helth.png"/></body></html>`)
	}
}

func main() {
	http.HandleFunc("/metrics", handleMetrics)
	http.HandleFunc("/health", handleHealth)
	http.ListenAndServe(":8282", nil)
}
