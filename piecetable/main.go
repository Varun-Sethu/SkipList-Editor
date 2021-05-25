package piecetable




// The piece table structure
// describes an edit made on the physical text buffer
type pieceDescriptor struct {
	editType 	uint8
	editStart 	uint
	editEnd 	uint
}
// edit types
const insertion uint8 = 1
const deletion  uint8 = 0