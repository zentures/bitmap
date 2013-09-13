/*
 * Copyright (c) 2013 Zhen, LLC. http://zhen.io. All rights reserved.
 * Use of this source code is governed by the Apache 2.0 license.
 *
 */

package ewah

import (
	"fmt"
)

type runningLengthWord struct {
	// m is the pointer to the marker word in the buffer
	m *int64

	// p is the index to the marker word in the buffer
	p int64

	// s is the size, or total number of words, this marker word represents
	s int64
}

// New creates a new running length word
// a is an array of 64-bit words
// p is the position in the array where the running length word is located
func newRunningLengthWord(a []int64, p int64) *runningLengthWord {
	//return &((*runningLengthWord)(&(a[p])))
	rlw := &runningLengthWord{
		m: &(a[p]),
		p: p,
	}

	rlw.s = int64(rlw.getNumberOfLiteralWords()) + rlw.getRunningLength()

	return rlw
}

func (this *runningLengthWord) reset(a []int64, p int64) *runningLengthWord {
	this.m = &(a[p])
	this.p = p
	this.s = int64(this.getNumberOfLiteralWords()) + this.getRunningLength()
	return this
}

func (this *runningLengthWord) getActualWord() int64 {
	return int64(*this.m)
}

// getNumberOfLiteralWords gets the number of literal words
func (this *runningLengthWord) getNumberOfLiteralWords() int32 {
	// logical shift right
	return int32(uint64(*this.m) >> uint32((1 + RunningLengthBits)))
}

// getRunningBit gets the running bit
func (this *runningLengthWord) getRunningBit() bool {
	return (int64(*this.m) & 1) != 0
}

// getRunningLength gets the running length
func (this *runningLengthWord) getRunningLength() int64 {
	// logical shift right
	return int64((uint64(*this.m) >> 1)) & LargestRunningLengthCount
}

// setNumberOfLiteralWords sets the number of literal words
func (this *runningLengthWord) setNumberOfLiteralWords(n int64) {
	*this.m = int64(*this.m) | NotRunningLengthPlusRunningBit
	*this.m = int64(*this.m) & ((n << uint64(RunningLengthBits + 1)) | RunningLengthPlusRunningBit)
}

// setRunningBit sets the running bit
func (this *runningLengthWord) setRunningBit(b bool) {
	if b {
		*this.m = int64(*this.m) | 1
	} else {
		*this.m = int64(*this.m) & ^1
	}
}

// setRunningLength sets the running length
func (this *runningLengthWord) setRunningLength(n int64) {
	//fmt.Printf("setRunningLength      n: %064b\n", uint64(n))
	//fmt.Printf("setRunningLength before: %064b\n", uint64(*this.m))
	//fmt.Printf("setRunningLength before: %064b\n", this.a[this.p])
	*this.m = int64(*this.m) | ShiftedLargestRunningLengthCount
	//fmt.Printf("setRunningLength shfitd: %064b\n", ShiftedLargestRunningLengthCount)
	//fmt.Printf("setRunningLength before: %064b\n", uint64(*this.m))
	*this.m = int64(*this.m) & ((n << 1) | NotShiftedLargestRunningLengthCount)
	//fmt.Printf("setRunningLength  after: %064b\n", uint64(*this.m))
}

// size returns the size in uncompressed words represented by this running length word
func (this *runningLengthWord) size() int64 {
	return this.s
}

func (this *runningLengthWord) String() string {
	return fmt.Sprintf("runningBit = %t, size = %d, runningLength = %d, numberOfLiteralWords = %d\n", this.getRunningBit(), this.size(), this.getRunningLength(), this.getNumberOfLiteralWords())
}
