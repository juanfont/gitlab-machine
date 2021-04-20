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

var vcdCmd = &cobra.Command{
	Use:   "vcd",
	Short: "Manage vCloud Director gitlab-machine driver",
}

var prepareVcdCmd = &cobra.Command{
	Use:   "prepare",
	Short: "Prepare a new instance of the vCloud Director executor",
	Long:  "",
	Run: func(cmd *cobra.Command, args []string) {
		vcd := getVcdDriver()
		e, _ := machine.NewExecutor(vcd)
		err := e.Prepare()
		fmt.Println(err)
	},
}

var runVcdCmd = &cobra.Command{
	Use:   "run PATH STAGE",
	Short: "Prepare a new instance of the vCloud Director executor",
	Long:  "",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) < 2 {
			return fmt.Errorf("Missing parameters")
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		vcd := getVcdDriver()
		e, _ := machine.NewExecutor(vcd)
		err := e.Run(args[0], args[1])
		fmt.Println(err)
	},
}

var cleanupVcdCmd = &cobra.Command{
	Use:   "cleanup",
	Short: "Prepare a new instance of the vCloud Director executor",
	Long:  "",
	Run: func(cmd *cobra.Command, args []string) {
		vcd := getVcdDriver()
		e, _ := machine.NewExecutor(vcd)
		err := e.CleanUp()
		fmt.Println(err)
	},
}

var executorCmd = &cobra.Command{
	Use:   "executor",
	Short: "executor - a Gitlab Custom Executor",
	Long: fmt.Sprintf(`
A custom executor for Gitlab
Juan Font Alonso <juanfontalonso@gmail.com> - 2021
https://gitlab.com/juanfont/gitlab-machine`),
}

func getVcdDriver() *vcd.VcdDriver {
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
		NumCpus:          viper.GetInt("vcd_num_cpus"),
		CoresPerSocket:   viper.GetInt("vcd_cores_per_socket"),
		MemorySizeMb:     viper.GetInt("vcd_memory_mb"),
		Description:      "Created by gitlab-machine",
		StorageProfile:   viper.GetString("vcd_storage_profile"),

		DefaultPassword: viper.GetString("vcd_default_password"), // I dont like this
	}

	machineName := fmt.Sprintf(
		"machine-%s-project-%s-concurrent-%s-job-%s",
		os.Getenv("CUSTOM_ENV_CI_RUNNER_ID"),
		os.Getenv("CUSTOM_ENV_CI_PROJECT_ID"),
		os.Getenv("CUSTOM_ENV_CI_CONCURRENT_PROJECT_ID"),
		os.Getenv("CUSTOM_ENV_CI_JOB_ID"),
	)

	vcd, _ := vcd.NewVcdDriver(cfg, machineName)
	return vcd
}

func main() {
	viper.SetConfigName("config")
	viper.AddConfigPath(".")
	viper.AddConfigPath("/opt/gitlab-machine")
	viper.AutomaticEnv()
	viper.ReadInConfig()

	executorCmd.AddCommand(versionCmd)
	executorCmd.AddCommand(vcdCmd)

	// vCloud Director implementations
	vcdCmd.AddCommand(prepareVcdCmd)
	vcdCmd.AddCommand(runVcdCmd)
	vcdCmd.AddCommand(cleanupVcdCmd)

	if err := executorCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}
