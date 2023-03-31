package lex

// Lexer based on Rob Pike's talk https://talks.golang.org/2011/lex.slide#1

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"
)

type Item struct {
	Typ  itemType // this Item's type
	Val  string   // the raw value of the Item
	Line int      // the line number within the input string
	pos  int      // the starting position, in bytes
}

func (i Item) String() string {
	switch i.Typ {
	case ItemEOF:
		return "EOF"
	case ItemError:
		return i.Val
	}
	if len(i.Val) > 10 {
		return fmt.Sprintf("%.10q...", i.Val)
	}
	return fmt.Sprintf("%q", i.Val)
}

type itemType string

// itemType identifies the type of lex items
const (
	ItemError                itemType = "ItemError"                // error occurred; value is text of error
	ItemEOF                  itemType = "ItemEOF"                  // end of file
	ItemSingleLineComment    itemType = "ItemSingleLineComment"    // A comment like --
	ItemMultiLineComment     itemType = "ItemMultiLineComment"     // A multiline comment like /* ... */
	ItemKeyword              itemType = "ItemKeyword"              // SQL language keyword like SELECT, INSERT, etc.
	ItemIdentifier           itemType = "ItemIdentifier"           // alphanumeric non-keyword identifier
	ItemBacktickedIdentifier itemType = "ItemBacktickedIdentifier" // '`users`'
	ItemOperator             itemType = "ItemOperator"             // operators like '=', '<>', etc.
	ItemLeftParen            itemType = "ItemLeftParen"            // '('
	ItemRightParen           itemType = "ItemRightParen"           // ')'
	ItemComma                itemType = "ItemComma"                // ','
	ItemDot                  itemType = "ItemDot"                  // '.'
	ItemStatementEnd         itemType = "ItemStatementEnd"         // ';'
	ItemNumber               itemType = "ItemNumber"               // simple number
	ItemString               itemType = "ItemString"               // quoted string (includes quotes)
)

const (
	KeywordFrom = "from"

	KeywordAs = "as"

	KeywordJoin  = "join"
	KeywordLeft  = "left"
	KeywordRight = "right"
	KeywordInner = "inner"
	KeywordOuter = "outer"
	KeywordFull  = "full"

	KeywordWhere = "where"
)

// keywords is a list of reserved SQL keywords
var keywords = map[string]struct{}{
	"all":                  {},
	"and":                  {},
	"any":                  {},
	"array":                {},
	"as":                   {},
	"asc":                  {},
	"assert_rows_modified": {},
	"at":                   {},
	"between":              {},
	"by":                   {},
	"case":                 {},
	"cast":                 {},
	"collate":              {},
	"contains":             {},
	"create":               {},
	"cross":                {},
	"cube":                 {},
	"current":              {},
	"default":              {},
	"define":               {},
	"desc":                 {},
	"distinct":             {},
	"else":                 {},
	"end":                  {},
	"enum":                 {},
	"escape":               {},
	"except":               {},
	"exclude":              {},
	"exists":               {},
	"extract":              {},
	"false":                {},
	"fetch":                {},
	"following":            {},
	"for":                  {},
	"from":                 {},
	"full":                 {},
	"group":                {},
	"grouping":             {},
	"groups":               {},
	"hash":                 {},
	"having":               {},
	"if":                   {},
	"ignore":               {},
	"in":                   {},
	"inner":                {},
	"intersect":            {},
	"interval":             {},
	"into":                 {},
	"is":                   {},
	"join":                 {},
	"lateral":              {},
	"left":                 {},
	"like":                 {},
	"limit":                {},
	"lookup":               {},
	"merge":                {},
	"natural":              {},
	"new":                  {},
	"no":                   {},
	"not":                  {},
	"null":                 {},
	"nulls":                {},
	"of":                   {},
	"on":                   {},
	"or":                   {},
	"order":                {},
	"outer":                {},
	"over":                 {},
	"partition":            {},
	"preceding":            {},
	"primary":              {},
	"proto":                {},
	"range":                {},
	"recursive":            {},
	"respect":              {},
	"right":                {},
	"rollup":               {},
	"rows":                 {},
	"select":               {},
	"set":                  {},
	"some":                 {},
	"struct":               {},
	"table":                {},
	"tablesample":          {},
	"then":                 {},
	"to":                   {},
	"treat":                {},
	"true":                 {},
	"unbounded":            {},
	"union":                {},
	"unnest":               {},
	"using":                {},
	"when":                 {},
	"where":                {},
	"window":               {},
	"with":                 {},
	"within":               {},
}

const (
	eof                    = -1
	singleLineCommentStart = "--"
	multiLineCommentStart  = "/*"
	multiLineCommentEnd    = "*/"
)

type Lexer struct {
	input string    // the string being scanned
	start int       // start position of this Item
	pos   int       // current position of the input
	width int       // width of last rune read from input
	line  int       // 1+number of newlines seen
	items chan Item // channel of scanned items
}

// stateFn represents the state of the scanner as a function
// that returns the next state.
type stateFn func(*Lexer) stateFn

// Lex creates a new Lexer
func Lex(input string) *Lexer {
	l := &Lexer{
		input: input,
		items: make(chan Item),
	}
	go l.run()
	return l
}

// NextItem returns the next Item from the input. The Lexer has to be
// drained (all items received until ItemEOF or ItemError) - otherwise
// the Lexer goroutine will leak.
func (l *Lexer) NextItem() Item {
	return <-l.items
}

func (l *Lexer) ReadAll() []Item {
	var items []Item
	for {
		item := l.NextItem()
		items = append(items, item)
		if item.Typ == ItemEOF || item.Typ == ItemError {
			break
		}
	}

	return items
}

// run runs the lexer - should be run in a separate goroutine.
func (l *Lexer) run() {
	for state := lexWhitespace; state != nil; {
		state = state(l)
	}
	close(l.items) // no more tokens will be delivered
}

func (l *Lexer) emit(t itemType) {
	l.items <- Item{
		Typ:  t,
		Val:  l.input[l.start:l.pos],
		pos:  l.pos,
		Line: l.line,
	}
	l.start = l.pos
}

// next advances to the next rune in input and returns it
func (l *Lexer) next() (r rune) {
	if l.pos >= len(l.input) {
		l.width = 0
		return eof
	}
	r, l.width = utf8.DecodeRuneInString(l.input[l.pos:])
	l.pos += l.width
	if r == '\n' {
		l.line++
	}
	return r
}

// ignore skips over the pending input before this point
func (l *Lexer) ignore() {
	l.start = l.pos
}

// backup steps back one rune. Can be called only once per call of next.
func (l *Lexer) backup() {
	if l.pos > 0 {
		r, w := utf8.DecodeLastRuneInString(l.input[:l.pos])
		l.pos -= w
		// Correct newline count.
		if r == '\n' {
			l.line--
		}
	}
}

// peek returns but does not consume the next run in the input.
func (l *Lexer) peek() rune {
	r := l.next()
	l.backup()
	return r
}

// errorf returns an error token and terminates the scan by passing back
// a nil pointer that will be the next state.
func (l *Lexer) errorf(format string, args ...interface{}) stateFn {
	l.items <- Item{
		Typ:  ItemError,
		Val:  fmt.Sprintf(format, args...),
		pos:  l.pos,
		Line: l.line,
	}
	return nil
}

// accept consumes the next rune if it's from the valid set.
func (l *Lexer) accept(valid string) bool {
	if strings.ContainsRune(valid, l.next()) {
		return true
	}
	l.backup()
	return false
}

// acceptRun consumes a run of runes from the valid set.
func (l *Lexer) acceptRun(valid string) {
	for strings.ContainsRune(valid, l.next()) {
	}
	l.backup()
}

// isSpace reports whether r is a whitespace character (space or end of line).
func isWhitespace(r rune) bool {
	return r == ' ' || r == '\t' || isNewline(r)
}

func isNewline(r rune) bool {
	return r == '\r' || r == '\n'
}

// isAlphaNumeric reports whether r is an alphabetic, digit, or underscore.
func isAlphaNumeric(r rune) bool {
	return r == '_' || unicode.IsLetter(r) || unicode.IsDigit(r)
}

// isOperator reports whether r is an operator.
func isOperator(r rune) bool {
	return r == '+' || r == '-' || r == '*' || r == '/' || r == '=' || r == '>' || r == '<' || r == '~' || r == '|' || r == '^' || r == '&' || r == '%'
}

func isDot(r rune) bool {
	return r == '.'
}

func isBacktick(r rune) bool {
	return r == '`'
}

func lexNumber(l *Lexer) stateFn {
	// Optional leading sign.
	l.accept("+-")
	// is it hex?
	digits := "0123456789"
	if l.accept("0") && l.accept("xX") {
		digits = "0123456789abcdefABCDEF"
	}
	l.acceptRun(digits)
	if l.accept(".") {
		l.acceptRun(digits)
	}
	if l.accept("eE") {
		l.accept("+-")
		l.acceptRun("0123456789")
	}
	if isAlphaNumeric(l.peek()) {
		l.next()
		return l.errorf("bad number syntax: %q", l.input[l.start:l.pos])
	}
	l.emit(ItemNumber)
	return lexWhitespace
}

func lexWhitespace(l *Lexer) stateFn {
	for {
		if strings.HasPrefix(l.input[l.pos:], singleLineCommentStart) {
			return lexSingleLineComment
		}

		if strings.HasPrefix(l.input[l.pos:], multiLineCommentStart) {
			return lexMultiLineComment
		}

		switch r := l.next(); {
		case isWhitespace(r):
			l.ignore()

		case r == eof:
			l.emit(ItemEOF)
			return nil

		case r == '(':
			l.next()
			l.emit(ItemLeftParen)
			return lexWhitespace

		case r == ')':
			l.next()
			l.emit(ItemRightParen)
			return lexWhitespace

		case r == ',':
			l.next()
			l.emit(ItemComma)
			return lexWhitespace

		case r == ';':
			l.next()
			l.emit(ItemStatementEnd)
			return lexWhitespace

		case isOperator(r):
			return lexOperator

		case isDot(r):
			l.emit(ItemDot)

		case isBacktick(r):
			return lexIdentifierWithBacktick

		case r == '"' || r == '\'':
			return lexString

		case unicode.IsDigit(r):
			return lexNumber

		case isAlphaNumeric(r) || r == '`':
			return lexIdentifierOrKeyword

		default:
			return l.errorf("unrecognized character in action: %#U", r)
		}
	}

}

func lexSingleLineComment(l *Lexer) stateFn {
	for {
		switch r := l.next(); {
		case r == eof:
			return l.errorf("eof found in middle of multi-line comment")
		case r == '\n' || r == '\r':
			l.backup() // do not emit the newline
			l.emit(ItemSingleLineComment)
			return lexWhitespace
		default:
			// absorb
		}
	}
}

func lexMultiLineComment(l *Lexer) stateFn {
	for {
		if strings.HasPrefix(l.input[l.pos:], multiLineCommentEnd) {
			l.acceptRun(multiLineCommentEnd)
			l.emit(ItemMultiLineComment)
			return lexWhitespace
		}
		switch r := l.next(); {
		case r == eof:
			return l.errorf("eof found in middle of multi-line comment")
		default:
			// absorb
		}
	}
}

func lexOperator(l *Lexer) stateFn {
	operators := "+-*/=><~|^&%"
	l.acceptRun(operators)
	l.emit(ItemOperator)
	return lexWhitespace
}

func lexString(l *Lexer) stateFn {
	// backup and retrieve which type of quote was used, " or '
	l.backup()
	quote := l.next()

	for {
		switch n := l.next(); {
		case n == eof || isNewline(n):
			l.backup()
			return l.errorf("unterminated quoted string")
		case n == '\\':
			if l.peek() == eof {
				return l.errorf("unterminated quoted string")
			}
		case n == quote:
			// TODO: detect triple quoted strings
			l.emit(ItemString)
			return lexWhitespace
		}
	}
}

func lexIdentifierOrKeyword(l *Lexer) stateFn {
	for {
		switch r := l.next(); {
		case isAlphaNumeric(r):
		// absorb.
		case isDot(r):
			// emit the identifier before the dot
			l.backup()
			l.emit(ItemIdentifier)

			// emit the dot
			l.next()
			l.emit(ItemDot)
		default:
			if r != eof {
				l.backup()
			}
			word := l.input[l.start:l.pos]
			if _, ok := keywords[strings.ToLower(word)]; ok {
				l.emit(ItemKeyword)
			} else {
				l.emit(ItemIdentifier)
			}
			return lexWhitespace
		}
	}
}

func lexIdentifierWithBacktick(l *Lexer) stateFn {
	for {
		switch r := l.next(); {
		case isAlphaNumeric(r):
		// absorb.
		case isBacktick(r):
			l.emit(ItemBacktickedIdentifier)
			return lexWhitespace
		default:
			l.backup()
			return l.errorf("unterminated backtick")
		}
	}
}
