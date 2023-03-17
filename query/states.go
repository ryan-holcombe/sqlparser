package query

import (
	"strings"

	"github.com/ryan-holcombe/sqlparser/lex"
	"github.com/ryan-holcombe/sqlparser/parse"
)

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

func sqlStatement(p *parse.Parser[Query]) parse.StateFn[Query] {
	next := p.MustNext()
	switch next.Typ {
	case lex.ItemMultiLineComment, lex.ItemSingleLineComment:
		return addComment(p, next.Val, sqlStatement)
	case lex.ItemKeyword:
		switch strings.ToUpper(next.Val) {
		case StatementSelect.String():
			p.Result.Stmt = StatementSelect
			return sqlColumns
		default:
			return p.Errorf("unsupported keyword found [%s]", next.Val)
		}
	default:
		return p.Errorf("unsupported next type [%v] found within [%s]", next.Typ, "sqlStatement")
	}
}

func sqlColumns(p *parse.Parser[Query]) parse.StateFn[Query] {
	var col Column
	for {
		switch next := p.MustNext(); {
		case next.Typ == lex.ItemComma: // a ',' indicates the end of a select statement item
			return addSelect(p, col, sqlColumns)
		case isKeyword(next, lex.KeywordFrom):
			return addSelect(p, col, sqlFrom)
		case isKeyword(next, lex.KeywordAs): // AS means the next identifier is an alias
			switch alias := p.MustNext(); alias.Typ {
			case lex.ItemIdentifier:
				col.Alias = alias.Val
				continue
			default:
				return p.Errorf("expected identifier, found [%v]", alias.Typ)
			}
		case next.Typ == lex.ItemIdentifier, next.Typ == lex.ItemBacktickedIdentifier:
			switch peek := p.MustPeek(); peek.Typ {
			case lex.ItemDot: // '.' indicates this is a table/column combo
				col.Table = next.Val
				p.Skip()
				continue
			case lex.ItemIdentifier: // AS is optional, assume this is the alias
				col.Column = next.Val
				col.Alias = peek.Val
				p.Skip()
				continue
			default:
				col.Column += next.Val
				continue
			}
		case next.Val == ColumnAsterisk: // '*' indicates a wildcard
			col.Column = ColumnAsterisk
			continue
		default:
			return p.Errorf("unsupported next type [%v] found within [%s]", next.Typ, "sqlColumns")
		}
	}
}

func sqlFrom(p *parse.Parser[Query]) parse.StateFn[Query] {
	var tbl Table
	for {
		switch next := p.MustNext(); {
		case next.Typ == lex.ItemComma: // store the current table and look for more in the FROM clause
			return addFrom(p, tbl, sqlFrom)
		case isKeyword(next, lex.KeywordWhere):
			return sqlWhere
		case isKeyword(next, lex.KeywordJoin, lex.KeywordInner, lex.KeywordOuter, lex.KeywordLeft, lex.KeywordRight, lex.KeywordFull):
			return sqlJoin
		case next.Typ == lex.ItemIdentifier, next.Typ == lex.ItemBacktickedIdentifier:
			switch {
			case tbl.Alias != "": // both name and alias are set, something is wrong
				return p.Errorf("unknown identifier in [%s], tbl name and alias are already set [%s %s]", "sqlFrom", tbl.Name, tbl.Alias)
			case tbl.Name != "": // name is already set, this is the table alias
				tbl.Alias = next.Val
				continue
			default:
				tbl.Name = next.Val
				continue
			}
		case next.Typ == lex.ItemEOF, next.Typ == lex.ItemStatementEnd: // end of SQL select statement
			return addFrom(p, tbl, nil)
		default:
			return p.Errorf("unsupported next type [%v] found within [%s]", next.Typ, "sqlFrom")
		}
	}
}

func sqlWhere(p *parse.Parser[Query]) parse.StateFn[Query] {
	return nil
}

func sqlJoin(p *parse.Parser[Query]) parse.StateFn[Query] {
	return nil
}
