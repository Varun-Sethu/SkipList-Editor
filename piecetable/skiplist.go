package piecetable

/*
	IMPLEMENTATION DETAILS:
		- The skip list is implemented as a normal linked list of piece descriptors (at the bottom level)
		- To reduce memory overhead all the levels above are a linked list of "entry descriptors"
		- Entry descriptors are a simple structure with a pointer to the linked list entry it describes and
		  a pointer to the next entry in its linked list level
 */

/*
	The core idea: Piece tables essentially define some partition of an edited file, each partition contains some
	payload data representing the value of the partition as well as the size of the partition, if for the levels
	of our Skip List we simply just describe arbitrary partitioning of the file them hopefully we get a decent complexity
	for traversing that list

	Fundamentally: the skip list is a layer of partitions each stacked on top of each other, layers of partitions
					align to some degree
 */



// SkipList is the definition of a skipList structure
type SkipList struct {
	topLevel *entry
	// the size of the document we are partitioning
	documentSize 	int
}

// NewSkipList takes a piece descriptor and constructs a skip list from it
func NewSkipList(descriptor *pieceDescriptor) *SkipList {
	return &SkipList{
		topLevel: &entry{
			size: descriptor.editSize,
			top: nil, bottom: nil, next: nil, prev: nil,
			payload: descriptor,
		},
		documentSize: descriptor.editSize,
	}
}



// just allocate a new level and make its size the document size :)
func (list *SkipList) newLevel() *entry {
	var newLevel = &entry{
		size: list.documentSize,
		top: nil,
		bottom: list.topLevel,
		next: nil,
		prev: nil,
	}
	list.topLevel = newLevel
	return newLevel
}



// entry just represents some section of a partition
type entry struct {
	// size of the entry
	size 	int

	// pointer data
	top 	*entry
	bottom 	*entry
	next 	*entry
	prev 	*entry

	// optional: pointer to payload data
	payload *pieceDescriptor
}


// search finds the smallest interval in the skip list containing our cursor
// searches all partitions
// returns the smallest entry and an integer with the "new offset"
func (list *SkipList) search(cursor int) (*entry, int) {

	// just iterate the list until we find what we are looking for :)
	curr := list.topLevel
	prev := list.topLevel
	accessCount := 0

	for curr != nil {
		for curr != nil && cursor > curr.size {
			cursor -= curr.size
			prev = curr
			curr = curr.next
			accessCount += 1
		}
		// if the cursor cant progress any further in this level of the list just hop down
		prev = curr
		if curr != nil {
			curr = curr.bottom
		}
	}

	// upon termination there are 2 distinct cases:
	// both curr and prev are nil indicating the cursor is out of bounds
	// or just a normal situation where we return prev
	if prev == nil {
		return nil, 0
	}
	return prev, cursor
}
















