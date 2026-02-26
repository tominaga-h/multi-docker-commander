package logger

import (
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
)

var (
	mu  sync.Mutex
	out io.Writer = os.Stdout
)

func SetOutput(w io.Writer) {
	mu.Lock()
	defer mu.Unlock()
	out = w
}

func prefix(projectName string) string {
	return fmt.Sprintf("[%s]", projectName)
}

func Start(projectName, cmd string) {
	mu.Lock()
	defer mu.Unlock()
	fmt.Fprintf(out, "🚀 %s Executing: %s\n", prefix(projectName), cmd)
}

func Success(projectName, cmd string) {
	mu.Lock()
	defer mu.Unlock()
	fmt.Fprintf(out, "✅ %s Completed: %s\n", prefix(projectName), cmd)
}

func Error(projectName, cmd string, err error) {
	mu.Lock()
	defer mu.Unlock()
	fmt.Fprintf(out, "❌ %s Failed: %s — %s\n", prefix(projectName), cmd, err)
}

func ProjectDone(projectName string) {
	mu.Lock()
	defer mu.Unlock()
	fmt.Fprintf(out, "✅ %s All commands completed\n", prefix(projectName))
}

func ProjectFailed(projectName string, err error) {
	mu.Lock()
	defer mu.Unlock()
	fmt.Fprintf(out, "❌ %s Aborted — %s\n", prefix(projectName), err)
}

func Output(projectName, output string) {
	mu.Lock()
	defer mu.Unlock()
	trimmed := strings.TrimRight(output, "\n")
	if trimmed == "" {
		return
	}
	for _, line := range strings.Split(trimmed, "\n") {
		fmt.Fprintf(out, "   %s %s\n", prefix(projectName), line)
	}
}
