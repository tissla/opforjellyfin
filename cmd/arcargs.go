package cmd

import (
	"strconv"
	"strings"
)

func positionalArcName(args []string) (string, bool) {
	if len(args) == 0 || allNumericRangeArgs(args) {
		return "", false
	}

	return strings.Join(args, " "), true
}

func allNumericRangeArgs(args []string) bool {
	for _, arg := range args {
		if !isNumericRangeArg(arg) {
			return false
		}
	}
	return true
}

func isNumericRangeArg(arg string) bool {
	if arg == "" {
		return false
	}

	if !strings.Contains(arg, "-") {
		_, err := strconv.Atoi(arg)
		return err == nil
	}

	parts := strings.Split(arg, "-")
	if len(parts) != 2 {
		return false
	}

	_, startErr := strconv.Atoi(parts[0])
	_, endErr := strconv.Atoi(parts[1])
	return startErr == nil && endErr == nil
}
