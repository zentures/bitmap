/*
 * Copyright (c) 2013 Zhen, LLC. http://zhen.io. All rights reserved.
 * Use of this source code is governed by the Apache 2.0 license.
 *
 */

package ewah

import (
	"github.com/zhenjl/bitmap"
	"math"
	"fmt"
	"errors"
)

const (
	// wordInBits is the constant representing the number of bits in a int64
	wordInBits int64 = 64

	// defaultBufferSize is a constant default memory allocation when the object is constructed
	defaultBufferSize int64 = 4
)

type Ewah struct {
	// actualSizeInWords is the number of words actually used in the buffer to represent the bitmap
	actualSizeInWords int64

	// sizeInBits is the number of total bits in the bitmap
	sizeInBits int64

	// buffer representing the bitmap
	buffer []int64

	// whether we adjust after some aggregation by adding in zeroes
	adjustContainerSizeWhenAggregating bool

	// getCursor remembers the last search position and try to search from there for the next one
	// It's an optimization for sequential Gets
	getCursor *cursor

	// The current (last) running length word
	//rlw *runningLengthWord
	rlwCursor *cursor
}

var _ bitmap.Bitmap = (*Ewah)(nil)
var _ BitmapStorage = (*Ewah)(nil)

func New() bitmap.Bitmap {
	ewah := new(Ewah)

	ewah.Reset()

	return ewah
}

// Set sets the bit at position i to true (1). The bits must be set in ascending order. For example, set(15)
// then set(7) will fail.
func (this *Ewah) Set(i int64) bitmap.Bitmap {
	// According to @lemire: https://github.com/lemire/javaewah/issues/23#issuecomment-23998948
	// In the current version, the range of allowable values for the set method is [0,Integer.MAX_VALUE - 64].
	// (If you use the 32-bit EWAH, the answer is slightly different [0,Integer.MAX_VALUE - 32].)
	// One concern about supporting very wide ranges is that bitmaps are not appropriate if the data is too sparse.
	// If you want to use a bitmap having few values over a wide range, it is wasted effort.
	// You are better off using a different data structure.
	if i > math.MaxInt32 - wordInBits || i < 0 {
		return nil
	}

	// If i is less than sizeInBits, then we are trying to set a previous bit, which is not allowed
	if i < this.sizeInBits {
		return nil
	}

	// Distance of the bit from the active word in the buffer
	// We want to know this so we can decide whether we need to add some empty words to pad the bitmap,
	// or update the bit in the current word
	dist := (i + wordInBits) / wordInBits - (this.sizeInBits + wordInBits - 1) / wordInBits
	//fmt.Println("ewah.go/Set: dist =", dist, "size =", this.sizeInBits)

	// Set the new size of the bitmap to the latest bit that's set (index is 0-based, thus +1)
	this.sizeInBits = i + 1

	// If the distance is greater than 0, that means we are not acting on the current active word
	if dist > 0 {
		// So we need to add some empty words if the distance is greater than 1
		// Basically adding dist-1 zero words to the bitmap
		if dist > 1 {
			this.fastAddStreamOfEmptyWords(false, dist-1)
		}

		// Once we padded the bitmap with empty words, then we can add a new literal word at the end
		//fmt.Println("ewah.go/Set: before addLiteralWord")
		this.addLiteralWord(int64(1) << uint64((i % wordInBits)))
		//fmt.Println("ewah.go/Set: after addLiteralWord")

		return this
	}

	// Now we know dist == 0 since it can't be < 0 (can't set a bit past the current active bit)
	if this.rlwCursor.rlw.getNumberOfLiteralWords() == 0 {
		this.rlwCursor.rlw.setRunningLength(this.rlwCursor.rlw.getRunningLength() - 1)
		this.addLiteralWord(1 << uint64(i % wordInBits))
		//fmt.Println("ewah.go/Set: after addLiteralWord inside numOfLiteralWords == 0")
		return this
	}

	this.buffer[this.actualSizeInWords - 1] |= 1 << uint64(i % wordInBits)
	if this.buffer[this.actualSizeInWords - 1] == ^0 {
		this.buffer[this.actualSizeInWords - 1] = 0
		this.actualSizeInWords -= 1
		this.rlwCursor.rlw.setNumberOfLiteralWords(int64(this.rlwCursor.rlw.getNumberOfLiteralWords()) - 1)
		this.addEmptyWord(true)
		//fmt.Println("ewah.go/Set: after addEmptyWord")
	}

	return this
}

func (this *Ewah) Get(i int64) bool {
	if i < 0 || i > this.sizeInBits {
		return false
	}

	wordi := i / wordInBits
	biti := uint64(i % wordInBits)

	if wordi <= this.getCursor.checked {
		this.getCursor.checked = 0
		this.getCursor.marker = 0
	}

	// If this is the first time, then cursor should have been initialized to 0, which means both are 0
	// If this is NOT the first time, then cursor should contain the last marker and words prior
	// If the word being checked is before the marker, then we reset to 0 and start again
	wordChecked := this.getCursor.checked
	marker := this.getCursor.marker

	// index to marker word
	m := newRunningLengthWord(this.buffer, marker)

	for wordChecked <= wordi {
		m.reset(this.buffer, marker)
		//fmt.Printf("ewah.go/Get: marker = %064b\n", m.getActualWord())
		numOfLiteralWords := int64(m.getNumberOfLiteralWords())
		runningLength := m.getRunningLength()
		wordChecked += runningLength

		if wordi < wordChecked {
			this.getCursor.marker = marker
			this.getCursor.checked = wordChecked - runningLength
			//fmt.Println("ewah.go/Get: cursor =", this.getCursor, ", i =", i, ", wordi =", wordi)
			return m.getRunningBit()
		}

		if wordi < wordChecked + numOfLiteralWords {
			//fmt.Printf("ewah.go/Get: index = %d\n", marker + (wordi - wordChecked) + 1)
			//fmt.Printf("ewah.go/Get: word = %064b\n", this.buffer[marker + (wordi - wordChecked) + 1])
			//fmt.Printf("ewah.go/Get: bit = %064b\n", this.buffer[marker + (wordi - wordChecked) + 1] & (int64(1) << biti))
			this.getCursor.marker = marker
			this.getCursor.checked = wordChecked - runningLength
			//fmt.Println("ewah.go/Get: cursor =", this.getCursor, ", i =", i, ", wordi =", wordi)
			return this.buffer[marker + (wordi - wordChecked) + 1] & (int64(1) << biti) != 0
		}
		wordChecked += numOfLiteralWords
		marker += numOfLiteralWords + 1
	}

	this.getCursor.marker = marker
	this.getCursor.checked = wordChecked
	//fmt.Println("ewah.go/Get: cursor =", this.getCursor, ", i =", i, ", wordi =", wordi)
	return false
}

func (this *Ewah) Swap(other *Ewah) bitmap.Bitmap {
	this.buffer, other.buffer = other.buffer, this.buffer
	this.rlwCursor, other.rlwCursor = other.rlwCursor, this.rlwCursor
	this.actualSizeInWords, other.actualSizeInWords = other.actualSizeInWords, this.actualSizeInWords
	this.sizeInBits, other.sizeInBits = other.sizeInBits, this.sizeInBits

	return this
}

// Returns the size in bits of the *uncompressed* bitmap represented by this compressed bitmap.
// Initially, the sizeInBits is zero. It is extended automatically when you set bits to true.
func (this *Ewah) Size() int64 {
	return this.sizeInBits
}

// Report the *compressed* size of the bitmap (equivalent to memory usage, after accounting for some overhead).
func (this *Ewah) SizeInBytes() int64 {
	return this.actualSizeInWords * (wordInBits / 8)
}

func (this *Ewah) SizeInWords() int64 {
	return this.actualSizeInWords
}

func (this *Ewah) Clear() {
	this.Reset()
}

func (this *Ewah) Reset() {
	if this.buffer == nil {
		this.buffer = make([]int64, defaultBufferSize)
	} else {
		this.buffer[0] = 0
	}

	/*
	if this.rlw == nil {
		this.rlw = newRunningLengthWord(this.buffer, 0)
	} else {
		this.rlwCursor.rlw.reset(this.buffer, 0)
	}
	*/

	if this.rlwCursor == nil {
		this.rlwCursor = newCursor(this.buffer, 0)
	} else {
		this.rlwCursor.reset(this.buffer, 0)
	}

	if this.getCursor == nil {
		this.getCursor = newCursor(this.buffer, 0)
	} else {
		this.getCursor.reset(this.buffer, 0)
	}

	this.actualSizeInWords = 1
	this.sizeInBits = 0
	this.adjustContainerSizeWhenAggregating = true

}

func (this *Ewah) Clone() bitmap.Bitmap {
	c := New().(*Ewah)
	c.reserve(int32(this.actualSizeInWords))
	copy(c.buffer, this.buffer)
	c.actualSizeInWords = this.actualSizeInWords
	c.sizeInBits = this.sizeInBits
	c.rlwCursor.rlw.reset(c.buffer, this.rlwCursor.rlw.p)

	return c
}

func (this *Ewah) Copy(other bitmap.Bitmap) bitmap.Bitmap {
	o := other.(*Ewah)
	this.buffer = make([]int64, o.SizeInWords())
	copy(this.buffer, o.buffer)
	this.actualSizeInWords = o.SizeInWords()
	this.sizeInBits = o.Size()
	this.rlwCursor.rlw.reset(this.buffer, o.rlwCursor.rlw.p)

	return this
}

func (this *Ewah) Equal(other bitmap.Bitmap) bool {
	if other == nil {
		return false
	}

	o := other.(*Ewah)
	if this.Size() != o.Size() {
		return false
	}

	for p, v := range this.buffer {
		if o.buffer[p] != v {
			return false
		}
	}
	return true
}

func (this *Ewah) Cardinality() int64 {
	counter := int64(0)

	// index to marker word
	marker := int64(0)

	for marker < this.actualSizeInWords {
		localrlw := newRunningLengthWord(this.buffer, marker)

		if localrlw.getRunningBit() {
			counter += wordInBits * localrlw.getRunningLength()
		}

		numOfLiteralWords := int64(localrlw.getNumberOfLiteralWords())

		//fmt.Printf("ewah.go/Cardinality: marker = %064b\n", localrlw.getActualWord())
		for j := int64(1); j <= numOfLiteralWords; j++ {
			//fmt.Println("ewah.go/Cardinality: literawords =", numOfLiteralWords, "marker =", marker, "j =", j)
			counter += int64(popcount_3(uint64(this.buffer[marker + j])))
		}

		marker += numOfLiteralWords + 1
	}

	return counter
}

func (this *Ewah) And(a bitmap.Bitmap) bitmap.Bitmap {
	return this.bitOp(a, this.andToContainer)
}

func (this *Ewah) AndNot(a bitmap.Bitmap) bitmap.Bitmap {
	return this.bitOp(a, this.andNotToContainer)
}

func (this *Ewah) Or(a bitmap.Bitmap) bitmap.Bitmap {
	return this.bitOp(a, this.orToContainer)
}

func (this *Ewah) Xor(a bitmap.Bitmap) bitmap.Bitmap {
	return this.bitOp(a, this.xorToContainer)
}

func (this *Ewah) Not() bitmap.Bitmap {
	marker := int64(0)
	m := newRunningLengthWord(this.buffer, marker)

	for marker < this.actualSizeInWords {
		m.reset(this.buffer, marker)
		numOfLiteralWords := int64(m.getNumberOfLiteralWords())
		m.setRunningBit(!m.getRunningBit())

		for i := int64(1); i <= numOfLiteralWords; i++ {
			this.buffer[marker + i] = ^this.buffer[marker + i]
		}

		// If this is the last word in the bitmap, we may need to do some special treatment since
		// it may not be fully populated.
		if marker+numOfLiteralWords+1 == this.actualSizeInWords {
			// If the last word is fully populated, then no need to do anything
			lastBits := this.sizeInBits % wordInBits
			if lastBits == 0 {
				break
			}

			// If there are no literal words (or all empty words) and the lastBits is not zero, this means
			// we need to make sure we break out the last empty word, and negate the populated portion of
			// the word
			if m.getNumberOfLiteralWords() == 0 {
				if m.getRunningLength() > 0 && m.getRunningBit() {
					m.setNumberOfLiteralWords(int64(m.getNumberOfLiteralWords())-1)
					this.addLiteralWord(int64(uint64(0) >> uint64(wordInBits - lastBits)))
				}

				break
			}

			this.buffer[marker + numOfLiteralWords] &= int64(^uint64(0) >> uint64(wordInBits - lastBits))
			break
		}

		marker += numOfLiteralWords + 1
	}

	return this
}

func (this *Ewah) PrintStats(details bool) {
	fmt.Println("actualSizeInWords =", this.actualSizeInWords, "words,", this.actualSizeInWords*wordInBits, "bits")
	fmt.Println("actualSizeInBits =", this.sizeInBits)
	fmt.Println("cardinality =", this.Cardinality())

	if details {
		this.printDetails()
	}
}

func (this *Ewah) printDetails() {
	//fmt.Println("                           0123456789012345678901234567890123456789012345678901234567890123")
	fmt.Println("                           3210987654321098765432109876543210987654321098765432109876543210")
	for i := int64(0); i < this.actualSizeInWords; i++ {
		fmt.Printf("%4d: %20d %064b\n", i, uint64(this.buffer[i]), uint64(this.buffer[i]))
	}
}

//
// Not-exported functions
//

func (this *Ewah) bitOp(a bitmap.Bitmap, f func(*Ewah, BitmapStorage)) bitmap.Bitmap {
	aEwah, ok := a.(*Ewah)
	if !ok {
		return nil
	}

	container := New().(*Ewah)
	container.reserve(int32(math.Max(float64(this.actualSizeInWords), float64(aEwah.actualSizeInWords))))

	f(aEwah, container)

	return container
}

func (this *Ewah) andToContainer(a *Ewah, container BitmapStorage) {
	// i and j may switch depending on the the bitwise operation
	i, j := a, this

	iCursor := newCursor(i.buffer, i.SizeInWords())
	jCursor := newCursor(j.buffer, j.SizeInWords())

	// Keep going thru the words until one of the cursors have reached the end (checked > size)
	//for iCursor.pointer < iCursor.size && jCursor.pointer < jCursor.size {
	for iCursor.rlwLiteralRemaining + iCursor.rlwEmptyRemaining > 0 && jCursor.rlwLiteralRemaining + jCursor.rlwEmptyRemaining > 0 {
		//fmt.Println("ewah.go/andToContainer2: inside 1st for loop\n--- iCursor =", iCursor, "\n--- jCursor =", jCursor)

		// For each of the marker words, keep moving thru them until both have gone through their
		// empty words
		for iCursor.rlwEmptyRemaining > 0 || jCursor.rlwEmptyRemaining > 0 {
			//fmt.Println("ewah.go/andToContainer2: inside 2nd for loop\n--- iCursor =", iCursor, "\n--- jCursor =", jCursor)


			// Predator is the one that has more empty words. Prey is the one with less.
			var prey, predator *cursor
			if iCursor.rlwEmptyRemaining < jCursor.rlwEmptyRemaining {
				prey, predator = iCursor, jCursor
			} else {
				prey, predator = jCursor, iCursor
			}

			if predator.rlw.getRunningBit() == false {
				// If predator's (one with more empty words) empty words are false, which means all these words
				// are 0, then the result of the AND operation will also be 0. So we insert the same number
				// of 0 words into the result
				container.addStreamOfEmptyWords(false, predator.rlwEmptyRemaining)

				// And we move both prey and predator forward by the same number of words
				//fmt.Printf("ewah.go/andToContainer2: prey.moveForward(%d)\n", predator.rlwEmptyRemaining)
				prey.moveForward(predator.rlwEmptyRemaining)
				//fmt.Printf("ewah.go/andToContainer2: predator.moveForward(%d)\n", predator.rlwEmptyRemaining)
				predator.moveForward(predator.rlwEmptyRemaining)
			} else {
				// If the predator's empty words are true, which means all these words are 1, then the result of
				// the AND operation will be the same as the prey's words. So we will essentially copy the prey's
				// words into the result set, up to the same number as the predator's running length. Prey may
				// not have enough remaining words to cover the full running length, so we need to get back the
				// total number that's been copied over.
				//fmt.Printf("ewah.go/andToContainer2: prey.copyForward(%d)\n", predator.rlwEmptyRemaining)
				index := prey.copyForward(container, predator.rlwEmptyRemaining, false)
				container.addStreamOfEmptyWords(false, predator.rlwEmptyRemaining - index)
				predator.moveForward(predator.rlwEmptyRemaining)
			}
		}

		// Now that we have gone through all the empty words, let's take care of the left over literal words
		leftOverLiterals := int64(math.Min(float64(iCursor.rlwLiteralRemaining), float64(jCursor.rlwLiteralRemaining)))
		//fmt.Printf("ewah.go/andToContainer2: leftOverLiterals = %d, i.rlwLiteralRemaining = %d, j.rlwLiteralRemaining = %d\n", leftOverLiterals, iCursor.rlwLiteralRemaining, jCursor.rlwLiteralRemaining)

		if leftOverLiterals > 0 {
			// for each of the left over literals, we will AND them and put the result in the contanier
			for k := int64(0); k < leftOverLiterals; k++ {
				container.add(iCursor.getLiteralWordAt(k) & jCursor.getLiteralWordAt(k))
			}

			// Move the cursors forward
			//fmt.Printf("ewah.go/andToContainer2: iCursor.moveForward(%d)\n", leftOverLiterals)
			iCursor.moveForward(leftOverLiterals)
			//fmt.Printf("ewah.go/andToContainer2: jCursor.moveForward(%d)\n", leftOverLiterals)
			jCursor.moveForward(leftOverLiterals)
		}

	}
	//fmt.Println("------------------------")
	//fmt.Println("ewah.go/andToContainer2: iCursor =", iCursor)
	//fmt.Println("ewah.go/andToContainer2: jCursor =", jCursor)

	// Adjust the result set size to the bigger of the two original bitmaps if needed, by padding 0's
	if this.adjustContainerSizeWhenAggregating {
		// Only one of the cursors should words left. So we check to see if iCursor has left over words.
		// If iCursor doesn't have anything left (checked >= size), then it must be jCursor that has left overs.
		iRemains := (iCursor.rlwEmptyRemaining + iCursor.rlwLiteralRemaining) != 0
		var remaining *cursor

		if iRemains {
			remaining = iCursor
		} else {
			remaining = jCursor
		}

		// For whatever number of words we have, they should all be 0's since this is an AND operation
		// So we just copy a bunch of 0 empty words over to the result container
		remaining.copyEmptyForward(container)

		// Then set the result container size to the max of the two bitmaps
		//fmt.Printf("ewah.go/andToContainer2: i.size = %d, j.size = %d\n", i.Size(), j.Size())
		container.setSizeInBits(int64(math.Max(float64(i.Size()), float64(j.Size()))))
	}
}

// Returns the cardinality of the result of a bitwise AND of the values of the current bitmap with some
// other bitmap. Avoids needing to allocate an intermediate bitmap to hold the result of the OR.
func (this *Ewah) andCardinality(a *Ewah) int32 {
	counter := newBitCounter()
	this.andToContainer(a, counter)
	return int32(counter.(*bitCounter).getCount())
}

func (this *Ewah) andNotToContainer(a *Ewah, container BitmapStorage) {
	i := NewEWAHIterator(this.buffer, this.actualSizeInWords)
	j := NewEWAHIterator(a.buffer, a.actualSizeInWords)

	rlwi := newBufferedRunningLengthWordIterator(i)
	rlwj := newBufferedRunningLengthWordIterator(j)

	for rlwi.size() > 0 && rlwj.size() > 0 {
		//fmt.Printf("ewah.go/andNotToContainer: rlwi.size = %d, rlwj.size = %d\n", rlwi.size(), rlwj.size())
		for rlwi.getRunningLength() > 0 || rlwj.getRunningLength() > 0 {
			i_is_prey := rlwi.getRunningLength() < rlwj.getRunningLength()
			var prey, predator *BufferedRunningLengthWordIterator

			if i_is_prey {
				prey = rlwi
				predator = rlwj
			} else {
				prey = rlwj
				predator = rlwi
			}

			//fmt.Println("ewah.go/andNotToContainer: i_is_prey =", i_is_prey)

			if (predator.getRunningBit() == true && i_is_prey) || (predator.getRunningBit() == false && !i_is_prey) {
				container.addStreamOfEmptyWords(false, predator.getRunningLength())
				prey.discardFirstWords(predator.getRunningLength())
				predator.discardFirstWords(predator.getRunningLength())
			} else if i_is_prey {
				//fmt.Println("ewah.go/andNotToContainer: predator.getRunningLength =", predator.getRunningLength())
				index := prey.discharge(container, predator.getRunningLength())
				container.addStreamOfEmptyWords(false, predator.getRunningLength() - index)
				//fmt.Println("ewah.go/andNotToContainer: i_is_prey index =", index)
				predator.discardFirstWords(predator.getRunningLength())
			} else {
				index := prey.dischargeNegated(container, predator.getRunningLength())
				container.addStreamOfEmptyWords(true, predator.getRunningLength() - index)
				predator.discardFirstWords(predator.getRunningLength())
			}
			//fmt.Println("----")
		}

		nbre_literal := int64(math.Min(float64(rlwi.getNumberOfLiteralWords()), float64(rlwj.getNumberOfLiteralWords())))
		if nbre_literal > 0 {
			for k := int32(0); k < int32(nbre_literal); k++ {
				//fmt.Printf("ewah.go/andNotToContainer: i = %064b\n", rlwi.getLiteralWordAt(k))
				//fmt.Printf("ewah.go/andNotToContainer: j = %064b\n", rlwj.getLiteralWordAt(k))
				//fmt.Printf("ewah.go/andNotToContainer:^j = %064b\n", uint64(^rlwj.getLiteralWordAt(k)))
				container.add(rlwi.getLiteralWordAt(k) &^ rlwj.getLiteralWordAt(k))
			}

			rlwi.discardFirstWords(nbre_literal)
			rlwj.discardFirstWords(nbre_literal)
		}
	}

	i_remains := rlwi.size() > 0
	var remaining *BufferedRunningLengthWordIterator

	if i_remains {
		remaining = rlwi
	} else {
		remaining = rlwj
	}

	if i_remains {
		remaining.dischargeContainer(container)
	} else if this.adjustContainerSizeWhenAggregating {
		remaining.dischargeAsEmpty(container)
	}

	if this.adjustContainerSizeWhenAggregating {
		container.setSizeInBits(int64(math.Max(float64(this.sizeInBits), float64(a.sizeInBits))))
	}
}

func (this *Ewah) andNotCardinality(a *Ewah) int32 {
	counter := newBitCounter()
	this.andNotToContainer(a, counter)
	return int32(counter.(*bitCounter).getCount())
}

func (this *Ewah) orToContainer(a *Ewah, container BitmapStorage) {
	i := NewEWAHIterator(a.buffer, a.actualSizeInWords)
	j := NewEWAHIterator(this.buffer, this.actualSizeInWords)

	rlwi := newBufferedRunningLengthWordIterator(i)
	rlwj := newBufferedRunningLengthWordIterator(j)

	for rlwi.size() > 0 && rlwj.size() > 0 {
		for rlwi.getRunningLength() > 0 || rlwj.getRunningLength() > 0 {
			i_is_prey := rlwi.getRunningLength() < rlwj.getRunningLength()
			var prey, predator *BufferedRunningLengthWordIterator

			if i_is_prey {
				prey = rlwi
				predator = rlwj
			} else {
				prey = rlwj
				predator = rlwi
			}

			if predator.getRunningBit() == true {
				container.addStreamOfEmptyWords(true, predator.getRunningLength())
				prey.discardFirstWords(predator.getRunningLength())
				predator.discardFirstWords(predator.getRunningLength())
			} else {
				index := prey.discharge(container, predator.getRunningLength())
				container.addStreamOfEmptyWords(false, predator.getRunningLength() - index)
				predator.discardFirstWords(predator.getRunningLength())
			}
		}

		nbre_literal := int64(math.Min(float64(rlwi.getNumberOfLiteralWords()), float64(rlwj.getNumberOfLiteralWords())))
		if nbre_literal > 0 {
			for k := int32(0); k < int32(nbre_literal); k++ {
				container.add(rlwi.getLiteralWordAt(k) | rlwj.getLiteralWordAt(k))
			}

			rlwi.discardFirstWords(nbre_literal)
			rlwj.discardFirstWords(nbre_literal)
		}
	}

	i_remains := rlwi.size() > 0
	var remaining *BufferedRunningLengthWordIterator

	if i_remains {
		remaining = rlwi
	} else {
		remaining = rlwj
	}

	remaining.dischargeContainer(container)
	container.setSizeInBits(int64(math.Max(float64(this.sizeInBits), float64(a.sizeInBits))))
}

func (this *Ewah) orCardinality(a *Ewah) int32 {
	counter := newBitCounter()
	this.orToContainer(a, counter)
	return int32(counter.(*bitCounter).getCount())
}

func (this *Ewah) xorToContainer(a *Ewah, container BitmapStorage) {
	i := NewEWAHIterator(a.buffer, a.actualSizeInWords)
	j := NewEWAHIterator(this.buffer, this.actualSizeInWords)

	rlwi := newBufferedRunningLengthWordIterator(i)
	rlwj := newBufferedRunningLengthWordIterator(j)

	for rlwi.size() > 0 && rlwj.size() > 0 {
		for rlwi.getRunningLength() > 0 || rlwj.getRunningLength() > 0 {
			i_is_prey := rlwi.getRunningLength() < rlwj.getRunningLength()
			var prey, predator *BufferedRunningLengthWordIterator

			if i_is_prey {
				prey = rlwi
				predator = rlwj
			} else {
				prey = rlwj
				predator = rlwi
			}

			if predator.getRunningBit() == false {
				index := prey.discharge(container, predator.getRunningLength())
				container.addStreamOfEmptyWords(false, predator.getRunningLength() - index)
				predator.discardFirstWords(predator.getRunningLength())
			} else {
				index := prey.dischargeNegated(container, predator.getRunningLength())
				container.addStreamOfEmptyWords(true, predator.getRunningLength() - index)
				predator.discardFirstWords(predator.getRunningLength())
			}
		}

		nbre_literal := int64(math.Min(float64(rlwi.getNumberOfLiteralWords()), float64(rlwj.getNumberOfLiteralWords())))
		if nbre_literal > 0 {
			for k := int32(0); k < int32(nbre_literal); k++ {
				container.add(rlwi.getLiteralWordAt(k) ^ rlwj.getLiteralWordAt(k))
			}

			rlwi.discardFirstWords(nbre_literal)
			rlwj.discardFirstWords(nbre_literal)
		}
	}

	i_remains := rlwi.size() > 0
	var remaining *BufferedRunningLengthWordIterator

	if i_remains {
		remaining = rlwi
	} else {
		remaining = rlwj
	}

	remaining.dischargeContainer(container)
	container.setSizeInBits(int64(math.Max(float64(this.sizeInBits), float64(a.sizeInBits))))
}

func (this *Ewah) xorCardinality(a *Ewah) int32 {
	counter := newBitCounter()
	this.xorToContainer(a, counter)
	return int32(counter.(*bitCounter).getCount())
}

// add is used to add words directly to the bitmap.
func (this *Ewah) add(newdata int64) {
	this.addSignificantBits(newdata, wordInBits)
}

// addWithSize adds words directly to the bitmap, but with the number of significant bits specified.
func (this *Ewah) addSignificantBits(newdata int64, bitsthatmatter int64) {
	//fmt.Printf("ewah.go/addSignificantBits:    %064b\n----\n", newdata)
	this.sizeInBits += bitsthatmatter
	if newdata == 0 {
		this.addEmptyWord(false)
	} else if newdata == ^1 {
		this.addEmptyWord(true)
	} else {
		this.addLiteralWord(newdata)
	}
}

// addEmptyWord adds an empty word of 1's or 0's to the bitmap. true: newdata==0; false: newdata== ~0
func (this *Ewah) addEmptyWord(v bool) {
	noLiteralWord := this.rlwCursor.rlw.getNumberOfLiteralWords() == 0
	runlen := this.rlwCursor.rlw.getRunningLength()

	if noLiteralWord && runlen == 0 {
		this.rlwCursor.rlw.setRunningBit(v)
	}

	if noLiteralWord && this.rlwCursor.rlw.getRunningBit() == v && runlen < LargestRunningLengthCount {
		this.rlwCursor.rlw.setRunningLength(runlen+1)
		return
	}

	this.pushBack(0)
	this.rlwCursor.rlw.reset(this.buffer, this.actualSizeInWords-1)
	this.rlwCursor.rlw.setRunningBit(v)
	this.rlwCursor.rlw.setRunningLength(1)
}

// addLiteralWord adds a literal word to the bitmap.
func (this *Ewah) addLiteralWord(newdata int64) {
	//fmt.Printf("ewah.go/addLiteralWord: newdata = %064b\n", newdata)
	numberSoFar := this.rlwCursor.rlw.getNumberOfLiteralWords()
	//fmt.Printf("ewah.go/addLiteralWord: numberSoFar = %d\n", numberSoFar)
	if numberSoFar >= LargestLiteralCount {
		this.pushBack(0)
		this.rlwCursor.rlw.reset(this.buffer, this.actualSizeInWords-1)
		this.rlwCursor.rlw.setNumberOfLiteralWords(1)
		this.pushBack(newdata)
	}
	this.rlwCursor.rlw.setNumberOfLiteralWords(int64(numberSoFar+1))
	//fmt.Printf("ewah.go/addLiteralWord: getNumberOfLiteralWords = %d\n", this.rlwCursor.rlw.getNumberOfLiteralWords())
	this.pushBack(newdata)
}

// addStreamOfLiteralWords adds several literal words at a time, might be faster
func (this *Ewah) addStreamOfLiteralWords(data []int64, start, number int32) {
	leftOverNumber := number

	for leftOverNumber > 0 {
		numberOfLiteralWords := this.rlwCursor.rlw.getNumberOfLiteralWords()
		whatWeCanAdd := int32(math.Min(float64(number), float64(LargestLiteralCount - numberOfLiteralWords)))

		this.rlwCursor.rlw.setNumberOfLiteralWords(int64(numberOfLiteralWords + whatWeCanAdd))
		leftOverNumber -= whatWeCanAdd
		this.pushBackMultiple(data, start, whatWeCanAdd)
		this.sizeInBits += int64(whatWeCanAdd) * wordInBits

		if leftOverNumber > 0 {
			this.pushBack(0)
			//this.rlwCursor.rlw.position = this.actualSizeInWords - 1
			this.rlwCursor.rlw.reset(this.buffer, this.actualSizeInWords-1)
		}
	}

	//fmt.Printf("ewah.go/addStreamOfLiteralWords: sizeinbits = %d\n", this.sizeInBits)
}

// addStreamOfEmptyWords adds several empty words at a time, might be faster
func (this *Ewah) addStreamOfEmptyWords(v bool, number int64) {
	if number == 0 {
		return
	}

	this.sizeInBits += number * wordInBits

	if this.rlwCursor.rlw.getRunningBit() != v && this.rlwCursor.rlw.size() == 0 {
		this.rlwCursor.rlw.setRunningBit(v)
	} else if this.rlwCursor.rlw.getNumberOfLiteralWords() != 0 || this.rlwCursor.rlw.getRunningBit() != v {
		this.pushBack(0)
		this.rlwCursor.rlw.reset(this.buffer, this.actualSizeInWords-1)
		if v {
			this.rlwCursor.rlw.setRunningBit(v)
		}
	}

	runlen := this.rlwCursor.rlw.getRunningLength()
	whatWeCanAdd := int64(math.Min(float64(number), float64(int64(LargestLiteralCount) - runlen)))

	this.rlwCursor.rlw.setRunningLength(runlen + whatWeCanAdd)
	number -= whatWeCanAdd

	for number >= LargestRunningLengthCount {
		this.pushBack(0)
		this.rlwCursor.rlw.reset(this.buffer, this.actualSizeInWords-1)
		if v {
			this.rlwCursor.rlw.setRunningBit(v)
		}

		this.rlwCursor.rlw.setRunningLength(LargestRunningLengthCount)
		number -= LargestRunningLengthCount
	}

	if number > 0 {
		this.pushBack(0)
		this.rlwCursor.rlw.reset(this.buffer, this.actualSizeInWords-1)
		if v {
			this.rlwCursor.rlw.setRunningBit(v)
		}
		this.rlwCursor.rlw.setRunningLength(number)
	}

	//fmt.Printf("ewah.go/addStreamOfEmptyWords: sizeinbits = %d\n", this.sizeInBits)
}

// fastAddStreamOfEmptyWords adds many zeroes and ones faster. This does not update sizeInBits
func (this *Ewah) fastAddStreamOfEmptyWords(v bool, number int64) {
	if this.rlwCursor.rlw.getRunningBit() != v && this.rlwCursor.rlw.size() == 0 {
		this.rlwCursor.rlw.setRunningBit(v)
	} else if this.rlwCursor.rlw.getNumberOfLiteralWords() != 0 || this.rlwCursor.rlw.getRunningBit() != v {
		this.pushBack(0)
		//this.rlwCursor.rlw.position = this.actualSizeInWords - 1
		this.rlwCursor.rlw.reset(this.buffer, this.actualSizeInWords-1)
		if v {
			this.rlwCursor.rlw.setRunningBit(v)
		}
	}

	runlen := this.rlwCursor.rlw.getRunningLength()
	whatWeCanAdd := int64(math.Min(float64(number), float64(int64(LargestLiteralCount) - runlen)))

	this.rlwCursor.rlw.setRunningLength(runlen + whatWeCanAdd)
	number -= whatWeCanAdd

	for number >= LargestRunningLengthCount {
		this.pushBack(0)
		//this.rlwCursor.rlw.position = this.actualSizeInWords - 1
		this.rlwCursor.rlw.reset(this.buffer, this.actualSizeInWords-1)
		if v {
			this.rlwCursor.rlw.setRunningBit(v)
		}

		this.rlwCursor.rlw.setRunningLength(LargestRunningLengthCount)
		number -= LargestRunningLengthCount
	}

	if number > 0 {
		this.pushBack(0)
		//this.rlwCursor.rlw.position = this.actualSizeInWords - 1
		this.rlwCursor.rlw.reset(this.buffer, this.actualSizeInWords-1)
		if v {
			this.rlwCursor.rlw.setRunningBit(v)
		}

		this.rlwCursor.rlw.setRunningLength(number)
	}
}

// addStreamOfNegatedLiteralWords is similar to addStreamOfLiteralWords except the words are negated
func (this *Ewah) addStreamOfNegatedLiteralWords(data []int64, start, number int32) {
	leftOverNumber := number
	for leftOverNumber > 0 {
		numberOfLiteralWords := this.rlwCursor.rlw.getNumberOfLiteralWords()
		whatWeCanAdd := int32(math.Min(float64(number), float64(LargestLiteralCount - numberOfLiteralWords)))

		this.rlwCursor.rlw.setNumberOfLiteralWords(int64(numberOfLiteralWords + whatWeCanAdd))
		leftOverNumber -= whatWeCanAdd
		this.negativePushBack(data, start, whatWeCanAdd)
		this.sizeInBits += int64(whatWeCanAdd) * wordInBits

		if leftOverNumber > 0 {
			this.pushBack(0)
			//this.rlwCursor.rlw.position = this.actualSizeInWords - 1
			this.rlwCursor.rlw.reset(this.buffer, this.actualSizeInWords-1)
		}
	}
}

func (this *Ewah) negativePushBack(data []int64, start, number int32) {
	negativeData := make([]int64, number)

	for i := int32(0); i < number; i++ {
		negativeData[i] = ^data[start + i]
	}

	this.pushBackMultiple(negativeData, 0, number)
}

// pushBack adds an element at the end
//
// This is a convenience method that calls push_back_multiple
func (this *Ewah) pushBack(data int64) {
	this.pushBackMultiple([]int64{data}, 0, 1)
}

// pushBack adds multiple element at the end
//
// This is the C++ vector pushBack description. Adds a new element at the end of the vector, after its
// current last element. The content of val is copied (or moved) to the new element.
//
// This effectively increases the container size by one, which causes an automatic reallocation of the
// allocated storage space if -and only if- the new vector size surpasses the current vector capacity.
func (this *Ewah) pushBackMultiple(data []int64, start, number int32) {
	// If the size of the bitmap is the same as the buffer length, that means the buffer is full, so we need
	// to allocate
	bufferCap := int32(cap(this.buffer))
	if this.actualSizeInWords == int64(bufferCap) {
		var newSize int32
		if bufferCap + number < 32768 {
			newSize = (bufferCap + number) * 2
		} else if (bufferCap + number) * 3 / 2 < (bufferCap + number) {
			// overflow
			newSize = math.MaxInt32
		} else {
			newSize = (bufferCap + number) * 3 / 2
		}
		oldBuffer := this.buffer
		this.buffer = make([]int64, newSize)
		copy(this.buffer, oldBuffer)
		this.rlwCursor.rlw.reset(this.buffer, this.rlwCursor.rlw.p)
		//this.rlwCursor.rlw.array = this.buffer
	}
	copy(this.buffer[this.actualSizeInWords:], data[start:start+number])
	this.actualSizeInWords += int64(number)
}

func (this *Ewah) setSizeInBits(size int64) error {
	if (size+wordInBits-1)/wordInBits != (this.sizeInBits+wordInBits-1)/wordInBits {
		return errors.New("ewah/setSizeInBits: You can only reduce the size of teh bitmap within the scope of the last word. To extend the bitmap, please call setSizeInBitsWithDefault(int32)")
	}

	this.sizeInBits = size
	//fmt.Println("ewah.go/setSizeInBits: size =", this.sizeInBits)
	return nil
}

// setSizeInBitsWithDefault changes the reported size in bits of the *uncompressed* bitmap represented
// by this compressed bitmap. It may change the underlying compressedb bitmap. It is not possible to reduce
// the sizeInBits, but it can be extended. The new bits are set to false or true depending on the
// value of the defaultValue
func (this *Ewah) setSizeInBitsWithDefault(size int64, defaultValue bool) bool {
	if size < this.sizeInBits {
		return false
	}

	if ! defaultValue {
		this.extendEmptyBits(this, this.sizeInBits, size)
	} else {
		for this.sizeInBits % wordInBits != 0 && this.sizeInBits < size {
			this.Set(this.sizeInBits)
		}

		this.addStreamOfEmptyWords(defaultValue, (size / wordInBits) - this.sizeInBits / wordInBits)

		for this.sizeInBits < size {
			this.Set(this.sizeInBits)
		}
	}

	this.sizeInBits = size
	return true

}

func (this *Ewah) toArray() []int {
	return nil
}

func (this *Ewah) extendEmptyBits(storage *Ewah, currentSize, newSize int64) {

}

func (this *Ewah) reserve(size int32) bitmap.Bitmap {
	if size > int32(len(this.buffer))	 {
		oldBuffer := this.buffer
		this.buffer = make([]int64, size)
		copy(this.buffer, oldBuffer)
		this.rlwCursor.reset(this.buffer, 0)
	}

	return this
}
