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

	// aggOption.AddFlags(cmd.Flags())

	// cmd.Flags().StringVar(&cfg.SynologyHost, "synology-host", "", "Synology NAS host")
	// cmd.Flags().IntVar(&cfg.SynologyPort, "synology-port", 5000, "Synology NAS port")
	// cmd.Flags().BoolVar(&cfg.SynologySSL, "synology-ssl", false, "Use SSL for Synology connection")
	// cmd.Flags().BoolVar(&cfg.ChapEnabled, "chap-enabled", true, "Enable CHAP authentication")
	// cmd.Flags().StringVar(&cfg.AdminUsername, "admin-username", "", "Synology admin username")
	// cmd.Flags().StringVar(&cfg.AdminPassword, "admin-password", "", "Synology admin password")

	return cmd
}
