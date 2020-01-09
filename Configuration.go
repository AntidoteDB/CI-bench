package main

import (
	"flag"
	"strconv"
	"fmt"
)

type Configuration struct {
	topology        string
	concurrent      []int //number of concurrent clients per host
	cpuProcs        int
	requests        int
	hosts           []string
	objects         int
	keyDistribution string
	objectType      string
	benchmarkType   string
	bashoBenchPath  string
	delay           int
	loss           int
	rate           int
	name            string
}

type clientsFlag []int
type hostsFlag []string

var (
	topology                    = flag.String("t", "dc2n1", "DC Topology")
	clients        				  clientsFlag
	cpuProcs                    = flag.Int("cpu", 4, "Maximum cores used")
	requests                    = flag.Int("r", 1000, "Number of requests per host")
	hosts                         hostsFlag
	objects                     = flag.Int("o", 5, "Number of objects used per request")
	keyDistribution             = flag.String("key", "paretoInt", "Key distribution")
	objectType                  = flag.String("object", "counter", "CRDT object")
	benchmarkType               = flag.String("b", "staticWrite", "Benchmark type")
	bashoBenchPath              = flag.String("bb", "", "BashoBench config path")
	delay                       = flag.Int("d", 0, "Network delay")
	loss                       = flag.Int("loss", 0, "Network loss in percentage")
	rate                       = flag.Int("rate", 0, "Network rate limit")
	name                        = flag.String("n", "", "Benchmark name")
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

	if len(clients) == 0 {
		clients = append(clients, 1)
	}
	if len(hosts) == 0 {
		hosts = append(hosts, "dc1n1:8087")
	}
	if *name == "" {
		*name = currentTimestamp()
	}

	configuration := Configuration{
		topology:        *topology,
		concurrent:      clients,
		cpuProcs:        *cpuProcs,
		requests:        *requests,
		hosts:           hosts,
		objects:         *objects,
		keyDistribution: *keyDistribution,
		objectType:      *objectType,
		benchmarkType:   *benchmarkType,
		bashoBenchPath:  *bashoBenchPath,
		delay:           *delay,
		loss:            *loss,
		rate:            *rate,
		name:			 *name,
	}
	return configuration
}
