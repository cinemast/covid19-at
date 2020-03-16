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
	
	_, err := ecdcExporter.GetMetrics()
	if err != nil {
		failures++
		errorResponse = errorResponse + "World stats are failing\n"
	}

	_, err = ministryExporter.GetMetrics()
	if err != nil {
		failures++
		errorResponse = errorResponse + err.Error()
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
