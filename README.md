# SkipList Editor

This project is just a small experiment in implementing a piece table with a skip list as opposed to a linked list, or a balanced search tree.

## Overview
The idea is to use a probabilistic data structure to manage entries into a piece table. A piece table essentially represents
a sequence of edits on some file, whenever an edit is made at a specific cursor we simply traverse the "linked list" of
edits until we reach that cursor and insert it directly after. The core advantage of a piece table is that we don't really need to shift
everything over in memory when inserting... [more info](https://en.wikipedia.org/wiki/Piece_table).

Fundamentally the linked list in a piece table represents some partition of the document into several sized chunks, eg. a document
may be partitioned as follows: [30] - [10] - [15] - [16]. When implementing the piece table with a skip list we essentially "layer on" several partitions of increasing span, each
element of a partition is then connected to the beginning of the span below it.  For example:
```go
[x] --- [115] ------------------------------------- [15] ----------- [ nil ]
[x] --- [90] --------------------- [25] ----------- [15] ----------- [ nil ]
[x] --- [30] --- [50] --- [10] --- [20] --- [5] --- [10] --- [5] --- [ nil ]
```
On the topmost level of the skip list the entire document has been partitioned into two intervals of sizes 115 and 15, beneath
it we see the partitioning: 90, 25, 15. The 90 interval is connected to the 115 interval above it as it marks the beginning of the interval spanned by the 115 entry.
finally, on the bottom row we have the entries which represent all the edits made in the document. 

### Searching
   - When searching for an interval in the skip list we start with a specific cursor value: n. Starting at the top row we repeatedly subtract interval sizes from n until we reach a pointer
      where subtracting will lead a negative number, we then go down instead (like a normal skip list search)
   - This leads to a log n lookup time in the average case.

## TODO:
 - [ ] When deleting over a span sometimes the skip list bugs out if there have been too many previous insertions