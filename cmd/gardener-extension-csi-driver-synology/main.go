package main

import (
	"context"
	"os"

	"github.com/gardener/gardener/cmd/utils"
	"github.com/metal-stack/gardener-extension-csi-driver-synology/cmd/gardener-extension-csi-driver-synology/app"
)

func main() {
	ctx := context.Background()

	cmd := app.NewControllerCommand(ctx)

	if err := cmd.Execute(); err != nil {
		utils.PrintError(err)
		os.Exit(1)
	}
}
