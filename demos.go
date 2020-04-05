package main

import (
	"fmt"
	"math/rand"
)

/*
 * Spread probability = probability of infection for each encounter between 2 people.
 */

const AlgorithmTypeDefault = 0
const AlgorithmTypeFaster = 1
const AlgorithmTypeLudicrous = 2

type DiseaseParameters struct {
	AlgorithmType           uint32  `json:"algorithm_type"`
	PopulationCount         uint32  `json:"population_count"`                       // Population size
	StartInfected           uint32  `json:"start_infected"`                         // Count of initially infected people
	CollapseThreshold       uint32  `json:"collapse_threshold"`                     // Health system collapse threshold - collapses if this many people are ill
	InteractionCircleCount  uint32  `json:"interaction_circle_count"`               // Dumbar's number
	AlwaysAsymptomaticRatio float32 `json:"asymptomatic_ratio"`                     // Percentage of people who are always asymptomatic
	IsolationRatio          float32 `json:"isolation_ratio"`                        // Probability than an infected person will go into isolation (quarantine)
	IsolationViolatorsRatio float32 `json:"isolation_violators_ratio"`              // Percentage of people who will ignore isolation rules
	AsymptomaticDays        uint32  `json:"asymptomatic_days"`                      // Number of days at which those who are not asymptomatic, look asymptomatic
	TotalDiseaseDays        uint32  `json:"total_disease_days"`                     // Total days the disease is present in a person
	RIsolationProb          float32 `json:"spread_prob_isolation"`                  // Spread probability of non-asymptomatic people in isolation
	RNotIsolationProb       float32 `json:"spread_prob_not_isolation"`              // Spread probability of non-asympromatic people not in isolation
	RAIsolationProb         float32 `json:"spread_prob_asymptomatic_isolation"`     // Spread probability for always asymptomatic people in isolation
	RANotIsolationProb      float32 `json:"spread_prob_asymptomatic_not_isolation"` // Spread probability for always asymptomatic people not in isolation
	RDeathNormal            float32 `json:"death_prob_normal"`                      // Death rate with a functioning healthcare system
	RDeathCollapse          float32 `json:"death_prob_collapse"`                    // Death rate without a functioning healthcare system
}

var defaultParams = DiseaseParameters{
	AlgorithmType:           AlgorithmTypeFaster,
	PopulationCount:         10_000_000,
	CollapseThreshold:       5000,
	StartInfected:           1000,
	InteractionCircleCount:  40,
	AlwaysAsymptomaticRatio: 0.5,
	IsolationRatio:          0.9,
	IsolationViolatorsRatio: 0.1,
	AsymptomaticDays:        13,
	TotalDiseaseDays:        25,
	RIsolationProb:          0.001,
	RNotIsolationProb:       0.4,
	RAIsolationProb:         0.0001,
	RANotIsolationProb:      0.001,
	RDeathNormal:            0.0008, // 0.02 / 25
	RDeathCollapse:          0.006,  // 0.15 / 25
}

/*
func (dp DiseaseParameters) AgeInfectionRate(age byte) float32 {
	return dp.AgeInfectionRatioBase + float32(1) - float32(1)/float32(100-age)
}
*/

const PERSON_STATUS_ALIVE = 1                // Person is alive
const PERSON_STATUS_INFECTED = 2             // Person is infected
const PERSON_STATUS_IN_ISOLATION = 4         // Person in isolation
const PERSON_STATUS_SYMPTOMATIC = 8          // Person is asymptomatic
const PERSON_STATUS_ALWAYS_ASYMPTOMATIC = 16 // Person is always asymptomatic
const PERSON_STATUS_ISOLATION_VIOLATOR = 32  // Person is an isolation violator
const PERSON_STATUS_IMMUNE = 64              // Person is immune (e.g. from previous illness)

type Person struct {
	Status       byte
	DaysInfected byte
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

func (p Person) IsSymptomatic() bool {
	return p.Status&PERSON_STATUS_SYMPTOMATIC != 0
}

func (p Person) IsAlwaysAsymptomatic() bool {
	return p.Status&PERSON_STATUS_ALWAYS_ASYMPTOMATIC != 0
}

func (p Person) IsIsolationViolator() bool {
	return p.Status&PERSON_STATUS_ISOLATION_VIOLATOR != 0
}

func (p Person) IsImmune() bool {
	return p.Status&PERSON_STATUS_IMMUNE != 0
}

type World struct {
	dParams    DiseaseParameters
	Population []Person
}

func NewWorld(dp DiseaseParameters) (w World) {
	w.dParams = dp
	w.Population = make([]Person, dp.PopulationCount)
	for i := range w.Population {
		status := byte(PERSON_STATUS_ALIVE)
		if rand.Float32() < dp.AlwaysAsymptomaticRatio {
			status |= PERSON_STATUS_ALWAYS_ASYMPTOMATIC
		}
		if rand.Float32() < dp.IsolationViolatorsRatio {
			status |= PERSON_STATUS_ISOLATION_VIOLATOR
		}
		w.Population[i].Status = status
	}
	for i := uint32(0); i < w.dParams.StartInfected; i++ {
		tgt := rand.Intn(int(w.dParams.PopulationCount))
		w.Population[tgt].Status |= PERSON_STATUS_INFECTED
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

func (w *World) TryInfect(src Person, tgt *Person) {
	if tgt.IsImmune() {
		return
	}
	prob := float32(0)
	if !src.IsSymptomatic() && !src.IsInIsolation() {
		prob = w.dParams.RANotIsolationProb
	} else if !src.IsSymptomatic() && src.IsInIsolation() {
		prob = w.dParams.RAIsolationProb
	} else if src.IsSymptomatic() && !src.IsInIsolation() {
		prob = w.dParams.RNotIsolationProb
	} else if src.IsSymptomatic() && src.IsInIsolation() {
		prob = w.dParams.RIsolationProb
	}
	if rand.Float32() < prob {
		tgt.Status |= PERSON_STATUS_INFECTED
		if !tgt.IsIsolationViolator() && rand.Float32() < w.dParams.IsolationRatio {
			tgt.Status |= PERSON_STATUS_IN_ISOLATION
		}
	}
}

func (w *World) NewDay() {
	// Calculate if the health system has collapsed
	nInfected := uint32(0)
	for _, p := range w.Population {
		if p.IsInfected() && p.IsAlive() {
			nInfected++
		}
	}
	rDeath := w.dParams.RDeathNormal
	if nInfected > w.dParams.CollapseThreshold {
		rDeath = w.dParams.RDeathCollapse
	}
	// Process each person
	for i, p := range w.Population {
		if p.IsAlive() && p.IsInfected() {
			if uint32(p.DaysInfected) >= w.dParams.TotalDiseaseDays {
				w.Population[i].Status &^= PERSON_STATUS_INFECTED
				w.Population[i].Status &^= PERSON_STATUS_IN_ISOLATION
				w.Population[i].Status |= PERSON_STATUS_IMMUNE
				continue
			}
			if rand.Float32() < rDeath {
				w.Population[i].Status &^= PERSON_STATUS_ALIVE
				continue
			}
			w.Population[i].DaysInfected++
			if uint32(w.Population[i].DaysInfected) > w.dParams.AsymptomaticDays && !p.IsAlwaysAsymptomatic() {
				w.Population[i].Status |= PERSON_STATUS_SYMPTOMATIC
			}
			p = w.Population[i]
			// Assumption: a person always has contact with the same people
			if w.dParams.AlgorithmType == AlgorithmTypeDefault {
				// Always a repeatable pseudo-random sequence from the same seed
				rSource := rand.NewSource(int64(i))
				r := rand.New(rSource)
				for j := uint32(0); j < w.dParams.InteractionCircleCount; j++ {
					tgt := r.Intn(int(w.dParams.PopulationCount))
					w.TryInfect(p, &w.Population[tgt])
				}
			} else if w.dParams.AlgorithmType == AlgorithmTypeFaster {
				// Home-grown LFSR
				s := uint32(i)
				for j := uint32(0); j < w.dParams.InteractionCircleCount; j++ {
					b := (s >> 0) ^ (s >> 2) ^ (s >> 6) ^ (s >> 7)
					s = (s >> 1) | (b << 31)
					tgt := s % w.dParams.PopulationCount
					w.TryInfect(p, &w.Population[tgt])
				}
			} else if w.dParams.AlgorithmType == AlgorithmTypeLudicrous {
				// Infects a sequential set of people in the array, but at a random location
				s := uint32(i)
				b := (s >> 0) ^ (s >> 2) ^ (s >> 6) ^ (s >> 7)
				s = (s >> 1) | (b << 31)
				for j := uint32(0); j < w.dParams.InteractionCircleCount; j++ {
					tgt := (s + j) % w.dParams.PopulationCount
					w.TryInfect(p, &w.Population[tgt])
				}
			} else {
				panic(fmt.Sprintf("Unknown algorithm: %d", w.dParams.AlgorithmType))
			}
		}
	}
}

type WorldStat struct {
	LiveCount      int32  `json:"live_count"`
	InfectedCount  uint32 `json:"infected_count"`
	DeadCount      uint32 `json:"dead_count"`
	IsolationCount uint32 `json:"isolation_count"`
	ImmuneCount    uint32 `json:"immune_count"`
}

func (w World) GetStat() (st WorldStat) {
	for _, p := range w.Population {
		if p.IsAlive() {
			st.LiveCount++
			if p.IsInfected() {
				st.InfectedCount++
			}
			if p.IsInIsolation() {
				st.IsolationCount++
			}
			if p.IsImmune() {
				st.ImmuneCount++
			}
		} else {
			st.DeadCount++
		}
	}
	return
}
