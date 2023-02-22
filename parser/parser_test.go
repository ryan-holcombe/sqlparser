package parser

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParse_SelectColumns(t *testing.T) {
	t.Run("asterisk", func(t *testing.T) {
		input := `SELECT * FROM users;`
		query, err := Parse(input)
		assert.NoError(t, err)
		assert.Equal(t, StatementSelect, query.Stmt)
		assert.Len(t, query.Selects, 1)
		assert.Equal(t, Column{Column: "*"}, query.Selects[0])
	})

	t.Run("alias", func(t *testing.T) {
		input := `SELECT name AS user_name FROM users;`
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
}
