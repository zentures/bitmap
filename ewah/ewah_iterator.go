/*
 * Copyright (c) 2013 Zhen, LLC. http://zhen.io. All rights reserved.
 * Use of this source code is governed by the Apache 2.0 license.
 *
 */

package ewah

// EWAHIterator represents a special type of efficient iterator iterating over (uncompressed) words of bits.
// It is not meant for end users.
type EWAHIterator struct {
	// The size in words.
	size int64

	// The current running length word.
	rlw *runningLengthWord

	// The pointer represent the location of the current running length word in the array of words
	// (embedded in the rlw attribute).
	pointer int32

	array []uint64
}

func NewEWAHIterator(a []uint64, sizeInWords int64) *EWAHIterator {
	return &EWAHIterator{
		size: sizeInWords,
		pointer: 0,
		array: a,
		rlw: newRunningLengthWord(a, 0),
	}
}

func (this *EWAHIterator) buffer() []uint64 {
	return this.array
}

// Position of the literal words represented by this running length word.
func (this *EWAHIterator) literalWords() int32 {

	//fmt.Println("ewah_iterator.go/literalWords: pointer =", this.pointer)
	return this.pointer - this.rlw.getNumberOfLiteralWords()
}

func (this *EWAHIterator) hasNext() bool {
	return int64(this.pointer) < this.size
}

func (this *EWAHIterator) next() *runningLengthWord {
	if int64(this.pointer) < this.size {
		this.rlw.reset(this.array, int64(this.pointer))
		this.pointer += this.rlw.getNumberOfLiteralWords() + 1
		return this.rlw
	}

	return nil
}
