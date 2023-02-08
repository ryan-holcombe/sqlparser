package sqlparser

import (
	"strings"
	"testing"
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

func assertItems(t *testing.T, items []Item, tokens ...interface{}) {
	for i, item := range items {
		if len(tokens) < i+1 {
			t.Fatalf("token for item at index %d, [%v] not found", i, item)
		}

		switch token := tokens[i].(type) {
		case string:
			if item.val != tokens[i] {
				t.Fatalf("expected item val [%s], got [%s]", tokens[i], item.val)
			}
		case itemType:
			if item.typ != token {
				t.Fatalf("expected item type [%v], got [%v]", tokens[i], item.typ)
			}
		case Item:
			if item != token {
				t.Fatalf("expected item [%v], got [%v]", tokens[i], item)
			}
		default:
			t.Fatalf("unsupported token type found: %v", token)
		}
	}
}

func TestLex_SimpleStatement(t *testing.T) {
	t.Run("missing semicolon", func(t *testing.T) {
		input := `SELECT * FROM users`
		assertItems(t, testExec(input), "SELECT", "*", "FROM", "users", itemEOF)
	})

	t.Run("with semicolon", func(t *testing.T) {
		input := `SELECT * FROM users;`
		assertItems(t, testExec(input), "SELECT", "*", "FROM", "users", itemStatementEnd, itemEOF)
	})

	t.Run("with dot", func(t *testing.T) {
		input := `SELECT * FROM users.person;`
		assertItems(t, testExec(input), "SELECT", "*", "FROM", "users", itemDot, "person", itemStatementEnd, itemEOF)
	})

	t.Run("with select field", func(t *testing.T) {
		input := `SELECT name FROM users.person;`
		assertItems(t, testExec(input), "SELECT", "name", "FROM", "users", itemDot, "person", itemStatementEnd, itemEOF)
	})

	t.Run("with AS keyword", func(t *testing.T) {
		input := `SELECT name AS userName FROM users.person;`
		assertItems(t, testExec(input), "SELECT", "name", "AS", "userName", "FROM", "users", itemDot, "person", itemStatementEnd, itemEOF)
	})

	t.Run("with AS keyword", func(t *testing.T) {
		input := `SELECT name AS userName, address FROM users.person;`
		assertItems(t, testExec(input), "SELECT", "name", "AS", "userName", itemComma, "address", "FROM", "users", itemDot, "person", itemStatementEnd, itemEOF)
	})

	t.Run("multi-line statement", func(t *testing.T) {
		input := strings.TrimSpace(`
SELECT name
	FROM users.person;
`)
		assertItems(t, testExec(input), "SELECT", "name", "FROM", "users", itemDot, "person", itemStatementEnd, itemEOF)
	})
}

func TestLex_ParseErrors(t *testing.T) {

}
