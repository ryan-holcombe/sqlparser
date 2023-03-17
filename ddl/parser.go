package ddl

import (
	"github.com/ryan-holcombe/sqlparser/collection"
	"github.com/ryan-holcombe/sqlparser/lex"
)

type parser struct {
	iter   *collection.Iterator[lex.Item]
	tables CreateTables
	err    error
}
