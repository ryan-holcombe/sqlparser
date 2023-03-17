package query

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParse_Comments(t *testing.T) {
	t.Run("single line", func(t *testing.T) {
		input := strings.TrimSpace(`-- this will select everything from the users table
SELECT * FROM users;`)
		query, err := Parse(input)
		assert.NoError(t, err)
		assert.Equal(t, StatementSelect, query.Stmt)
		assert.Len(t, query.Comments, 1)
		assert.Equal(t, "-- this will select everything from the users table", query.Comments[0])
	})

	t.Run("multi-line", func(t *testing.T) {
		input := strings.TrimSpace(`/* this will select everything from the users table
and also the addresses table
*/
SELECT * FROM users;`)
		query, err := Parse(input)
		assert.NoError(t, err)
		assert.Equal(t, StatementSelect, query.Stmt)
		assert.Len(t, query.Comments, 1)
		assert.Equal(t, "/* this will select everything from the users table\nand also the addresses table\n*/", query.Comments[0])
		assert.Equal(t, StatementSelect, query.Stmt)
	})
}

func TestParse_SelectColumns(t *testing.T) {
	t.Run("asterisk", func(t *testing.T) {
		input := `SELECT * FROM users;`
		query, err := Parse(input)
		assert.NoError(t, err)
		assert.Equal(t, StatementSelect, query.Stmt)
		assert.Len(t, query.Selects, 1)
		assert.Equal(t, Column{Column: "*"}, query.Selects[0])
	})

	t.Run("alias with AS", func(t *testing.T) {
		input := `SELECT name AS user_name FROM users;`
		query, err := Parse(input)
		assert.NoError(t, err)
		assert.Equal(t, StatementSelect, query.Stmt)
		assert.Len(t, query.Selects, 1)
		assert.Equal(t, Column{Column: "name", Alias: "user_name"}, query.Selects[0])
	})

	t.Run("alias without AS", func(t *testing.T) {
		input := `SELECT name user_name FROM users;`
		query, err := Parse(input)
		assert.NoError(t, err)
		assert.Equal(t, StatementSelect, query.Stmt)
		assert.Len(t, query.Selects, 1)
		assert.Equal(t, Column{Column: "name", Alias: "user_name"}, query.Selects[0])
	})

	t.Run("backticked column", func(t *testing.T) {
		input := "SELECT `name` AS user_name FROM users;"
		query, err := Parse(input)
		assert.NoError(t, err)
		assert.Equal(t, StatementSelect, query.Stmt)
		assert.Len(t, query.Selects, 1)
		assert.Equal(t, Column{Column: "`name`", Alias: "user_name"}, query.Selects[0])
	})

	t.Run("table and column", func(t *testing.T) {
		input := `SELECT user.name FROM users;`
		query, err := Parse(input)
		assert.NoError(t, err)
		assert.Equal(t, StatementSelect, query.Stmt)
		assert.Len(t, query.Selects, 1)
		assert.Equal(t, Column{Table: "user", Column: "name"}, query.Selects[0])
	})

	t.Run("backticked table and column", func(t *testing.T) {
		input := "SELECT `user`.`name` FROM users;"
		query, err := Parse(input)
		assert.NoError(t, err)
		assert.Equal(t, StatementSelect, query.Stmt)
		assert.Len(t, query.Selects, 1)
		assert.Equal(t, Column{Table: "`user`", Column: "`name`"}, query.Selects[0])
	})

	t.Run("backticked table and asterisk", func(t *testing.T) {
		input := "SELECT `user`.* FROM users;"
		query, err := Parse(input)
		assert.NoError(t, err)
		assert.Equal(t, StatementSelect, query.Stmt)
		assert.Len(t, query.Selects, 1)
		assert.Equal(t, Column{Table: "`user`", Column: "*"}, query.Selects[0])
	})

	t.Run("multiple selects", func(t *testing.T) {
		input := `SELECT user.id, user.*, name FROM users;`
		query, err := Parse(input)
		assert.NoError(t, err)
		assert.Equal(t, StatementSelect, query.Stmt)
		assert.Len(t, query.Selects, 3)
		assert.Equal(t, Column{Table: "user", Column: "id"}, query.Selects[0])
		assert.Equal(t, Column{Table: "user", Column: "*"}, query.Selects[1])
		assert.Equal(t, Column{Column: "name"}, query.Selects[2])
	})

	t.Run("missing comma between selects", func(t *testing.T) {
		input := `SELECT user.id user.* name FROM users;`
		_, err := Parse(input)
		assert.Error(t, err)
	})
}

func TestParse_Froms(t *testing.T) {
	t.Run("simple", func(t *testing.T) {
		input := `SELECT * FROM users;`
		query, err := Parse(input)
		assert.NoError(t, err)
		assert.Equal(t, StatementSelect, query.Stmt)
		assert.Len(t, query.Froms, 1)
		assert.Equal(t, Table{Name: "users"}, query.Froms[0])
	})

	t.Run("with alias", func(t *testing.T) {
		input := `SELECT * FROM users U;`
		query, err := Parse(input)
		assert.NoError(t, err)
		assert.Equal(t, StatementSelect, query.Stmt)
		assert.Len(t, query.Froms, 1)
		assert.Equal(t, Table{Name: "users", Alias: "U"}, query.Froms[0])
	})

	t.Run("invalid", func(t *testing.T) {
		input := `SELECT * FROM users U user;`
		_, err := Parse(input)
		assert.Error(t, err)
	})

	t.Run("multiple tables", func(t *testing.T) {
		input := `SELECT * FROM users U, people;`
		query, err := Parse(input)
		assert.NoError(t, err)
		assert.Equal(t, StatementSelect, query.Stmt)
		assert.Len(t, query.Froms, 2)
		assert.Equal(t, Table{Name: "users", Alias: "U"}, query.Froms[0])
		assert.Equal(t, Table{Name: "people"}, query.Froms[1])
	})
}
