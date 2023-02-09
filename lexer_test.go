package sqlparser

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testExec(input string) []Item {
	var items []Item
	l := Lex(input)
	for {
		item := l.NextItem()
		items = append(items, item)
		if item.typ == itemEOF || item.typ == itemError {
			break
		}
	}
	return items
}

func requireItems(t *testing.T, items []Item, tokens ...interface{}) {
	for i, item := range items {
		require.Greater(t, len(tokens), i, "token for item at index %d, [%v] not found", i, item)

		switch token := tokens[i].(type) {
		case string:
			require.Equal(t, item.val, tokens[i], "index %d: expected item val [%s], got [%s]", i, tokens[i], item.val)
		case itemType:
			require.Equal(t, item.typ, tokens[i], "index %d: expected item type [%v], got [%v]", i, tokens[i], item.typ)
		case Item:
			require.Equal(t, item, token, "index %d: expected item [%v], got [%v]", i, tokens[i], item)
		default:
			require.Failf(t, "unsupported token type", "found: %v", token)
		}
	}
}

func TestLex_SimpleStatement(t *testing.T) {
	t.Run("missing semicolon", func(t *testing.T) {
		input := `SELECT * FROM users`
		requireItems(t, testExec(input), "SELECT", "*", "FROM", "users", itemEOF)
	})

	t.Run("with semicolon", func(t *testing.T) {
		input := `SELECT * FROM users;`
		requireItems(t, testExec(input), "SELECT", "*", "FROM", "users", itemStatementEnd, itemEOF)
	})

	t.Run("with dot", func(t *testing.T) {
		input := `SELECT * FROM users.person;`
		requireItems(t, testExec(input), "SELECT", "*", "FROM", "users", itemDot, "person", itemStatementEnd, itemEOF)
	})

	t.Run("with select field", func(t *testing.T) {
		input := `SELECT name FROM users.person;`
		requireItems(t, testExec(input), "SELECT", "name", "FROM", "users", itemDot, "person", itemStatementEnd, itemEOF)
	})

	t.Run("with AS keyword", func(t *testing.T) {
		input := `SELECT name AS userName FROM users.person;`
		requireItems(t, testExec(input), "SELECT", "name", "AS", "userName", "FROM", "users", itemDot, "person", itemStatementEnd, itemEOF)
	})

	t.Run("with AS keyword", func(t *testing.T) {
		input := `SELECT name AS userName, address FROM users.person;`
		requireItems(t, testExec(input), "SELECT", "name", "AS", "userName", itemComma, "address", "FROM", "users", itemDot, "person", itemStatementEnd, itemEOF)
	})

	t.Run("multi-line statement", func(t *testing.T) {
		input := strings.TrimSpace(`
SELECT name
	FROM users.person;
`)
		requireItems(t, testExec(input), "SELECT", "name", "FROM", "users", itemDot, "person", itemStatementEnd, itemEOF)
	})

	t.Run("with backtick", func(t *testing.T) {
		input := "SELECT name FROM `users`.person;"
		requireItems(t, testExec(input), "SELECT", "name", "FROM", itemBacktickedIdentifier, itemDot, "person", itemStatementEnd, itemEOF)
	})

	t.Run("where clause", func(t *testing.T) {
		input := "SELECT name FROM users.person WHERE name = 'Bob';"
		requireItems(t, testExec(input), "SELECT", "name", "FROM", "users", itemDot, "person", "WHERE", "name", "=", "'Bob'", itemStatementEnd, itemEOF)
	})
}

func TestLex_ParseErrors(t *testing.T) {
	t.Run("unterminated backtick", func(t *testing.T) {
		input := "SELECT name FROM `users WHERE foo = 'bar';"
		requireItems(t, testExec(input), "SELECT", "name", "FROM", itemError)
	})

	t.Run("error records correct line number", func(t *testing.T) {
		input := strings.TrimSpace(`
SELECT name
FROM users
WHERE foo = 'bar
AND bar = 'baz';
`)
		items := testExec(input)
		requireItems(t, items, "SELECT", "name", "FROM", "users", "WHERE", "foo", "=", itemError)
		assert.Equal(t, 2, items[7].line)
	})
}
