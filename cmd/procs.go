package cmd

import (
	"fmt"
	"os"
	"strings"

	"mdc/internal/pidfile"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/spf13/cobra"
)

var procsCmd = &cobra.Command{
	Use:   "procs [config-name]",
	Short: "List background processes managed by mdc",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		var allData map[string]map[string][]pidfile.Entry
		var err error

		if len(args) == 1 {
			configName := args[0]
			projects, loadErr := pidfile.LoadAll(configName)
			if loadErr != nil {
				fmt.Fprintln(os.Stderr, loadErr)
				os.Exit(1)
			}
			if len(projects) > 0 {
				allData = map[string]map[string][]pidfile.Entry{
					configName: projects,
				}
			}
		} else {
			allData, err = pidfile.LoadAllConfigs()
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
		}

		if len(allData) == 0 {
			fmt.Println("No background processes found.")
			return
		}

		t := table.NewWriter()
		t.SetOutputMirror(os.Stdout)
		t.AppendHeader(table.Row{"CONFIG", "PROJECT", "COMMAND", "DIR", "PID", "STATUS"})
		for configName, projects := range allData {
			for projectName, entries := range projects {
				for _, e := range entries {
					status := "Dead"
					if pidfile.IsRunning(e.PID) {
						status = "Running"
					}
					t.AppendRow(table.Row{configName, projectName, e.Command, shortenHome(e.Dir), e.PID, status})
				}
			}
		}
		t.Render()
	},
}

func shortenHome(path string) string {
	home, err := os.UserHomeDir()
	if err != nil || home == "" {
		return path
	}
	if path == home {
		return "~"
	}
	if strings.HasPrefix(path, home+"/") {
		return "~" + path[len(home):]
	}
	return path
}

func init() {
	rootCmd.AddCommand(procsCmd)
}
