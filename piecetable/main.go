package piecetable




// The piece table structure
// describes an edit made on the physical text buffer
type pieceDescriptor struct {
	bufferSource bool

	bufferStart int
	editSize	int
}
const original bool = false
const changes  bool = true

// The actual PieceTable structure :)
type PieceTable struct {
	originalBuffer 	[]byte
	editBuffer		[]byte

	changesTable   *SkipList
}
