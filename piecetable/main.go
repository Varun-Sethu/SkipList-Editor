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




// NewPieceTable returns a new PieceTable implementation
func NewPieceTable(originalBuf string) *PieceTable {
	initialDescriptor := &pieceDescriptor{bufferStart: 0,
						bufferSource: original, editSize: len(originalBuf)}

	return &PieceTable{
		originalBuffer: []byte(originalBuf),
		editBuffer: []byte{},
		changesTable: NewSkipList(initialDescriptor),
	}
}


// Insert just adds a chunk of text to the piece table at the specified cursor
func (table *PieceTable) Insert(addition string, cursor int) {
	newDescriptor := &pieceDescriptor{bufferSource: changes,
						bufferStart: len(table.editBuffer), editSize: len(addition)}
	table.editBuffer = append(table.editBuffer, []byte(addition)...)
	table.changesTable.Insert(newDescriptor, cursor)
}


// DeleteRange just deletes a range of words from the piece table
func (table *PieceTable) DeleteRange(start, end int) {
	table.changesTable.DeleteRange(start - 1, end)
}



// Stringify just reads everything in the underlying skip list and returns a string
func (table *PieceTable) Stringify() string {
	// Identify the base
	curr := table.changesTable.topLevel
	for curr.bottom != nil {
		curr = curr.bottom
	}


	// iterate and print
	var output string = ""
	for curr != nil {
		s := curr.payload.bufferStart
		e := s + curr.payload.editSize
		if curr.payload.bufferSource == original {
			output += string(table.originalBuffer[s:e])
		} else {
			output += string(table.editBuffer[s:e])
		}
		curr = curr.next
	}
	return output
}






