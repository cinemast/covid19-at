package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
)

var logger = log.New(os.Stdout, "covid19-at", 0)
var mp = newMetadataProvider()
var exporters = []Exporter{
	newHealthMinistryExporter(),
	newSocialMinistryExporter(mp),
	newEcdcExporter(mp),
	newMathdroExporter(),
}

func handleMetrics(w http.ResponseWriter, _ *http.Request) {
	for _, e := range exporters {
		metrics, err := e.GetMetrics()
		if err == nil {
			writeMetrics(metrics, w)
		}
	}
}

func handleHealth(w http.ResponseWriter, _ *http.Request) {
	errors := make([]error, 0)
	for _, e := range exporters {
		errors = append(errors, e.Health()...)
	}

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
