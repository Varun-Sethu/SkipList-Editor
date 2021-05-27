package piecetable

import (
	"fmt"
	"testing"
)



// TestPieceTableInsertion tests the insertion function
func TestPieceTableInsertion(t *testing.T) {
	table := NewPieceTable("hello world!")
	table.Insert(" editor", 5)
	table.Insert(" pasta is tasty", 5)
	table.Insert(" and not nice", 15)
	
	fmt.Print(table.changesTable.visualiseList(), "\n")
	fmt.Printf("%s\n\n", table.Stringify())
}