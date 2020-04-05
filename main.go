package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"math/rand"
)

var isPrint = flag.Bool("p", false, "Prints the default values as JSON and exits")
var loadFile = flag.String("l", "", "Loads parameters from JSON file (see output of -p")
var randomSeed = flag.Int64("r", 42, "Random seed, for repeatability")

func main() {
	flag.Parse()

	if *isPrint {
		param, err := json.MarshalIndent(defaultParams, "", "  ")
		if err != nil {
			panic(err)
		}
		fmt.Printf("%s\n", string(param))
		return
	}

	params := defaultParams

	if *loadFile != "" {
		data, err := ioutil.ReadFile(*loadFile)
		if err != nil {
			fmt.Println(err)
			return
		}
		err = json.Unmarshal(data, &params)
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Printf("; loaded params from %s\n", *loadFile)
	}

	rand.Seed(*randomSeed)

	jsonParam, err := json.Marshal(params)
	if err != nil {
		panic(err)
	}
	fmt.Printf("; %s\n", string(jsonParam))
	fmt.Printf("day, live_count, infected_count, dead_count, isolation_count, immune_count\n")
	w := NewWorld(params)
	for i := 0; ; i++ {
		st := w.GetStat()
		fmt.Printf("%4d, %8d, %8d, %8d, %8d, %8d\n", i, st.LiveCount, st.InfectedCount, st.DeadCount, st.IsolationCount, st.ImmuneCount)
		if st.InfectedCount == 0 {
			break
		}
		w.NewDay()
	}
}
