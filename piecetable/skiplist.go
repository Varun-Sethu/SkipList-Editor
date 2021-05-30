package piecetable

import (
	"math/bits"
	"math/rand"
	"strconv"
	"time"
)

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

// constants involved with level generation
const (
	maxLevel = 25
)



// SkipList is the definition of a skipList structure
type SkipList struct {
	topLevel *entry
	// the size of the document we are partitioning
	documentSize 	int
}

// NewSkipList takes a piece descriptor and constructs a skip list from it
func NewSkipList(descriptor *pieceDescriptor) *SkipList {
	rand.Seed(time.Now().UnixNano())
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
	list.topLevel.top = newLevel
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





// visualiseList returns a string representation of the skip list (mostly for debugging)
func (list *SkipList) visualiseList() string {
	// Start at the top level, scan across and go down
	currentLevel := list.topLevel
	curr := currentLevel
	outBuffer := ""

	for currentLevel != nil {
		for curr != nil {
			outBuffer += " " + strconv.FormatUint(uint64(curr.size), 10) + " "
			curr = curr.next
		}
		outBuffer += "\n"
		currentLevel = currentLevel.bottom
		curr = currentLevel
	}

	return outBuffer
}


// search finds the smallest interval in the skip list containing our cursor
// searches all partitions
// returns the smallest entry and an integer with the "new offset"
// to speed up operations search also supports an "offset" value that adds an offset every time it goes down a node
// essentially correcting an insertion while finding where to insert
func (list *SkipList) search(cursor int, offset int) (*entry, int) {

	// just iterate the list until we find what we are looking for :)
	curr := list.topLevel
	prev := list.topLevel
	prevPrev := prev

	for curr != nil {
		for curr != nil && cursor >= curr.size {
			cursor -= curr.size

			prevPrev = prev
			prev = curr
			curr = curr.next

		}
		// if the cursor cant progress any further in this level of the list just hop down
		prevPrev = prev
		prev = curr
		if curr != nil {
			if curr.bottom != nil {
				curr.size += offset
			}
			curr = curr.bottom
		}
	}

	// upon termination there are 2 distinct cases:
	// if curr is nil indicating the cursor is out of bounds, to suppport editing past document bounds we simply
	// return the edge node in the document
	// or just a normal situation where we return prev
	if prev == nil {
		return prevPrev, prevPrev.size
	}
	return prev, cursor
}


// Insert inserts a new PieceDescriptor into the skip list at a specific cursor
func (list *SkipList) Insert(descriptor *pieceDescriptor, cursor int) {
	// locate the "interval" that currently contains our cursor
	// also attain "the offset" into that interval (this is returned by the search function)
	interval, cursor := list.search(cursor, descriptor.editSize)

	// cases:
	// if the cursor is 0: insert in front of interval
	// if the cursor is interval.size - 1 then we insert immediately after interval
	// otherwise: we split interval into two pieces and insert accordingly
	// the new entry we are creating
	var newEntry *entry
	// adding a layer is conditional... we only perform it if we inserted a value AFTER internal (to make life a little easy)
	var addLayer bool = true

	switch cursor {
		case 0:
			newEntry = &entry{size: descriptor.editSize, next: interval, prev: interval.prev, top: interval.top,
						payload: descriptor}
			// if the interval has a parent just "disown?" that parent and make the newEntry the new child
			// we only inherit interval's parent if we insert in front of it :)
			if interval.top != nil { interval.top.bottom = newEntry; interval.top = nil
			} else if interval.prev == nil { list.topLevel = newEntry }

			// if interval follows a previous entry then just set its "next" entry to newEntry
			if interval.prev != nil { interval.prev.next = newEntry }
			interval.prev = newEntry
			addLayer = false

			break
		case interval.size:
			newEntry = &entry{size: descriptor.editSize, next: interval.next, prev: interval, top: nil,
							payload: descriptor}
			// if interval actually has a next value update its information
			if interval.next != nil {
				interval.next.prev = newEntry
			}
			interval.next = newEntry

			break
		default:
			newEntry = &entry{size: descriptor.editSize, next: nil, prev: interval, top: nil,
							payload: descriptor}
			// split in two: shrink interval and allocate a new entry to follow "newEntry"
			intervalSecondHalf := &entry{size: interval.payload.editSize - cursor, next: interval.next, prev: newEntry, top: nil,
										payload: &pieceDescriptor{bufferSource: interval.payload.bufferSource,
											bufferStart: interval.payload.bufferStart + cursor, editSize: interval.size - cursor}}
			newEntry.next = intervalSecondHalf
			if interval.next != nil {
				interval.next.prev = intervalSecondHalf
			}
			interval.next = newEntry
			interval.size = cursor; interval.payload.editSize = cursor

			break
	}

	// correct the list by adding layers and fixing weights
	list.documentSize += descriptor.editSize
	if addLayer {
		list.probabilityInsert(newEntry)
	}

}


// fixList performs two functions: given a specific entry and an offset it will correct all the interval sizes
// for the parents of that entry, it will also deal with "the bubbling" to ensure we get a logarithmic average
// case complexity; note: the offset can also be 0 :)
func (list *SkipList) fixList(target *entry, offset int) {

	// repeatedly go backwards: whenever we "go up a level" we need to add the offset to the
	// level we pop up to to correct the interval ranges
	curr := target
	for curr != nil {
		if curr.top == nil {
			curr = curr.prev
		} else {
			curr = curr.top
			curr.size += offset
		}
	}
	list.documentSize += offset

}


// randomLevel prior to the insertion of an entry into the skip list randomLevel will generate the height of that entry
// within the skip list: this seems to perform better than repeatedly generating random numbers (should only call rand once)
func (list *SkipList) randomLevel() uint {
	// minimises calls to rand: https://ticki.github.io/blog/skip-lists-done-right/
	var levels uint = maxLevel
	var x uint64 = rand.Uint64() & ((1 << (maxLevel-1)) - 1)
	var firstSet uint = uint(bits.TrailingZeros64(x))

	// return based on the values
	if firstSet <= maxLevel {
		levels = firstSet
	}
	return levels
}


// probabilityInsert probabilistically inserts an entry into the skip list (bubbling it up)
func (list *SkipList) probabilityInsert(target *entry) {

	// randomly generate the level of this node prior to "upgrading"
	nodeLevel := list.randomLevel()

	// continuously create new levels in our skip list until we reach the desired node level
	for i := uint(0) ; i < nodeLevel; i++ {
		// The core idea is: keep going backwards until we find a node that we can climb up
		curr := target
		var spanningRange int = 0
		for curr.prev != nil && curr.top == nil {
			curr = curr.prev
			spanningRange += curr.size
		}
		curr = curr.top

		// now that we have found it there are two cases: it has no parent indicating we need to allocate a new level
		// or we insert directly after it
		if curr == nil { curr = list.newLevel() }

		// finally insert our new allocation node directly after curr
		newAllocation := &entry{prev: curr, top: nil,
							bottom: target, next: curr.next}
		if curr.next != nil { curr.next.prev = newAllocation }
		curr.next = newAllocation
		target.top = newAllocation

		// just correct the sizes of new allocation and curr
		newAllocation.size = curr.size - spanningRange
		curr.size = spanningRange
		target = newAllocation
	}

}


// DeleteRange takes two cursors and deletes all values between those cursors
func (list *SkipList) DeleteRange(start, end int) {

	deletionSize := end - start

	// locate the two entries where the cursors belong
	lowerBound, start 	:= list.search(start, 0)
	upperBound, end   	:= list.search(end, 0)


	// based off the cursor values we need to update the bounds accordingly
	if lowerBound != upperBound {
		if start > 0 { lowerBound.payload.editSize = start; lowerBound.size = start
				lowerBound = lowerBound.next }
		if end != upperBound.size { upperBound.payload.bufferStart += end
				upperBound.payload.editSize -= end; upperBound.size -= end
				upperBound = upperBound.prev}
	} else { // if the lower bound does equal the upper bound then the operation is simple: split in two and return
		// first: allocate a new node containing the upper bound information
		newUpper := &entry{prev: upperBound, next: upperBound.next,
						size: upperBound.size - end,
						payload: &pieceDescriptor{
							editSize: upperBound.size - end, bufferStart: upperBound.payload.bufferStart + end,
							bufferSource: upperBound.payload.bufferSource},
						top: nil, bottom: nil}

		// then update the lower bound information
		lowerBound.payload.editSize = start; lowerBound.size = start

		if lowerBound.next != nil { lowerBound.next.prev = newUpper }
		lowerBound.next = newUpper
		list.fixList(lowerBound, -deletionSize); return
	}

	// now that the bounds have been fixed ( and if the deletion spans a single entry then performed )
	// we simply iterate from lowerBound to upperBound and delete all the nodes :)
	curr := lowerBound
	lowerBoundPre := lowerBound.prev
	for curr != upperBound.next {
		next := curr.next
		list.deleteEntry(curr)
		curr = next
	}
	list.fixList(lowerBoundPre, -deletionSize)
}


// deleteEntry just removes an entry from a skip list
// note: it assumes that the entry is the smallest partition (it is at the bottom)
// in detail deleteEntry removes a partition by connecting its next pointer to its prev pointer
// furthermore its prev value inherits its span
func (list *SkipList) deleteEntry(target *entry) {

	// recursively delete the top value
	if target.top != nil {
		list.deleteEntry(target.top)
	}


	// connect the adjacent values to target.prev
	if target.prev != nil { target.prev.next = target.next }
	if target.next != nil { target.next.prev = target.prev }
	if target.bottom != nil { target.bottom.top = target.prev }
	if target.top != nil { target.top.bottom = target.prev }

	// finally connect target.prev to the adjacent values
	// 2 cases: target is not on the edge (predecessor inherits) or target is (next value inherits)
	if target.prev != nil {

		if target.bottom != nil { target.prev.size += target.size }
		// we should have delete target's "top value" so it will always be nil :)
		target.top = nil

	} else {
		// we are deleting something from the very edge, there are 2 cases here:
		// no other entries exist in which we delete the row OR: the next value takes up target's position

		if target.next == nil {
			// delete row
			list.topLevel = target.bottom

		} else {
			// allow the next value to inherent target's details
			if target.top == nil { list.topLevel = target.next }
			target.next.bottom = target.bottom
			if target.bottom != nil { target.next.size += target.size; target.bottom.top = target.next }
		}
	}
}