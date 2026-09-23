package main

import (
	"bufio"
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

var unquotedNameRegex = regexp.MustCompile(`[a-z_]{1}[0-9a-z_]+`)
var	quotedNameRegex = regexp.MustCompile(`"[0-9a-zA-Z-_! $]+"`)
var InvalidSyntaxError = errors.New("invalid syntax")

func createScriptScanner(script string) (scriptScanner *bufio.Scanner) {
	scriptScanner = bufio.NewScanner(strings.NewReader(script))
	scriptScanner.Split(splitFunc)
	return scriptScanner
}

func splitFunc(data []byte, atEOF bool) (advance int, token []byte, err error) {
	advance, token, err = bufio.ScanWords(data, atEOF)
	return
}

func BuildInsertScriptStruct(script string, scriptScanner *bufio.Scanner) (InsertScript, error) {
	insertScript := InsertScript{}
	// at End Of String
	atEOS := false

	if strings.ToLower(scriptScanner.Text()) != "insert" {
		return insertScript, InvalidSyntaxError
	}
	
	atEOS = scriptScanner.Scan()
	if atEOS || strings.ToLower(scriptScanner.Text()) != "into" {
		return insertScript, InvalidSyntaxError
	}

	atEOS = scriptScanner.Scan()
	if atEOS {
		return insertScript, InvalidSyntaxError
	}

	// TODO: fix logic, quoted table name may have inner spaces

	return insertScript, nil
}

func extractQuotedName(scriptScanner *bufio.Scanner, delimiter rune) (name string) {
	// if current word also ends in ', the name doesn't contain spaces and is returned as is
	// TODO: implement logic for case ends with comma, rather than single quote
	if scriptScanner.Text()[len(scriptScanner.Text()) - 1] == '\'' {
		name = scriptScanner.Text()
		scriptScanner.Scan()
		return
	}

	// if the name contain spaces, join the words into a single quoted string
	// TODO: complete function
	name = ""
	return
} 

func validateName(fragment string) bool {
	quotedMatch := quotedNameRegex.Match([]byte(fragment))
	unquotedMatch := unquotedNameRegex.Match([]byte(fragment))
	return quotedMatch || unquotedMatch
}

func main() {
}
