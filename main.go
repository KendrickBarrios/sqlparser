package main

import (
	"errors"
	"regexp"
	"strings"
)

const (
	indexAfterInsertInto = 12
)

var unquotedTableNameRegex = regexp.MustCompile(`[a-z_]{1}[0-9a-z_]+`)
var	quotedTableNameRegex = regexp.MustCompile(`[0-9a-zA-Z-_! $]+`)
var InvalidSyntaxError = errors.New("invalid syntax")

func ParseSqlIntoJson(script string) (string, error) {
	script = strings.Trim(script, " ")
	sliceSplitByQuotes := splitScriptByQuotes(script)
	err := validateQuotePairs(len(sliceSplitByQuotes))

	if err != nil {
		return "", err
	} 

	err = validateBeginsWithInsertInto(sliceSplitByQuotes[0])

	if err != nil {
		return "", err
	}

	validTableName := validateTableName(sliceSplitByQuotes[0])

	if !validTableName {
		return "", InvalidSyntaxError
	}

	sliceSplitByQuotes[0] = removeInsertInto(sliceSplitByQuotes[0])

	_, err = extractTableName(sliceSplitByQuotes[0])
	
	return "", nil
}

func splitScriptByQuotes(script string) []string {
	return strings.Split(script, "'")
}

func validateQuotePairs(length int) error {
	if length % 2 != 0 {
		return nil
	}

	return InvalidSyntaxError
}

func validateBeginsWithInsertInto(fragment string) error {
	if strings.Index(strings.ToLower(fragment), "insert into ") == 0 {
		return nil
	}

	return InvalidSyntaxError
}

func removeInsertInto(fragment string) string {
	return fragment[indexAfterInsertInto:]
}

func validateTableName(fragment string) bool {
	var slice = make([]string, 0)

	if strings.Contains(fragment, "\"") {
		slice = strings.Split(fragment, "\"")
	}

	if len(slice) % 2 == 0 {
		return false
	}

	if len(slice) > 1 {
		return quotedTableNameRegex.Match([]byte(slice[1]))
	} else {
		return unquotedTableNameRegex.Match([]byte(strings.ToLower(fragment)))
	}
}

func extractTableName(fragment string) (string, error) {
	// TODO: implement function
	return "", nil
}

func main() {
}
