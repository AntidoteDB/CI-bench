package main

import (
	antidote "github.com/AntidoteDB/antidote-go-client"
	"fmt"
	"time"
	"regexp"
	"strconv"
)

type worker struct {
	host       antidote.Host
	benchmark  Benchmark
	queue      chan RequestConfiguration
	resultChan chan RequestResult
	bucket     antidote.Bucket
}

var (
	codeRegexp = regexp.MustCompile("[0-9]+")
)

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

		for { //repeat aborted requests
			err := worker.benchmark.function(client, &request.objects)
			if err == nil || parseErrorCode(err) != 3 { //code 3: aborted
				break
			}
		}

		duration := time.Since(start)

		result := RequestResult{latency: duration, failed: err != nil}
		if err != nil {
			result.errorCode = parseErrorCode(err)
		}

		worker.resultChan <- result
	}
}

func parseErrorCode(error error) int {
	codeString := codeRegexp.FindString(error.Error())
	if codeString == "" {
		return 0 //unknown
	}
	code, err := strconv.Atoi(codeString)
	if err != nil {
		return 0 //unknown
	}
	return code
}
