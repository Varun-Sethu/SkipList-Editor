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
	documentSize 	uint
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
	size 	uint

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
func (list *SkipList) search(cursor uint, offset uint) (*entry, uint) {

	// just iterate the list until we find what we are looking for :)
	curr := list.topLevel
	prev := list.topLevel

	for curr != nil {
		for curr != nil && cursor >= curr.size {
			cursor -= curr.size

			prev = curr
			curr = curr.next
		}
		// if the cursor cant progress any further in this level of the list just hop down
		prev = curr
		if curr != nil {
			if curr.bottom != nil {
				curr.size += offset
			}
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


// Insert inserts a new PieceDescriptor into the skip list at a specific cursor
func (list *SkipList) Insert(descriptor *pieceDescriptor, cursor uint) {
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
		case interval.size - 1:
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
// deleteMode indicates if we are fixing a list after a deletion... if this is the case then we dont do any bubbling
func (list *SkipList) fixList(target *entry, offset uint) {

	// repeatedly go backwards: whenever we "go up a level" we need to add the offset to the
	// level we pop up to to correct the interval ranges, only bother if the offset is non zero though
	if offset != 0 {
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
	var i uint = 0
	for ; i < nodeLevel; i++ {
		// The core idea is: keep going backwards until we find a node that we can climb up
		curr := target
		var spanningRange uint = 0
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
func (list *SkipList) DeleteRange(start, end uint) {
	// locate the two entries where the cursors belong
	entryS, cStart := list.search(start, 0)
	entryE, cEnd   := list.search(end, 0)

	// since these return entries our cursors are in there are 2 cases: the two cursors span the same entry
	// or the span a set of entries

	// To illustrate deletion consider the following
	/*
		[] --- [] ---- []
		[] - [] - [] - []
	 */
	// case where deletion range spans a set of entries
	var deleteEntry = true
	if entryS != entryE {
		// to deal with partial span of an entry we simply just update the cursor values for that entry
		if cStart > 0 {
			entryS.size -= entryS.size - cStart
			entryS.payload.editSize = entryS.size
			entryS = entryS.next
		}

		// Since we are essentially trying to delete the entire bottom row we can bubble up "offset" changes from entryS
		// iterate from entryS -> entryE and start chomping boii
		curr := entryS.next
		for curr != entryE.next {
			cop := curr.next
			list.deleteEntry(curr)
			curr = cop
		}
	// case where it spans a single entry
	} else {
		// 3 situations:
		// the requested deletion range spans to the end of the range
		// or the requested deletion range is in between a range: we deal with them independently
		if cEnd == entryE.size { // spans to end
			// we just need to update the size of entryS to accommodate for this :)
			entryS.size = cStart
			entryS.payload.editSize = cStart
			deleteEntry = false
		} else if cStart == 0 { // spans from start
			entryS.size = cEnd
			entryS.payload.bufferStart = cEnd
			deleteEntry = false
		} else { // spans across
			// just split entryE into two separate pieces
			// [cStart] -> [void] -> [cEnd]
			split := &entry{size: cEnd, next: entryS.next, prev: entryE, top: nil, bottom: nil}
			entryS.next = split
			// construct an associated descriptor for the split now :)
			split.payload = &pieceDescriptor{
				bufferSource: entryS.payload.bufferSource, bufferStart: entryS.payload.bufferStart + cEnd,
				editSize: cEnd}
		}
	}

	list.fixList(entryS, -(end - start))
	// I'm really sorry... I generally hate flags but yeah
	if deleteEntry {
		list.deleteEntry(entryS)
	}

}


// deleteEntry just removes an entry from a skip list
// note: it assumes that the entry is the smallest partition (it is at the bottom)
// furthermore it assumes all partition offsets have already been updated prior to deletion
func (list *SkipList) deleteEntry(target *entry) {
	if target.top != nil {
		// recursively delete parent
		list.deleteEntry(target.top)
	}


	// when deleting an entry there are two possible cases:
	// the node we are deleting is at the leftmost edge of the skip list, in which me make its successor
	// the new "big range" otherwise the predecessor inherits the span
	if target.prev == nil {
		// two cases: we are deleting a row where there is no suitable replacement on the same row (replacement is bottom)
		// or the other thing ^ opposite of that where replacement is the next value
		replacement := target.next
		// no suitable replacement on the same row so just make the replacement the bottom
		if target.next == nil { replacement = target.bottom }
		if replacement == nil {return}

		// now once again there are 2 cases :P; we have no reasonable parent to attach to (we connect to list.topLevel)
		// or we just connect to out parent
		if target.top == nil { list.topLevel = replacement; replacement.top = nil }

		if target.next != nil {
			parent := target.top
			child := target.bottom
			if parent != nil { parent.bottom = replacement }
			replacement.top = parent
			// now just insert the kid
			if child != nil {
				child.top = replacement
				// if replacement had a child delete it
				if replacement.bottom != nil {
					replacement.bottom.top = nil
				}
			}
			replacement.bottom = child
			// shit just works aight?
			if target.bottom != nil { replacement.size += target.size }
			target.next.prev = replacement
		}
		replacement.prev = nil
	} else {
		// the other case is when there is a reasonable predecessor, to deal with this case we "merge" the spanning
		// range of this entry into the previous entry :)
		replacement := target.prev
		child  		:= target.bottom

		// we only bother fixing corrections if the parent is not nil
		replacement.next = target.next
		if target.next != nil {
			target.next.prev = replacement
		}

		// if we aren't at the bottom layer just merge :)
		if child != nil {
			replacement.size += target.size
		}
	}


	// more edge cases coz im bad
	if list.topLevel.next == nil {
		list.topLevel.size = list.documentSize
	}
}






