/*
 * Copyright (c) 2013 Zhen, LLC. http://zhen.io. All rights reserved.
 * Use of this source code is governed by the Apache 2.0 license.
 *
 */

package ewah

import (
	"math"
)

type BufferedRunningLengthWordIterator struct {
	buffer []uint64
	literalWordStartPosition int32
	brlw *BufferedRunningLengthWord
	iterator *EWAHIterator
}

var _ RLWIterator = (*BufferedRunningLengthWordIterator)(nil)

func newBufferedRunningLengthWordIterator(iterator *EWAHIterator) *BufferedRunningLengthWordIterator {
	n := iterator.next()
	brlw := newBufferedRunningLengthWord(n.getActualWord())
	pos := iterator.literalWords() + brlw.literalWordOffset;
	//fmt.Println("literalWords() =", iterator.literalWords(), "literalWordOffset =",  brlw.literalWordOffset, "pos =", pos)

	return &BufferedRunningLengthWordIterator{
		iterator: iterator,
		brlw: brlw,
		literalWordStartPosition: pos,
		buffer: iterator.buffer(),
	}
}

func (this *BufferedRunningLengthWordIterator) discardFirstWords(x int64) {
	for x > 0 {
		if this.brlw.runningLength > x {
			this.brlw.runningLength -= x
			return
		}

		x -= this.brlw.runningLength
		this.brlw.runningLength = 0
		toDiscard := int64(math.Min(float64(x), float64(this.brlw.numberOfLiteralWords)))
		this.literalWordStartPosition += int32(toDiscard)
		this.brlw.numberOfLiteralWords -= int32(toDiscard)
		x -= toDiscard

		if x > 0 || this.brlw.size() == 0 {
			if !this.iterator.hasNext() {
				break
			}

			this.brlw.reset(this.iterator.next().getActualWord())
			this.literalWordStartPosition = this.iterator.literalWords()
		}
	}
}

func (this *BufferedRunningLengthWordIterator) getNumberOfLiteralWords() int32 {
	return this.brlw.getNumberOfLiteralWords()
}

func (this *BufferedRunningLengthWordIterator) getRunningLength() int64 {
	return this.brlw.getRunningLength()
}

func (this *BufferedRunningLengthWordIterator) getRunningBit() bool {
	return this.brlw.getRunningBit()
}

func (this *BufferedRunningLengthWordIterator) size() int64 {
	return this.brlw.size()
}

func (this *BufferedRunningLengthWordIterator) getLiteralWordAt(index int32) uint64 {
	//fmt.Println("getLiteralWordAt this.literalWordStartPosition + index =", this.literalWordStartPosition + index)
	return this.buffer[this.literalWordStartPosition + index]
}

func (this *BufferedRunningLengthWordIterator) next() bool {
	if !this.iterator.hasNext() {
		this.brlw.numberOfLiteralWords = 0
		this.brlw.runningLength = 0
		return false
	}

	this.brlw.reset(this.iterator.next().getActualWord())
	this.literalWordStartPosition = this.iterator.literalWords()
	return true
}

func (this *BufferedRunningLengthWordIterator) discharge(container BitmapStorage, max int64) int64 {
	return this.dischargeInternal(container, max, this.writeLiteralWords)
}

func (this *BufferedRunningLengthWordIterator) dischargeNegated(container BitmapStorage, max int64) int64 {
	return this.dischargeInternal(container, max, this.writeNegatedLiteralWords)
}

func (this *BufferedRunningLengthWordIterator) dischargeContainer(container BitmapStorage) {
	this.brlw.literalWordOffset = this.literalWordStartPosition - this.iterator.literalWords()
	this.dischargeIterate(this.brlw, this.iterator, container)
}

func (this *BufferedRunningLengthWordIterator) dischargeIterate(initWord *BufferedRunningLengthWord, iter *EWAHIterator, container BitmapStorage) {
	rlw := initWord

	for {
		runningLength := rlw.getRunningLength()
		container.addStreamOfEmptyWords(rlw.getRunningBit(), runningLength)
		container.addStreamOfLiteralWords(iter.buffer(), iter.literalWords() + rlw.literalWordOffset, rlw.getNumberOfLiteralWords())

		if !iter.hasNext() {
			break
		}

		rlw = newBufferedRunningLengthWord(iter.next().getActualWord())
	}
}

func (this *BufferedRunningLengthWordIterator) dischargeInternal(container BitmapStorage, max int64, write func(int32, BitmapStorage)) int64 {
	index := int64(0)

	for index < max && this.brlw.size() > 0 {
		pl := this.brlw.getRunningLength()
		if index + pl > max {
			pl = max - index
		}

		container.addStreamOfEmptyWords(this.brlw.getRunningBit(), pl)
		index += pl
		pd := this.brlw.getNumberOfLiteralWords()

		if int64(pd) + index > max {
			pd = int32(max - index)
		}

		write(pd, container)
		this.discardFirstWords(pl + int64(pd))
		index += int64(pd)
	}

	return index
}

func (this *BufferedRunningLengthWordIterator) dischargeAsEmpty(container BitmapStorage) {
	for this.brlw.size() > 0 {
		container.addStreamOfEmptyWords(false, this.brlw.size())
		this.discardFirstWords(this.brlw.size())
	}
}

func (this *BufferedRunningLengthWordIterator) dischargeRemaining(container BitmapStorage) {
	this.brlw.literalWordOffset = this.literalWordStartPosition - this.iterator.literalWords()
	runningLengthWord := this.brlw

	for {
		runningLength := runningLengthWord.getRunningLength()
		container.addStreamOfEmptyWords(runningLengthWord.getRunningBit(), runningLength)
		container.addStreamOfLiteralWords(this.iterator.buffer(), this.iterator.literalWords() + runningLengthWord.literalWordOffset, runningLengthWord.getNumberOfLiteralWords())

		if !this.iterator.hasNext() {
			break
		}

		runningLengthWord = newBufferedRunningLengthWord(this.iterator.next().getActualWord())
	}
}

func (this *BufferedRunningLengthWordIterator) writeLiteralWords(numWords int32, container BitmapStorage) {
	container.addStreamOfLiteralWords(this.buffer, this.literalWordStartPosition, numWords)
}

func (this *BufferedRunningLengthWordIterator) writeNegatedLiteralWords(numWords int32, container BitmapStorage) {
	container.addStreamOfNegatedLiteralWords(this.buffer, this.literalWordStartPosition, numWords)
}


