package main

import (
	"os"
	"fmt"
)


func writeStatisticsToFile(benchmarkName string, resourceStatistics *[]ResourceStatistics) error {
	for _, stats := range *resourceStatistics {
		fileName := "/output/" + benchmarkName + "-" + stats.Container + ".csv"

		if _, err := os.Stat(fileName); os.IsNotExist(err) {
			if err := createCSVFile(fileName, "avgmem;maxmem;cpu;trans;rec;read;write\n"); err != nil {
				return fmt.Errorf("error creating file %v: %v", fileName, err)
			}
		}

		f, err := os.OpenFile(fileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return fmt.Errorf("error writing to file %v: %v", fileName, err)
		}
		defer f.Close()

		s := fmt.Sprintf("%.f;%.f;%.f;%.f;%.f;%.f;%.f\n", stats.AvgMem, stats.MaxMem, stats.Cpu, stats.NetTransmitted, stats.NetReceived, stats.DiskRead, stats.DiskWrite)
		_, err = f.WriteString(s)
		if err != nil {
			return fmt.Errorf("error writing to file %v: %v", fileName, err)
		}
	}
	return nil
}

func writeResultToFile(benchmarkName string, result BenchmarkResult) error{
	fileName := "/output/" + benchmarkName + ".csv"

	if _, err := os.Stat(fileName); os.IsNotExist(err) {
		if err := createCSVFile(fileName, "min;max;avg;rps;failed\n"); err != nil {
			return fmt.Errorf("error creating file %v: %v", fileName, err)
		}
	}

	f, err := os.OpenFile(fileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("error writing to file %v: %v", fileName, err)
	}
	defer f.Close()

	s := fmt.Sprintf("%s;%s;%s;%.2f;%d\n", result.min.String(), result.max.String(), result.avg.String(), result.rps, result.failed)
	_, err = f.WriteString(s)
	return fmt.Errorf("error writing to file %v: %v", fileName, err)
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

