package main

import (
	"fmt"
	"io"
	"strings"
)

//MetricExporter defines all required functions for a metric exporter
type MetricExporter interface {
	GetMetrics() ([]Metric, error)
}

//Metric struct used for the exporter
type Metric struct {
	Name string
	//Descrition string
	Tags *map[string]string
	Value uint64
}

//WriteMetrics exprots a slice of metrics to a writer
func WriteMetrics(metrics []Metric, w io.Writer) error {
	for _, m := range metrics {
		_, err := io.WriteString(w, FormatMetric(m))
		if err != nil {
			return err
		}
	}
	return nil
}

//FormatMetric converts the metric to a string
func FormatMetric(m Metric) string {
	tags := []string{}
	if m.Tags != nil {
		for k,v := range *m.Tags {
			tags = append(tags, k + `="` + v + `"`)
		}
		return fmt.Sprintf("%s{%s} %d\n", m.Name, strings.Join(tags, ","), m.Value)
	}
	return fmt.Sprintf("%s %d\n", m.Name, m.Value)
}