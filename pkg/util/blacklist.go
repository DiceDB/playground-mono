package helpers

import (
	"errors"
	"strings"
)

var blocklistedCommands = map[string]bool{
	"FLUSHALL": true, "FLUSHDB": true, "DUMP": true, "ABORT": true, "AUTH": true,
	"CONFIG": true, "SAVE": true, "BGSAVE": true, "BGREWRITEAOF": true, "RESTORE": true,
	"MULTI": true, "EXEC": true, "DISCARD": true, "QWATCH": true, "QUNWATCH": true,
	"LATENCY": true, "CLIENT": true, "SLEEP": true, "PERSIST": true,
}

// BlockListedCommand checks if a command is blocklisted
func BlockListedCommand(cmd string) error {
	if blocklistedCommands[strings.ToUpper(cmd)] {
		return errors.New("command is blocklisted")
	}
	return nil
}
