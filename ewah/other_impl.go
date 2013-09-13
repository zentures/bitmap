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

func (this *Ewah) Cardinality2() int64 {
	counter := int64(0)
	i := NewEWAHIterator(this.buffer, this.actualSizeInWords)

	for i.hasNext() {
		localrlw := i.next()
		if localrlw.getRunningBit() {
			counter += wordInBits * localrlw.getRunningLength()
		}

		for j := int32(0); j < localrlw.getNumberOfLiteralWords(); j++ {
			counter += int64(popcount_3(uint64(i.buffer()[i.literalWords()]+int64(j))))
		}
	}

	return counter
}

func (this *Ewah) Cardinality3() int64 {
	counter := int64(0)

	// index to marker word
	marker := int64(0)

	for marker < this.actualSizeInWords {
		localrlw := newRunningLengthWord(this.buffer, marker)

		if localrlw.getRunningBit() {
			counter += wordInBits * localrlw.getRunningLength()
		}

		numOfLiteralWords := int64(localrlw.getNumberOfLiteralWords())

		for j := int64(1); j <= numOfLiteralWords; j++ {
			counter += int64(popcount_4(uint64(this.buffer[marker + j])))
		}

		marker += numOfLiteralWords + 1
	}

	return counter
}

func (this *Ewah) Cardinality4() int64 {
	counter := int64(0)
	i := NewEWAHIterator(this.buffer, this.actualSizeInWords)

	for i.hasNext() {
		localrlw := i.next()
		if localrlw.getRunningBit() {
			counter += wordInBits * localrlw.getRunningLength()
		}

		for j := int32(0); j < localrlw.getNumberOfLiteralWords(); j++ {
			counter += int64(popcount_4(uint64(i.buffer()[i.literalWords()]+int64(j))))
		}
	}

	return counter
}

func (this *Ewah) Get1(i int64) bool {
	if i < 0 || i > this.sizeInBits {
		return false
	}

	wordi := i / wordInBits
	biti := uint64(i % wordInBits)

	wordChecked := int64(0)
	marker := int64(0)

	// index to marker word
	m := newRunningLengthWord(this.buffer, marker)

	for wordChecked <= wordi {
		m.reset(this.buffer, marker)
		//fmt.Printf("ewah.go/Get: marker = %064b\n", m.getActualWord())
		numOfLiteralWords := int64(m.getNumberOfLiteralWords())
		wordChecked += m.getRunningLength()

		if wordi < wordChecked {
			return m.getRunningBit()
		}

		if wordi < wordChecked + numOfLiteralWords {
			//fmt.Printf("ewah.go/Get: index = %d\n", marker + (wordi - wordChecked) + 1)
			//fmt.Printf("ewah.go/Get: word = %064b\n", this.buffer[marker + (wordi - wordChecked) + 1])
			//fmt.Printf("ewah.go/Get: bit = %064b\n", this.buffer[marker + (wordi - wordChecked) + 1] & (int64(1) << biti))
			return this.buffer[marker + (wordi - wordChecked) + 1] & (int64(1) << biti) != 0
		}
		wordChecked += numOfLiteralWords
		marker += numOfLiteralWords + 1
	}

	return false
}

func (this *Ewah) Get2(i int64) bool {
	if i < 0 || i > this.sizeInBits {
		return false
	}

	wordChecked := int64(0)
	wordi := i / wordInBits
	biti := uint64(i % wordInBits)

	// index to marker word
	iter := NewEWAHIterator(this.buffer, this.actualSizeInWords)

	for m := iter.next(); m != nil && wordChecked <= wordi; m = iter.next() {
		//fmt.Printf("ewah.go/Get3: %064b\n", m.getActualWord())
		//fmt.Println(m)
		//fmt.Printf("ewah.go/Get3: wordChecked = %d, wordi = %d, biti = %d\n", wordChecked, wordi, biti)
		numOfLiteralWords := int64(m.getNumberOfLiteralWords())
		wordChecked += m.getRunningLength()
		//fmt.Printf("ewah.go/Get3: wordChecked = %d, wordi = %d, biti = %d\n", wordChecked, wordi, biti)

		if wordi < wordChecked {
			return m.getRunningBit()
		}

		if wordi < wordChecked + numOfLiteralWords {
			//fmt.Printf("ewah.go/Get3: index = %d, literalwords = %d\n", int64(iter.literalWords()) + (wordi - wordChecked), iter.literalWords())
			//fmt.Printf("ewah.go/Get3: %064b\n", this.buffer[int64(iter.literalWords()) + (wordi - wordChecked)])
			//fmt.Printf("ewah.go/Get3: %064b\n", this.buffer[int64(iter.literalWords()) + (wordi - wordChecked)] & (int64(1) << biti))
			return this.buffer[int64(iter.literalWords()) + (wordi - wordChecked)] & (int64(1) << biti) != 0
		}
		wordChecked += numOfLiteralWords
	}

	return false
}

func (this *Ewah) Get3(i int64) bool {
	if i < 0 || i > this.sizeInBits {
		return false
	}

	wordChecked := int64(0)
	j := newBufferedRunningLengthWordIterator(NewEWAHIterator(this.buffer, this.actualSizeInWords))
	wordi := i/wordInBits
	biti := i%wordInBits

	for wordChecked <= wordi {
		brlw := j.brlw
		wordChecked += brlw.getRunningLength()

		if wordi < wordChecked {
			return brlw.getRunningBit()
		}

		if int64(wordi) < wordChecked + int64(brlw.getNumberOfLiteralWords()) {
			w := j.getLiteralWordAt(int32(wordi - wordChecked))
			return (w & (int64(1) << uint64(biti))) != 0
		}

		wordChecked += int64(brlw.getNumberOfLiteralWords())
		j.next()
	}

	return false;
}
func (this *Ewah) And2(a bitmap.Bitmap) bitmap.Bitmap {
	return this.bitOp(a, this.andToContainer2)
}

func (this *Ewah) andToContainer2(a *Ewah, container BitmapStorage) {
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
				container.addStreamOfEmptyWords(false, predator.getRunningLength())
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
				container.add(rlwi.getLiteralWordAt(k) & rlwj.getLiteralWordAt(k))
			}

			rlwi.discardFirstWords(nbre_literal)
			rlwj.discardFirstWords(nbre_literal)
		}
	}

	if this.adjustContainerSizeWhenAggregating {
		i_remains := rlwi.size() > 0
		var remaining *BufferedRunningLengthWordIterator

		if i_remains {
			remaining = rlwi
		} else {
			remaining = rlwj
		}

		remaining.dischargeAsEmpty(container)
		container.setSizeInBits(int64(math.Max(float64(this.sizeInBits), float64(a.sizeInBits))))
	}
}

func (this *Ewah) Not2() bitmap.Bitmap {
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

func (this *Ewah) AndNot2(a bitmap.Bitmap) bitmap.Bitmap {
	return this.bitOp(a, this.andNotToContainer2)
}

func (this *Ewah) andNotToContainer2(a *Ewah, container BitmapStorage) {
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

func (this *Ewah) Or2(a bitmap.Bitmap) bitmap.Bitmap {
	return this.bitOp(a, this.orToContainer2)
}

func (this *Ewah) orToContainer2(a *Ewah, container BitmapStorage) {
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

func (this *Ewah) Xor2(a bitmap.Bitmap) bitmap.Bitmap {
	return this.bitOp(a, this.xorToContainer2)
}

func (this *Ewah) xorToContainer2(a *Ewah, container BitmapStorage) {
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

