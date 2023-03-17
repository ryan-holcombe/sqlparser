package parse

import (
	"errors"
	"fmt"

	"github.com/ryan-holcombe/sqlparser/collection"
	"github.com/ryan-holcombe/sqlparser/lex"
)

var errNoToken = errors.New("no lex token found, when one was expected")

type StateFn[V any] func(*Parser[V]) StateFn[V]

type Parser[V any] struct {
	Result *V
	iter   *collection.Iterator[lex.Item]
	err    error
}

func (p *Parser[V]) HasError() bool {
	return p.err != nil
}

func (p *Parser[V]) Error(err error) StateFn[V] {
	p.err = errors.Join(p.err, err)
	return nil
}

func (p *Parser[V]) Errorf(format string, args ...interface{}) StateFn[V] {
	return p.Error(fmt.Errorf(format, args...))
}

func (p *Parser[V]) MustNext() lex.Item {
	item, ok := p.iter.Next()
	if !ok {
		p.Error(errNoToken)
		return lex.Item{}
	}
	return item
}

func (p *Parser[V]) MustPeek() lex.Item {
	item, ok := p.iter.Peek()
	if !ok {
		p.Error(errNoToken)
		return lex.Item{}
	}
	return item
}

func (p *Parser[V]) Skip() {
	p.MustNext()
}

func (p *Parser[V]) Get() (*V, error) {
	return p.Result, p.err
}

func NewParser[V any](in string) *Parser[V] {
	var result V

	// read all the tokens
	l := lex.Lex(in)
	items := l.ReadAll()

	return &Parser[V]{
		Result: &result,
		iter:   collection.NewIterator(items...),
	}
}
