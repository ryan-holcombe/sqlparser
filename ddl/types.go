package ddl

type CreateTable struct {
}

type CreateTables []CreateTable

/*
type CreateTable struct {
	Name              ID
	Columns           []ColumnDef
	Constraints       []TableConstraint
	PrimaryKey        []KeyPart
	Interleave        *Interleave
	RowDeletionPolicy *RowDeletionPolicy

	Position Position // position of the "CREATE" token
}

type ColumnDef struct {
	Name    ID
	Type    Type
	NotNull bool

	Default   Expr // set if this column has a default value
	Generated Expr // set of this is a generated column

	Options ColumnOptions

	Position Position // position of the column name
}

type Type struct {
	Array bool
	Base  TypeBase // Bool, Int64, Float64, Numeric, String, Bytes, Date, Timestamp
	Len   int64    // if Base is String or Bytes; may be MaxLen
}
*/
