package parser

import (
	"strings"

	"github.com/ryan-holcombe/sqlparser/lex"
)

type stateFn func(*parser) stateFn

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

func sqlStatement(p *parser) stateFn {
	next := p.mustNext()
	switch next.Typ {
	case lex.ItemMultiLineComment, lex.ItemSingleLineComment:
		return p.addComment(next.Val, sqlStatement)
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
		switch next := p.mustNext(); {
		case next.Typ == lex.ItemComma: // a ',' indicates the end of a select statement item
			return p.addSelect(col, sqlColumns)
		case isKeyword(next, lex.KeywordFrom):
			return p.addSelect(col, sqlFrom)
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
				p.skip()
				continue
			case lex.ItemIdentifier: // AS is optional, assume this is the alias
				col.Column = next.Val
				col.Alias = peek.Val
				p.skip()
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
	var tbl Table
	for {
		switch next := p.mustNext(); {
		case next.Typ == lex.ItemComma: // store the current table and look for more in the FROM clause
			return p.addFrom(tbl, sqlFrom)
		case isKeyword(next, lex.KeywordWhere):
			return sqlWhere
		case isKeyword(next, lex.KeywordJoin, lex.KeywordInner, lex.KeywordOuter, lex.KeywordLeft, lex.KeywordRight, lex.KeywordFull):
			return sqlJoin
		case next.Typ == lex.ItemIdentifier, next.Typ == lex.ItemBacktickedIdentifier:
			switch {
			case tbl.Alias != "": // both name and alias are set, something is wrong
				return p.errorf("unknown identifier in [%s], tbl name and alias are already set [%s %s]", "sqlFrom", tbl.Name, tbl.Alias)
			case tbl.Name != "": // name is already set, this is the table alias
				tbl.Alias = next.Val
				continue
			default:
				tbl.Name = next.Val
				continue
			}
		case next.Typ == lex.ItemEOF, next.Typ == lex.ItemStatementEnd: // end of SQL select statement
			return p.addFrom(tbl, nil)
		default:
			return p.errorf("unsupported next type [%v] found within [%s]", next.Typ, "sqlFrom")
		}
	}
}

func sqlWhere(p *parser) stateFn {
	return nil
}

func sqlJoin(p *parser) stateFn {
	return nil
}
