package ddl

import (
	"strings"

	"github.com/ryan-holcombe/sqlparser/lex"
	"github.com/ryan-holcombe/sqlparser/parse"
)

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

func addColumn(p *parse.Parser[CreateTable], column TableColumn, next parse.StateFn[CreateTable]) parse.StateFn[CreateTable] {
	if !parse.Validate(&column) {
		return p.Errorf("invalid table column found")
	}
	p.Result.Columns = append(p.Result.Columns, column)
	return next
}

func isKeyword(item lex.Item, keywords ...string) bool {
	if item.Typ != lex.ItemKeyword {
		return false
	}

	for _, k := range keywords {
		if strings.EqualFold(k, item.Val) {
			return true
		}
	}

	return false
}

func isIdentifier(item lex.Item, keywords ...string) bool {
	if item.Typ != lex.ItemIdentifier {
		return false
	}

	for _, k := range keywords {
		if strings.EqualFold(k, item.Val) {
			return true
		}
	}

	return false
}
