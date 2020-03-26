package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
)

var logger = log.New(os.Stdout, "covid19-at", 0)
var mp = newMetadataProvider()
var he = newHealthMinistryExporter()
var se = newSocialMinistryExporter(mp)
var exporters = []Exporter{
	he,
	se,
	newEcdcExporter(mp),
	newMathdroExporter(),
}

var a = newApi(he, se)

func writeJson(w http.ResponseWriter, f func() (interface{}, error)) {
	result, err := f()
	if err != nil {
		w.WriteHeader(500)
		w.Write([]byte(err.Error()))
	} else {
		bytes, err := json.Marshal(result)
		if err != nil {
			w.WriteHeader(500)
			w.Write([]byte(err.Error()))
		} else {
			w.Header().Add("Content-type", "application/json; charset=utf-8")
			w.Write(bytes)
		}
	}
}

func handleApiBundesland(w http.ResponseWriter, _ *http.Request) {
	writeJson(w, func() (interface{}, error) { return a.GetBundeslandStat() })
}

func handleApiBezirk(w http.ResponseWriter, _ *http.Request) {
	writeJson(w, func() (interface{}, error) { return a.GetBezirkStat() })
}

func handleApiTotal(w http.ResponseWriter, _ *http.Request) {
	writeJson(w, func() (interface{}, error) { return a.GetOverallStat() })
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
	http.HandleFunc("/api/bundesland", handleApiBundesland)
	http.HandleFunc("/api/bezirk", handleApiBezirk)
	http.HandleFunc("/api/total", handleApiTotal)
	http.ListenAndServe(":8282", nil)
}
