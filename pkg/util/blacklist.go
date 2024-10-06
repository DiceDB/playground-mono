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

// BlockListedCommand checks if a command is blocklisted
func BlockListedCommand(cmd string) error {
	for _, blacklistedCmd := range blacklistedCommands {
		if strings.ToUpper(cmd) == blacklistedCmd {
			return errors.New("command is blocklisted")
		}
	}
	return nil
}
