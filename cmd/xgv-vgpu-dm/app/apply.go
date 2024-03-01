package app

import (
	"fmt"
	"os"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var applyFlags = Flags{}

func applyWrapper() error {
	err := CheckFlags(&applyFlags)
	if err != nil {
		return err
	}

	log.Debugf("Parsing config file...")
	spec, err := ParseConfigFile(&applyFlags)
	if err != nil {
		return fmt.Errorf("failed to parse config file: %v", err)
	}

	VGPUConfig, err := GetSelectedVGPUConfig(&applyFlags, spec)
	log.Debugf("Selecting specific vgpu configuration: %v", VGPUConfig)
	if err != nil {
		return fmt.Errorf("failed to select vgpu config: %v", err)
	}

	err = AssertVGPUConfig()
	if err != nil {
		log.Infoln("Apply vGPU device configuration...")
		return nil
	}

	log.Infof("Selected vGPU device configuration successfully applied")
	return nil
}

var applyCmd = &cobra.Command{
	Use:   "apply",
	Short: "Apply changes (if necessary) for a specific vGPU device configuration from a configuration file",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := applyWrapper(); err != nil {
			log.Errorln(err)
			return err
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(applyCmd)
	applyCmd.PersistentFlags().StringVarP(&applyFlags.ConfigFile, "config-file", "f", os.Getenv("XGV_VGPU_DM_CONFIG_FILE"), "Path to the configuration file")
	applyCmd.PersistentFlags().StringVarP(&applyFlags.SelectedConfig, "selected-config", "c", os.Getenv("XGV_VGPU_DM_SELECTED_CONFIG"), "The label of the vgpu-config from the config file to apply to the node")
}
