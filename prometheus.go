package main

import (
	"fmt"
	"io"
	"strings"
)

type metrics []metric

type metric struct {
	Name  string
	Tags  *map[string]string
	Value float64
}

type CovidStat struct {
	location string
	infected uint64
	deaths   uint64
}

type Exporter interface {
	GetMetrics() (metrics, error)
	Health() []error
}

func writeMetrics(metrics metrics, w io.Writer) error {
	for _, m := range metrics {
		_, err := io.WriteString(w, formatMetric(m))
		if err != nil {
			return err
		}
	}
	return nil
}

func formatMetric(m metric) string {
	tags := []string{}
	if m.Tags != nil {
		for k, v := range *m.Tags {
			tags = append(tags, k+`="`+v+`"`)
		}
		return fmt.Sprintf("%s{%s} %f\n", m.Name, strings.Join(tags, ","), m.Value)
	}
	return fmt.Sprintf("%s %f\n", m.Name, m.Value)
}

func (metrics metrics) findMetric(metricName string, tagMatch string) *metric {
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

func (metrics metrics) checkMetric(metricName, tagMatch string, checkFunction func(x float64) bool) error {
	metric := metrics.findMetric(metricName, tagMatch)
	if metric == nil {
		return fmt.Errorf("Could not find metric %s / (%s)", metricName, tagMatch)
	}
	if !checkFunction((*metric).Value) {
		return fmt.Errorf("Check metric for metric %s / (%s) failed with value: %f", metricName, tagMatch, (*metric).Value)
	}
	return nil
}
