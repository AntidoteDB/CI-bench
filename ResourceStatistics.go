package main

import (
	"time"
	"github.com/google/cadvisor/info/v1"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"fmt"
	cadvisor "github.com/google/cadvisor/client"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"context"
)

type ResourceStatistics struct {
	Container      string
	AvgMem         float64
	MaxMem         float64
	Cpu            float64
	NetReceived    float64
	NetTransmitted float64
	DiskRead       float64
	DiskWrite      float64
}

func startStats() (string, error) {
	image := "google/cadvisor"

	containerConfig := &container.Config{
		Image: image,
	}

	hostConfig := &container.HostConfig{
		Mounts: []mount.Mount{
			{
				Type:   mount.TypeBind,
				Source: "/",
				Target: "/rootfs",
				ReadOnly: true,
			},
			{
				Type:   mount.TypeBind,
				Source: "/var/run",
				Target: "/var/run",
				ReadOnly: true,
			},
			{
				Type:   mount.TypeBind,
				Source: "/sys",
				Target: "/sys",
				ReadOnly: true,
			},
			{
				Type:   mount.TypeBind,
				Source: "/var/lib/docker/",
				Target: "/var/lib/docker",
				ReadOnly: true,
			},
			{
				Type:   mount.TypeBind,
				Source: "/dev/disk/",
				Target: "/dev/disk",
				ReadOnly: true,
			},
		},
	}

	id, err := startContainer(image, containerConfig, hostConfig)

	if err != nil {
		return id, err
	}

	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return id, err
	}
	if err := cli.NetworkConnect(ctx, "benchmark-net", id, &network.EndpointSettings{Aliases: []string{"cadvisor"}}); err != nil {
			return "", err
	}
	return id, nil
}


func collectStats(start time.Time, end time.Time, dbContainer *[]DbContainer) (*[]ResourceStatistics, error) {
	cadvisorClient, err := cadvisor.NewClient("http://cadvisor:8080/")
	if err != nil {
		return nil, err
	}

	options := &v1.ContainerInfoRequest{
		NumStats: -1,
		Start: start,
		End: end,
	}

	resourceStatistics := make([]ResourceStatistics, len(*dbContainer))

	for i,c := range *dbContainer {
		containerInfo, err := cadvisorClient.ContainerInfo("/docker/" + c.Id, options)

		if err != nil {
			return nil, err
		}

		if len(containerInfo.Stats) == 0 {
			continue
		}

		var mem uint64
		var maxMem uint64

		first := containerInfo.Stats[0]
		last := containerInfo.Stats[len(containerInfo.Stats) - 1]

		for _,s := range containerInfo.Stats {
			mem += s.Memory.Usage
			if s.Memory.Usage > maxMem {
				maxMem = s.Memory.Usage
			}

		}

		var eth0first, eth0last *v1.InterfaceStats
		for _,interf := range first.Network.Interfaces {
			if interf.Name == "eth0" {
				eth0first = &interf
				break
			}
		}
		for _,interf := range last.Network.Interfaces {
			if interf.Name == "eth0" {
				eth0last = &interf
				break
			}
		}

		var read, write uint64
		for _, io := range last.DiskIo.IoServiceBytes {
			read += io.Stats["Read"]
			write += io.Stats["Write"]
		}
		for _, io := range first.DiskIo.IoServiceBytes {
			read -= io.Stats["Read"]
			write -= io.Stats["Write"]
		}

		containerStatistics := ResourceStatistics{
			Container: c.Name,
			AvgMem: float64(mem) / float64(len(containerInfo.Stats)),
			MaxMem: float64(maxMem),
			Cpu: float64(last.Cpu.Usage.User - first.Cpu.Usage.User) * 1e-9,
			DiskRead: float64(read),
			DiskWrite: float64(write),
		}

		if eth0first !=nil && eth0last !=nil {
			containerStatistics.NetReceived = float64(eth0last.RxBytes) - float64(eth0first.RxBytes)
			containerStatistics.NetTransmitted = float64(eth0last.TxBytes) - float64(eth0first.TxBytes)
		}

		printStatistics(containerStatistics)
		resourceStatistics[i] = containerStatistics
	}
	return &resourceStatistics, nil
}

const (
	KILOBYTE = 1e3
	MEGABYTE = 1e6
	GIGABYTE = 1e9
	TERABYTE = 1e12
)

func formatBytes(bytes float64) string {
	unit := ""
	value := bytes

	switch {
	case bytes >= TERABYTE:
		unit = "TB"
		value = value / TERABYTE
	case bytes >= GIGABYTE:
		unit = "GB"
		value = value / GIGABYTE

	case bytes >= MEGABYTE:
		unit = "MB"
		value = value / MEGABYTE
	case bytes >= KILOBYTE:
		unit = "kB"
		value = value / KILOBYTE
	default:
		unit = "B"
	}
	return fmt.Sprintf("%.2f%s", value, unit)
}

func printStatistics(statistics ResourceStatistics) {
	fmt.Printf("Container: %v\n", statistics.Container)
	fmt.Printf("Avg Memory usage %s\n", formatBytes(statistics.AvgMem))
	fmt.Printf("Max Memory usage %s\n", formatBytes(statistics.MaxMem))
	fmt.Printf("CPU usage %.2fs\n", statistics.Cpu) //cpu seconds
	fmt.Printf("Network received %s\n", formatBytes(statistics.NetReceived))
	fmt.Printf("Network transmitted %s\n", formatBytes(statistics.NetTransmitted))
	fmt.Printf("Disk read %s\n", formatBytes(statistics.DiskRead))
	fmt.Printf("Disk write %s\n", formatBytes(statistics.DiskWrite))
}
