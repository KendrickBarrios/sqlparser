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
		} else if currentRune == quoteDelimiter {
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

func isUnquotedDelimiter(r rune) bool {
	switch r {
	case ';', ',', '(', ')', '[', ']', ' ', '\t', '\n', '\v', '\f', '\r':
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

func buildInsertScriptStruct(scanner *bufio.Scanner) (InsertScript, error) {
	insertScript := InsertScript{}
	var err error

	token, hasTokensLeft := getNextToken(scanner)
	err = tokenMatchesAny(hasTokensLeft, strings.ToUpper(token), "INSERT")
	if err != nil {
		return insertScript, err
	}

	token, hasTokensLeft = getNextToken(scanner)
	err = tokenMatchesAny(hasTokensLeft, strings.ToUpper(token), "INTO")
	if err != nil {
		return insertScript, err
	}

	token, hasTokensLeft = getNextToken(scanner)
	err = validateNameMatchesRegex(hasTokensLeft, token)
	if err != nil {
		return insertScript, err
	}
	insertScript.tableName = token

	token, hasTokensLeft = getNextToken(scanner)
	err = tokenMatchesAny(hasTokensLeft, token, "(")
	if err != nil {
		return insertScript, err
	}

	for {
		token, hasTokensLeft = getNextToken(scanner)
		err = validateNameMatchesRegex(hasTokensLeft, token)
		if err != nil {
			return insertScript, err
		}
		insertScript.columns = append(insertScript.columns, token)
		token, hasTokensLeft = getNextToken(scanner)
		err = tokenMatchesAny(hasTokensLeft, token, ",", ")")
		if err != nil {
			return insertScript, err
		}
		if token == ")" {
			break
		}
	}

	token, hasTokensLeft = getNextToken(scanner)
	err = tokenMatchesAny(hasTokensLeft, strings.ToUpper(token), "VALUES")
	if err != nil {
		return insertScript, err
	}

	var isTokenSemicolon bool
	for !isTokenSemicolon {
		var values []string
		token, hasTokensLeft = getNextToken(scanner)
		err = tokenMatchesAny(hasTokensLeft, token, "(")
		if err != nil {
			return insertScript, err
		}
		for {
			token, hasTokensLeft = getNextToken(scanner)
			if !hasTokensLeft {
				return insertScript, InvalidSyntaxError
			}
			values = append(values, token)
			token, hasTokensLeft = getNextToken(scanner)
			err = tokenMatchesAny(hasTokensLeft, token, ",", ")")
			if err != nil {
				return insertScript, err
			}
			if token == ")" {
				break
			}
		}
		if len(insertScript.columns) != len(values) {
			return insertScript, InvalidSyntaxError
		}
		insertScript.rows = append(insertScript.rows, values)
		token, hasTokensLeft = getNextToken(scanner)
		if token == ";" {
			err = validateSemicolonRules(scanner)
			return insertScript, err
		}
		err = tokenMatchesAny(hasTokensLeft, token, ",")
		if err != nil {
			return insertScript, err
		}
	}

	return insertScript, nil
}

// get next token and return true when reaching the last token
func getNextToken(scanner *bufio.Scanner) (token string, hasTokensLeft bool)  {
	hasTokensLeft = scanner.Scan()
	token = scanner.Text()
	return
}

// return error if actual token doesn't match any of the expected, or if it's the last
func tokenMatchesAny(hasTokensLeft bool, actualToken string, expectedTokens ...string) error {
	for _, expected := range expectedTokens {
		if actualToken == expected && hasTokensLeft {
			return nil
		}
	}
	return InvalidSyntaxError
}

func validateNameMatchesRegex(hasTokensLeft bool, token string) error {
	matchIndexes := make([]int, 2)
	if token[0] == '"' {
		matchIndexes = quotedNameRegex.FindStringIndex(token)
	} else {
		matchIndexes = unquotedNameRegex.FindStringIndex(token)
	}
	if matchIndexes == nil || matchIndexes[0] != 0 || matchIndexes[1] != len(token) || !hasTokensLeft {
		return InvalidSyntaxError
	}
	return nil
}

func validateSemicolonRules(scanner *bufio.Scanner) error {
	hasTokensLeft := scanner.Scan()
	if hasTokensLeft {
		return InvalidSyntaxError
	}
	return nil
}

func (insertScript InsertScript) buildJsonArray() string {
	var builder strings.Builder
	builder.WriteString("[\n")
	for i, row := range insertScript.rows {
		builder.WriteString("\t{\n")
		for j, value := range row {
			builder.WriteString("\t\t")
			fmt.Fprintf(&builder, "\"%s\": %s", insertScript.columns[j], value)
			if j == len(row) - 1 {
				builder.WriteString("\n")
			} else {
				builder.WriteString(",\n")
			}
		}
		builder.WriteString("\t}")
		if i == len(insertScript.rows) - 1 {
			builder.WriteString("\n")
		} else {
			builder.WriteString(",\n")
		}
	}
	builder.WriteString("]\n")
	return builder.String()
}

func main() {
	script :=
	`INSERT INTO "facility" (name, capacity, type, state, is_active) VALUES
	('O-201', 30, 'CLASSROOM', 'AVAILABLE', true),
	('O-102', 30, 'CLASSROOM', 'AVAILABLE', true),
	('Quirófano 1', 6, 'OPERATINGROOM', 'AVAILABLE', true),
	('Quirófano 2', 6, 'OPERATINGROOM', 'AVAILABLE', true),
	('Clínica 1', 8, 'CLINIC', 'AVAILABLE', true),
	('Clínica 2', 10, 'CLINIC', 'AVAILABLE', true),
	('Laboratorio 1', 20, 'LABORATORY', 'AVAILABLE', true),
	('Laboratorio 2', 20, 'LABORATORY', 'AVAILABLE', true);`
	scanner := createScriptScanner(script)
	insertScript, err := buildInsertScriptStruct(scanner)

	if err != nil {
		fmt.Println(err)
	}
	json := insertScript.buildJsonArray()
	fmt.Println(json)
}
