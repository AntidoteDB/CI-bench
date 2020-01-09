package main

import (
	"os"
	"bufio"
	"fmt"
	"html/template"
	"os/exec"
)

type ReportData struct {
	Name 		string
	Summary BenchmarkResult
	ResourceStatistics []ResourceStatistics
}

func generateReport(name string, summary *BenchmarkResult, resourceStatistics *[]ResourceStatistics) error {
	path := "/output/report/" +  name
	fileName := path + "/report.html"

	data := ReportData{
		Name:      name,
		Summary: *summary,
		ResourceStatistics: *resourceStatistics,
	}

	if err := os.MkdirAll(path, 0775); err != nil {
		return err
	}
	f, err := os.OpenFile(fileName, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("!error writing to file %v: %v", fileName, err)
	}
	defer f.Close()

	w := bufio.NewWriter(f)

	tmpl, err := template.New("report.html").Funcs(template.FuncMap{
		"ByteFormat": formatBytes, //TODO
	}).ParseFiles("./report/report.html")
	if err != nil {
		return fmt.Errorf("error parsing html template: %v", err)
	}
	err = tmpl.Execute(w, data)
	if err != nil {
		return fmt.Errorf("error parsing html template: %v", err)
	}

	err = w.Flush()
	if err != nil {
		return fmt.Errorf("error writing to file %v: %v", fileName, err)
	}
	return nil
}

func generateVis(name string, resultFile string) error {
	outPath := "/output/report/" + name + "/img"
	if err := os.MkdirAll(outPath, 0775); err != nil {
		return err
	}
	cmd := exec.Command("python3", "/go/src/benchmark/report/vis.py", "-o", outPath, "-i", resultFile)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}