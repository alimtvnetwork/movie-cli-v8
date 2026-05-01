// Package cmd — movie_condition.go
//
// Tokeniser + parser + parameterised SQL builder for the condition expression
// grammar used by `movie rm` and `movie move <selector> <dest>`.
//
// Spec: spec/08-app/10-remove-move-rescan/03-condition-expression-grammar.md
//
// Public API:
//
//	BuildConditionSQL(expr string) (whereClause string, args []any, err error)
//
// The returned `whereClause` is suitable for appending after `WHERE` in a
// SELECT against the Media table; `IsDeleted = 0` is always appended so
// soft-deleted rows are never matched.
package cmd

import (
	"strconv"
	"strings"

	"github.com/alimtvnetwork/movie-cli-v7/apperror"
)

// ---- field map -------------------------------------------------------------

type fieldSpec struct {
	column string // SQL column (or sentinel "GENRE")
	kind   string // "num" | "text" | "size" | "genre"
}

var conditionFields = map[string]fieldSpec{
	"rating":     {"TmdbRating", "num"},
	"r":          {"TmdbRating", "num"},
	"year":       {"Year", "num"},
	"y":          {"Year", "num"},
	"duration":   {"Runtime", "num"},
	"d":          {"Runtime", "num"},
	"size":       {"FileSizeMb", "size"},
	"s":          {"FileSizeMb", "size"},
	"resolution": {"Resolution", "text"},
	"res":        {"Resolution", "text"},
	"genre":      {"GENRE", "genre"},
	"g":          {"GENRE", "genre"},
}

var validOps = map[string]bool{
	"<": true, "<=": true, "=": true, ">=": true, ">": true, "!=": true,
}

// ---- public entry ----------------------------------------------------------

// BuildConditionSQL parses expr and returns a parameterised WHERE clause.
func BuildConditionSQL(expr string) (string, []any, error) {
	expr = strings.TrimSpace(expr)
	if expr == "" {
		return "", nil, apperror.New("empty condition expression")
	}
	tokens, err := tokenizeCondition(expr)
	if err != nil {
		return "", nil, err
	}
	return buildWhereFromTokens(tokens)
}

// ---- tokeniser -------------------------------------------------------------

type condToken struct {
	kind string // "field" | "op" | "value" | "logic"
	text string
}

func tokenizeCondition(expr string) ([]condToken, error) {
	runes := []rune(expr)
	var out []condToken
	i := 0
	for i < len(runes) {
		i = skipSpaces(runes, i)
		if i >= len(runes) {
			break
		}
		tok, next, err := readNextToken(runes, i)
		if err != nil {
			return nil, err
		}
		out = append(out, tok)
		i = next
	}
	return out, nil
}

func skipSpaces(r []rune, i int) int {
	for i < len(r) && (r[i] == ' ' || r[i] == '\t') {
		i++
	}
	return i
}

func readNextToken(r []rune, i int) (condToken, int, error) {
	c := r[i]
	if c == '"' || c == '\'' {
		return readQuoted(r, i)
	}
	if isOpStart(c) {
		return readOperator(r, i)
	}
	return readWord(r, i), advanceWord(r, i), nil
}

func isOpStart(c rune) bool {
	return c == '<' || c == '>' || c == '=' || c == '!'
}

func readQuoted(r []rune, i int) (condToken, int, error) {
	quote := r[i]
	i++
	start := i
	for i < len(r) && r[i] != quote {
		i++
	}
	if i >= len(r) {
		return condToken{}, 0, apperror.New("unterminated quoted value")
	}
	val := string(r[start:i])
	return condToken{kind: "value", text: val}, i + 1, nil
}

func readOperator(r []rune, i int) (condToken, int, error) {
	if i+1 < len(r) && (r[i+1] == '=') {
		op := string(r[i : i+2])
		return condToken{kind: "op", text: op}, i + 2, nil
	}
	op := string(r[i])
	if !validOps[op] {
		return condToken{}, 0, apperror.New("invalid operator: " + op)
	}
	return condToken{kind: "op", text: op}, i + 1, nil
}

func advanceWord(r []rune, i int) int {
	for i < len(r) && !isWordBoundary(r[i]) {
		i++
	}
	return i
}

func isWordBoundary(c rune) bool {
	return c == ' ' || c == '\t' || isOpStart(c)
}

func readWord(r []rune, i int) condToken {
	end := advanceWord(r, i)
	w := string(r[i:end])
	upper := strings.ToUpper(w)
	if upper == "AND" || upper == "OR" {
		return condToken{kind: "logic", text: upper}
	}
	return condToken{kind: "word", text: w}
}

// ---- parser / SQL builder --------------------------------------------------

func buildWhereFromTokens(tokens []condToken) (string, []any, error) {
	if len(tokens) < 3 {
		return "", nil, apperror.New("condition needs field op value")
	}
	var parts []string
	var args []any
	i := 0
	for i < len(tokens) {
		clause, vals, next, err := parseSingleTerm(tokens, i)
		if err != nil {
			return "", nil, err
		}
		parts = append(parts, clause)
		args = append(args, vals...)
		i = next
		if i >= len(tokens) {
			break
		}
		if tokens[i].kind != "logic" {
			return "", nil, apperror.New("expected AND/OR, got: " + tokens[i].text)
		}
		parts = append(parts, tokens[i].text)
		i++
	}
	where := strings.Join(parts, " ") + " AND IsDeleted = 0"
	return where, args, nil
}

func parseSingleTerm(t []condToken, i int) (string, []any, int, error) {
	if i+2 >= len(t) {
		return "", nil, 0, apperror.New("incomplete term near end of expression")
	}
	fieldTok, opTok, valTok := t[i], t[i+1], t[i+2]
	if fieldTok.kind != "word" {
		return "", nil, 0, apperror.New("expected field, got: " + fieldTok.text)
	}
	spec, ok := conditionFields[strings.ToLower(fieldTok.text)]
	if !ok {
		return "", nil, 0, apperror.New("unknown field: " + fieldTok.text)
	}
	if opTok.kind != "op" {
		return "", nil, 0, apperror.New("expected operator, got: " + opTok.text)
	}
	clause, args, err := buildClause(spec, opTok.text, valTok.text)
	if err != nil {
		return "", nil, 0, err
	}
	return clause, args, i + 3, nil
}

func buildClause(spec fieldSpec, op, raw string) (string, []any, error) {
	if spec.kind == "genre" {
		return buildGenreClause(op, raw)
	}
	if spec.kind == "size" {
		return buildSizeClause(spec.column, op, raw)
	}
	if spec.kind == "num" {
		n, err := strconv.ParseFloat(raw, 64)
		if err != nil {
			return "", nil, apperror.Wrap("numeric value: "+raw, err)
		}
		return spec.column + " " + op + " ?", []any{n}, nil
	}
	return spec.column + " " + op + " ?", []any{raw}, nil
}

func buildGenreClause(op, raw string) (string, []any, error) {
	if op != "=" && op != "!=" {
		return "", nil, apperror.New("genre supports only = / !=")
	}
	notKw := ""
	if op == "!=" {
		notKw = "NOT "
	}
	sql := notKw + `EXISTS (
		SELECT 1 FROM MediaGenre mg
		JOIN Genre g ON g.GenreId = mg.GenreId
		WHERE mg.MediaId = Media.MediaId AND g.Name = ?
	)`
	return sql, []any{raw}, nil
}

func buildSizeClause(col, op, raw string) (string, []any, error) {
	mb, err := parseSizeToMB(raw)
	if err != nil {
		return "", nil, err
	}
	return col + " " + op + " ?", []any{mb}, nil
}

func parseSizeToMB(raw string) (float64, error) {
	upper := strings.ToUpper(strings.TrimSpace(raw))
	mult := 1.0
	switch {
	case strings.HasSuffix(upper, "GB"):
		mult = 1024
		upper = strings.TrimSuffix(upper, "GB")
	case strings.HasSuffix(upper, "MB"):
		upper = strings.TrimSuffix(upper, "MB")
	}
	n, err := strconv.ParseFloat(strings.TrimSpace(upper), 64)
	if err != nil {
		return 0, apperror.Wrap("size value: "+raw, err)
	}
	return n * mult, nil
}
