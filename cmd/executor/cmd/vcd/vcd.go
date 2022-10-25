package vcdcmd

import (
	"fmt"
	"os"

	"github.com/juanfont/gitlab-machine/pkg/drivers/vcd"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var VcdCmd = &cobra.Command{
	Use:   "vcd",
	Short: "Manage vCloud Director gitlab-machine driver",
}

func init() {
	VcdCmd.AddCommand(prepareVcdCmd)
	VcdCmd.AddCommand(runVcdCmd)
	VcdCmd.AddCommand(cleanupVcdCmd)
	VcdCmd.AddCommand(shellVcdCmd)
}

func getVcdDriver() *vcd.VcdDriver {
	cfg := vcd.VcdDriverConfig{
		VcdURL:           viper.GetString("drivers.vcd.url"),
		VcdOrg:           viper.GetString("drivers.vcd.org"),
		VcdVdc:           viper.GetString("drivers.vcd.vdc"),
		VcdInsecure:      viper.GetBool("drivers.vcd.insecure"),
		VcdUser:          viper.GetString("drivers.vcd.user"),
		VcdPassword:      viper.GetString("drivers.vcd.password"),
		VcdOrgVDCNetwork: viper.GetString("drivers.vcd.vdc_network"),
		Catalog:          viper.GetString("drivers.vcd.catalog"),
		Template:         viper.GetString("drivers.vcd.template"),
		NumCpus:          viper.GetInt("drivers.vcd.num_cpus"),
		CoresPerSocket:   viper.GetInt("drivers.vcd.cores_per_socket"),
		MemorySizeMb:     viper.GetInt("drivers.vcd.memory_mb"),
		Description:      "Created by gitlab-machine",
		StorageProfile:   viper.GetString("drivers.vcd.storage_profile"),

		DefaultPassword: viper.GetString("drivers.vcd.default_password"), // I dont like this
	}

	machineName := fmt.Sprintf(
		"gitlab-machine-%s-project-%s-concurrent-%s-job-%s",
		os.Getenv("CUSTOM_ENV_CI_RUNNER_ID"),
		os.Getenv("CUSTOM_ENV_CI_PROJECT_ID"),
		os.Getenv("CUSTOM_ENV_CI_CONCURRENT_PROJECT_ID"),
		os.Getenv("CUSTOM_ENV_CI_JOB_ID"),
	)

	vcd, _ := vcd.NewVcdDriver(cfg, machineName)
	return vcd
}
