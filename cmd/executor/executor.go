package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/rs/zerolog/log"

	machine "github.com/juanfont/gitlab-machine"
	"github.com/juanfont/gitlab-machine/pkg/drivers/vcd"
	"github.com/rs/zerolog"
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
		vcdDriver := getVcdDriver()
		e, _ := machine.NewExecutor(vcdDriver)

		if vcdDriver.MOTD() != "" {
			log.Info().Msg(vcdDriver.MOTD())
		}

		err := e.Prepare()
		if err != nil {
			log.Fatal().Err(err).Msg("Error preparing executor")
		}
	},
}

var runVcdCmd = &cobra.Command{
	Use:   "run PATH STAGE",
	Short: "Prepare a new instance of the vCloud Director executor",
	Long:  "",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) < 2 {
			return fmt.Errorf("missing parameters")
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		vcd := getVcdDriver()
		e, _ := machine.NewExecutor(vcd)
		err := e.Run(args[0], args[1])
		if err != nil {
			log.Fatal().Err(err).Msg("Error running the command")
		}
	},
}

var cleanupVcdCmd = &cobra.Command{
	Use:   "cleanup",
	Short: "Remove the current executor",
	Long:  "",
	Run: func(cmd *cobra.Command, args []string) {
		vcd := getVcdDriver()
		e, _ := machine.NewExecutor(vcd)
		err := e.CleanUp()
		if err != nil {
			log.Fatal().Err(err).Msg("Error cleaning up executor")
		}
	},
}

var shellVcdCmd = &cobra.Command{
	Use:   "shell cmd",
	Short: "Opens a shell with the specified command",
	Long:  "",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return fmt.Errorf("missing parameters")
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		vcd := getVcdDriver()
		e, _ := machine.NewExecutor(vcd)
		err := e.Shell(args[0])
		if err != nil {
			log.Fatal().Err(err).Msg("Error creating executor")
		}
	},
}

var executorCmd = &cobra.Command{
	Use:   "executor",
	Short: "executor - a Gitlab Custom Executor",
	Long: `
A custom executor for Gitlab
Juan Font Alonso <juanfontalonso@gmail.com> - 2022
https://gitlab.com/juanfont/gitlab-machine`,
}

func getVcdDriver() *vcd.VcdDriver {
	cfg := vcd.VcdDriverConfig{
		VcdURL:           viper.GetString("vcd.url"),
		VcdOrg:           viper.GetString("vcd.org"),
		VcdVdc:           viper.GetString("vcd.vdc"),
		VcdInsecure:      viper.GetBool("vcd.insecure"),
		VcdUser:          viper.GetString("vcd.user"),
		VcdPassword:      viper.GetString("vcd.password"),
		VcdOrgVDCNetwork: viper.GetString("vcd.vdc_network"),
		Catalog:          viper.GetString("vcd.catalog"),
		Template:         viper.GetString("vcd.template"),
		NumCpus:          viper.GetInt("vcd.num_cpus"),
		CoresPerSocket:   viper.GetInt("vcd.cores_per_socket"),
		MemorySizeMb:     viper.GetInt("vcd.memory_mb"),
		Description:      "Created by gitlab-machine",
		StorageProfile:   viper.GetString("vcd.storage_profile"),

		DefaultPassword: viper.GetString("vcd.default_password"), // I dont like this

		MOTD: viper.GetString("vcd.motd"),
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

func main() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Logger = log.Output(zerolog.ConsoleWriter{
		Out:        os.Stdout,
		TimeFormat: time.RFC3339,
		NoColor:    false,
	})

	viper.SetConfigName("config")
	viper.AddConfigPath(".")
	viper.AddConfigPath("/opt/gitlab-machine")
	ex, err := os.Executable()
	if err != nil {
		panic(err)
	}

	exPath := filepath.Dir(ex)
	viper.AddConfigPath(exPath)
	viper.AutomaticEnv()
	viper.ReadInConfig()

	logLevel := viper.GetString("log_level")
	if logLevel == "" {
		logLevel = "info"
	}
	level, err := zerolog.ParseLevel(logLevel)
	if err != nil {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	} else {
		zerolog.SetGlobalLevel(level)
	}

	executorCmd.AddCommand(versionCmd)
	executorCmd.AddCommand(vcdCmd)

	// vCloud Director implementations
	vcdCmd.AddCommand(prepareVcdCmd)
	vcdCmd.AddCommand(runVcdCmd)
	vcdCmd.AddCommand(cleanupVcdCmd)
	vcdCmd.AddCommand(shellVcdCmd)

	if err := executorCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}
