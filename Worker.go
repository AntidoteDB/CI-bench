package main

import (
	antidote "github.com/AntidoteDB/antidote-go-client"
	"fmt"
	"time"
)

type worker struct {
	host       antidote.Host
	benchmark  Benchmark
	queue      chan RequestConfiguration
	resultChan chan RequestResult
	bucket     antidote.Bucket
}

func newWorker(host antidote.Host, benchmark Benchmark, queue chan RequestConfiguration, resultChan chan RequestResult, bucket antidote.Bucket) *worker {
	return &worker{host, benchmark, queue, resultChan, bucket}
}

func (worker *worker) run() {
	client, err := antidote.NewClient(worker.host)

	if err != nil {
		fmt.Println("Error creating Client.")
		return
	}
	defer client.Close()

	for request := range worker.queue {
		start := time.Now()

		err := worker.benchmark.function(client, &request.objects)

		duration := time.Since(start)
		worker.resultChan <- RequestResult{latency: duration, failed: err != nil}
	}
}
