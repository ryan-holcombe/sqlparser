package ddl

type ColumnType string

func (s ColumnType) String() string {
	return string(s)
}

const (
	ColumnTypeBool      ColumnType = "BOOL"
	ColumnTypeInt64     ColumnType = "INT64"
	ColumnTypeFloat64   ColumnType = "FLOAT64"
	ColumnTypeNumeric   ColumnType = "NUMERIC"
	ColumnTypeString    ColumnType = "STRING"
	ColumnTypeBytes     ColumnType = "BYTES"
	ColumnTypeDate      ColumnType = "DATE"
	ColumnTypeTimestamp ColumnType = "TIMESTAMP"
	ColumnTypeJSON      ColumnType = "JSON"
)

type CreateTable struct {
	Name     string
	Comments []string
	Columns  []TableColumn
}

type TableColumn struct {
	Name     string
	BaseType string // base type: BOOL, INT64, FLOAT64, NUMERIC, STRING, BYTES, DATE, TIMESTAMP, JSON
	TypeSize string // the number in parentheses: STRING(10), STRING(MAX)
	Array    bool
	NotNull  bool
}
