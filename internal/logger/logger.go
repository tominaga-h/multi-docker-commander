package logger

import (
	"fmt"
	"io"
	"math/rand/v2"
	"os"
	"strings"
	"sync"

	"github.com/jedib0t/go-pretty/v6/text"
	"golang.org/x/term"
)

var (
	mu  sync.Mutex
	out io.Writer = os.Stdout

	projectColors = map[string]text.Colors{}
	colorPalette = []text.Color{
		text.FgHiGreen,
		text.FgHiMagenta,
		text.FgHiYellow,
		text.FgHiBlue,
		text.FgHiRed,
		text.FgGreen,
		text.FgMagenta,
		text.FgYellow,
		text.FgBlue,
		text.FgRed,
	}
	shuffledPalette []text.Color
	shuffleIdx      int
)

func init() {
	shufflePalette()
}

func shufflePalette() {
	shuffledPalette = make([]text.Color, len(colorPalette))
	copy(shuffledPalette, colorPalette)
	rand.Shuffle(len(shuffledPalette), func(i, j int) {
		shuffledPalette[i], shuffledPalette[j] = shuffledPalette[j], shuffledPalette[i]
	})
	shuffleIdx = 0
}

const defaultBorderWidth = 60

func SetOutput(w io.Writer) {
	mu.Lock()
	defer mu.Unlock()
	out = w
}

func ResetColors() {
	mu.Lock()
	defer mu.Unlock()
	projectColors = map[string]text.Colors{}
	shufflePalette()
}

func getProjectColor(projectName string) text.Colors {
	if c, ok := projectColors[projectName]; ok {
		return c
	}
	c := text.Colors{shuffledPalette[shuffleIdx%len(shuffledPalette)]}
	shuffleIdx++
	projectColors[projectName] = c
	return c
}

func prefix(projectName string) string {
	return getProjectColor(projectName).Sprintf("%s", projectName)
}

func colorCmd(cmd string) string {
	return text.Colors{text.FgCyan}.Sprint(cmd)
}

func colorPID(pid int) string {
	return text.Colors{text.FgYellow}.Sprintf("%d", pid)
}

func terminalWidth() int {
	if w, _, err := term.GetSize(int(os.Stdout.Fd())); err == nil && w > 0 {
		return w
	}
	return defaultBorderWidth
}

func outputBorder() string {
	return strings.Repeat("=", terminalWidth())
}

func writef(format string, args ...any) {
	mu.Lock()
	defer mu.Unlock()
	_, _ = fmt.Fprintf(out, format, args...)
}

func Border() {
	writef("%s\n", outputBorder())
}

func Start(projectName, cmd string) {
	writef("🚀 [%s] Executing: %s\n", prefix(projectName), colorCmd(cmd))
}

func Success(projectName, cmd string) {
	writef("✅ [%s] Completed: %s\n", prefix(projectName), colorCmd(cmd))
}

func Error(projectName, cmd string, err error) {
	writef("❌ [%s] Failed: %s — %s\n", prefix(projectName), colorCmd(cmd), err)
}

func Background(projectName, cmd string, pid int) {
	writef("🔄 [%s] Background: %s (PID: %s)\n", prefix(projectName), colorCmd(cmd), colorPID(pid))
}

func Stop(projectName, cmd string, pid int) {
	writef("🛑 [%s] Stopping: %s (PID: %s)\n", prefix(projectName), colorCmd(cmd), colorPID(pid))
}

func Stopped(projectName string) {
	writef("✅ [%s] Stopped successfully\n", prefix(projectName))
}

func ProjectDone(projectName string) {
	writef("✅ [%s] All commands completed\n", prefix(projectName))
}

func ProjectFailed(projectName string, err error) {
	writef("❌ [%s] Aborted — %s\n", prefix(projectName), err)
}

func Output(projectName, output string) {
	mu.Lock()
	defer mu.Unlock()
	trimmed := strings.TrimRight(output, "\n")
	if trimmed == "" {
		return
	}
	border := outputBorder()
	_, _ = fmt.Fprintf(out, "   [%s] %s\n", prefix(projectName), border)
	for _, line := range strings.Split(trimmed, "\n") {
		_, _ = fmt.Fprintf(out, "   [%s] %s\n", prefix(projectName), line)
	}
	_, _ = fmt.Fprintf(out, "   [%s] %s\n", prefix(projectName), border)
}
