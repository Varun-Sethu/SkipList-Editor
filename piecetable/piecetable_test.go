package piecetable

import (
	"fmt"
	"math/rand"
	"testing"
	"time"
)



// TestPieceTableInsertion tests the insertion function
func TestPieceTableInsertion(t *testing.T) {
	table := NewPieceTable("hello world!")
	table.Insert(" editor", 5)
	table.Insert(" pasta is tasty", 5)
	table.Insert(" and not nice", 15)

	//fmt.Print(table.changesTable.visualiseList(), "\n")
	//fmt.Printf("%s\n\n", table.Stringify())
}


// BenchmarkPieceTableInsert tests the piece table on how quickly it can insert
func BenchmarkPieceTableInsert(b *testing.B) {
	for _, size := range []int{1, 10, 100, 1000, 10000, 100000} {
		benchmarkPieceTableInsert(b, size)
	}
}


// sub method for insertion
func benchmarkPieceTableInsert(b *testing.B, size int) {
	text := "abcd"

	rand.Seed(time.Now().UnixNano())

	// insert a bunch of times
	b.Run(fmt.Sprintf("%d", size), func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			b.StopTimer()
			table := NewPieceTable("hello world!")
			b.StartTimer()

			for j := 0; j < size; j++ {
				b.StopTimer()
				cursor := rand.Intn(table.changesTable.documentSize - 1) + 1
				b.StartTimer()

				table.Insert(text, cursor)
			}

		}
	})

}