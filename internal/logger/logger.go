package logger

import (
	"fmt"
	"strings"
	"sync"
)

var mu sync.Mutex

func prefix(projectName string) string {
	return fmt.Sprintf("[%s]", projectName)
}

func Start(projectName, cmd string) {
	mu.Lock()
	defer mu.Unlock()
	fmt.Printf("🚀 %s Executing: %s\n", prefix(projectName), cmd)
}

func Success(projectName, cmd string) {
	mu.Lock()
	defer mu.Unlock()
	fmt.Printf("✅ %s Completed: %s\n", prefix(projectName), cmd)
}

func Error(projectName, cmd string, err error) {
	mu.Lock()
	defer mu.Unlock()
	fmt.Printf("❌ %s Failed: %s — %s\n", prefix(projectName), cmd, err)
}

func ProjectDone(projectName string) {
	mu.Lock()
	defer mu.Unlock()
	fmt.Printf("✅ %s All commands completed\n", prefix(projectName))
}

func ProjectFailed(projectName string, err error) {
	mu.Lock()
	defer mu.Unlock()
	fmt.Printf("❌ %s Aborted — %s\n", prefix(projectName), err)
}

func Output(projectName, output string) {
	mu.Lock()
	defer mu.Unlock()
	trimmed := strings.TrimRight(output, "\n")
	if trimmed == "" {
		return
	}
	for _, line := range strings.Split(trimmed, "\n") {
		fmt.Printf("   %s %s\n", prefix(projectName), line)
	}
}
