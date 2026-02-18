package main

import (
	"os"

	logger "github.com/gardener/gardener/pkg/logger"
	runtimelog "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager/signals"

	"github.com/metal-stack/gardener-extension-csi-driver-synology/cmd/gardener-extension-csi-driver-synology/app"
)

func main() {
	runtimelog.SetLogger(logger.MustNewZapLogger(logger.InfoLevel, logger.FormatJSON))
	cmd := app.NewControllerCommand(signals.SetupSignalHandler())

	if err := cmd.Execute(); err != nil {
		runtimelog.Log.Error(err, "error executing the main controller command")
		os.Exit(1)
	}
}
