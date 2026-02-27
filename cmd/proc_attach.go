package cmd

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"mdc/internal/logger"
	"mdc/internal/pidfile"

	"github.com/spf13/cobra"
)

var (
	tailLines int
	noFollow  bool
)

var procAttachCmd = &cobra.Command{
	Use:   "attach <PID>",
	Short: "Stream stdout/stderr of a background process",
	Long: `Attach to a background process and stream its log output.
Press Ctrl-C to detach (the process keeps running).`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		pid, err := strconv.Atoi(args[0])
		if err != nil {
			fmt.Fprintf(os.Stderr, "invalid PID: %s\n", args[0])
			os.Exit(1)
		}

		configName, projectName, entry, err := pidfile.FindByPID(pid)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}

		logPath, err := pidfile.ProcLogFilePath(configName, projectName, pid)
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to resolve log path: %v\n", err)
			os.Exit(1)
		}

		if _, err := os.Stat(logPath); os.IsNotExist(err) {
			fallback, fbErr := pidfile.ProcLogTmpPath(configName, projectName)
			if fbErr == nil {
				if _, stErr := os.Stat(fallback); stErr == nil {
					logPath = fallback
				}
			}
			if _, err := os.Stat(logPath); os.IsNotExist(err) {
				fmt.Fprintf(os.Stderr, "log file not found: %s\n", logPath)
				fmt.Fprintln(os.Stderr, "This process may have been started before log capture was enabled.")
				os.Exit(1)
			}
		}

		logger.Attach(projectName, entry.Command, pid)

		f, err := os.Open(logPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to open log file: %v\n", err)
			os.Exit(1)
		}
		defer func() { _ = f.Close() }()

		if tailLines > 0 {
			seekToLastNLines(f, tailLines)
		}

		if noFollow {
			if _, err := io.Copy(os.Stdout, f); err != nil {
				fmt.Fprintf(os.Stderr, "read error: %v\n", err)
				os.Exit(1)
			}
			return
		}

		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		defer signal.Stop(sigCh)

		reader := bufio.NewReader(f)
		for {
			line, err := reader.ReadBytes('\n')
			if err == nil {
				_, _ = os.Stdout.Write(line)
				continue
			}

			if err != io.EOF {
				fmt.Fprintf(os.Stderr, "\nread error: %v\n", err)
				break
			}

			if len(line) > 0 {
				_, _ = os.Stdout.Write(line)
			}

			if !pidfile.IsRunning(pid) {
				remaining, _ := io.ReadAll(reader)
				if len(remaining) > 0 {
					_, _ = os.Stdout.Write(remaining)
				}
				logger.ProcessExited(projectName, pid)
				break
			}

			select {
			case <-sigCh:
				logger.Detach(projectName)
				return
			case <-time.After(100 * time.Millisecond):
			}
		}
	},
}

// seekToLastNLines positions the file so that the last n lines remain
// to be read. Scans backward from EOF in chunks for memory efficiency.
func seekToLastNLines(f *os.File, n int) {
	const chunkSize = 8192

	fi, err := f.Stat()
	if err != nil || fi.Size() == 0 {
		return
	}

	size := fi.Size()
	buf := make([]byte, chunkSize)
	newlines := 0
	offset := size
	firstChunk := true

	for offset > 0 {
		readSize := int64(chunkSize)
		if readSize > offset {
			readSize = offset
		}
		offset -= readSize

		nRead, err := f.ReadAt(buf[:readSize], offset)
		if err != nil && err != io.EOF {
			_, _ = f.Seek(0, io.SeekStart)
			return
		}

		startIdx := nRead - 1
		if firstChunk && nRead > 0 && buf[nRead-1] == '\n' {
			startIdx = nRead - 2
		}
		firstChunk = false

		for i := startIdx; i >= 0; i-- {
			if buf[i] == '\n' {
				newlines++
				if newlines >= n {
					_, _ = f.Seek(offset+int64(i)+1, io.SeekStart)
					return
				}
			}
		}
	}

	_, _ = f.Seek(0, io.SeekStart)
}

func init() {
	procAttachCmd.Flags().IntVar(&tailLines, "tail", 0, "Number of lines to show from the end of the log")
	procAttachCmd.Flags().BoolVar(&noFollow, "no-follow", false, "Print existing log and exit without streaming")
	procCmd.AddCommand(procAttachCmd)
}
