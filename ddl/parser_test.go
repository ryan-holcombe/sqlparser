package ddl

import "testing"

func TestParse_CreateTable(t *testing.T) {
	input := `CREATE TABLE users (
    user_id int PRIMARY KEY,
    username STRING(MAX) NOT NULL,
    password STRING(MAX) NOT NULL
);`
}
