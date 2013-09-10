/*
 * Copyright (c) 2013 Zhen, LLC. http://zhen.io. All rights reserved.
 * Use of this source code is governed by the Apache 2.0 license.
 *
 */

package ewah

import "fmt"

type BufferedRunningLengthWord struct {
	actualWord int64
	literalWordOffset int32
	numberOfLiteralWords int32
	runningBit bool
	runningLength int64
}

func newBufferedRunningLengthWord(a int64) *BufferedRunningLengthWord {
	rlw := newRunningLengthWord([]int64{a}, 0)

	return &BufferedRunningLengthWord {
		actualWord: a,
		literalWordOffset: 0,
		numberOfLiteralWords: rlw.getNumberOfLiteralWords(),
		runningBit: rlw.getRunningBit(),
		runningLength: rlw.getRunningLength(),
	}
}

func (this *BufferedRunningLengthWord) getActualWord() int64 {
	return this.actualWord
}

func (this *BufferedRunningLengthWord) discardFirstWords(x int64) {
	if this.runningLength >= x {
		this.runningLength -= x
		return
	}

	x -= this.runningLength
	this.runningLength = 0
	this.literalWordOffset += int32(x)
	this.numberOfLiteralWords -= int32(x)
}

func (this *BufferedRunningLengthWord) getNumberOfLiteralWords() int32 {
	return this.numberOfLiteralWords
}

func (this *BufferedRunningLengthWord) setNumberOfLiteralWords(number int32) {
	this.numberOfLiteralWords = number
}

func (this *BufferedRunningLengthWord) getRunningBit() bool {
	return this.runningBit
}

func (this *BufferedRunningLengthWord) setRunningBit(b bool) {
	this.runningBit = b
}

func (this *BufferedRunningLengthWord) getRunningLength() int64 {
	return this.runningLength
}

func (this *BufferedRunningLengthWord) setRunningLength(number int64) {
	this.runningLength = number
}

func (this *BufferedRunningLengthWord) size() int64 {
	return this.runningLength + int64(this.numberOfLiteralWords)
}

// reset resets the values using the provided word
func (this *BufferedRunningLengthWord) reset(a int64) {
	rlw := newRunningLengthWord([]int64{a}, 0)
	this.actualWord = a
	this.literalWordOffset = 0
	this.numberOfLiteralWords = rlw.getNumberOfLiteralWords()
	this.runningBit = rlw.getRunningBit()
	this.runningLength = rlw.getRunningLength()
}

func (this *BufferedRunningLengthWord) String() string {
	return fmt.Sprintf("%064b\nrunning bit = %t, running length = %d, literal words = %d\n", this.getActualWord(), this.getRunningBit(), this.getRunningLength(), this.getNumberOfLiteralWords())
}

