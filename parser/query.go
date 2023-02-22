package parser

type Statement string

func (s Statement) String() string {
	return string(s)
}

const (
	StatementSelect Statement = "SELECT"
)

const (
	ColumnAsterisk = "*"
)

type Query struct {
	Comments []string
	Stmt     Statement
	Selects  []Column
}

type Column struct {
	Table  string
	Column string
	Alias  string
}

func (q *Query) addComment(comment string) {
	q.Comments = append(q.Comments, comment)
}

func (q *Query) addSelect(column Column) {
	q.Selects = append(q.Selects, column)
}
