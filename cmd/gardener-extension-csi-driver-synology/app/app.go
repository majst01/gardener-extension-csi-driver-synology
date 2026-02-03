package app

import (
	"context"
	"fmt"

	extensioncontroller "github.com/gardener/gardener/extensions/pkg/controller"
	"github.com/gardener/gardener/extensions/pkg/controller/cmd"
	"github.com/gardener/gardener/extensions/pkg/util"
	"github.com/metal-stack/gardener-extension-csi-driver-synology/pkg/apis/config"
	"github.com/metal-stack/gardener-extension-csi-driver-synology/pkg/constants"
	"github.com/metal-stack/gardener-extension-csi-driver-synology/pkg/controller/healthcheck"
	"github.com/metal-stack/gardener-extension-csi-driver-synology/pkg/controller/lifecycle"
	"github.com/spf13/cobra"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

// NewControllerCommand creates a new command for running the Synology CSI extension controller
func NewControllerCommand(ctx context.Context) *cobra.Command {
	var (
		restOpts = &cmd.RESTOptions{}
		mgrOpts  = &cmd.ManagerOptions{
			LeaderElection:          true,
			LeaderElectionID:        cmd.LeaderElectionNameID(constants.ServiceName),
			LeaderElectionNamespace: util.GetEnvOrDefault("LEADER_ELECTION_NAMESPACE", ""),
		}
		cfg = &config.Configuration{
			SynologyPort: 5000,
		}

		aggOption = cmd.NewOptionAggregator(
			restOpts,
			mgrOpts,
		)
	)

	cmd := &cobra.Command{
		Use:   constants.ServiceName,
		Short: "Synology CSI Extension Controller",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := aggOption.Complete(); err != nil {
				return fmt.Errorf("error completing options: %w", err)
			}

			util.ApplyClientConnectionConfigurationToRESTConfig(
				&extensioncontroller.ClientConnection{
					QPS:   100.0,
					Burst: 130,
				},
				restOpts.Completed().Config,
			)

			mgr, err := manager.New(restOpts.Completed().Config, mgrOpts.Completed().Options())
			if err != nil {
				return fmt.Errorf("could not instantiate manager: %w", err)
			}

			if err := extensioncontroller.AddToScheme(mgr.GetScheme()); err != nil {
				return fmt.Errorf("could not update manager scheme: %w", err)
			}

			if err := lifecycle.AddToManager(mgr, cfg); err != nil {
				return fmt.Errorf("could not add lifecycle controller: %w", err)
			}

			if err := healthcheck.AddToManager(mgr); err != nil {
				return fmt.Errorf("could not add healthcheck controller: %w", err)
			}

			if err := mgr.Start(ctx); err != nil {
				return fmt.Errorf("error running manager: %w", err)
			}

			return nil
		},
	}

	aggOption.AddFlags(cmd.Flags())

	cmd.Flags().StringVar(&cfg.SynologyHost, "synology-host", "", "Synology NAS host")
	cmd.Flags().IntVar(&cfg.SynologyPort, "synology-port", 5000, "Synology NAS port")
	cmd.Flags().BoolVar(&cfg.SynologySSL, "synology-ssl", false, "Use SSL for Synology connection")
	cmd.Flags().BoolVar(&cfg.ChapEnabled, "chap-enabled", true, "Enable CHAP authentication")
	cmd.Flags().StringVar(&cfg.AdminUsername, "admin-username", "", "Synology admin username")
	cmd.Flags().StringVar(&cfg.AdminPassword, "admin-password", "", "Synology admin password")

	return cmd
}
