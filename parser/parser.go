package parser

import (
	"errors"
	"fmt"

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

func (p *parser) addComment(comment string, next stateFn) stateFn {
	if !p.validate(&comment) {
		return p.errorf("invalid comment found")
	}
	p.query.Comments = append(p.query.Comments, comment)
	return next
}

func (p *parser) addSelect(column Column, next stateFn) stateFn {
	if !p.validate(&column) {
		return p.errorf("invalid select column found")
	}
	p.query.Selects = append(p.query.Selects, column)
	return next
}

func (p *parser) addFrom(from Table, next stateFn) stateFn {
	if !p.validate(&from) {
		return p.errorf("invalid table found in FROM clause")
	}
	p.query.Froms = append(p.query.Froms, from)
	return next
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

func (p *parser) skip() {
	p.mustNext()
}

func (p *parser) validate(typ interface{}) bool {
	if valid, ok := typ.(Validator); ok {
		return valid.Valid()
	}

	return true
}

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
