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

func handleHealth(w http.ResponseWriter, r *http.Request) {
	failures := 0
	errorResponse := ""
	
	worldStats, err := ecdcExporter.GetMetrics()
	if err != nil || len(worldStats) < 200 {
		failures++
		errorResponse = errorResponse + "World stats are failing\n"
	}

	for _,m := range worldStats {
		country := (*m.Tags)["country"]
		if locationProvider.GetLocation(country) == nil {
			failures++
			errorResponse = errorResponse + "Could not find location for country " + country + "\n"
		}
	}

	ministryStats, err := ministryExporter.GetMetrics()
	if err != nil {
		failures++
		errorResponse = errorResponse + err.Error() + "\n"
	}

	if len(ministryStats) < 14 {
		failures++
		errorResponse = errorResponse + "Missing ministry stats\n"
	}


	if failures > 0 {
		w.WriteHeader(http.StatusInternalServerError)
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
