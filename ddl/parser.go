package ddl

import (
	"github.com/ryan-holcombe/sqlparser/parse"
)

type CreateTable struct {
	Comments []string
}

func Parse(in string) (*CreateTable, error) {
	p := parse.NewParser[CreateTable](in)
	for state := createTable; state != nil && !p.HasError(); {
		state = state(p)
	}

	return p.Get()
}

func addComment(p *parse.Parser[CreateTable], comment string, next parse.StateFn[CreateTable]) parse.StateFn[CreateTable] {
	if !parse.Validate(&comment) {
		return p.Errorf("invalid comment found")
	}
	p.Result.Comments = append(p.Result.Comments, comment)
	return next
}
