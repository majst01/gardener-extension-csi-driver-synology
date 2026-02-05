package app

import (
	"context"
	"fmt"

	"github.com/metal-stack/gardener-extension-csi-driver-synology/pkg/constants"
	"github.com/spf13/cobra"
)

// NewControllerCommand creates a new command for running the Synology CSI extension controller
func NewControllerCommand(ctx context.Context) *cobra.Command {
	options := NewOptions()

	cmd := &cobra.Command{
		Use:           constants.ExtensionName,
		Short:         "Synology CSI Extension Controller",
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := options.optionAggregator.Complete(); err != nil {
				return fmt.Errorf("error completing options: %w", err)
			}

			if err := options.heartbeatOptions.Validate(); err != nil {
				return err
			}

			cmd.SilenceUsage = true

			return options.run(ctx)
		},
	}

	options.optionAggregator.AddFlags(cmd.Flags())

	return cmd
}
