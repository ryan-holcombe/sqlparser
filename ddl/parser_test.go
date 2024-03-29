package ddl

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParse_CreateTable(t *testing.T) {
	t.Run("simple table", func(t *testing.T) {
		input := strings.TrimSpace(`CREATE TABLE users (
    user_id int PRIMARY KEY,
    username varchar(MAX) NOT NULL,
    password varchar(MAX) NOT NULL);`)
		table, err := Parse(input)
		assert.NoError(t, err)
		_ = table
	})
}
