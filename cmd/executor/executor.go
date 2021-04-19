package main

import (
	"fmt"
	"os"

	executor "github.com/juanfont/gitlab-windows-custom-executor"
	"github.com/juanfont/gitlab-windows-custom-executor/drivers"
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
		cfg := drivers.VcdDriverConfig{
			VcdURL:           "",
			VcdOrg:           "",
			VcdVdc:           "",
			VcdInsecure:      false,
			VcdUser:          "",
			VcdPassword:      "",
			VcdOrgVDCNetwork: "",
			Catalog:          "",
			Template:         "",
			NumCpus:          0,
			CoresPerSocket:   0,
			MemorySizeMb:     0,
			VAppHREF:         "",
			VMHREF:           "",
			Description:      "",
			StorageProfile:   "",
		}
		vcd, _ := drivers.NewVcdDriver(cfg)
		e, _ := executor.NewExecutor(vcd)
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
