package main

import (
	"errors"
	"regexp"
	"strings"
)

const (
	indexAfterInsertInto = 12
)

type InsertScript struct {
	tableName string
	columns []string
	rows [][]string
}

var unquotedTableNameRegex = regexp.MustCompile(`[a-z_]{1}[0-9a-z_]+`)
var	quotedTableNameRegex = regexp.MustCompile(`[0-9a-zA-Z-_! $]+`)
var InvalidSyntaxError = errors.New("invalid syntax")

func BuildInsertScriptStruct(script string) (InsertScript, error) {
	insertScript := InsertScript{}
	script = strings.Trim(script, " ")
	sliceSplitByQuotes := splitScriptByQuotes(script)
	err := validateQuotePairs(len(sliceSplitByQuotes))

	if err != nil {
		return insertScript, err
	} 

	err = validateBeginsWithInsertInto(sliceSplitByQuotes[0])

	if err != nil {
		return insertScript, err
	}

	sliceSplitByQuotes[0] = removeInsertInto(sliceSplitByQuotes[0])
	isTableNameQuoted := verifyIfTableNameIsQuoted(sliceSplitByQuotes[0])
	validTableName := validateTableName(sliceSplitByQuotes[0], isTableNameQuoted)

	if !validTableName {
		return insertScript, InvalidSyntaxError
	}

	insertScript.tableName = extractTableName(sliceSplitByQuotes[0], isTableNameQuoted)
	
	return insertScript, nil
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

func verifyIfTableNameIsQuoted(fragment string) bool {
	return strings.Contains(fragment, "\"")
}

func validateTableName(fragment string, isTableNameQuoted bool) bool {
	if isTableNameQuoted {
		slicedFragment := strings.Split(fragment, "\"")
		if len(slicedFragment) % 2 == 0 {
			return false
		}
		return quotedTableNameRegex.Match([]byte(slicedFragment[1]))
	}

	return unquotedTableNameRegex.Match([]byte(strings.ToLower(fragment)))
}

func extractTableName(fragment string, isTableNameQuoted bool) string {
	if isTableNameQuoted {
		return strings.Split(fragment, "\"")[1]
	}

	trimmedFragment := strings.Trim(fragment, " ")
	splitFragment := strings.Split(trimmedFragment, " ")
	return splitFragment[0]
}

func main() {
}
