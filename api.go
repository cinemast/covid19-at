package main

type apiLocaiton struct {
	Lat  float64
	Long float64
}

type bundeslandStat struct {
	Name          string
	Location      apiLocaiton
	Population    uint64
	Infected      uint64
	Dead          uint64
	Hospitalized  uint64
	IntensiveCare uint64
}

type bezirkStat struct {
	Name       string
	Location   apiLocaiton
	Population uint64
	Infected   uint64
}

type overallStat struct {
	TotalInfected            uint64
	TotalDead                uint64
	TotalHospitalized        uint64
	TotalIntensiveCare       uint64
	AgeDistributionInfection map[string]uint64
}

type api struct {
	he *healthMinistryExporter
	se *socialMinistryExporter
}

func newApi(he *healthMinistryExporter, se *socialMinistryExporter) *api {
	return &api{he, se}
}

func (a *api) GetOverallStat() (overallStat, error) {
	r := overallStat{}
	ageStats, err := a.he.getAgeStat()
	if err != nil {
		return overallStat{}, err
	}
	r.AgeDistributionInfection = ageStats
	bundeslandStats, err := a.GetBundeslandStat()

	d, err2 := a.he.getSimpleData()
	if len(err2) != 0 {
		return overallStat{}, err2[0]
	}
	sumInfect := uint64(0)
	sumDead := uint64(0)
	sumHospitalized := uint64(0)
	sumIntensiveCare := uint64(0)

	for _, v := range bundeslandStats {
		sumInfect += v.Infected
		sumDead += v.Dead
		sumHospitalized += v.Hospitalized
		sumIntensiveCare += v.IntensiveCare
	}
	r.TotalDead = sumDead
	r.TotalInfected = sumInfect
	r.TotalHospitalized = sumHospitalized
	r.TotalIntensiveCare = sumIntensiveCare

	confirmed := d.findMetric("cov19_confirmed", "")
	if confirmed != nil {
		r.TotalInfected = uint64(confirmed.Value)
	}

	return r, nil
}

func (a *api) GetBezirkStat() ([]bezirkStat, error) {
	return a.he.getBezirkStat()
}

func (a *api) GetBundeslandStat() ([]bundeslandStat, error) {
	hospitalStat, err := a.se.getHospitalizedStats()
	if err != nil {
		return nil, err
	}

	bundeslandStats, err := a.se.getBundeslandStats()
	if err != nil {
		return nil, err
	}

	result := make([]bundeslandStat, 0)
	for k, v := range bundeslandStats {
		data := a.se.mp.getMetadata(k)
		hospitalized := uint64(0)
		intensiveCare := uint64(0)
		if v, ok := hospitalStat[k]; ok {
			hospitalized = v.Hospitalized
			intensiveCare = v.IntensiveCare
		}
		result = append(result, bundeslandStat{
			Name:          k,
			Infected:      v.infected,
			Dead:          v.deaths,
			Population:    data.population,
			Hospitalized:  hospitalized,
			IntensiveCare: intensiveCare,
			Location:      apiLocaiton{Lat: data.location.lat, Long: data.location.long},
		})
	}
	return result, nil
}
