package cmds

import (
	"fmt"
	"strconv"
	"strings"
)

var nilTuple = "(nil)"
var emptyList = "(empty list or set)"
var invalidString = "(error) invalid type"

// getRender retrieves the appropriate callback for the command
func GetRender(commandName string) DiceDBCallback {
	commandUpper := strings.ToUpper(strings.TrimSpace(commandName))
	callback, exists := command2callback[commandUpper]
	if !exists {
		return renderListOrString
	}

	// If the type assertion fails, return a default fallback function
	return callback
}

// Render functions

func renderBulkString(value interface{}) interface{} {
	if value == nil {
		return nilTuple
	}
	result, ok := value.(int64)
	if ok {
		return renderInt(result)
	}
	return (fmt.Sprintf("%v", value))
}

func renderInt(value interface{}) interface{} {
	if value == nil {
		return nilTuple
	}
	if intValue, ok := value.(int64); ok {
		return fmt.Sprintf("(integer) %d", intValue)
	}
	return invalidString
}

func renderList(value interface{}) interface{} {
	items, ok := value.([]interface{})
	if !ok {
		return invalidString
	}

	var builder strings.Builder
	for i, item := range items {
		strItem := fmt.Sprintf("%v", item)
		quoted := doubleQuotes(strItem)
		builder.WriteString(fmt.Sprintf("%d) %s\n", i+1, quoted))
	}
	return builder.String()
}

func renderListOrString(value interface{}) interface{} {
	if items, ok := value.([]interface{}); ok {
		return renderList(items)
	}
	return renderBulkString(value)
}

func renderStringOrInt(value interface{}) interface{} {
	if intValue, ok := value.(int); ok {
		return renderInt(intValue)
	}
	return renderBulkString(value)
}

func renderSimpleString(value interface{}) interface{} {
	if value == nil {
		return nilTuple
	}
	text := fmt.Sprintf("%v", value)
	return text
}
func renderHashPairs(value interface{}) interface{} {
	items, ok := value.([]interface{})
	if len(items) == 0 {
		return emptyList
	}
	if !ok || len(items)%2 != 0 {
		return "(error) invalid hash pair format"
	}

	var builder strings.Builder
	indexWidth := len(strconv.Itoa(len(items) / 2))
	for i := 0; i < len(items); i += 2 {
		key := fmt.Sprintf("%v", items[i])
		value := fmt.Sprintf("%v", items[i+1])

		// Format the index and key
		indexStr := fmt.Sprintf("%*d) ", indexWidth, i/2+1)
		builder.WriteString(indexStr)
		builder.WriteString(key + "\n")

		// Format the value, ensuring correct indentation
		// and preserving quotes if necessary
		if strings.Contains(value, "\"") {
			value = fmt.Sprintf("%q", value)
		}
		valueStr := strings.Repeat(" ", len(indexStr)) + value
		builder.WriteString(valueStr + "\n")
	}
	return builder.String()
}

func commandHscan(value interface{}) interface{} {
	scanResult, ok := value.([]interface{})
	if !ok || len(scanResult) < 2 {
		return "(error) invalid type or format"
	}

	cursor := fmt.Sprintf("%v", scanResult[0])
	items, ok := scanResult[1].([]interface{})
	if !ok {
		return "(error) invalid scan items format"
	}

	var builder strings.Builder
	builder.WriteString(fmt.Sprintf("(cursor) %s\n", cursor))
	renderedItems := renderHashPairs(items)
	builder.WriteString(fmt.Sprintf("%s", renderedItems))

	return builder.String()
}

// RenderMembers renders a list of set or sorted set members
func renderMembers(value interface{}) interface{} {
	items, ok := value.([]interface{})
	if !ok {
		return invalidString
	}

	var builder strings.Builder
	indexWidth := len(strconv.Itoa(len(items)))
	for i, item := range items {
		member := fmt.Sprintf("%v", item)
		indexStr := fmt.Sprintf("%*d) ", indexWidth, i+1)
		builder.WriteString(indexStr)
		builder.WriteString(member + "\n")
	}

	return builder.String()
}

func doubleQuotes(value string) string {
	escaped := strings.ReplaceAll(value, `"`, `\"`)
	return fmt.Sprintf(`"%q"`, escaped)
}
