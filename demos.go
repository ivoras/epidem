package main

import (
	"math/rand"
)

type DiseaseParameters struct {
	PopulationCount         uint32  `json:"population_count"`
	InteractionCircleCount  uint32  `json:"interaction_circle_count"`               // Dumbar's number
	AlwaysAsymptomaticRatio float32 `json:"asymptomatic_ratio"`                     // Percentage of people who are always asymptomatic
	IsolationViolatorsRatio float32 `json:"isolation_violators_ratio"`              // Percentage of people who will ignore isolation rules
	AsymptomaticDays        uint32  `json:"asymptomatic_days"`                      // Number of days at which those who are not asymptomatic, look asymptomatic
	RIsolationDay           float32 `json:"spread_rate_isolation"`                  // Spread rate per day of non-asymptomatic people in isolation
	RNotIsolationDay        float32 `json:"spread_rate_not_isolation"`              // Spread rate per day of non-asympromatic people not in isolation
	RAIsolationDay          float32 `json:"spread_rate_asymptomatic_isolation"`     // Spread rate per day for always asymptomatic people in isolation
	RANotIsolationDay       float32 `json:"spread_rate_asymptomatic_not_isolation"` // Spread rate per day for always asymptomatic people not in isolation
	//	AgeInfectionRatioBase   float32 `json:"age_infection_rate_base"`                // Baseline (minimal) infection rate added to 1/(100-x)
	RDeathNormal   float32 `json:"death_rate_normal"`   // Death rate with a functioning healthcare system
	RDeathCollapse float32 `json:"death_rate_collapse"` // Death rate without a functioning healthcare system
}

var defaultParams = DiseaseParameters{
	PopulationCount:         10_000_000,
	InteractionCircleCount:  100,
	AlwaysAsymptomaticRatio: 0.5,
	IsolationViolatorsRatio: 0.1,
	AsymptomaticDays:        14,
	RIsolationDay:           0.001,
	RNotIsolationDay:        2,
	RAIsolationDay:          0.0001,
	RANotIsolationDay:       0.001,
	RDeathNormal:            0.02,
	RDeathCollapse:          0.1,
}

/*
func (dp DiseaseParameters) AgeInfectionRate(age byte) float32 {
	return dp.AgeInfectionRatioBase + float32(1) - float32(1)/float32(100-age)
}
*/

const PERSON_STATUS_ALIVE = 1                // Person is alive
const PERSON_STATUS_INFECTED = 2             // Person is infected
const PERSON_STATUS_IN_ISOLATION = 4         // Person in isolation
const PERSON_STATUS_ASYMPTOMATIC = 8         // Person is asymptomatic
const PERSON_STATUS_ALWAYS_ASYMPTOMATIC = 16 // Person is always asymptomatic
const PERSON_STATUS_ISOLATION_VIOLATOR = 32  // Person is an isolation violator

type Person struct {
	Status byte
}

func (p Person) IsAlive() bool {
	return p.Status&PERSON_STATUS_ALIVE != 0
}

func (p Person) IsInfected() bool {
	return p.Status&PERSON_STATUS_INFECTED != 0
}

func (p Person) IsInIsolation() bool {
	return p.Status&PERSON_STATUS_IN_ISOLATION != 0
}

func (p Person) IsAsymptompatic() bool {
	return p.Status&PERSON_STATUS_ASYMPTOMATIC != 0
}

func (p Person) IsAlwaysAsymptomatic() bool {
	return p.Status&PERSON_STATUS_ALWAYS_ASYMPTOMATIC != 0
}

func (p Person) IsIsolationViolator() bool {
	return p.Status&PERSON_STATUS_ISOLATION_VIOLATOR != 0
}

type World struct {
	dParams    DiseaseParameters
	Population []Person
}

func NewWorld(dp DiseaseParameters) (w World) {
	w.dParams = dp
	w.Population = make([]Person, dp.PopulationCount)
	for i := range w.Population {
		if rand.Float32() < dp.AlwaysAsymptomaticRatio {
			w.Population[i].Status = PERSON_STATUS_ALWAYS_ASYMPTOMATIC
		}
		if rand.Float32() < dp.IsolationViolatorsRatio {
			w.Population[i].Status ^= PERSON_STATUS_ISOLATION_VIOLATOR
		}
	}
	return
}

func (w *World) DeadCount() (count uint32) {
	for _, p := range w.Population {
		if !p.IsAlive() {
			count++
		}
	}
	return
}

func (w *World) TryInfect(src, tgt *Person) {
	tgt.Status ^= PERSON_STATUS_INFECTED
}

func (w *World) NewDay() {
	// Process new infections
	for i, p := range w.Population {
		if p.IsAlive() && p.IsInfected() {
			// Assumption: a person always has contact with the same InteractionCircleCount people
			rSource := rand.NewSource(int64(i))
			r := rand.New(rSource)
			for j := uint32(0); j < w.dParams.InteractionCircleCount; j++ {
				tgt := r.Intn(int(w.dParams.PopulationCount))
				w.TryInfect(&p, &w.Population[tgt])
			}
		}
	}
}
