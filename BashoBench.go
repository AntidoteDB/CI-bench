package main

import (
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/client"
	"github.com/docker/docker/api/types"
	"context"
	"github.com/docker/docker/api/types/network"
)

func runBashoBench(configPath string) error {
	image := "antidotedb:bashobench"

	containerConfig := &container.Config{
		Image: image,
		Cmd: []string{"bash", "-c", "cd /opt && ./basho_bench antidote_pb.config && cat tests/current/update-only-txn_latencies.csv"},
	}

	hostConfig := &container.HostConfig{
		Mounts: []mount.Mount{
			{
				Type:   mount.TypeBind,
				Source: configPath,
				Target: "/opt/antidote_pb.config",
				ReadOnly: true,
			},
		},
	}

	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())

	if err != nil {
		return  err
	}

	resp, err := cli.ContainerCreate(ctx, containerConfig, hostConfig, nil, nil, "")

	if err != nil {
		return  err
	}

	if err := cli.NetworkConnect(ctx, "benchmark_default", resp.ID, &network.EndpointSettings{}); err != nil {
		return err
	}

	if err := cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		return  err
	}
	outputLog(resp.ID)
	stopContainer(resp.ID)
	return nil
}
