package cmd

import (
	"fmt"
	"os"

	"mdc/internal/config"
	"mdc/internal/runner"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
	"github.com/spf13/cobra"
)

var psCmd = &cobra.Command{
	Use:   "ps [config-name]",
	Short: "Show container status for all projects",
	Long: `Show running container status.
If config-name is given, runs "docker compose ps" in each project directory.
Without arguments, runs "docker ps" to show all containers on the host.`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			containers, err := runner.CollectDockerPS()
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
			printDockerPSTable(containers)
			return
		}

		cfg, err := config.Load(args[0])
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}

		results := runner.CollectPS(cfg)
		printPSTable(results)
	},
}

func printPSTable(results []runner.ProjectContainers) {
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"PROJECT", "NAME", "PORTS", "STATUS"})

	hasRows := false
	var warnings []string

	for _, r := range results {
		if r.Err != nil {
			warnings = append(warnings, fmt.Sprintf("  ⚠️  %s: %s", r.ProjectName, r.Err))
			continue
		}
		for _, c := range r.Containers {
			hasRows = true
			status := colorizeState(c.State, c.Status)
			t.AppendRow(table.Row{r.ProjectName, c.Name, c.Ports, status})
		}
	}

	if hasRows {
		t.Render()
	} else if len(warnings) == 0 {
		fmt.Println("No running containers found.")
	}

	for _, w := range warnings {
		fmt.Fprintln(os.Stderr, w)
	}
}

func printDockerPSTable(containers []runner.ContainerInfo) {
	if len(containers) == 0 {
		fmt.Println("No running containers found.")
		return
	}

	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"NAME", "PORTS", "STATUS"})

	for _, c := range containers {
		status := colorizeState(c.State, c.Status)
		t.AppendRow(table.Row{c.Name, c.Ports, status})
	}
	t.Render()
}

func colorizeState(state, status string) string {
	switch state {
	case "running":
		return text.Colors{text.FgGreen}.Sprint(status)
	default:
		return text.Colors{text.FgRed}.Sprint(status)
	}
}

func init() {
	rootCmd.AddCommand(psCmd)
}
