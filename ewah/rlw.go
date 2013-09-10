/*
 * Copyright (c) 2013 Zhen, LLC. http://zhen.io. All rights reserved.
 * Use of this source code is governed by the Apache 2.0 license.
 *
 */

package ewah

import (
	"fmt"
)

const (
	RunningLengthBits int32 = 32
	LiteralBits int32 = 64 - 1 - RunningLengthBits
	LargestLiteralCount int32 = int32((1 << uint32(LiteralBits)) - 1)
	LargestRunningLengthCount int64 = (int64(1) << uint32(RunningLengthBits)) - 1
	RunningLengthPlusRunningBit int64 = (int64(1) << uint32(RunningLengthBits + 1)) - 1
	ShiftedLargestRunningLengthCount int64 = LargestRunningLengthCount << 1
	NotRunningLengthPlusRunningBit int64 = ^RunningLengthPlusRunningBit
	NotShiftedLargestRunningLengthCount int64 = ^ShiftedLargestRunningLengthCount
)

type runningLengthWord struct {
	*int64
}

// New creates a new running length word
// a is an array of 64-bit words
// p is the position in the array where the running length word is located
func newRunningLengthWord(a []int64, p int64) *runningLengthWord {
	//return &((*runningLengthWord)(&(a[p])))
	return &runningLengthWord{
		int64: &(a[p]),
	}
}

func (this *runningLengthWord) reset(a []int64, p int64) *runningLengthWord {
	this.int64 = &(a[p])
	return this
}

func (this *runningLengthWord) getActualWord() int64 {
	return int64(*this.int64)
}

// getNumberOfLiteralWords gets the number of literal words
func (this *runningLengthWord) getNumberOfLiteralWords() int32 {
	// logical shift right
	return int32(uint64(*this.int64) >> uint32((1 + RunningLengthBits)))
}

// getRunningBit gets the running bit
func (this *runningLengthWord) getRunningBit() bool {
	return (int64(*this.int64) & 1) != 0
}

// getRunningLength gets the running length
func (this *runningLengthWord) getRunningLength() int64 {
	// logical shift right
	return int64((uint64(*this.int64) >> 1)) & LargestRunningLengthCount
}

// setNumberOfLiteralWords sets the number of literal words
func (this *runningLengthWord) setNumberOfLiteralWords(n int64) {
	*this.int64 = int64(*this.int64) | NotRunningLengthPlusRunningBit
	*this.int64 = int64(*this.int64) & ((n << uint64(RunningLengthBits + 1)) | RunningLengthPlusRunningBit)
}

// setRunningBit sets the running bit
func (this *runningLengthWord) setRunningBit(b bool) {
	if b {
		*this.int64 = int64(*this.int64) | 1
	} else {
		*this.int64 = int64(*this.int64) & ^1
	}
}

// setRunningLength sets the running length
func (this *runningLengthWord) setRunningLength(n int64) {
	//fmt.Printf("setRunningLength      n: %064b\n", uint64(n))
	//fmt.Printf("setRunningLength before: %064b\n", uint64(*this.int64))
	//fmt.Printf("setRunningLength before: %064b\n", this.a[this.p])
	*this.int64 = int64(*this.int64) | ShiftedLargestRunningLengthCount
	//fmt.Printf("setRunningLength shfitd: %064b\n", ShiftedLargestRunningLengthCount)
	//fmt.Printf("setRunningLength before: %064b\n", uint64(*this.int64))
	*this.int64 = int64(*this.int64) & ((n << 1) | NotShiftedLargestRunningLengthCount)
	//fmt.Printf("setRunningLength  after: %064b\n", uint64(*this.int64))
}

// size returns the size in uncompressed words represented by this running length word
func (this *runningLengthWord) size() int64 {
	return this.getRunningLength() + int64(this.getNumberOfLiteralWords())
}

func (this *runningLengthWord) String() string {
	return fmt.Sprintf("runningBit = %t, size = %d, runningLength = %d, numberOfLiteralWords = %d\n", this.getRunningBit(), this.size(), this.getRunningLength(), this.getNumberOfLiteralWords())
}
