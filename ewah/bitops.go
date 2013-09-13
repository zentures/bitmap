/*
 * Copyright (c) 2013 Zhen, LLC. http://zhen.io. All rights reserved.
 * Use of this source code is governed by the Apache 2.0 license.
 *
 */

package ewah

import (
	"github.com/zhenjl/bitmap"
	"math"
)

func (this *Ewah) bitOp(f func(*Ewah, BitmapStorage), a ...bitmap.Bitmap) bitmap.Bitmap {
	n := len(a)
	b, ok := a[0].(*Ewah)
	if !ok {
		return nil
	}

	ans := New().(*Ewah)
	tmp := New().(*Ewah)
	ans.reserve(int32(math.Max(float64(this.actualSizeInWords), float64(b.actualSizeInWords))))
	tmp.reserve(int32(math.Max(float64(this.actualSizeInWords), float64(b.actualSizeInWords))))

	this.andToContainer(b, ans)

	for i := 1; i < n; i++ {
		b, ok := a[i].(*Ewah)
		if !ok {
			return nil
		}

		ans.andToContainer(b, tmp)
		tmp.Swap(ans)
		tmp.Reset()
	}

	return ans
}

func (this *Ewah) And(a ...bitmap.Bitmap) bitmap.Bitmap {
	n := len(a)
	b, ok := a[0].(*Ewah)
	if !ok {
		return nil
	}

	ans := New().(*Ewah)
	tmp := New().(*Ewah)
	ans.reserve(int32(math.Max(float64(this.actualSizeInWords), float64(b.actualSizeInWords))))
	tmp.reserve(int32(math.Max(float64(this.actualSizeInWords), float64(b.actualSizeInWords))))

	this.andToContainer(b, ans)

	for i := 1; i < n; i++ {
		b, ok := a[i].(*Ewah)
		if !ok {
			return nil
		}

		ans.andToContainer(b, tmp)
		tmp.Swap(ans)
		tmp.Reset()
	}

	return ans
}

func (this *Ewah) AndNot(a ...bitmap.Bitmap) bitmap.Bitmap {
	n := len(a)
	b, ok := a[0].(*Ewah)
	if !ok {
		return nil
	}

	ans := New().(*Ewah)
	tmp := New().(*Ewah)
	ans.reserve(int32(math.Max(float64(this.actualSizeInWords), float64(b.actualSizeInWords))))
	tmp.reserve(int32(math.Max(float64(this.actualSizeInWords), float64(b.actualSizeInWords))))

	this.andNotToContainer(b, ans)

	for i := 1; i < n; i++ {
		b, ok := a[i].(*Ewah)
		if !ok {
			return nil
		}

		ans.andNotToContainer(b, tmp)
		tmp.Swap(ans)
		tmp.Reset()
	}

	return ans
}

func (this *Ewah) Or(a ...bitmap.Bitmap) bitmap.Bitmap {
	n := len(a)
	b, ok := a[0].(*Ewah)
	if !ok {
		return nil
	}

	ans := New().(*Ewah)
	tmp := New().(*Ewah)
	ans.reserve(int32(math.Max(float64(this.actualSizeInWords), float64(b.actualSizeInWords))))
	tmp.reserve(int32(math.Max(float64(this.actualSizeInWords), float64(b.actualSizeInWords))))

	this.orToContainer(b, ans)

	for i := 1; i < n; i++ {
		b, ok := a[i].(*Ewah)
		if !ok {
			return nil
		}

		ans.orToContainer(b, tmp)
		tmp.Swap(ans)
		tmp.Reset()
	}

	return ans
}

func (this *Ewah) Xor(a ...bitmap.Bitmap) bitmap.Bitmap {
	n := len(a)
	b, ok := a[0].(*Ewah)
	if !ok {
		return nil
	}

	ans := New().(*Ewah)
	tmp := New().(*Ewah)
	ans.reserve(int32(math.Max(float64(this.actualSizeInWords), float64(b.actualSizeInWords))))
	tmp.reserve(int32(math.Max(float64(this.actualSizeInWords), float64(b.actualSizeInWords))))

	this.xorToContainer(b, ans)

	for i := 1; i < n; i++ {
		b, ok := a[i].(*Ewah)
		if !ok {
			return nil
		}

		ans.xorToContainer(b, tmp)
		tmp.Swap(ans)
		tmp.Reset()
	}

	return ans
}

func (this *Ewah) Not() bitmap.Bitmap {
	c := newCursor(this.buffer, this.actualSizeInWords)

	for !c.end() {
		//fmt.Printf("bitops.go/Not2: c.marker = %d, sizeInWords = %d, literals = %d\n",
		// c.marker, this.actualSizeInWords, c.rlwLiteralRemaining)
		c.setRunningBit(!c.getRunningBit())

		for i := int64(1); i <= c.rlwLiteralRemaining; i++ {
			//fmt.Printf("bitops.go/Not2: i = %d, buffer before = %064b\n", i, uint64(this.buffer[c.marker+i]))
			this.buffer[c.marker + i] = ^this.buffer[c.marker + i]
			//fmt.Printf("bitops.go/Not2: buffer  after = %064b\n", uint64(this.buffer[c.marker+i]))
		}

		// If this is the last word in the bitmap, we may need to do some special treatment since
		// it may not be fully populated.
		if c.marker+c.rlwLiteralRemaining+1 == this.actualSizeInWords {
			// If the last word is fully populated, then no need to do anything
			lastBits := this.sizeInBits % wordInBits
			if lastBits == 0 {
				break
			}

			// If there are no literal words (or all empty words) and the lastBits is not zero, this means
			// we need to make sure we break out the last empty word, and negate the populated portion of
			// the word
			if c.getNumberOfLiteralWords() == 0 {
				if c.getRunningLength() > 0 && c.getRunningBit() {
					c.setNumberOfLiteralWords(int64(c.getNumberOfLiteralWords())-1)
					this.addLiteralWord(int64(uint64(0) >> uint64(wordInBits - lastBits)))
				}

				break
			}

			this.buffer[c.marker + c.rlwLiteralRemaining] &= int64(^uint64(0) >> uint64(wordInBits - lastBits))
			break
		}

		if c.nextMarker() != nil {
			break
		}
	}

	return this
}

func (this *Ewah) andToContainer(a *Ewah, container BitmapStorage) {
	// i and j may switch depending on the the bitwise operation
	i, j := a, this

	iCursor := newCursor(i.buffer, i.SizeInWords())
	jCursor := newCursor(j.buffer, j.SizeInWords())

	// Keep going thru the words until one of the cursors have reached the end (checked > size)
	for iCursor.rlwRemaining() > 0 && jCursor.rlwRemaining() > 0 {
		//fmt.Println("bitops.go/andToContainer2: inside 1st for loop\n--- iCursor =", iCursor, "\n--- jCursor =", jCursor)

		// For each of the marker words, keep moving thru them until both have gone through their empty words
		for iCursor.rlwEmptyRemaining > 0 || jCursor.rlwEmptyRemaining > 0 {
			//fmt.Println("bitops.go/andToContainer2: inside 2nd for loop\n--- iCursor =", iCursor, "\n--- jCursor =", jCursor)


			// Predator is the one that has more empty words. Prey is the one with less.
			var prey, predator *cursor
			if iCursor.rlwEmptyRemaining < jCursor.rlwEmptyRemaining {
				prey, predator = iCursor, jCursor
			} else {
				prey, predator = jCursor, iCursor
			}

			if predator.getRunningBit() == false {
				// If predator's (one with more empty words) empty words are false, which means all these words
				// are 0, then the result of the AND operation will also be 0. So we insert the same number
				// of 0 words into the result
				container.addStreamOfEmptyWords(false, predator.rlwEmptyRemaining)

				// And we move both prey and predator forward by the same number of words
				//fmt.Printf("bitops.go/andToContainer2: prey.moveForward(%d)\n", predator.rlwEmptyRemaining)
				prey.moveForward(predator.rlwEmptyRemaining)
				//fmt.Printf("bitops.go/andToContainer2: predator.moveForward(%d)\n", predator.rlwEmptyRemaining)
				predator.moveForward(predator.rlwEmptyRemaining)
			} else {
				// If the predator's empty words are true, which means all these words are 1, then the result of
				// the AND operation will be the same as the prey's words. So we will essentially copy the prey's
				// words into the result set, up to the same number as the predator's running length. Prey may
				// not have enough remaining words to cover the full running length, so we need to get back the
				// total number that's been copied over.
				//fmt.Printf("bitops.go/andToContainer2: prey.copyForward(%d)\n", predator.rlwEmptyRemaining)
				index := prey.copyForward(container, predator.rlwEmptyRemaining, false)
				container.addStreamOfEmptyWords(false, predator.rlwEmptyRemaining - index)
				predator.moveForward(predator.rlwEmptyRemaining)
			}
		}

		// Now that we have gone through all the empty words, let's take care of the left over literal words
		leftOverLiterals := int64(math.Min(float64(iCursor.rlwLiteralRemaining), float64(jCursor.rlwLiteralRemaining)))
		//fmt.Printf("bitops.go/andToContainer2: leftOverLiterals = %d, i.rlwLiteralRemaining = %d, j.rlwLiteralRemaining = %d\n",
		//	leftOverLiterals, iCursor.rlwLiteralRemaining, jCursor.rlwLiteralRemaining)

		if leftOverLiterals > 0 {
			// for each of the left over literals, we will AND them and put the result in the contanier
			for k := int64(0); k < leftOverLiterals; k++ {
				container.add(iCursor.getLiteralWordAt(k) & jCursor.getLiteralWordAt(k))
			}

			// Move the cursors forward
			//fmt.Printf("bitops.go/andToContainer2: iCursor.moveForward(%d)\n", leftOverLiterals)
			iCursor.moveForward(leftOverLiterals)
			//fmt.Printf("bitops.go/andToContainer2: jCursor.moveForward(%d)\n", leftOverLiterals)
			jCursor.moveForward(leftOverLiterals)
		}

	}
	//fmt.Println("------------------------")
	//fmt.Println("bitops.go/andToContainer2: iCursor =", iCursor)
	//fmt.Println("bitops.go/andToContainer2: jCursor =", jCursor)

	// Adjust the result set size to the bigger of the two original bitmaps if needed, by padding 0's
	if this.adjustContainerSizeWhenAggregating {
		// Only one of the cursors should words left. So we check to see if iCursor has left over words.
		// If iCursor doesn't have anything left (checked >= size), then it must be jCursor that has left overs.
		iRemains := iCursor.rlwRemaining() > 0
		var remaining *cursor

		if iRemains {
			remaining = iCursor
		} else {
			remaining = jCursor
		}

		// For whatever number of words we have, they should all be 0's since this is an AND operation
		// So we just copy a bunch of 0 empty words over to the result container
		remaining.copyForwardEmpty(container)

		// Then set the result container size to the max of the two bitmaps
		//fmt.Printf("bitops.go/andToContainer2: i.size = %d, j.size = %d\n", i.Size(), j.Size())
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
	// i and j may switch depending on the the bitwise operation
	i, j := this, a

	iCursor := newCursor(i.buffer, i.SizeInWords())
	jCursor := newCursor(j.buffer, j.SizeInWords())

	// Keep going thru the words until one of the cursors have reached the end (checked > size)
	for iCursor.rlwRemaining() > 0 && jCursor.rlwRemaining() > 0 {

		// For each of the marker words, keep moving thru them until both have gone through their empty words
		for iCursor.rlwEmptyRemaining > 0 || jCursor.rlwEmptyRemaining > 0 {

			// Predator is the one that has more empty words. Prey is the one with less.
			var prey, predator *cursor
			i_is_prey := iCursor.rlwEmptyRemaining < jCursor.rlwEmptyRemaining
			if i_is_prey {
				prey, predator = iCursor, jCursor
			} else {
				prey, predator = jCursor, iCursor
			}

			if (predator.getRunningBit() == true && i_is_prey) || (predator.getRunningBit() == false && !i_is_prey) {
				container.addStreamOfEmptyWords(false, predator.rlwEmptyRemaining)
				prey.moveForward(predator.rlwEmptyRemaining)
				predator.moveForward(predator.rlwEmptyRemaining)
			} else if i_is_prey {
				index := prey.copyForward(container, predator.rlwEmptyRemaining, false)
				container.addStreamOfEmptyWords(false, predator.rlwEmptyRemaining - index)
				predator.moveForward(predator.rlwEmptyRemaining)
			} else {
				index := prey.copyForward(container, predator.rlwEmptyRemaining, true)
				container.addStreamOfEmptyWords(true, predator.rlwEmptyRemaining - index)
				predator.moveForward(predator.rlwEmptyRemaining)
			}
		}

		leftOverLiterals := int64(math.Min(float64(iCursor.rlwLiteralRemaining), float64(jCursor.rlwLiteralRemaining)))

		if leftOverLiterals > 0 {
			for k := int64(0); k < leftOverLiterals; k++ {
				container.add(iCursor.getLiteralWordAt(k) &^ jCursor.getLiteralWordAt(k))
			}

			iCursor.moveForward(leftOverLiterals)
			jCursor.moveForward(leftOverLiterals)
		}
	}


	iRemains := iCursor.rlwRemaining() > 0
	var remaining *cursor

	if iRemains {
		remaining = iCursor
	} else {
		remaining = jCursor
	}

	if iRemains {
		remaining.copyForwardRemaining(container)
	} else if this.adjustContainerSizeWhenAggregating {
		remaining.copyForwardEmpty(container)
	}

	if this.adjustContainerSizeWhenAggregating {
		container.setSizeInBits(int64(math.Max(float64(i.Size()), float64(j.Size()))))
	}
}

func (this *Ewah) andNotCardinality(a *Ewah) int32 {
	counter := newBitCounter()
	this.andNotToContainer(a, counter)
	return int32(counter.(*bitCounter).getCount())
}

func (this *Ewah) orToContainer(a *Ewah, container BitmapStorage) {
	i, j := a, this

	iCursor := newCursor(i.buffer, i.SizeInWords())
	jCursor := newCursor(j.buffer, j.SizeInWords())

	// Keep going thru the words until one of the cursors have reached the end (checked > size)
	for iCursor.rlwRemaining() > 0 && jCursor.rlwRemaining() > 0 {
		//fmt.Println("bitops.go/orToContainer: i =", iCursor)
		//fmt.Println("bitops.go/orToContainer: j =", jCursor)
		// For each of the marker words, keep moving thru them until both have gone through their empty words
		for iCursor.rlwEmptyRemaining > 0 || jCursor.rlwEmptyRemaining > 0 {
			// Predator is the one that has more empty words. Prey is the one with less.
			var prey, predator *cursor
			if iCursor.rlwEmptyRemaining < jCursor.rlwEmptyRemaining {
				prey, predator = iCursor, jCursor
			} else {
				prey, predator = jCursor, iCursor
			}

			if predator.getRunningBit() == true {
				container.addStreamOfEmptyWords(true, predator.rlwEmptyRemaining)
				prey.moveForward(predator.rlwEmptyRemaining)
				predator.moveForward(predator.rlwEmptyRemaining)
			} else {
				index := prey.copyForward(container, predator.rlwEmptyRemaining, false)
				container.addStreamOfEmptyWords(false, predator.rlwEmptyRemaining - index)
				predator.moveForward(predator.rlwEmptyRemaining)
			}
		}

		// Now that we have gone through all the empty words, let's take care of the left over literal words
		leftOverLiterals := int64(math.Min(float64(iCursor.rlwLiteralRemaining), float64(jCursor.rlwLiteralRemaining)))

		if leftOverLiterals > 0 {
			for k := int64(0); k < leftOverLiterals; k++ {
				container.add(iCursor.getLiteralWordAt(k) | jCursor.getLiteralWordAt(k))
			}

			// Move the cursors forward
			iCursor.moveForward(leftOverLiterals)
			jCursor.moveForward(leftOverLiterals)
		}
	}

	// Adjust the result set size to the bigger of the two original bitmaps if needed, by padding 0's
	if this.adjustContainerSizeWhenAggregating {
		// Only one of the cursors should words left. So we check to see if iCursor has left over words.
		// If iCursor doesn't have anything left (checked >= size), then it must be jCursor that has left overs.
		iRemains := iCursor.rlwRemaining() > 0
		var remaining *cursor

		if iRemains {
			remaining = iCursor
		} else {
			remaining = jCursor
		}

		remaining.copyForwardRemaining(container)
		container.setSizeInBits(int64(math.Max(float64(i.Size()), float64(j.Size()))))
	}
}

func (this *Ewah) orCardinality(a *Ewah) int32 {
	counter := newBitCounter()
	this.orToContainer(a, counter)
	return int32(counter.(*bitCounter).getCount())
}

func (this *Ewah) xorToContainer(a *Ewah, container BitmapStorage) {
	i, j := a, this

	iCursor := newCursor(i.buffer, i.SizeInWords())
	jCursor := newCursor(j.buffer, j.SizeInWords())

	// Keep going thru the words until one of the cursors have reached the end (checked > size)
	for iCursor.rlwRemaining() > 0 && jCursor.rlwRemaining() > 0 {
		//fmt.Println("bitops.go/orToContainer: i =", iCursor)
		//fmt.Println("bitops.go/orToContainer: j =", jCursor)
		// For each of the marker words, keep moving thru them until both have gone through their empty words
		for iCursor.rlwEmptyRemaining > 0 || jCursor.rlwEmptyRemaining > 0 {
			// Predator is the one that has more empty words. Prey is the one with less.
			var prey, predator *cursor
			if iCursor.rlwEmptyRemaining < jCursor.rlwEmptyRemaining {
				prey, predator = iCursor, jCursor
			} else {
				prey, predator = jCursor, iCursor
			}

			if predator.getRunningBit() == false {
				index := prey.copyForward(container, predator.rlwEmptyRemaining, false)
				container.addStreamOfEmptyWords(false, predator.rlwEmptyRemaining - index)
				predator.moveForward(predator.rlwEmptyRemaining)
			} else {
				index := prey.copyForward(container, predator.rlwEmptyRemaining, true)
				container.addStreamOfEmptyWords(true, predator.rlwEmptyRemaining - index)
				predator.moveForward(predator.rlwEmptyRemaining)
			}
		}

		// Now that we have gone through all the empty words, let's take care of the left over literal words
		leftOverLiterals := int64(math.Min(float64(iCursor.rlwLiteralRemaining), float64(jCursor.rlwLiteralRemaining)))

		if leftOverLiterals > 0 {
			for k := int64(0); k < leftOverLiterals; k++ {
				container.add(iCursor.getLiteralWordAt(k) ^ jCursor.getLiteralWordAt(k))
			}

			// Move the cursors forward
			iCursor.moveForward(leftOverLiterals)
			jCursor.moveForward(leftOverLiterals)
		}
	}

	iRemains := iCursor.rlwRemaining() > 0
	var remaining *cursor

	if iRemains {
		remaining = iCursor
	} else {
		remaining = jCursor
	}

	remaining.copyForwardRemaining(container)
	container.setSizeInBits(int64(math.Max(float64(i.Size()), float64(j.Size()))))
}

func (this *Ewah) xorCardinality(a *Ewah) int32 {
	counter := newBitCounter()
	this.xorToContainer(a, counter)
	return int32(counter.(*bitCounter).getCount())
}
