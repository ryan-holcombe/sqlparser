package query

import (
	"github.com/ryan-holcombe/sqlparser/parse"
)

func Parse(in string) (*Query, error) {
	p := parse.NewParser[Query](in)
	for state := sqlStatement; state != nil && !p.HasError(); {
		state = state(p)
	}

	return p.Get()
}

func addSelect(p *parse.Parser[Query], column Column, next parse.StateFn[Query]) parse.StateFn[Query] {
	if !parse.Validate(&column) {
		return p.Errorf("invalid select column found")
	}
	p.Result.Selects = append(p.Result.Selects, column)
	return next
}

func addFrom(p *parse.Parser[Query], from Table, next parse.StateFn[Query]) parse.StateFn[Query] {
	if !parse.Validate(&from) {
		return p.Errorf("invalid table found in FROM clause")
	}
	p.Result.Froms = append(p.Result.Froms, from)
	return next
}

func addComment(p *parse.Parser[Query], comment string, next parse.StateFn[Query]) parse.StateFn[Query] {
	if !parse.Validate(&comment) {
		return p.Errorf("invalid comment found")
	}
	p.Result.Comments = append(p.Result.Comments, comment)
	return next
}
