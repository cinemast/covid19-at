package main

import (
	"encoding/json"
	"fmt"
)

type healthMinistryExporter struct {
	mp  *metadataProvider
	url string
}

type ministryStat []struct {
	Label string
	Y     uint64
	Z     uint64
}

func newHealthMinistryExporter() *healthMinistryExporter {
	return &healthMinistryExporter{mp: newMetadataProviderWithFilename("bezirke.csv"), url: "https://info.gesundheitsministerium.at/data"}
}

func checkTags(result metrics, field string) []error {
	errors := make([]error, 0)
	for _, s := range result {
		if len(*s.Tags) != 4 {
			errors = append(errors, fmt.Errorf("Missing tags for: %s", (*s.Tags)[field]))
		}
	}
	return errors
}

func (h *healthMinistryExporter) GetMetrics() (metrics, error) {
	metrics := make(metrics, 0)

	result, _ := h.getSimpleData()
	metrics = append(metrics, result...)

	result, err := h.getAgeMetrics()
	metrics = append(metrics, result...)

	result, err = h.getGeschlechtsVerteilung()
	metrics = append(metrics, result...)

	result, err = h.getBundeslandInfections()
	metrics = append(metrics, result...)

	result, err = h.getBezirke()
	metrics = append(metrics, result...)
	result, err = h.getBundeslandHealedDeaths()
	metrics = append(metrics, result...)
	return metrics, err
}

func (h *healthMinistryExporter) Health() []error {
	errors := make([]error, 0)
	result, err := h.getBezirke()
	if err != nil {
		errors = append(errors, err)
	}
	if len(result) < 10 {
		errors = append(errors, fmt.Errorf("Not enough Bezirke Results: %d", len(result)))
	}
	errors = append(errors, checkTags(result, "bezirk")...)

	result, err = h.getBundeslandInfections()
	if err != nil {
		errors = append(errors, err)
	}
	if len(result) != 27 {
		errors = append(errors, fmt.Errorf("Missing Bundesland result %d", len(result)))
	}
	errors = append(errors, checkTags(result, "province")...)

	result, err = h.getAgeMetrics()
	if err != nil {
		errors = append(errors, err)
	}
	if len(result) < 4 {
		errors = append(errors, fmt.Errorf("Missing age metrics"))
	}

	result, err = h.getGeschlechtsVerteilung()
	if err != nil {
		errors = append(errors, err)
	}
	if len(result) != 2 {
		errors = append(errors, fmt.Errorf("Geschlechtsverteilung failed"))
	}

	result, err2 := h.getSimpleData()
	errors = append(errors, err2...)

	if len(result) < 3 {
		errors = append(errors, fmt.Errorf("Could not find \"Bestätigte Fälle\""))
	}

	return errors
}

func (h *healthMinistryExporter) getTags(location string, fieldName string, data *metaData) *map[string]string {
	if data != nil {
		return &map[string]string{fieldName: location, "country": "Austria", "longitude": ftos(data.location.long), "latitude": ftos(data.location.lat)}
	}
	return &map[string]string{fieldName: location, "country": "Austria"}
}

func (h *healthMinistryExporter) getBezirke() (metrics, error) {
	arrayString, err := readArrayFromGet(h.url + "/Bezirke.js")
	if err != nil {
		return nil, err
	}
	bezirkeStats := ministryStat{}
	err = json.Unmarshal([]byte(arrayString), &bezirkeStats)
	if err != nil {
		return nil, err
	}
	result := make(metrics, 0)
	for _, s := range bezirkeStats {
		data := h.mp.getMetadata(s.Label)
		tags := h.getTags(s.Label, "bezirk", data)
		result = append(result, metric{"cov19_bezirk_infected", tags, float64(s.Y)})
		if data != nil {
			result = append(result, metric{"cov19_bezirk_infected_100k", tags, float64(infection100k(s.Y, data.population))})
		}
	}
	return result, nil
}

func (h *healthMinistryExporter) getBezirkStat() ([]bezirkStat, error) {
	arrayString, err := readArrayFromGet(h.url + "/Bezirke.js")
	if err != nil {
		return nil, err
	}
	bezirkeStats := ministryStat{}
	err = json.Unmarshal([]byte(arrayString), &bezirkeStats)
	if err != nil {
		return nil, err
	}
	result := make([]bezirkStat, 0)
	for _, s := range bezirkeStats {
		data := h.mp.getMetadata(s.Label)
		result = append(result, bezirkStat{s.Label, apiLocaiton{Lat: data.location.lat, Long: data.location.long}, data.population, s.Y})
	}
	return result, nil
}

func mapBundeslandLabel(label string) string {
	switch label {
	case "Ktn":
		return "Kärnten"
	case "NÖ":
		return "Niederösterreich"
	case "OÖ":
		return "Oberösterreich"
	case "Sbg":
		return "Salzburg"
	case "Stmk":
		return "Steiermark"
	case "T":
		return "Tirol"
	case "V":
		return "Vorarlberg"
	case "W":
		return "Wien"
	case "Bgld":
		return "Burgenland"
	}
	return "unknown"
}

func (h *healthMinistryExporter) getBundeslandInfections() (metrics, error) {
	arrayString, err := readArrayFromGet(h.url + "/Bundesland.js")
	if err != nil {
		return nil, err
	}
	provinceStats := ministryStat{}
	err = json.Unmarshal([]byte(arrayString), &provinceStats)
	if err != nil {
		return nil, err
	}
	result := make(metrics, 0)
	for _, s := range provinceStats {
		s.Label = mapBundeslandLabel(s.Label)
		data := h.mp.getMetadata(s.Label)
		tags := h.getTags(s.Label, "province", data)
		result = append(result, metric{"cov19_detail", tags, float64(s.Y)})
		if data != nil {
			result = append(result, metric{"cov19_detail_infected_per_100k", tags, float64(infection100k(s.Y, data.population))})
			result = append(result, metric{"cov19_detail_infection_rate", tags, float64(infectionRate(s.Y, data.population))})
		}
	}
	return result, nil
}

func (h *healthMinistryExporter) getBundeslandHealedDeaths() (metrics, error) {
	arrayString, err := readArrayFromGet(h.url + "/GenesenTodesFaelleBL.js")
	if err != nil {
		return nil, err
	}
	provinceStats := ministryStat{}
	err = json.Unmarshal([]byte(arrayString), &provinceStats)
	if err != nil {
		return nil, err
	}
	result := make(metrics, 0)
	for _, s := range provinceStats {
		data := h.mp.getMetadata(s.Label)
		tags := h.getTags(s.Label, "province", data)
		result = append(result, metric{"cov19_detail_healed", tags, float64(s.Y)})
		result = append(result, metric{"cov19_detail_dead", tags, float64(s.Z)})
	}
	return result, nil
}

func (h *healthMinistryExporter) getAgeStat() (map[string]uint64, error) {
	arrayString, err := readArrayFromGet(h.url + "/Altersverteilung.js")
	if err != nil {
		return nil, err
	}
	ageStats := ministryStat{}
	err = json.Unmarshal([]byte(arrayString), &ageStats)
	if err != nil {
		return nil, err
	}
	result := make(map[string]uint64)
	for _, s := range ageStats {
		result[s.Label] = s.Y
	}
	return result, nil
}

func (h *healthMinistryExporter) getAgeMetrics() (metrics, error) {
	arrayString, err := readArrayFromGet(h.url + "/Altersverteilung.js")
	if err != nil {
		return nil, err
	}
	ageStats := ministryStat{}
	err = json.Unmarshal([]byte(arrayString), &ageStats)
	if err != nil {
		return nil, err
	}
	result := make(metrics, 0)
	for _, s := range ageStats {
		tags := &map[string]string{"country": "Austria", "group": s.Label}
		result = append(result, metric{"cov19_age_distribution", tags, float64(s.Y)})
	}
	return result, nil
}

func (h *healthMinistryExporter) getGeschlechtsVerteilung() (metrics, error) {
	arrayString, err := readArrayFromGet(h.url + "/Geschlechtsverteilung.js")
	if err != nil {
		return nil, err
	}
	ageStats := ministryStat{}
	err = json.Unmarshal([]byte(arrayString), &ageStats)
	if err != nil {
		return nil, err
	}
	result := make(metrics, 0)
	for _, s := range ageStats {
		tags := &map[string]string{"country": "Austria", "sex": s.Label}
		result = append(result, metric{"cov19_sex_distribution", tags, float64(s.Y)})
	}
	return result, nil
}

func addVarIfValid(errors []error, result metrics, url string, varName string, metricName string) ([]error, metrics) {
	value, err := readJsVarFromGet(url, varName)
	if err != nil {
		errors = append(errors, err)
	} else {
		result = append(result, metric{metricName, nil, atof(value)})
	}
	return errors, result
}

func (h *healthMinistryExporter) getSimpleData() (metrics, []error) {
	errors := make([]error, 0)
	result := make(metrics, 0)

	errors, result = addVarIfValid(errors, result, h.url+"/SimpleData.js", "Erkrankungen", "cov19_confirmed")
	errors, result = addVarIfValid(errors, result, h.url+"/Genesen.js", "dpGenesen", "cov19_healed")
	errors, result = addVarIfValid(errors, result, h.url+"/VerstorbenGemeldet.js", "dpTotGemeldet", "cov19_dead")
	errors, result = addVarIfValid(errors, result, h.url+"/GesamtzahlNormalbettenBel.js", "dpGesNBBel", "cov19_hospitalized")
	errors, result = addVarIfValid(errors, result, h.url+"/GesamtzahlIntensivBettenBel.js", "dpGesIBBel", "cov19_intensive_care")
	errors, result = addVarIfValid(errors, result, h.url+"/GesamtzahlTestungen.js", "dpGesTestungen", "cov19_tests")

	return result, errors
}
