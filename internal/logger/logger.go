package logger

import (
	"fmt"
	"os"
	"strings"
)

func Info(component, msg string, kv ...any) {
	fmt.Fprintf(os.Stdout, "[INFO] [%s] %s%s\n", component, msg, formatKV(kv))
}

func Warn(component, msg string, kv ...any) {
	fmt.Fprintf(os.Stdout, "[WARN] [%s] %s%s\n", component, msg, formatKV(kv))
}

func Error(component, msg string, kv ...any) {
	fmt.Fprintf(os.Stderr, "[ERROR] [%s] %s%s\n", component, msg, formatKV(kv))
}

func Fatal(component, msg string, kv ...any) {
	fmt.Fprintf(os.Stderr, "[FATAL] [%s] %s%s\n", component, msg, formatKV(kv))
	os.Exit(1)
}

func formatKV(kv []any) string {
	if len(kv) == 0 {
		return ""
	}
	var out strings.Builder
	out.WriteString(" |")
	for i := 0; i < len(kv); i += 2 {
		if i+1 < len(kv) {
			fmt.Fprintf(&out, " %v=%v", kv[i], kv[i+1])
		}
	}
	return out.String()
}
