package piecetable

import (
	"math/rand"
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
 */

// TODO: make not ugly

// type definitions
// defines an entry descriptor, contains a partitioning number that indicates where this "split" occurs
type entryDescriptor struct {
	partitionSize uint

	// next descriptor is the next descriptor in the linked list
	// child descriptor is the child in the level beneath the descriptor's level
	nextDescriptor  *entryDescriptor
	prevDescriptor  *entryDescriptor
	childDescriptor *entryDescriptor
	parentDescriptor *entryDescriptor
}

// delete is a method that removes a entryDescriptor from a linked list of descriptors, also deletes any
// descriptors above it recursively
// TODO: implement deletion from the skiplist level
func (descriptor *entryDescriptor) delete() {
	// might segfault :); correction: will segfault
	if descriptor.parentDescriptor != nil {
		descriptor.parentDescriptor.delete()
	}

	descriptor.prevDescriptor.nextDescriptor = descriptor.nextDescriptor
}


// describes a level in the skipList
type skipListLevel struct {
	levelHead *entryDescriptor
	// should be null
	levelTail *entryDescriptor
}

// definition of a skipList structure
type SkipList struct {
	topLevel *skipListLevel
	// at level 1: the index of the partition entry is the index of the pieceDescriptor
	payloadData []pieceDescriptor
}


// TODO: fix how data is tied to nodes within the skip list



// NewSkipList allocates and returns a new skipList
func NewSkipList(originalEdit pieceDescriptor) *SkipList {
	// when we allocate a skipList it only has one level (the base level)
	var partitionRepresentation entryDescriptor = entryDescriptor{
		partitionSize:    originalEdit.editStart - originalEdit.editEnd,
		nextDescriptor:   nil,
		childDescriptor:  nil,
		prevDescriptor:   nil,
		parentDescriptor: nil,
	}

	return &SkipList{
		topLevel: &skipListLevel{&partitionRepresentation, nil},
		payloadData: []pieceDescriptor{originalEdit},
	}
}


// searchSkipList returns the pieceDescriptor that contains a specified cursor
// also returns the "modified" value of the cursor
func (list *SkipList) searchSkipList(cursor uint) (*entryDescriptor, uint) {

	// continuously iterate every node in the level of interest until we reach the base of the linked list
	var currentPartition *entryDescriptor = list.topLevel.levelHead
	for currentPartition.childDescriptor != nil {
		for currentPartition != nil && cursor > currentPartition.partitionSize {
			cursor -= currentPartition.partitionSize
			currentPartition = currentPartition.nextDescriptor
		}

		// if the current partition is nil that means the requested cursor was out of bounds
		if currentPartition == nil {
			return nil, 0
		}
		currentPartition = currentPartition.childDescriptor
	}



	return currentPartition, cursor
}


// InsertDescriptor adds a new piece descriptor to the skip list at a specific cursor, partitioning is then handled
// from there
func (list *SkipList) InsertDescriptor(edit pieceDescriptor, cursor uint) error {

	var currentPartition *entryDescriptor
	currentPartition, cursor = list.searchSkipList(cursor)

	// now that the "smallest partition" contain our cursor has been identified we need to deal with it accordingly...
	// either: split the node into two or insert directly after
	var insertAfter bool = currentPartition.nextDescriptor == nil ||
							currentPartition.nextDescriptor.partitionSize == cursor
	// the partitioning we are about to insert
	var newPartitioning entryDescriptor = entryDescriptor{
		prevDescriptor: currentPartition,
		childDescriptor: nil,
		parentDescriptor: nil,
		nextDescriptor:  currentPartition.nextDescriptor,
		partitionSize: edit.editStart - edit.editEnd,
	}

	// trivial :P
	if insertAfter {
		// allocate a new partition entry
		currentPartition.nextDescriptor = &newPartitioning
		list.bubbleUpPartition(&newPartitioning)
	} else {
		// we need to split the current partition into two and insert at the split
		var subsequentPartitioning = entryDescriptor{
			childDescriptor: nil,
			parentDescriptor: nil,
			nextDescriptor:  currentPartition.nextDescriptor,
			partitionSize: currentPartition.partitionSize - cursor,
		}
		// split and insert
		newPartitioning.nextDescriptor = &subsequentPartitioning
		currentPartition.nextDescriptor = &newPartitioning
		currentPartition.partitionSize = cursor

		// bubble up values in the skip list
		list.bubbleUpPartition(&newPartitioning)
		list.bubbleUpPartition(&subsequentPartitioning)
	}

	return nil
}



// bubbleUpPartition essentially bubbles up the passed descriptor to higher levels in the skip list
func (list *SkipList) bubbleUpPartition(descriptor *entryDescriptor) {

	// seeding the generated based on the current system time
	rand.Seed(time.Now().UnixNano())

	for v := rand.Intn(1); v == 1;  {
		// continuously go backwards, adding up the total size of the partition spanned by our piece descriptor
		// and the descriptor "connected" to the partition above
		var currentDescriptor *entryDescriptor = descriptor
		var culledPartitionSize uint = 0
		for currentDescriptor.parentDescriptor != nil {
			currentDescriptor = currentDescriptor.prevDescriptor
			culledPartitionSize += currentDescriptor.partitionSize
		}
		currentDescriptor = currentDescriptor.parentDescriptor

		// two situations: the current descriptor is null indicating a new level needs to be added
		// or its not null in which we just add the new partitioning after it
		var newPartitionEntry entryDescriptor = entryDescriptor{
			partitionSize:    nil,
			nextDescriptor:   nil,
			prevDescriptor:   nil,
			childDescriptor:  descriptor,
			parentDescriptor: nil,
		}

		if currentDescriptor != nil {
			newPartitionEntry.partitionSize = currentDescriptor.partitionSize - culledPartitionSize
			newPartitionEntry.prevDescriptor = currentDescriptor
			newPartitionEntry.nextDescriptor = currentDescriptor.nextDescriptor
			currentDescriptor.partitionSize = culledPartitionSize
			currentDescriptor.nextDescriptor = &newPartitionEntry
			descriptor = &newPartitionEntry
		} else {
			// a new level in the piece table needs to be constructed
			var newLevel *skipListLevel = &skipListLevel{
				levelHead: &newPartitionEntry,
				levelTail: nil,
			}
			list.topLevel = newLevel
		}
	}
}



// DeleteDescriptors removes all the entry descriptors between cursorStart and cursorEnd
// also splits up entry descriptors contained within this range
func (list *SkipList) DeleteDescriptors(cursorStart, cursorEnd uint) {

	// obtain the two ranges as well as the modified cursors
	lowRange, cursorStart  := list.searchSkipList(cursorStart)
	highRange, cursorEnd := list.searchSkipList(cursorEnd)

	// shave off the edges of low and high range
	// deal with cursorStart first
	lowRange.partitionSize = cursorStart
	lowRange = lowRange.nextDescriptor
	// now shave off the highRange and cursorEnd
	highRange.partitionSize -= cursorEnd
	highRange = highRange.prevDescriptor


	// iterate over the linked list of entry descriptors
	for lowRange != highRange {
		var nextLow = lowRange.nextDescriptor
		lowRange.delete()
		lowRange = nextLow
	}

}






















