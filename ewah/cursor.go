/*
 * Copyright (c) 2013 Zhen, LLC. http://zhen.io. All rights reserved.
 * Use of this source code is governed by the Apache 2.0 license.
 *
 */

package ewah

import (
	"math"
	"fmt"
	"errors"
)

// cursor is a struct that keeps track of the last marker checked.
// Reference: http://drum.lib.umd.edu/bitstream/1903/544/2/CS-TR-2286.1.pdf - section 3.1
// Take a page from the skiplist search with finger concept
// For sequential checks, this should speed it up dramatically. If the check is previous to the cursor,
// then we just start from the beginning (at least for now.)
type cursor struct {
	// buffer is a slice pointing to the original data
	buffer []int64

	// size is the size of the buffer, in words
	bsize int64

	// marker is the position of the last marker (runningLengthWord) word checked
	marker int64

	// checked is the total number of words that's been checked (or moved forward)
	checked int64

	// rlw is the current running length word, basically buffer[marker]
	//rlw *runningLengthWord

	// rlwEmptyRemaining is the number of empty words remaining (unchecked) for this running length (marker) word
	rlwEmptyRemaining int64

	// rlwLiteralRemaining is the number of literal words remaining (unchecked) for this running length (marker) word
	rlwLiteralRemaining int64

	// rlwLiteralChecked is the number of literal words checked for this running length word (or marker word)
	rlwLiteralChecked int64

	// pointer points to the last location in the array
	pointer int64
}

func newCursor(a []int64, s int64) *cursor {
	//fmt.Println("cursor.go/New")
	f := new(cursor)
	f.reset(a, s)
	return f
}

func (this *cursor) reset(a []int64, s int64) {
	//fmt.Println("cursor.go/reset")
	this.buffer = a
	this.bsize = s
	this.marker = 0
	this.checked = 0
	this.pointer = 0
	//this.rlw = newRunningLengthWord(a, 0)
	this.rlwEmptyRemaining = this.getRunningLength()
	this.rlwLiteralRemaining = int64(this.getNumberOfLiteralWords())
	this.rlwLiteralChecked = 0
}

func (this *cursor) resetMarker(a []int64, s int64, m int64) {
	this.buffer = a
	this.bsize = s
	this.marker = m

	// WARNING: this might cause bugs in the future. Once you reset the marker, we can no longer treat
	// the number of words checked as valid since we really don't know how many words there were before
	this.checked = 0

	this.pointer = 0
	//this.rlw = newRunningLengthWord(a, 0)
	this.rlwEmptyRemaining = this.getRunningLength()
	this.rlwLiteralRemaining = int64(this.getNumberOfLiteralWords())
	this.rlwLiteralChecked = 0
}

func (this *cursor) nextMarker() error {
	//fmt.Println("cursor.go/nextMarker")
	this.marker += int64(this.getNumberOfLiteralWords())+1
	this.pointer = this.marker
	if this.end() {
		return errors.New("cursor.go/nextMarker: No more markers in this buffer")
	}

	//fmt.Printf("cursor.go/nextMarker: marker = %064b\n", uint64(this.buffer[this.marker]))

	//this.rlw.reset(this.buffer, this.marker)
	this.rlwEmptyRemaining = this.getRunningLength()
	this.rlwLiteralRemaining = this.getNumberOfLiteralWords()
	this.rlwLiteralChecked = 0

	//fmt.Printf("cursor.go/nextMarker: cursor = %v\n", this)

	return nil
}

func (this *cursor) rlwRemaining() int64 {
	return this.rlwEmptyRemaining + this.rlwLiteralRemaining
}

func (this *cursor) end() bool {
	if this.pointer+1 >= this.bsize {
		return true
	}

	return false
}

// moveForward moves the cursor forward by X words, effectively discarding them
func (this *cursor) moveForward(x int64) int64 {
	//fmt.Printf("cursor.go/moveForward: 1.x = %d, rlwEmptyRemaining = %d\n", x, this.rlwEmptyRemaining)
	a := x

	for x > 0 {
		// We are trying to move forward by x words. If the remaining empty words in this RLW is more than x,
		// it means we have still more empty words then we just move the rlwEmptyRemaining forward, and move on.
		if this.rlwEmptyRemaining > x {
			this.rlwEmptyRemaining -= x
			x = 0
			break
		}
		//fmt.Printf("cursor.go/moveForward: 2.x = %d, emptyRemain = %d\n", x, this.rlwEmptyRemaining)

		// If we don't have enough empty words to cover x, then we just move forward by the number of empty
		// words left, which means we have no more empty words remaining to check for this marker word.
		x -= this.rlwEmptyRemaining
		this.rlwEmptyRemaining = 0
		//fmt.Printf("cursor.go/moveForward: 3.x = %d, emptyRemain = %d\n", x, this.rlwEmptyRemaining)

		// Given that we have more words, we have to figure out how many literal words we need to move forward.
		// So we need to figure out if we have enough literal words to cover x.
		// Basically we are moving forward "n" words, which is the minimum of x or numOfLiteralWords
		// If x is greater, then we just move forward and discard all the literal words.
		// If we have more literal words, then we just move forward x words
		literalRemaining := int64(this.getNumberOfLiteralWords()) - this.rlwLiteralChecked
		n := int64(math.Min(float64(x), float64(literalRemaining)))
		//fmt.Printf("cursor.go/moveForward: literalRemaining = %d, n = %d, cursor = %v\n", literalRemaining, n, this)
		this.rlwLiteralChecked += n
		this.pointer += n
		this.rlwLiteralRemaining -= n
		//fmt.Println("cursor.go/moveForward: cursor =", this)

		// If n == x, then x becomes 0; if n < x, then x is greater than 0.
		// n cannot be greater than x, given the above min(), so x should never be < 0
		x -= n

		// If we have exhausted the current marker word, or if we still haven't moved forward enough,
		// then we should go to the next marker and continue from there
		if x > 0 || this.rlwRemaining() == 0 {
			// If we are at the end then break
			//fmt.Printf("cursor.go/moveForward: marker = %d, literalWords = %d\n", this.marker, this.getNumberOfLiteralWords())
			if this.end() {
				break
			}

			// Otherwise we go to the next marker word and start the process again
			if this.nextMarker() != nil {
				break
			}
		}
	}

	this.checked += a-x
	//fmt.Printf("cursor.go/moveForward: 4.x = %d, a = %d, cursor = %v\n", x, a, this)
	return a-x
}

// copyForward copies X words of the buffer into the container, and moves forward to the next word
func (this *cursor) copyForward(container BitmapStorage, max int64, negated bool) int64 {
	// index keeps track of the number of words we have copied so far
	index := int64(0)

	// If the words we have copied is less than max, and there are still words remaining in the marker,
	// then we will continue to loop and copy
	for index < max && this.rlwEmptyRemaining + this.rlwLiteralRemaining > 0 {
		// First we will copy all the empty words over first. If there are more empty words than we need,
		// then we will only copy up to max.
		pl := this.rlwEmptyRemaining
		if index + pl > max {
			pl = max - index
		}

		// Copy the words into the result set with the same 0 or 1 setting
		container.addStreamOfEmptyWords(this.getRunningBit(), pl)

		// Update the index to reflect the number of words copied
		index += pl

		// Now we copy the remaining literal words. If there are more literal words than we need, then we
		// just copy up to max
		pd := this.rlwLiteralRemaining
		if pd + index > max {
			pd = max - index
		}

		// Copy the literal words into the container, starting at the next unchecked position
		start := this.marker + int64(this.getNumberOfLiteralWords()) - this.rlwLiteralRemaining + 1
		if !negated {
			container.addStreamOfLiteralWords(this.buffer, int32(start), int32(pd))
		} else {
			container.addStreamOfNegatedLiteralWords(this.buffer, int32(start), int32(pd))
		}
		
		// Now that we have copied the words, move the cursor forward
		this.moveForward(pl + int64(pd))

		// Update the index to reflect the number of words copied
		index += int64(pd)
	}

	return index
}

func (this *cursor) copyForwardEmpty(container BitmapStorage) int64 {
	n := int64(0)
	//fmt.Printf("cursor.go/copyEmptyForward: bsize = %d, pointer = %d cursor = %v\n", this.bsize, this.pointer, this)

	for s := this.rlwRemaining(); s > 0; s = this.rlwRemaining() {
		//fmt.Println("cursor.go/copyEmptyForward: s =", s, "cursor =", this)
		container.addStreamOfEmptyWords(false, s)
		this.moveForward(s)
		n += s
	}

	return n
}

// Copy the remaining words in the bitmap into the result container
func (this *cursor) copyForwardRemaining(container BitmapStorage) int64 {
	n := int64(0)
	for {
		//fmt.Printf("cursor.go/copyForwardRemaining: n = %d, cursor = %v\n", n, this)
		//fmt.Printf("cursor.go/copyForwardRemaining: marker = %064b\n", uint64(this.buffer[this.marker]))
		container.addStreamOfEmptyWords(this.getRunningBit(), this.rlwEmptyRemaining)
		n += this.rlwEmptyRemaining

		container.addStreamOfLiteralWords(this.buffer, int32(this.pointer)+1, int32(this.rlwLiteralRemaining))
		n += this.rlwLiteralRemaining

		this.moveForward(this.rlwRemaining())
		//fmt.Printf("cursor.go/copyForwardRemaining: after moveForward: n = %d, cursor = %v\n", n, this)

		if this.end() {
			break
		}
	}

	return n
}

func (this *cursor) getLiteralWordAt(k int64) int64 {
	n := this.marker + int64(this.getNumberOfLiteralWords()) - this.rlwLiteralRemaining + 1 + k
	if n >= this.bsize {
		fmt.Printf("cursor.go/getLiteralWordAt: ERROR cursor = %v\n", this)
	}
	//fmt.Printf("cursor.go/getLiteralWordAt: k = %d, n = %d, %064b\n", k, n, uint64(this.buffer[n]))
	return this.buffer[n]
}

func (this *cursor) String() string {
	return fmt.Sprintf("Size = %d, marker = %d, checked = %d, emptyRemain = %d, literalRamin = %d, literalChecked = %d, pointer = %d\n",
		this.bsize, this.marker, this.checked,
		this.rlwEmptyRemaining, this.rlwLiteralRemaining, this.rlwLiteralChecked,
		this.pointer)
}


func (this *cursor) getMarkerWord() int64 {
	return this.buffer[this.marker]
}

// getNumberOfLiteralWords gets the number of literal words
func (this *cursor) getNumberOfLiteralWords() int64 {
	// logical shift right
	return int64(uint64(this.buffer[this.marker]) >> uint32((1 + RunningLengthBits)))
}

// getRunningBit gets the running bit
func (this *cursor) getRunningBit() bool {
	return (int64(this.buffer[this.marker]) & 1) != 0
}

// getRunningLength gets the running length
func (this *cursor) getRunningLength() int64 {
	// logical shift right
	return int64((uint64(this.buffer[this.marker]) >> 1)) & LargestRunningLengthCount
}

// setNumberOfLiteralWords sets the number of literal words
func (this *cursor) setNumberOfLiteralWords(n int64) {
	this.buffer[this.marker] |= NotRunningLengthPlusRunningBit
	this.buffer[this.marker] &= (n << uint64(RunningLengthBits + 1)) | RunningLengthPlusRunningBit
}

// setRunningBit sets the running bit
func (this *cursor) setRunningBit(b bool) {
	if b {
		this.buffer[this.marker] |= 1
	} else {
		this.buffer[this.marker] &= ^1
	}
}

// setRunningLength sets the running length
func (this *cursor) setRunningLength(n int64) {
	//fmt.Printf("setRunningLength      n: %064b\n", uint64(n))
	//fmt.Printf("setRunningLength before: %064b\n", uint64(*this.m))
	//fmt.Printf("setRunningLength before: %064b\n", this.a[this.p])
	this.buffer[this.marker] |= ShiftedLargestRunningLengthCount
	//fmt.Printf("setRunningLength shfitd: %064b\n", ShiftedLargestRunningLengthCount)
	//fmt.Printf("setRunningLength before: %064b\n", uint64(*this.m))
	this.buffer[this.marker] &= (n << 1) | NotShiftedLargestRunningLengthCount
	//fmt.Printf("setRunningLength  after: %064b\n", uint64(*this.m))
}

// size returns the size in uncompressed words represented by this running length word
func (this *cursor) size() int64 {
	return this.getNumberOfLiteralWords() + this.getRunningLength()
}
