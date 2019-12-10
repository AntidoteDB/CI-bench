package main

import (
	"os/exec"
	"os"
	"context"
)

func delay(ctx context.Context) {
	cmd := exec.CommandContext(ctx, "pumba", "netem", "--duration", "1h", "delay", "--time", "10", "re2:^dc")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Start()
}
