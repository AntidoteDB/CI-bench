package main

import (
	"os/exec"
	"os"
	"time"
	"github.com/docker/docker/client"
	"context"
	"errors"
	"fmt"
	"github.com/docker/docker/api/types"
	"strconv"
	"sort"
)

type DbContainer struct {
	Id string
	Name string
	Dc int
	Node int
	IPAddress string
}

func startDB(composePath string) error {
	cmd := exec.Command("docker-compose", "-f", composePath, "-p", "benchmark", "up", "-d")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func stopDB(composePath string) {
	cmd := exec.Command("docker-compose", "-f", composePath, "-p", "benchmark", "down")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Run()
}

func waitForStart() error {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		panic(err)
	}

	waitBodyC, errC := cli.ContainerWait(ctx, "benchmark_link-dcs_1", "not-running")

	select {
	case err := <-errC :
		fmt.Printf("Error connecting Dcs: %v \n", err)
		return err
	case waitBody := <- waitBodyC :
		if waitBody.StatusCode != 0 {
			return errors.New(fmt.Sprintf("Could not connect Dcs. Error Code: %v", waitBody.StatusCode))
		}
	}
	return nil
}

func getDbContainer() (*[]DbContainer, error) {
	cli, _ := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	ctx := context.Background()
	container, _ := cli.ContainerList(ctx, types.ContainerListOptions{})
	res := make([]DbContainer, 0)

	for _,c := range container {
		if _, ok := c.Labels["eu.antidotedb.dc"]; ok == true {
			dc, err := strconv.Atoi(c.Labels["eu.antidotedb.dc"])
			if err != nil {
				return nil, fmt.Errorf("could not parse dc for container: %v, %v", c.Names, err)
			}
			node, err := strconv.Atoi(c.Labels["eu.antidotedb.node"])
			if err != nil {
				return nil, fmt.Errorf("could not parse node for container: %v, %v", c.Names, err)
			}
			dbContainer := DbContainer{
				Id: c.ID,
				Name: c.Labels["eu.antidotedb.name"],
				Dc: dc,
				Node: node,
			}
			res = append(res, dbContainer)

			for _, network := range c.NetworkSettings.Networks {
				dbContainer.IPAddress = network.IPAddress
				break

			}
			if dbContainer.IPAddress == "" {
				return nil, fmt.Errorf("could not get ip address for container: %v", c.Names)
			}
		}
	}

	//sort result by dc and node
	sort.Slice(res, func(i, j int) bool {
		return res[i].Name < res[j].Name
	})
	return &res, nil
}