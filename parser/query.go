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

type Validator interface {
	Valid() bool
}

type Query struct {
	Comments []string
	Stmt     Statement
	Selects  []Column
	Froms    []Table
}

type Column struct {
	Table  string
	Column string
	Alias  string
}

func (c Column) Valid() bool {
	return c.Column != ""
}

type Table struct {
	Name  string
	Alias string
}

func (t Table) Valid() bool {
	return t.Name != ""
}
