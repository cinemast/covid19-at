package main

import (
	"fmt"
	"net/http"
)

var locationProvider = NewLocationProvider()
var ministryExporter = NewMinistryExporter(locationProvider)
var ecdcExporter = NewEcdcExporter(locationProvider)

func handleMetrics(w http.ResponseWriter, r *http.Request) {
	austriaStats, _ := ministryExporter.GetMetrics()
	if austriaStats != nil {
		WriteMetrics(austriaStats, w)
	}
	worldStats, _ := ecdcExporter.GetMetrics()
	if worldStats != nil {
		WriteMetrics(worldStats, w)
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
		if locationProvider.GetLocation(country) == nil {
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

	err = ministryStats.CheckMetric("cov19_confirmed", "", func(x uint64) bool { return x > 1000 })
	if err != nil {
		errors = append(errors, err)
	}

	err = ministryStats.CheckMetric("cov19_tests", "", func(x uint64) bool { return x > 10000 })
	if err != nil {
		errors = append(errors, err)
	}

	err = ministryStats.CheckMetric("cov19_healed", "", func(x uint64) bool { return x > 5 })
	if err != nil {
		errors = append(errors, err)
	}
	return errors
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
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
