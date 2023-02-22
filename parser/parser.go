package parser

import (
	"errors"
	"fmt"
	"strings"

	"github.com/ryan-holcombe/sqlparser/collection"
	"github.com/ryan-holcombe/sqlparser/lex"
)

var errNoToken = errors.New("no lex token found, when one was expected")

type parser struct {
	iter  *collection.Iterator[lex.Item]
	query *Query
	err   error
}

func (p *parser) run() (*Query, error) {
	for state := sqlStatement; state != nil && p.err == nil; {
		state = state(p)
	}

	return p.query, p.err
}

func (p *parser) error(err error) stateFn {
	p.err = errors.Join(p.err, err)
	return nil
}

func (p *parser) errorf(format string, args ...interface{}) stateFn {
	return p.error(fmt.Errorf(format, args...))
}

func (p *parser) mustNext() lex.Item {
	item, ok := p.iter.Next()
	if !ok {
		p.error(errNoToken)
		return lex.Item{}
	}
	return item
}

func (p *parser) mustPeek() lex.Item {
	item, ok := p.iter.Peek()
	if !ok {
		p.error(errNoToken)
		return lex.Item{}
	}
	return item
}

type stateFn func(*parser) stateFn

func Parse(in string) (*Query, error) {
	l := lex.Lex(in)

	// read all the tokens
	items := l.ReadAll()

	p := &parser{
		iter:  collection.NewIterator(items...),
		query: &Query{},
	}

	return p.run()
}

func isKeyword(item lex.Item, keyword string) bool {
	if item.Typ != lex.ItemKeyword {
		return false
	}

	return strings.ToLower(keyword) == strings.ToLower(item.Val)
}

func sqlStatement(p *parser) stateFn {
	next := p.mustNext()
	switch next.Typ {
	case lex.ItemMultiLineComment, lex.ItemSingleLineComment:
		p.query.addComment(next.Val)
		return sqlStatement
	case lex.ItemKeyword:
		switch strings.ToUpper(next.Val) {
		case StatementSelect.String():
			p.query.Stmt = StatementSelect
			return sqlColumns
		default:
			return p.errorf("unsupported keyword found [%s]", next.Val)
		}
	default:
		return p.errorf("unsupported next type [%v] found within [%s]", next.Typ, "sqlStatement")
	}
}

func sqlColumns(p *parser) stateFn {
	var col Column
	for {
		next := p.mustNext()
		switch {
		case next.Typ == lex.ItemComma: // a ',' indicates the end of a select statement item
			p.query.addSelect(col)
			return sqlColumns
		case isKeyword(next, lex.KeywordFrom):
			p.query.addSelect(col)
			return sqlFrom
		case isKeyword(next, lex.KeywordAs): // AS means the next identifier is an alias
			switch alias := p.mustNext(); alias.Typ {
			case lex.ItemIdentifier:
				col.Alias = alias.Val
				continue
			default:
				return p.errorf("expected identifier, found [%v]", alias.Typ)
			}
		case next.Typ == lex.ItemIdentifier, next.Typ == lex.ItemBacktickedIdentifier:
			switch peek := p.mustPeek(); peek.Typ {
			case lex.ItemDot: // '.' indicates this is a table/column combo
				col.Table = next.Val
				p.mustNext()
				continue
			default:
				col.Column += next.Val
				continue
			}
		case next.Val == ColumnAsterisk: // '*' indicates a wildcard
			col.Column = ColumnAsterisk
			continue
		default:
			return p.errorf("unsupported next type [%v] found within [%s]", next.Typ, "sqlColumns")
		}
	}
}

func sqlFrom(p *parser) stateFn {
	return nil
}
