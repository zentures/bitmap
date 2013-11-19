/*
 * Copyright (c) 2013 Zhen, LLC. http://zhen.io. All rights reserved.
 * Use of this source code is governed by the Apache 2.0 license.
 *
 */

package ewah

import (
	"errors"
	"fmt"
	"math"
)

// cursor is a struct that keeps track of the last marker checked.
// Reference: http://drum.lib.umd.edu/bitstream/1903/544/2/CS-TR-2286.1.pdf - section 3.1
// Take a page from the skiplist search with finger concept
// For sequential checks, this should speed it up dramatically. If the check is previous to the cursor,
// then we just start from the beginning (at least for now.)
type cursor struct {
	// buffer is a slice pointing to the original data
	buffer []uint64

	// size is the size of the buffer, in words
	bsize int64

	// marker is the position of the last marker (runningLengthWord) word checked
	marker int64

	// emptyChecked is the number of uncompressed empty words checked for this marker word
	emptyChecked int64

	// literalChecked is the number of uncompressed literal words checked for this marker word
	literalChecked int64

	// totalChecked is the total number of uncompressed words that's been checked for the whole bitmap
	totalChecked int64

	// Keep track of these so we don't have to do bitwise op every time
	emptyCnt     int64
	literalCnt   int64
	emptyWordBit bool
}

func newCursor(a []uint64, s int64) *cursor {
	f := new(cursor)
	f.resetMarker(a, s, 0)
	return f
}

func (this *cursor) reset(a []uint64, s int64) {
	this.resetMarker(a, s, 0)
}

// quickUpdate only updates the buffer and buffer size without changing anything else
func (this *cursor) quickUpdate(a []uint64, s int64) {
	this.buffer = a
	this.bsize = s

	this.updateMarkerCounts()
}

func (this *cursor) updateMarkerCounts() {
	this.emptyCnt = int64((this.buffer[this.marker] >> 1) & LargestRunningLengthCount)
	this.literalCnt = int64(this.buffer[this.marker] >> uint32((1 + RunningLengthBits)))
	this.emptyWordBit = (int64(this.buffer[this.marker]) & 1) != 0
}

func (this *cursor) resetMarker(a []uint64, s int64, m int64) {
	this.buffer = a
	this.bsize = s
	this.marker = m

	// WARNING: this might cause bugs in the future. Once you reset the marker, we can no longer treat
	// the number of words checked as valid since we really don't know how many words there were before
	this.totalChecked = 0

	this.emptyChecked = 0
	this.literalChecked = 0

	this.updateMarkerCounts()
}

func (this *cursor) nextMarker() error {
	if this.end() {
		return errors.New("cursor.go/nextMarker: No more markers in this buffer")
	}

	this.marker += this.literalCount() + 1
	this.emptyChecked = 0
	this.literalChecked = 0

	this.updateMarkerCounts()

	return nil
}

// moveForward moves the cursor forward by X words, effectively discarding them
func (this *cursor) moveForward(x int64) (int64, error) {
	a := x

	for x > 0 {
		// We are trying to move forward by x words. If the remaining empty words in this marker is more than x,
		// it means we have still more empty words then we just move the emptyChecked forward, and move on.
		if this.emptyRemaining() > x {
			this.emptyChecked += x
			x = 0
			break
		}

		// If we don't have enough empty words to cover x, then we just move forward by the number of empty
		// words left, which means we have fully checked all the empty words for this marker.
		x -= this.emptyRemaining()
		this.emptyChecked = this.emptyCount()

		// Given that we have more words, we have to figure out how many literal words we need to move forward.
		// So we need to figure out if we have enough literal words to cover x.
		// Basically we are moving forward "n" words, which is the minimum of x or numOfLiteralWords
		// If x is greater, then we just move forward and discard all the literal words.
		// If we have more literal words, then we just move forward x words
		n := int64(math.Min(float64(x), float64(this.literalRemaining())))
		this.literalChecked += n

		// If n == x, then x becomes 0; if n < x, then x is greater than 0.
		// n cannot be greater than x, given the above min(), so x should never be < 0
		x -= n

		// If we have exhausted the current marker word, or if we still haven't moved forward enough,
		// then we should go to the next marker and continue from there
		if x > 0 || this.markerRemaining() == 0 {
			// If we are at the end then break
			//if this.end() {
			//	break
			//}

			// Otherwise we go to the next marker word and start the process again
			// If there's no next marker then it's the end
			if this.nextMarker() != nil {
				break
			}
		}
	}

	this.totalChecked += a - x
	return a - x, nil
}

// copyForward copies X words of the buffer into the container, and moves forward to the next word
func (this *cursor) copyForward(container BitmapStorage, max int64, negated bool) (int64, error) {
	if container == nil {
		return 0, errors.New("cursor:copyForward: container is nil")
	}

	//fmt.Printf("cursor.go/copyForward: max = %d\n", max)
	// index keeps track of the number of words we have copied so far
	index := int64(0)

	// If the words we have copied is less than max, and there are still words remaining in the marker,
	// then we will continue to loop and copy
	for index < max && this.markerRemaining() > 0 {
		var pl, pd int64

		// First we will copy all the empty words over first. If there are more empty words than we need,
		// then we will only copy up to max.
		if pl = this.emptyRemaining(); pl > 0 {
			if index+pl > max {
				pl = max - index
			}

			// Copy the words into the result set with the same 0 or 1 setting
			container.addStreamOfEmptyWords(this.emptyBit(), pl)

			// Update the index to reflect the number of words copied
			index += pl
		}
		//fmt.Printf("cursor.go/copyForward: pl = %d\n", pl)

		// Now we copy the remaining literal words. If there are more literal words than we need, then we
		// just copy up to max
		if pd = this.literalRemaining(); pd > 0 {
			if pd+index > max {
				pd = max - index
			}

			// Copy the literal words into the container, starting at the next unchecked position
			start := this.marker + this.literalChecked + 1
			if !negated {
				container.addStreamOfLiteralWords(this.buffer, int32(start), int32(pd))
			} else {
				container.addStreamOfNegatedLiteralWords(this.buffer, int32(start), int32(pd))
			}

			// Update the index to reflect the number of words copied
			index += pd
		}
		//fmt.Printf("cursor.go/copyForward: pd = %d\n", pd)

		// Now that we have copied the words, move the cursor forward
		if _, err := this.moveForward(pl + pd); err != nil {
			return index, err
		}
	}

	//fmt.Printf("cursor.go/copyForward: index = %d\n", index)

	return index, nil
}

func (this *cursor) copyForwardEmpty(container BitmapStorage) (int64, error) {
	if container == nil {
		return 0, errors.New("cursor:copyForwardEmpty: container is nil")
	}

	n := int64(0)

	for s := this.markerRemaining(); s > 0; s = this.markerRemaining() {
		container.addStreamOfEmptyWords(false, s)
		n += s

		if _, err := this.moveForward(s); err != nil {
			return n, err
		}
	}

	return n, nil
}

// Copy the remaining words in the bitmap into the result container
func (this *cursor) copyForwardRemaining(container BitmapStorage) (int64, error) {
	if container == nil {
		return 0, errors.New("cursor:copyForwardRemaining: container is nil")
	}

	n := int64(0)

	for {
		container.addStreamOfEmptyWords(this.emptyBit(), this.emptyRemaining())
		n += this.emptyRemaining()

		container.addStreamOfLiteralWords(this.buffer, int32(this.marker+this.literalChecked)+1, int32(this.literalRemaining()))
		n += this.literalRemaining()

		this.moveForward(this.markerRemaining())

		if this.end() {
			break
		}

	}

	return n, nil
}

func (this *cursor) getLiteralWordAt(k int64) uint64 {
	n := this.marker + this.literalChecked + 1 + k
	if n >= this.bsize {
		fmt.Printf("cursor.go/getLiteralWordAt: ERROR cursor = %v\n", this)
	}
	return this.buffer[n]
}

func (this *cursor) String() string {
	return fmt.Sprintf("Buffer size = %d, marker = %d, totalChecked = %d, literalChecked = %d, literalTotal = %d, emptyChecked = %d, emptyTotal = %d",
		this.bsize, this.marker, this.totalChecked, this.literalChecked, this.literalCount(), this.emptyChecked, this.emptyCount())
}

func (this *cursor) end() bool {
	if this.marker+this.literalChecked+1 >= this.bsize {
		return true
	}

	return false
}

func (this *cursor) markerWord() uint64 {
	return this.buffer[this.marker]
}

func (this *cursor) markerRemaining() int64 {
	return this.emptyRemaining() + this.literalRemaining()
}

func (this *cursor) literalCount() int64 {
	//return int64(this.buffer[this.marker] >> uint32((1 + RunningLengthBits)))
	return this.literalCnt
}

func (this *cursor) emptyBit() bool {
	//return (int64(this.buffer[this.marker]) & 1) != 0
	return this.emptyWordBit
}

func (this *cursor) emptyCount() int64 {
	//return int64((this.buffer[this.marker] >> 1) & LargestRunningLengthCount)
	return this.emptyCnt
}

func (this *cursor) literalRemaining() int64 {
	return this.literalCnt - this.literalChecked
}

func (this *cursor) emptyRemaining() int64 {
	return this.emptyCnt - this.emptyChecked
}

func (this *cursor) setLiteralCount(n int64) {
	this.buffer[this.marker] |= NotRunningLengthPlusRunningBit
	this.buffer[this.marker] &= (uint64(n) << uint64(RunningLengthBits+1)) | RunningLengthPlusRunningBit
}

func (this *cursor) setEmptyBit(b bool) {
	if b {
		this.buffer[this.marker] |= uint64(1)
	} else {
		this.buffer[this.marker] &= ^uint64(1)
	}
}

func (this *cursor) setEmptyCount(n int64) {
	this.buffer[this.marker] |= ShiftedLargestRunningLengthCount
	this.buffer[this.marker] &= (uint64(n) << 1) | NotShiftedLargestRunningLengthCount
}

// size returns the size in uncompressed words represented by this running length word
func (this *cursor) size() int64 {
	return this.emptyCnt + this.literalCnt
}
