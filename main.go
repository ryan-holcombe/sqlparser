package sqlparser

import "fmt"

func main() {
	l := Lex("hi {{ id }}")
	for {
		item := l.NextItem()
		fmt.Println(item)
		if item.typ == itemEOF || item.typ == itemError {
			break
		}
	}
}
