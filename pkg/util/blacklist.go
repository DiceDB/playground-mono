package helpers

import (
	"errors"
	"strings"
)

var blacklistedCommands = []string{
	"FLUSHALL", "FLUSHDB", "DUMP", "ABORT", "AUTH", "CONFIG", "SAVE", "BGSAVE",
	"BGREWRITEAOF", "RESTORE", "MULTI", "EXEC", "DISCARD", "QWATCH", "QUNWATCH",
	"LATENCY", "CLIENT", "SLEEP", "PERSIST",
}

// IsBlacklistedCommand checks if a command is blacklisted
func IsBlacklistedCommand(cmd string) error {
	for _, blacklistedCmd := range blacklistedCommands {
		if strings.ToUpper(cmd) == blacklistedCmd {
			return errors.New("command is blacklisted")
		}
	}
	return nil
}
