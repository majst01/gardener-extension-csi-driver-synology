package main

import (
	"context"
	"os"

	"github.com/metal-stack/gardener-extension-csi-driver-synology/cmd/gardener-extension-csi-driver-synology/app"
	runtimelog "sigs.k8s.io/controller-runtime/pkg/log"
)

func main() {
	ctx := context.Background()

	cmd := app.NewControllerCommand(ctx)

	if err := cmd.Execute(); err != nil {
		runtimelog.Log.Error(err, "error executing the main controller command")
		os.Exit(1)
	}
}
