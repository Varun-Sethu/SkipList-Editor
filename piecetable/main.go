package piecetable

import "bytes"

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
	originalBuffer 	*bytes.Buffer
	editBuffer		*bytes.Buffer

	changesTable   *SkipList
}




// NewPieceTable returns a new PieceTable implementation
func NewPieceTable(originalBuf string) *PieceTable {
	initialDescriptor := &pieceDescriptor{bufferStart: 0,
						bufferSource: original, editSize: len(originalBuf)}

	return &PieceTable{
		originalBuffer: bytes.NewBuffer([]byte(originalBuf)),
		editBuffer: bytes.NewBuffer([]byte{}),
		changesTable: NewSkipList(initialDescriptor),
	}
}


// Insert just adds a chunk of text to the piece table at the specified cursor
func (table *PieceTable) Insert(addition string, cursor int) {

	if table.changesTable.topLevel == nil {
		// instead of inserting we create a new entry with the appropriate buffer
		*table = *NewPieceTable(addition)
		return
	}

	newDescriptor := &pieceDescriptor{bufferSource: changes,
		bufferStart: table.editBuffer.Len(), editSize: len(addition)}
	table.editBuffer.WriteString(addition)
	table.changesTable.Insert(newDescriptor, cursor)
}


// DeleteRange just deletes a range of words from the piece table
func (table *PieceTable) DeleteRange(start, end int) {
	table.changesTable.DeleteRange(start, end)
}



// Stringify just reads everything in the underlying skip list and returns a string
func (table *PieceTable) Stringify() string {
	// Identify the base
	curr := table.changesTable.topLevel
	if curr == nil { return "" }

	for curr.bottom != nil {
		curr = curr.bottom
	}


	// iterate and print
	var output string = ""
	for curr != nil {
		s := curr.payload.bufferStart
		e := s + curr.payload.editSize
		if curr.payload.bufferSource == original {
			output += string(table.originalBuffer.Bytes()[s:e])
		} else {
			output += string(table.editBuffer.Bytes()[s:e])
		}
		curr = curr.next
	}
	return output
}






