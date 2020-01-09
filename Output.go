package main

import (
	"os"
	"fmt"
	"bufio"
)


func writeStatisticSummaryToFile(benchmarkName string, resourceStatistics *[]ResourceStatistics) error {
	for _, stats := range *resourceStatistics {
		fileName := "/output/" + benchmarkName + "-" + stats.Container + ".csv"

		if _, err := os.Stat(fileName); os.IsNotExist(err) {
			if err := createCSVFile(fileName, "avgmem,maxmem,cpu,trans,rec,read,write\n"); err != nil {
				return fmt.Errorf("error creating file %v: %v", fileName, err)
			}
		}

		f, err := os.OpenFile(fileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return fmt.Errorf("error writing to file %v: %v", fileName, err)
		}
		defer f.Close()

		s := fmt.Sprintf("%.f,%.f,%.f,%.f,%.f,%.f,%.f\n", stats.AvgMem, stats.MaxMem, stats.Cpu, stats.NetTransmitted, stats.NetReceived, stats.DiskRead, stats.DiskWrite)
		_, err = f.WriteString(s)
		if err != nil {
			return fmt.Errorf("error writing to file %v: %v", fileName, err)
		}
	}
	return nil
}

func writeResultSummaryToFile(benchmarkName string, result BenchmarkResult) error {
	fileName := "/output/" + benchmarkName + ".csv"

	if _, err := os.Stat(fileName); os.IsNotExist(err) {
		if err := createCSVFile(fileName, "min,max,avg,rps,failed\n"); err != nil {
			return fmt.Errorf("error creating file %v: %v", fileName, err)
		}
	}

	f, err := os.OpenFile(fileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("error writing to file %v: %v", fileName, err)
	}
	defer f.Close()

	s := fmt.Sprintf("%d,%d,%d,%.2f,%d\n", result.Min.Milliseconds(), result.Max.Milliseconds(), result.Avg.Milliseconds(), result.Rps, result.Failed)
	_, err = f.WriteString(s)
	if err != nil {
		return fmt.Errorf("error writing to file %v: %v", fileName, err)
	}
	return nil
}

func writeResultsToFile(benchmarkName string, runId string, results *[]RequestResult) error {
	fileName := "/tmp/results_" + benchmarkName + "_" + runId + ".csv"

	if _, err := os.Stat(fileName); os.IsNotExist(err) {
		if err := createCSVFile(fileName, "start,latency,failed,code\n"); err != nil {
			return fmt.Errorf("error creating file %v: %v", fileName, err)
		}
	}

	f, err := os.OpenFile(fileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("error writing to file %v: %v", fileName, err)
	}
	defer f.Close()

	w := bufio.NewWriter(f)
	for _,result := range *results {
		s := fmt.Sprintf("%d,%d,%t,%d\n", result.start, result.latency.Nanoseconds(), result.failed, result.errorCode)
		_, err := w.WriteString(s)
		if err != nil {
			return fmt.Errorf("error writing to file %v: %v", fileName, err)
		}
	}

	err = w.Flush()
	if err != nil {
		return fmt.Errorf("error writing to file %v: %v", fileName, err)
	}
	return nil
}


func createCSVFile(fileName string, header string) error {
	f, err := os.Create(fileName)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.WriteString(header)
	return err
}

