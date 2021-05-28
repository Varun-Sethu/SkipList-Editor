package piecetable

import (
	"fmt"
	"math/rand"
	"testing"
	"time"
)



// TestPieceTableInsertion tests the insertion function
func TestPieceTableInsertion(t *testing.T) {
	table := NewPieceTable("hello world")
	table.Insert("abcd", 1)
	table.Insert("jamie says: ", 0)
	table.Insert("abcd", 10)
	table.Insert("abcd", 11)
	table.Insert("abcd", 16)
	table.Insert("abcd", 30)
	table.Insert("abcd", 7)
	table.Insert("abcd", 39)

	fmt.Print("\n", table.changesTable.visualiseList(), "\n")
	fmt.Printf("%s\n\n", table.Stringify())
	//fmt.Printf("%d\n", table.changesTable.documentSize)
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
				cursor := uint(rand.Intn(int(table.changesTable.documentSize - 1)) + 1)
				b.StartTimer()
				table.Insert(text, cursor)
			}

		}
	})

}