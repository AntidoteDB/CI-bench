package main

import (
	"fmt"
	antidote "github.com/AntidoteDB/antidote-go-client"
	"sync"
	"time"
	"runtime"
	"strings"
	"strconv"
	"os"
	"encoding/binary"
	"math/rand"
)

type BenchmarkResult struct {
	Requests int
	Clients  int
	Min      time.Duration
	Max      time.Duration
	Avg      time.Duration
	Rps      float64
	Failed   int
}

type RequestConfiguration struct {
	objects BObject
}

type RequestResult struct {
	start         int64
	latency       time.Duration
	failed        bool
	errorCode     int
}

var Topologys = map[string]string {
	"dc1n2": "compose/dc1n2/docker-compose.yml",
	"dc2n1": "compose/dc2n1/docker-compose.yml",
	"dc2n2": "compose/dc2n2/docker-compose.yml",
}

func main() { os.Exit(mainReturnWithCode()) }

func mainReturnWithCode() int {
	fmt.Println("Init Benchmark.")

	configuration := loadConfiguration()
	runtime.GOMAXPROCS(configuration.cpuProcs)

	composePath, ok := Topologys[configuration.topology]
	if !ok {
		fmt.Println("Illegal dc topology type: " + configuration.objectType)
		return 1
	}

	defer stopDB(composePath)
	err := startDB(composePath)
	if err != nil {
		fmt.Println(err)
		return 1
	}

	idStats, err := startStats()
	if err != nil {
		fmt.Println(err)
		return 1
	}
	defer stopContainer(idStats)

	fmt.Println("Wait for DCs to connect.")
	if err := waitForStart(); err != nil {
		fmt.Println(err)
		return 1
	}

	if configuration.delay != 0 {
		fmt.Printf("Apply network delay %dms\n", configuration.delay)
		idDelay, err := applyDelay(configuration.delay)
		if err != nil {
			fmt.Println(err)
			return 1
		}
		defer stopContainer(idDelay)
	}
	if configuration.loss != 0 {
		fmt.Printf("Apply network loss %d%%\n", configuration.loss)
		idLoss, err := applyLoss(configuration.loss)
		if err != nil {
			fmt.Println(err)
			return 1
		}
		defer stopContainer(idLoss)
	}
	if configuration.rate != "" {
		fmt.Printf("Apply network rate %s\n", configuration.rate)
		idRate, err := applyRate(configuration.rate)
		if err != nil {
			fmt.Println(err)
			return 1
		}
		defer stopContainer(idRate)
	}

	dbContainer, err := getDbContainer()
	if err != nil {
		fmt.Println(err)
		return 1
	}


	fmt.Println("Start Benchmarks.")

	if configuration.bashoBenchPath != "" {
		err = runBashoBench(configuration.bashoBenchPath)
		if err != nil {
			fmt.Println(err)
			return 1
		}
	} else {
		_, ok := BObjects[configuration.objectType]
		if !ok {
			fmt.Println("Illegal object type: " + configuration.objectType)
			return 1
		}
		_, ok = Benchmarks[configuration.benchmarkType]
		if !ok {
			fmt.Println("Illegal benchmark type: " + configuration.benchmarkType)
			return 1
		}

		for _, c := range configuration.concurrent {
			if err := runBenchmark(c, configuration, dbContainer); err != nil {
				fmt.Println(err)
				return 1
			}
		}
	}
	fmt.Println("done.")
	return 0
}

func runBenchmark(concurrent int, configuration Configuration, dbContainer *[]DbContainer) error {
	runId := currentTimestamp()

	benchmark := Benchmarks[configuration.benchmarkType]

	//generate random bucket for each benchmark
	bucketKey := make([]byte, 8)
	rand.Read(bucketKey)
	bucket := antidote.Bucket{Bucket: bucketKey}

	queue := make(chan RequestConfiguration, configuration.requests)
	resultChannel := make(chan RequestResult, configuration.requests)

	createObject := BObjects["counter"]

	keys := GenerateKeys(configuration.keyDistribution, configuration.requests*configuration.objects)
	if keys == nil {
		fmt.Println("Illegal key distribution: " + configuration.keyDistribution)
		os.Exit(1)
	}

	hosts := make([]antidote.Host, len(configuration.hosts))
	for i, host := range configuration.hosts {
		h := strings.Split(host, ":")
		port, err := strconv.Atoi(h[1])
		if err != nil {
			fmt.Println("Error parsing port: " + h[1])
			continue
		}
		hosts[i] = antidote.Host{Name: h[0], Port: port}
		err = testConnection(hosts[i])
		if err != nil {
			fmt.Println("Error creating Client.")
			return err
		}
	}

	if benchmark.init != nil {
		uniqueMap := make(map[uint64]struct{}, len(*keys))
		uniques := make([]antidote.Key, len(*keys))
		i := 0
		for _, v := range *keys {
			intValue := binary.LittleEndian.Uint64(v)
			if _, ok := uniqueMap[intValue]; !ok {
				uniqueMap[intValue] = struct{}{}
				uniques[i] = v
				i++
			}
		}
		uniques = uniques[:i]
		object := createObject(&bucket, uniques, false, true)
		client, err := antidote.NewClient(hosts[0])

		if err != nil {
			fmt.Println("Error creating Client.")
			return err
		}
		benchmark.init(client, &object)
		client.Close()
	}

	for i := 0; i < configuration.requests; i++ {
		queue <- RequestConfiguration{createObject(&bucket, (*keys)[i*configuration.objects:(i+1)*configuration.objects], benchmark.read, benchmark.write)}
	}
	close(queue)

	wg := sync.WaitGroup{}
	start := time.Now()

	for _, host := range hosts {
		for i := 0; i < concurrent; i++ {
			wg.Add(1)
			worker := newWorker(host, benchmark, queue, resultChannel, bucket)
			go func() {
				worker.run()
				wg.Done()
			}()
		}
	}
	go func() {
		wg.Wait()
		close(resultChannel)
	}()

	failedCount := 0
	min := time.Duration(1<<63 - 1)
	max := time.Duration(0)
	sum := time.Duration(0)

	results := make([]RequestResult, 0)

	for result := range resultChannel {
		latency := result.latency
		sum += latency

		if min > latency {
			min = latency
		}
		if max < latency {
			max = latency
		}
		if result.failed {
			failedCount++
		}
		results = append(results, result)
	}

	end := time.Since(start)
	endTime := time.Now()
	avg := time.Duration(float64(sum.Nanoseconds()) / float64(configuration.requests))
	rps := (float64(configuration.requests) / float64(end.Nanoseconds())) * (1e9)

	result := BenchmarkResult{
		Requests: configuration.requests,
		Clients:  concurrent,
		Min:      min,
		Max:      max,
		Avg:      avg,
		Rps:      rps,
		Failed:   failedCount,
	}

	printBenchmarkResult(result)
	if err := writeResultSummaryToFile(configuration.name + "-" + strconv.Itoa(concurrent), result); err != nil {
		return err
	}
	if err := writeResultsToFile(configuration.name + "-" + strconv.Itoa(concurrent), runId, &results); err != nil {
		return err
	}
	resourceStatistics, err := collectStats(start, endTime, dbContainer)
	if err != nil {
		return err
	}
	if err := writeStatisticSummaryToFile(configuration.name + "-" + strconv.Itoa(concurrent), resourceStatistics); err != nil {
		return err
	}
	if err := generateReport(configuration.name + "-" + strconv.Itoa(concurrent), &result, resourceStatistics); err != nil {
		return err
	}
	if err := generateVis(configuration.name + "-" + strconv.Itoa(concurrent),"/tmp/results_" + configuration.name + "-" + strconv.Itoa(concurrent) + "_" + runId + ".csv"); err != nil {
		return err
	}
	return nil
}

func testConnection(host antidote.Host) error {
	_, err := antidote.NewClient(host)
	return err
}

func printBenchmarkResult(result BenchmarkResult) {
	fmt.Printf("Clients: %d\n", result.Clients)
	fmt.Printf("Number of Requests: %d\n", result.Requests)
	fmt.Println("Min: " + result.Min.String())
	fmt.Println("Max: " + result.Max.String())
	fmt.Println("Avg: " + result.Avg.String())
	fmt.Printf("Rps: %.2f\n", result.Rps)
	fmt.Printf("Failed Requests: %d\n", result.Failed)
}
