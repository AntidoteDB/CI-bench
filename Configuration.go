package main

import (
	"flag"
	"strconv"
	"fmt"
)

type Configuration struct {
	concurrent      []int //number of concurrent clients per host
	cpuProcs        int
	requests        int
	hosts           []string
	objects         int
	keyDistribution string
	objectType      string
	benchmarkType   string
}

type clientsFlag []int
type hostsFlag []string

var (
	clients         clientsFlag = []int{1, 5, 10, 20, 30, 50, 100}
	cpuProcs                    = flag.Int("cpu", 4, "Maximum cores used")
	requests                    = flag.Int("r", 10000, "Number of requests per host")
	hosts           hostsFlag   = []string{"127.0.0.1:8087"}
	objects                     = flag.Int("o", 5, "Number of objects used per request")
	keyDistribution             = flag.String("key", "paretoInt", "Key distribution")
	objectType                  = flag.String("object", "counter", "CRDT object")
	benchmarkType               = flag.String("b", "staticWrite", "Benchmark type")
)

func (i *clientsFlag) String() string {
	return fmt.Sprintf("%v", *i)
}

func (i *clientsFlag) Set(value string) error {
	v, err := strconv.Atoi(value)
	if err == nil {
		*i = append(*i, v)
	}
	return err
}

func (i *hostsFlag) String() string {
	return fmt.Sprintf("%v", *i)
}

func (i *hostsFlag) Set(value string) error {
	*i = append(*i, value)
	return nil
}

func loadConfiguration() Configuration {
	flag.Var(&clients, "c", "Concurrent clients")
	flag.Var(&hosts, "h", "Concurrent clients")
	flag.Parse()

	configuration := Configuration{
		concurrent:      clients,
		cpuProcs:        *cpuProcs,
		requests:        *requests,
		hosts:           hosts,
		objects:         *objects,
		keyDistribution: *keyDistribution,
		objectType:      *objectType,
		benchmarkType:   *benchmarkType,
	}
	return configuration
}
