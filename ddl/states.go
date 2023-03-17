package ddl

import (
	"strings"

	"github.com/ryan-holcombe/sqlparser/lex"
	"github.com/ryan-holcombe/sqlparser/parse"
)

func createTable(p *parse.Parser[CreateTable]) parse.StateFn[CreateTable] {
	next := p.MustNext()
	switch next.Typ {
	case lex.ItemMultiLineComment, lex.ItemSingleLineComment:
		return addComment(p, next.Val, createTable)
	case lex.ItemKeyword:
		switch strings.ToUpper(next.Val) {
		case StatementCreate.String():
			return tableColumns
		default:
			return p.Errorf("unsupported keyword found [%s]", next.Val)
		}
	default:
		return p.Errorf("unsupported next type [%v] found within [%s]", next.Typ, "createTable")
	}
}

func tableColumns(p *parse.Parser[CreateTable]) parse.StateFn[CreateTable] {
	return nil
}
