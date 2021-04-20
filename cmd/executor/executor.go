package main

import (
	"fmt"
	"os"

	machine "github.com/juanfont/gitlab-machine"
	"github.com/juanfont/gitlab-machine/drivers/vcd"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const version = "0.1"

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version.",
	Long:  "The version of the executor.",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(version)
	},
}

var prepareCmd = &cobra.Command{
	Use:   "prepare",
	Short: "Prepare a new instance of the executor.",
	Long:  "",
	Run: func(cmd *cobra.Command, args []string) {
		cfg := vcd.VcdDriverConfig{
			VcdURL:           viper.GetString("vcd_url"),
			VcdOrg:           viper.GetString("vcd_org"),
			VcdVdc:           viper.GetString("vcd_vdc"),
			VcdInsecure:      viper.GetBool("vcd_insecure"),
			VcdUser:          viper.GetString("vcd_user"),
			VcdPassword:      viper.GetString("vcd_password"),
			VcdOrgVDCNetwork: viper.GetString("vcd_vdc_network"),
			Catalog:          viper.GetString("vcd_catalog"),
			Template:         viper.GetString("vcd_template"),
			NumCpus:          0,
			CoresPerSocket:   0,
			MemorySizeMb:     0,
			Description:      "Create by the gitlab-custom-executor",
			StorageProfile:   "",
		}

		machineName := fmt.Sprintf(
			"runner-%s-project-%s-concurrent-%s-job-%s",
			os.Getenv("CUSTOM_ENV_CI_RUNNER_ID"),
			os.Getenv("CUSTOM_ENV_CI_PROJECT_ID"),
			os.Getenv("CUSTOM_ENV_CI_CONCURRENT_PROJECT_ID"),
			os.Getenv("CUSTOM_ENV_CI_JOB_ID"),
		)

		vcd, _ := vcd.NewVcdDriver(cfg, machineName)
		e, _ := machine.NewExecutor(vcd)
		e.Prepare()
	},
}

var executorCmd = &cobra.Command{
	Use:   "executor",
	Short: "executor - a Gitlab Custom Executor",
	Long: fmt.Sprintf(`
A custom executor for Gitlab
Juan Font Alonso <juanfontalonso@gmail.com> - 2021
https://gitlab.com/juanfont/gitlab-custom-executor`),
}

func main() {
	viper.SetConfigName("config")
	viper.AddConfigPath(".")
	viper.AutomaticEnv()
	viper.ReadInConfig()

	executorCmd.AddCommand(versionCmd)
	executorCmd.AddCommand(prepareCmd)

	if err := executorCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}
