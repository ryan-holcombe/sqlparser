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
		switch {
		case isKeyword(next, "create"):
			return createTable
		case isKeyword(next, "table"):
			// found CREATE TABLE
			return tableName
		default:
			return p.Errorf("unsupported keyword found [%s]", next.Val)
		}
	default:
		return p.Errorf("unsupported next type [%v] found within [%s]", next.Typ, "createTable")
	}
}

func tableName(p *parse.Parser[CreateTable]) parse.StateFn[CreateTable] {
	next := p.MustNext()
	switch next.Typ {
	case lex.ItemIdentifier:
		p.Result.Name = next.Val
		if next := p.MustNext(); next.Typ != lex.ItemLeftParen {
			return p.Errorf("expected left parenthesis after table name, found [%s] instead", next.Val)
		} else {
			return tableColumns
		}
	default:
		return p.Errorf("expected identifier to define table, found [%s] instead", next.Val)
	}
}

func tableColumns(p *parse.Parser[CreateTable]) parse.StateFn[CreateTable] {
	var column TableColumn
	for {
		next := p.MustNext()
		switch {
		case isKeyword(next, "not"):
			peek := p.MustPeek()
			if isKeyword(peek, "null") {
				column.NotNull = true
				p.Skip()
			} else {
				return p.Errorf("unsupported next type [%v] found while parsing the not null for [%s]", peek.Typ, column.Name)
			}
		case isKeyword(next, "primary"):
			peek := p.MustPeek()
			if isIdentifier(peek, "key") {
				//column.PrimaryKey = true
				p.Skip()
			} else {
				return p.Errorf("unsupported next type [%v] found while parsing the primary key for [%s]", peek.Typ, column.Name)
			}
		case next.Typ == lex.ItemIdentifier && column.Name == "":
			column.Name = next.Val
		case next.Typ == lex.ItemIdentifier:
			column.BaseType = strings.ToUpper(next.Val)
		case next.Typ == lex.ItemLeftParen:
			next := p.MustNext()
			peek := p.MustPeek()
			if next.Typ == lex.ItemIdentifier && peek.Typ == lex.ItemRightParen {
				column.TypeSize = strings.ToUpper(next.Val)
				p.Skip()
			} else {
				return p.Errorf("unsupported next type [%v] found while parsing the type size for [%s]", next.Typ, column.Name)
			}
		case next.Typ == lex.ItemComma:
			// add the column to the table and look for another
			return addColumn(p, column, tableColumns)
		case next.Typ == lex.ItemRightParen:
			// add the column to the table and return
			return addColumn(p, column, nil)
		default:
			return p.Errorf("unsupported next type [%v] found within [%s]", next.Typ, "tableColumns")
		}

	}
}
