package main

import (
	"fmt"
	"io"
	"strings"
)

//Metrics contains a slice of Metric
type Metrics []Metric

//MetricExporter defines all required functions for a metric exporter
type MetricExporter interface {
	GetMetrics() (Metrics, error)
}

//Metric struct used for the exporter
type Metric struct {
	Name string
	//Descrition string
	Tags  *map[string]string
	Value uint64
}

//WriteMetrics exprots a slice of metrics to a writer
func WriteMetrics(metrics Metrics, w io.Writer) error {
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
		for k, v := range *m.Tags {
			tags = append(tags, k+`="`+v+`"`)
		}
		return fmt.Sprintf("%s{%s} %d\n", m.Name, strings.Join(tags, ","), m.Value)
	}
	return fmt.Sprintf("%s %d\n", m.Name, m.Value)
}

//FindMetric finds a metric by name and tagMatch (k=v)
func (metrics Metrics) FindMetric(metricName string, tagMatch string) *Metric {
	for _, m := range metrics {
		if m.Name == metricName && tagMatch == "" {
			return &m
		} else if m.Name == metricName {
			for k, v := range *m.Tags {
				if fmt.Sprintf("%s=%s", k, v) == tagMatch {
					return &m
				}
			}
		}
	}
	return nil
}

func (metrics Metrics) CheckMetric(metricName, tagMatch string, checkFunction func(x uint64) bool) error {
	metric := metrics.FindMetric(metricName, tagMatch)
	if metric == nil {
		return fmt.Errorf("Could not find metric %s / (%s)", metricName, tagMatch)
	}
	if !checkFunction((*metric).Value) {
		return fmt.Errorf("Check metric for metric %s / (%s) failed with value: %d", metricName, tagMatch, (*metric).Value)
	}
	return nil
}
