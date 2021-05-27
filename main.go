package main

import (
	"fmt"
	"skiplist-editor/piecetable"
)

func main() {
	table := piecetable.NewPieceTable("hello world!")
	table.Insert(" editor", 5)
	table.Insert(" pasta is tasty", 5)
	table.Insert(" and not nice", 15)
	fmt.Printf("%s\n", table.Stringify())
}
