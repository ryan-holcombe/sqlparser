package lex

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testExec(input string) []Item {
	l := Lex(input)
	return l.ReadAll()
}

func requireItems(t *testing.T, items []Item, tokens ...interface{}) {
	for i, item := range items {
		require.Greater(t, len(tokens), i, "token for item at index %d, [%v] not found", i, item)

		switch token := tokens[i].(type) {
		case string:
			require.Equal(t, tokens[i], item.Val, "index %d: expected item val [%s], got [%s]", i, tokens[i], item.Val)
		case itemType:
			require.Equal(t, tokens[i], item.Typ, "index %d: expected item type [%v], got [%v]", i, tokens[i], item.Typ)
		case Item:
			require.Equal(t, token, item, "index %d: expected item [%v], got [%v]", i, tokens[i], item)
		default:
			require.Failf(t, "unsupported token type", "found: %v", token)
		}
	}
}

func TestLex_SimpleStatement(t *testing.T) {
	t.Run("missing semicolon", func(t *testing.T) {
		input := `SELECT * FROM users`
		requireItems(t, testExec(input), "SELECT", "*", "FROM", "users", ItemEOF)
	})

	t.Run("with semicolon", func(t *testing.T) {
		input := `SELECT * FROM users;`
		requireItems(t, testExec(input), "SELECT", "*", "FROM", "users", ItemStatementEnd, ItemEOF)
	})

	t.Run("with dot", func(t *testing.T) {
		input := `SELECT * FROM users.person;`
		requireItems(t, testExec(input), "SELECT", "*", "FROM", "users", ItemDot, "person", ItemStatementEnd, ItemEOF)
	})

	t.Run("with select field", func(t *testing.T) {
		input := `SELECT name FROM users.person;`
		requireItems(t, testExec(input), "SELECT", "name", "FROM", "users", ItemDot, "person", ItemStatementEnd, ItemEOF)
	})

	t.Run("with AS keyword", func(t *testing.T) {
		input := `SELECT name AS userName FROM users.person;`
		requireItems(t, testExec(input), "SELECT", "name", "AS", "userName", "FROM", "users", ItemDot, "person", ItemStatementEnd, ItemEOF)
	})

	t.Run("with AS keyword", func(t *testing.T) {
		input := `SELECT name AS userName, address FROM users.person;`
		requireItems(t, testExec(input), "SELECT", "name", "AS", "userName", ItemComma, "address", "FROM", "users", ItemDot, "person", ItemStatementEnd, ItemEOF)
	})

	t.Run("multi-line statement", func(t *testing.T) {
		input := strings.TrimSpace(`
SELECT name
	FROM users.person;
`)
		requireItems(t, testExec(input), "SELECT", "name", "FROM", "users", ItemDot, "person", ItemStatementEnd, ItemEOF)
	})

	t.Run("with backtick", func(t *testing.T) {
		input := "SELECT name FROM `users`.person;"
		requireItems(t, testExec(input), "SELECT", "name", "FROM", ItemBacktickedIdentifier, ItemDot, "person", ItemStatementEnd, ItemEOF)
	})
}

func TestLex_Comments(t *testing.T) {
	t.Run("single line", func(t *testing.T) {
		input := strings.TrimSpace(`-- this is a comment
SELECT name
	FROM users.person;
`)
		requireItems(t, testExec(input), "-- this is a comment", "SELECT", "name", "FROM", "users", ItemDot, "person", ItemStatementEnd, ItemEOF)

	})

	t.Run("multi line", func(t *testing.T) {
		input := strings.TrimSpace(`/* this is a comment
that spans multiple lines
*/
SELECT name
	FROM users.person;
`)
		requireItems(t, testExec(input), "/* this is a comment\nthat spans multiple lines\n*/", "SELECT", "name", "FROM", "users", ItemDot, "person", ItemStatementEnd, ItemEOF)

	})
}

func TestLex_SimpleStatementWithWhereClause(t *testing.T) {
	t.Run("single quotes", func(t *testing.T) {
		input := "SELECT name FROM users.person WHERE name = 'Bob';"
		requireItems(t, testExec(input), "SELECT", "name", "FROM", "users", ItemDot, "person", "WHERE", "name", "=", "'Bob'", ItemStatementEnd, ItemEOF)
	})

	t.Run("double quotes", func(t *testing.T) {
		input := "SELECT * FROM `database`.`table` WHERE `column` = " + `"value";`
		requireItems(t, testExec(input), "SELECT", "*", "FROM", "`database`", ItemDot, "`table`", "WHERE", "`column`", "=", `"value"`, ItemStatementEnd, ItemEOF)
	})
}

func TestLex_ParseErrors(t *testing.T) {
	t.Run("unterminated backtick", func(t *testing.T) {
		input := "SELECT name FROM `users WHERE foo = 'bar';"
		requireItems(t, testExec(input), "SELECT", "name", "FROM", ItemError)
	})

	t.Run("error records correct line number", func(t *testing.T) {
		input := strings.TrimSpace(`
SELECT name
FROM users
WHERE foo = 'bar
AND bar = 'baz';
`)
		items := testExec(input)
		requireItems(t, items, "SELECT", "name", "FROM", "users", "WHERE", "foo", "=", ItemError)
		assert.Equal(t, 2, items[7].Line)
	})
}
