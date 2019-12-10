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
	"context"
)

type BenchmarkResult struct {
	configuration Configuration
	clients       int
	min           time.Duration
	max           time.Duration
	avg           time.Duration
	rps           float64
	failed        int
}

type RequestConfiguration struct {
	objects BObject
}

type RequestResult struct {
	latency time.Duration
	failed  bool
	errorCode     int
}

func main() { os.Exit(mainReturnWithCode()) }

func mainReturnWithCode() int {
	fmt.Println("Init Benchmark.")

	configuration := loadConfiguration()
	runtime.GOMAXPROCS(configuration.cpuProcs)

	composePath := "compose/docker-compose.yml"
	defer stopDB(composePath)
	startDB(composePath)

	id, _ :=startStats()
	defer stopStats(id)
	fmt.Println(id)

	fmt.Println("Wait for DCs to connect.")
	if err := waitForStart(); err != nil {
		fmt.Println(err)
		return 1
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	delay(ctx)

	dbContainer, err := getDbContainer()
	if err != nil {
		fmt.Println(err)
		return 1
	}


	fmt.Println("Start Benchmarks.")
	start := time.Now()
	_, ok := BObjects[configuration.objectType]
	if !ok {
		fmt.Println("Illegal object type: " + configuration.objectType)
		os.Exit(1)
	}
	_, ok = Benchmarks[configuration.benchmarkType]
	if !ok {
		fmt.Println("Illegal benchmark type: " + configuration.benchmarkType)
		os.Exit(1)
	}

	for _, c := range configuration.concurrent {
		runBenchmark(c, configuration)
	}
	end := time.Now()

	collectStats(start, end, dbContainer)
	fmt.Println("done.")
	return 0
}

func runBenchmark(concurrent int, configuration Configuration) {
	benchmark := Benchmarks[configuration.benchmarkType]

	//generate random bucket for each benchmark
	bucketKey := make([]byte, 8)
	rand.Read(bucketKey)
	bucket := antidote.Bucket{Bucket: bucketKey}

	queue := make(chan RequestConfiguration, configuration.requests)
	results := make(chan RequestResult, configuration.requests)

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
			return
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
			worker := newWorker(host, benchmark, queue, results, bucket)
			go func() {
				worker.run()
				wg.Done()
			}()
		}
	}
	go func() {
		wg.Wait()
		close(results)
	}()

	failedCount := 0
	min := time.Duration(1<<63 - 1)
	max := time.Duration(0)
	sum := time.Duration(0)

	for result := range results {
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
	}

	end := time.Since(start)
	avg := time.Duration(float64(sum.Nanoseconds()) / float64(configuration.requests))
	rps := (float64(configuration.requests) / float64(end.Nanoseconds())) * (1e9)

	result := BenchmarkResult{
		configuration: configuration,
		clients:       concurrent,
		min:           min,
		max:           max,
		avg:           avg,
		rps:           rps,
		failed:        failedCount,
	}

	printBenchmarkResult(result)
}

func printBenchmarkResult(result BenchmarkResult) {
	fmt.Printf("Clients: %d\n", result.clients)
	fmt.Printf("Number of Requests: %d\n", result.configuration.requests)
	fmt.Println("Min: " + result.min.String())
	fmt.Println("Max: " + result.max.String())
	fmt.Println("Avg: " + result.avg.String())
	fmt.Printf("Rps: %.2f\n", result.rps)
	fmt.Printf("Failed Requests: %d\n", result.failed)
}
