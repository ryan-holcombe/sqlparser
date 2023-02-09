package sqlparser

// Lexer based on Rob Pike's talk https://talks.golang.org/2011/lex.slide#1

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"
)

type Item struct {
	typ  itemType // this Item's type
	val  string   // the raw value of the Item
	pos  int      // the starting position, in bytes
	line int      // the line number within the input string
}

func (i Item) String() string {
	switch i.typ {
	case itemEOF:
		return "EOF"
	case itemError:
		return i.val
	}
	if len(i.val) > 10 {
		return fmt.Sprintf("%.10q...", i.val)
	}
	return fmt.Sprintf("%q", i.val)
}

type itemType int

// itemType identifies the type of lex items
const (
	itemError itemType = iota // error occurred; value is text of error
	itemEOF

	itemSingleLineComment    // A comment like --
	itemMultiLineComment     // A multiline comment like /* ... */
	itemKeyword              // SQL language keyword like SELECT, INSERT, etc.
	itemIdentifier           // alphanumeric non-keyword identifier
	itemBacktickedIdentifier // '`users`'
	itemOperator             // operators like '=', '<>', etc.
	itemLeftParen            // '('
	itemRightParen           // ')'
	itemComma                // ','
	itemDot                  // '.'
	itemStatementEnd         // ';'
	itemNumber               // simple number
	itemString               // quoted string (includes quotes)
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
// drained (all items received until itemEOF or itemError) - otherwise
// the Lexer goroutine will leak.
func (l *Lexer) NextItem() Item {
	return <-l.items
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
		typ:  t,
		val:  l.input[l.start:l.pos],
		pos:  l.pos,
		line: l.line,
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
		typ:  itemError,
		val:  fmt.Sprintf(format, args...),
		pos:  l.pos,
		line: l.line,
	}
	return nil
}

// accept consumes the next rune if it's from the valid set.
func (l *Lexer) accept(valid string) bool {
	if strings.IndexRune(valid, l.next()) >= 0 {
		return true
	}
	l.backup()
	return false
}

// acceptRun consumes a run of runes from the valid set.
func (l *Lexer) acceptRun(valid string) {
	for strings.IndexRune(valid, l.next()) >= 0 {
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
	l.emit(itemNumber)
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
			l.emit(itemEOF)
			return nil

		case r == '(':
			l.next()
			l.emit(itemLeftParen)
			return lexWhitespace

		case r == ')':
			l.next()
			l.emit(itemRightParen)
			return lexWhitespace

		case r == ',':
			l.next()
			l.emit(itemComma)
			return lexWhitespace

		case r == ';':
			l.next()
			l.emit(itemStatementEnd)
			return lexWhitespace

		case isOperator(r):
			return lexOperator

		case isDot(r):
			l.emit(itemDot)

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
		if strings.HasPrefix(l.input[l.pos:], multiLineCommentEnd) {
			l.emit(itemMultiLineComment)
			return lexWhitespace
		}
		switch r := l.next(); {
		case r == eof:
			return l.errorf("eof found in middle of multi-line comment")
		case r == '\n' || r == '\r':
			l.emit(itemSingleLineComment)
			return lexWhitespace
		default:
			// absorb
		}
	}
}

func lexMultiLineComment(l *Lexer) stateFn {
	for {
		if strings.HasPrefix(l.input[l.pos:], multiLineCommentEnd) {
			l.emit(itemMultiLineComment)
			return lexWhitespace
		}
		switch r := l.next(); {
		case r == eof:
			return l.errorf("eof found in middle of multi-line comment")
		default:
			l.ignore()
		}
	}
}

func lexOperator(l *Lexer) stateFn {
	operators := "+-*/=><~|^&%"
	l.acceptRun(operators)
	l.emit(itemOperator)
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
			l.emit(itemString)
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
			l.emit(itemIdentifier)

			// emit the dot
			l.next()
			l.emit(itemDot)
		default:
			if r != eof {
				l.backup()
			}
			word := l.input[l.start:l.pos]
			if _, ok := keywords[strings.ToLower(word)]; ok {
				l.emit(itemKeyword)
			} else {
				l.emit(itemIdentifier)
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
			l.emit(itemBacktickedIdentifier)
			return lexWhitespace
		default:
			l.backup()
			return l.errorf("unterminated backtick")
		}
	}
}
