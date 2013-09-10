/*
 * Copyright (c) 2013 Zhen, LLC. http://zhen.io. All rights reserved.
 * Use of this source code is governed by the Apache 2.0 license.
 *
 */

package ewah


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


func (this *Ewah) Get2(i int64) bool {
	if i < 0 || i > this.sizeInBits {
		return false
	}

	wordChecked := int64(0)
	wordi := i / this.wordInBits
	biti := uint64(i % this.wordInBits)

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
	wordi := i/this.wordInBits
	biti := i%this.wordInBits

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
