package piecetable

import (
	"fmt"
	"testing"
)



// just initialise the skip list we are testing all operations on
var testList *SkipList = nil
// The constructed skip list is rather simple... it looks like:
/*
	[x] --- [115] ------------------------------------- [15] ----------- [ nil ]
	[x] --- [90] --------------------- [25] ----------- [15] ----------- [ nil ]
	[x] --- [30] --- [50] --- [10] --- [20] --- [5] --- [10] --- [5] --- [ nil ]
*/





// TestSearch just tests searching against a skiplist
func TestSearch(t *testing.T) {

	// get the bottom most level of the testList
	payloadBase := testList.topLevel.bottom.bottom
	// offsets of this base will be used for verification

	// tests defines a map of cursors and the size of the appropriate interval they belong to (smallest interval)
	tests := map[uint]*entry{
		89: payloadBase.next.next,
		25: payloadBase,
		130: payloadBase.next.next.next.next.next.next,
		10000: nil,
		95: payloadBase.next.next.next,
		129: payloadBase.next.next.next.next.next.next,
	}

	// for each test just assert that the two corresponding values in the map are the same
	// also assert that this is the smallest interval (no child pointer)
	for cursor, expectedInterval := range tests {
		containingInterval, _ := testList.search(cursor, 0)

		if containingInterval != expectedInterval {
			t.Errorf("Wrong interval... Expected: %p got %p", expectedInterval, containingInterval)
		}
	}

}



// TestInsert just tests the insertion of an element into the skip list
// note: it assumes "search" is valid and working
func TestInsert(t *testing.T) {
	testDescriptor := &pieceDescriptor{editSize: 30, bufferStart: 10, bufferSource: changes}
	testList.Insert(testDescriptor, 15)

	if out, _ := testList.search(16, 0); out.payload != testDescriptor {
		t.Errorf("Insertion failed... \nexpected: %v\n got: %v\n", *testDescriptor, *out.payload)
	}
	fmt.Print(testList.visualiseList())
}


// TestDelete just tests the deletion function for the skiplist
func TestDelete(t *testing.T) {
	fmt.Print(testList.visualiseList(), "\n")
	testList.DeleteRange(0, 130)
	fmt.Print(testList.visualiseList(), "\n")
}




// at bottom coz ugly :)
// just build the skip list in question... sorry this method is ugly :( I didnt want to use normal insertion
// functions as we "technically" dont know if they're correct yet
func init() {

	// initialize the
	testList = NewSkipList(&pieceDescriptor{
		bufferSource: original, editSize: 30, bufferStart: 0,
	})

	listTail := testList.topLevel
	var layerTwoConnections = []*entry{listTail}

	// the values we are inserting in the first layer
	layerOne := []uint{30, 50, 10, 20, 5, 10, 5}
	for i, val := range layerOne[1:] {
		listTail.next = &entry{size: val, next: nil, top: nil, bottom: nil, prev: listTail, payload: nil}
		testList.documentSize += val
		listTail = listTail.next

		// append to layer two connections if the index is 0, 3 or 5
		if i == 2 || i == 4 {
			layerTwoConnections = append(layerTwoConnections, listTail)
		}
	}


	// the values in the second layer
	listTail = testList.newLevel()
	listTail.size = 90
	listTail.bottom = layerTwoConnections[0]

	// create new entries
	listTail.next = &entry{size: 25, next: nil, prev: listTail, bottom: layerTwoConnections[1], top: nil}
	listTail.next.next = &entry{size: 15, next: nil, prev: listTail.next, bottom: layerTwoConnections[2], top: nil}
	// update the entries underneath them
	layerTwoConnections[0].top = listTail
	layerTwoConnections[1].top = listTail.next
	layerTwoConnections[2].top = listTail.next.next


	// just connect up the third layer normally, this is a cbbs :(
	thirdListTail := testList.newLevel()
	thirdListTail.size = 115
	thirdListTail.bottom = listTail
	listTail.top = thirdListTail

	thirdListTail.next = &entry{size: 15, next: nil, top: nil, bottom: listTail.next.next, prev: thirdListTail, payload: nil}
	listTail.next.next.top = thirdListTail.next


	fmt.Printf("Constructed Skip List\n")
}







