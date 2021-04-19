package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

var headscaleCmd = &cobra.Command{
	Use:   "executor",
	Short: "executor - a Gitlab Custom Executor",
	Long: fmt.Sprintf(`
A custom executor for Gitlab
Juan Font Alonso <juanfontalonso@gmail.com> - 2021
https://gitlab.com/juanfont/gitlab-windows-custom-executor`),
}
