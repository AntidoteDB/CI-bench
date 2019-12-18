package main

import (
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/api/types/container"
	"strconv"
)

func applyDelay(delay int) (string, error) {
	image := "gaiaadm/pumba"

	containerConfig := &container.Config{
		Image: image,
		Cmd: []string{"netem", "--duration", "1h", "delay", "--time", strconv.Itoa(delay), "re2:^dc"},
	}

	hostConfig := &container.HostConfig{
		Mounts: []mount.Mount{
			{
				Type:   mount.TypeBind,
				Source: "/var/run/docker.sock",
				Target: "/var/run/docker.sock",
				ReadOnly: false,
			},
		},
	}

	return startContainer(image, containerConfig, hostConfig)
}

func applyLoss(loss int) (string, error) {
	image := "gaiaadm/pumba"

	containerConfig := &container.Config{
		Image: image,
		Cmd: []string{"netem", "--duration", "1h", "loss", "--percent", strconv.Itoa(loss), "re2:^dc"},
	}

	hostConfig := &container.HostConfig{
		Mounts: []mount.Mount{
			{
				Type:   mount.TypeBind,
				Source: "/var/run/docker.sock",
				Target: "/var/run/docker.sock",
				ReadOnly: false,
			},
		},
	}

	return startContainer(image, containerConfig, hostConfig)
}

func applyRate(rate int) (string, error) {
	image := "gaiaadm/pumba"

	containerConfig := &container.Config{
		Image: image,
		Cmd: []string{"netem", "--duration", "1h", "rate", "--rate", strconv.Itoa(rate), "re2:^dc"},
	}

	hostConfig := &container.HostConfig{
		Mounts: []mount.Mount{
			{
				Type:   mount.TypeBind,
				Source: "/var/run/docker.sock",
				Target: "/var/run/docker.sock",
				ReadOnly: false,
			},
		},
	}

	return startContainer(image, containerConfig, hostConfig)
}
