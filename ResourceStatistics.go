package main

import (
	"time"
	"github.com/google/cadvisor/info/v1"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/go-connections/nat"
	"github.com/docker/docker/api/types/mount"
	"fmt"
	cadvisor "github.com/google/cadvisor/client"
)

func startStats() (string, error) {
	image := "google/cadvisor"

	containerConfig := &container.Config{
		Image: image,
		ExposedPorts: nat.PortSet{
			"8080/tcp": struct{}{},
		},
	}

	hostConfig := &container.HostConfig{
		PortBindings: nat.PortMap{
			"8080/tcp": []nat.PortBinding{
				{
					HostIP: "0.0.0.0",
					HostPort: "8080",
				},
			},
		},
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

	return startContainer(image, containerConfig, hostConfig)
}


func collectStats(start time.Time, end time.Time, dbContainer *[]DbContainer) {
	cadvisorClient, err := cadvisor.NewClient("http://localhost:8080/")
	if err != nil {
		panic(err)
	}

	options := &v1.ContainerInfoRequest{
		NumStats: -1,
		Start: start,
		End: end,
	}

	for _,c := range *dbContainer {
		fmt.Printf("Container: %v\n", c.Name)
		containerInfo, err := cadvisorClient.ContainerInfo("/docker/" + c.Id, options)

		if err != nil {
			panic(err)
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



		fmt.Printf("Avg Memory usage %s\n", formatBytesFloat(float64(mem) / float64(len(containerInfo.Stats))))
		fmt.Printf("Max Memory usage %s\n", formatBytes(maxMem))
		fmt.Printf("Max Memory usage %s\n", formatBytes(last.Memory.MaxUsage))

		fmt.Printf("CPU usage %.2fs\n", float64(last.Cpu.Usage.User - first.Cpu.Usage.User) * 1e-9) //cpu seconds

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

		if eth0first !=nil && eth0last !=nil {
			fmt.Printf("Network recieved %s\n", formatBytes(eth0last.RxBytes - eth0first.RxBytes))
			fmt.Printf("Network transmitted %s\n", formatBytes(eth0last.TxBytes - eth0first.TxBytes))
		}


		//for _, io := range first.DiskIo.IoServiceBytes {
		//		//	fmt.Printf("%v\n", io.Device)
		//		//	fmt.Printf("%v\n", io.Stats["Read"])
		//		//	fmt.Printf("%v\n", io.Stats["Write"])
		//		//}
		//		//for _, io := range last.DiskIo.IoServiceBytes {
		//		//	fmt.Printf("%v\n", io.Device)
		//		//	fmt.Printf("%v\n", io.Stats["Read"])
		//		//	fmt.Printf("%v\n", io.Stats["Write"])
		//		//}

		var read, write uint64
		for _, io := range last.DiskIo.IoServiceBytes {
			read += io.Stats["Read"]
			write += io.Stats["Write"]
		}
		for _, io := range first.DiskIo.IoServiceBytes {
			read -= io.Stats["Read"]
			write -= io.Stats["Write"]
		}
		fmt.Printf("Disk read %s\n", formatBytes(read))
		fmt.Printf("Disk write %s\n", formatBytes(write))
	}
}

const (
	KILOBYTE = 1e3
	MEGABYTE = 1e6
	GIGABYTE = 1e9
	TERABYTE = 1e12
)

func formatBytes(bytes uint64) string {
	return formatBytesFloat(float64(bytes))
}

func formatBytesFloat(bytes float64) string {
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
