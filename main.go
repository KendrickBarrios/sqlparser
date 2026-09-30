package main

import (
	"bufio"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"unicode/utf8"
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
	scriptScanner.Split(SplitSqlToken)
	return scriptScanner
}

// function that splits by unquoted delimiters or quotes (single or double)
func SplitSqlToken(data []byte, atEOF bool) (advance int, token []byte, err error) {

	// ommit all spaces or equivalent characters (such as tabulation or new lines)
	start := 0
	for runeWidth := 0; start < len(data); start += runeWidth {
		var currentRune rune
		currentRune, runeWidth = utf8.DecodeRune(data[start:])
		if !isSpace(currentRune) {
			break
		}
	}

	firstRune, firstRuneWidth := utf8.DecodeRune(data[start:])

	// if the first rune is an unquoted delimiter, return it as the token
	if isUnquotedDelimiter(firstRune) {
		return start + firstRuneWidth, data[start:start + firstRuneWidth], nil
	}

	isQuoted := isQuote(firstRune)
	var quoteDelimiter rune
	offset := 0

	if isQuoted {
		offset++
		quoteDelimiter = firstRune
	}

	// if the word is quoted, end at the next quote, if unquoted, end at the unquoted delimiter
	for runeWidth, i := 0, start + offset; i < len(data); i += runeWidth {
		var currentRune rune
		currentRune, runeWidth = utf8.DecodeRune(data[i:])
		if !isQuoted {
			nextRune, _ := utf8.DecodeRune(data[i + runeWidth:])
			if isUnquotedDelimiter(nextRune) {
				return i + runeWidth, data[start:i + runeWidth], nil
			}
		} else if isRuneEqualToQuoteDelimiter(currentRune, quoteDelimiter) {
			// takes one more character to include the delimiter
			return i + runeWidth, data[start:i + runeWidth], nil
		}
	}

	if atEOF && len(data) > start {
		return len(data), data[start:], nil
	}

	return start, nil, nil
}

func isSpace(r rune) bool {
	switch r {
	case ' ', '\t', '\n', '\v', '\f', '\r':
		return true
	default:
		return false
	}
}

func isQuote(r rune) bool {
	switch r {
	case '\'', '"':
		return true
	default:
		return false
	}
}

func isUnquotedDelimiter(r rune) bool {
	switch r {
	case ';', ',', '(', ')', '[', ']', ' ', '\t', '\n', '\v', '\f', '\r':
		return true
	default:
		return false
	}
}

func isRuneEqualToQuoteDelimiter(r, delimiter rune) bool {
	return r == delimiter
}

func validateScript(_ string, _ *bufio.Scanner) error {
	// TODO: fix logic, quoted table name may have inner spaces

	// at End Of String
	// atEOS := false

	return nil
}

func useSplitSqlTokenFunc(scriptScanner *bufio.Scanner) {
	hasTokensLeft := true
	for true {
		hasTokensLeft = scriptScanner.Scan()
		if !hasTokensLeft {
			break
		}
		fmt.Println(scriptScanner.Text())
	}
}

func main() {
	script :=
	`INSERT INTO facility (id, name, capacity, type, state, is_active) VALUES
	(1, 'O-201', 30, 'CLASSROOM', 'AVAILABLE', true),
	(2, 'O-102', 30, 'CLASSROOM', 'AVAILABLE', true),
	(3, 'Quirófano 1', 6, 'OPERATINGROOM', 'AVAILABLE', true),
	(4, 'Quirófano 2', 6, 'OPERATINGROOM', 'AVAILABLE', true),
	(5, 'Clínica 1', 8, 'CLINIC', 'AVAILABLE', true),
	(6, 'Clínica 2', 10, 'CLINIC', 'AVAILABLE', true),
	(7, 'Laboratorio 1', 20, 'LABORATORY', 'AVAILABLE', true),
	(8, 'Laboratorio 2', 20, 'LABORATORY', 'AVAILABLE', true);`
	scanner := createScriptScanner(script)
	useSplitSqlTokenFunc(scanner)
}
