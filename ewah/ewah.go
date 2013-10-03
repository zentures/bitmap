/*
 * Copyright (c) 2013 Zhen, LLC. http://zhen.io. All rights reserved.
 * Use of this source code is governed by the Apache 2.0 license.
 *
 */

package ewah

import (
	"github.com/reducedb/bitmap"
	"math"
	"fmt"
	"errors"
)

const (
	// wordInBits is the constant representing the number of bits in a int64
	wordInBits int64 = 64

	// defaultBufferSize is a constant default memory allocation when the object is constructed
	defaultBufferSize uint64 = 4

	RunningLengthBits int32 = 32
	LiteralBits int32 = 64 - 1 - RunningLengthBits
	LargestLiteralCount uint64 = (uint64(1) << uint32(LiteralBits)) - 1
	LargestRunningLengthCount uint64 = (uint64(1) << uint32(RunningLengthBits)) - 1
	RunningLengthPlusRunningBit uint64 = (uint64(1) << uint32(RunningLengthBits + 1)) - 1
	ShiftedLargestRunningLengthCount uint64 = LargestRunningLengthCount << 1
	NotRunningLengthPlusRunningBit uint64 = ^RunningLengthPlusRunningBit
	NotShiftedLargestRunningLengthCount uint64 = ^ShiftedLargestRunningLengthCount

)

type Ewah struct {
	// actualSizeInWords is the number of words actually used in the buffer to represent the bitmap
	actualSizeInWords int64

	// sizeInBits is the number of total bits in the bitmap
	sizeInBits int64

	// buffer representing the bitmap
	buffer []uint64

	// whether we adjust after some aggregation by adding in zeroes
	adjustContainerSizeWhenAggregating bool

	// getCursor remembers the last search position and try to search from there for the next one
	// It's an optimization for sequential Gets
	getCursor *cursor

	// setCursor remembers the last set position and move forward from there
	setCursor *cursor
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
		this.addLiteralWord(uint64(1) << uint64((i % wordInBits)))

		return this
	}

	// Now we know dist == 0 since it can't be < 0 (can't set a bit past the current active bit)
	if this.setCursor.literalCount() == 0 {
		this.setCursor.setEmptyCount(this.setCursor.emptyCount() - 1)
		this.addLiteralWord(1 << uint64(i % wordInBits))
		return this
	}

	this.buffer[this.actualSizeInWords - 1] |= 1 << uint64(i % wordInBits)
	if this.buffer[this.actualSizeInWords - 1] == ^uint64(0) {
		this.buffer[this.actualSizeInWords - 1] = 0
		this.actualSizeInWords -= 1
		this.setCursor.setLiteralCount(this.setCursor.literalCount() - 1)
		this.addEmptyWord(true)

		// Be a good citizen and update the cursors
		this.getCursor.quickUpdate(this.buffer, this.actualSizeInWords)
		this.setCursor.quickUpdate(this.buffer, this.actualSizeInWords)
	}

	return this
}

func (this *Ewah) Get(i int64) bool {
	if i < 0 || i > this.sizeInBits {
		return false
	}

	wordToCheck := i / wordInBits
	bitInWord := uint64(i % wordInBits)

	// If the word to check is before the the words already checked then let's update the buffer
	if wordToCheck < this.getCursor.totalChecked {
		//fmt.Printf("ewah.go/Get: reset ---> wordToCheck = %d, bitInWord = %d, size = %d\n---> %v\n", wordToCheck, bitInWord, this.SizeInWords(), this.getCursor)
		this.getCursor.reset(this.buffer, this.actualSizeInWords)
	}

	//fmt.Printf("ewah.go/Get: ---> wordToCheck = %d, bitInWord = %d\n", wordToCheck, bitInWord)
	//this.printDetails()

	for this.getCursor.totalChecked <= wordToCheck && !this.getCursor.end() {
		//fmt.Println("ewah.go/Get: cursor =", this.getCursor)

		emptyRemaining := this.getCursor.emptyRemaining()

		if wordToCheck < this.getCursor.totalChecked + emptyRemaining {
			//fmt.Println("ewah.go/Get: inside empty words,", this.getCursor.emptyBit())
			return this.getCursor.emptyBit()
		}

		this.getCursor.moveForward(emptyRemaining)

		//fmt.Printf("ewah.go/Get: after move forward %d empty words, cursor = %v\n", emptyRemaining, this.getCursor)

		literalRemaining := this.getCursor.literalRemaining()

		if wordToCheck < this.getCursor.totalChecked + literalRemaining {
			//i := this.getCursor.marker + (wordToCheck - this.getCursor.totalChecked) + 1
			//fmt.Printf("%4d: %20d %064b\n", i, uint64(this.buffer[i]), uint64(this.buffer[i]))
			b := this.buffer[this.getCursor.marker + (wordToCheck - this.getCursor.totalChecked) + 1] & (uint64(1) << bitInWord) != 0
			//fmt.Printf("ewah.go/Get: inside literal words and it's %t\n", b)
			return b
		}

		this.getCursor.moveForward(literalRemaining)
		//fmt.Printf("ewah.go/Get: after move forward %d literal words, cursor = %v\n", literalRemaining, this.getCursor)
	}

	return false
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
	this.actualSizeInWords = 1
	this.sizeInBits = 0
	this.adjustContainerSizeWhenAggregating = true

	if this.buffer == nil {
		this.buffer = make([]uint64, defaultBufferSize)
	} else {
		this.buffer[0] = 0
	}

	if this.setCursor == nil {
		this.setCursor = newCursor(this.buffer, this.actualSizeInWords)
	} else {
		this.setCursor.reset(this.buffer, this.actualSizeInWords)
	}

	if this.getCursor == nil {
		this.getCursor = newCursor(this.buffer, this.actualSizeInWords)
	} else {
		this.getCursor.reset(this.buffer, this.actualSizeInWords)
	}

}

func (this *Ewah) Swap(other *Ewah) bitmap.Bitmap {
	this.buffer, other.buffer = other.buffer, this.buffer
	this.actualSizeInWords, other.actualSizeInWords = other.actualSizeInWords, this.actualSizeInWords
	this.sizeInBits, other.sizeInBits = other.sizeInBits, this.sizeInBits

	s1, s2 := this.setCursor.marker, other.setCursor.marker
	g1, g2 := this.getCursor.marker, other.getCursor.marker

	this.setCursor.resetMarker(this.buffer, this.actualSizeInWords, s2)
	this.getCursor.resetMarker(this.buffer, this.actualSizeInWords, g2)
	other.setCursor.resetMarker(other.buffer, other.actualSizeInWords, s1)
	other.getCursor.resetMarker(other.buffer, other.actualSizeInWords, g1)

	return this
}

func (this *Ewah) Clone() bitmap.Bitmap {
	c := New().(*Ewah)
	c.reserve(int32(this.actualSizeInWords))
	copy(c.buffer, this.buffer)
	c.actualSizeInWords = this.actualSizeInWords
	c.sizeInBits = this.sizeInBits

	c.setCursor.resetMarker(c.buffer, c.actualSizeInWords, this.setCursor.marker)
	c.getCursor.resetMarker(c.buffer, c.actualSizeInWords, this.getCursor.marker)
	c.getCursor.totalChecked = this.getCursor.totalChecked

	return c
}

func (this *Ewah) Copy(other bitmap.Bitmap) bitmap.Bitmap {
	o := other.(*Ewah)
	this.buffer = make([]uint64, o.SizeInWords())
	copy(this.buffer, o.buffer)
	this.actualSizeInWords = o.SizeInWords()
	this.sizeInBits = o.Size()

	this.setCursor.resetMarker(this.buffer, this.actualSizeInWords, o.setCursor.marker)
	this.getCursor.resetMarker(this.buffer, this.actualSizeInWords, o.getCursor.marker)
	this.getCursor.totalChecked = o.getCursor.totalChecked

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

	//fmt.Printf("ewah.go/Equal: sizeInWords = %d, this.cap = %d, o.cap = %d\n", this.SizeInWords(), cap(this.buffer), cap(o.buffer))

	for i := int64(0); i < this.actualSizeInWords; i++ {
		if o.buffer[i] != this.buffer[i] {
			return false
		}
	}
	return true
}

func (this *Ewah) Cardinality() int64 {
	n := int64(0)
	c := newCursor(this.buffer, this.actualSizeInWords)

	for !c.end() {
		//fmt.Printf("ewah.go/Cardinality: cursor = %v\n", c)
		if c.emptyBit() {
			n += wordInBits * c.emptyCount()
		}

		//fmt.Printf("ewah.go/Cardinality: #literalWords = %d\n", c.literalCount())
		for j := int64(0); j < c.literalCount(); j++ {
			n += int64(popcount_3(uint64(c.getLiteralWordAt(j))))
		}

		if c.nextMarker() != nil {
			break
		}
	}

	return n
}

func (this *Ewah) PrintStats(details bool) {
	fmt.Printf("actualSizeInWords = %d, actualSizeInBits = %d, cardinality = %d\n", this.SizeInWords(), this.Size(), this.Cardinality())

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


// add is used to add words directly to the bitmap.
func (this *Ewah) add(newdata uint64) {
	this.addSignificantBits(newdata, wordInBits)
}

// addWithSize adds words directly to the bitmap, but with the number of significant bits specified.
func (this *Ewah) addSignificantBits(newdata uint64, bitsthatmatter int64) {
	//fmt.Printf("ewah.go/addSignificantBits:    %064b\n----\n", newdata)
	this.sizeInBits += bitsthatmatter
	if newdata == 0 {
		this.addEmptyWord(false)
	} else if newdata == ^uint64(1) {
		this.addEmptyWord(true)
	} else {
		this.addLiteralWord(newdata)
	}
}

// addEmptyWord adds an empty word of 1's or 0's to the bitmap. true: newdata==0; false: newdata== ~0
func (this *Ewah) addEmptyWord(v bool) {
	noLiteralWord := this.setCursor.literalCount() == 0
	runlen := this.setCursor.emptyCount()

	if noLiteralWord && runlen == 0 {
		this.setCursor.setEmptyBit(v)
	}

	if noLiteralWord && this.setCursor.emptyBit() == v && uint64(runlen) < LargestRunningLengthCount {
		this.setCursor.setEmptyCount(runlen+1)
		return
	}

	this.pushback(0)
	this.setCursor.resetMarker(this.buffer, this.actualSizeInWords, this.actualSizeInWords-1)
	this.setCursor.setEmptyBit(v)
	this.setCursor.setEmptyCount(1)
}

// addLiteralWord adds a literal word to the bitmap.
func (this *Ewah) addLiteralWord(newdata uint64) {
	//fmt.Printf("ewah.go/addLiteralWord: newdata = %064b\n", newdata)
	numberSoFar := this.setCursor.literalCount()
	//fmt.Printf("ewah.go/addLiteralWord: numberSoFar = %d\n", numberSoFar)
	if uint64(numberSoFar) >= LargestLiteralCount {
		this.pushback(0)
		this.setCursor.resetMarker(this.buffer, this.actualSizeInWords, this.actualSizeInWords-1)
		this.setCursor.setLiteralCount(1)
		this.pushback(newdata)
	}
	this.setCursor.setLiteralCount(numberSoFar+1)
	//fmt.Printf("ewah.go/addLiteralWord: getNumberOfLiteralWords = %d\n", this.setCursor.literalCount())
	this.pushback(newdata)
}

// addStreamOfLiteralWords adds several literal words at a time, might be faster
func (this *Ewah) addStreamOfLiteralWords(data []uint64, start, number int32) {
	leftOverNumber := int64(number)

	for leftOverNumber > 0 {
		numberOfLiteralWords := this.setCursor.literalCount()
		whatWeCanAdd := int64(math.Min(float64(leftOverNumber), float64(LargestLiteralCount - uint64(numberOfLiteralWords))))

		this.setCursor.setLiteralCount(numberOfLiteralWords + whatWeCanAdd)
		leftOverNumber -= whatWeCanAdd

		//fmt.Printf("ewah.go/addStreamOfLiteralWords: #ofLiteral = %d, leftOver = %d, whatWeCanAdd = %d\n", numberOfLiteralWords, leftOverNumber, whatWeCanAdd)
		this.pushbackMultiple(data, start, int32(whatWeCanAdd))
		this.sizeInBits += whatWeCanAdd * wordInBits

		if leftOverNumber > 0 {
			this.pushback(0)
			this.setCursor.resetMarker(this.buffer, this.actualSizeInWords, this.actualSizeInWords-1)
		}
	}

	//fmt.Printf("ewah.go/addStreamOfLiteralWords: sizeinbits = %d, actuaSizeInWords = %d\n", this.sizeInBits, this.actualSizeInWords)
}

// addStreamOfEmptyWords adds several empty words at a time, might be faster
func (this *Ewah) addStreamOfEmptyWords(v bool, number int64) {
	if number == 0 {
		return
	}

	this.sizeInBits += number * wordInBits

	if this.setCursor.emptyBit() != v && this.setCursor.size() == 0 {
		this.setCursor.setEmptyBit(v)
	} else if this.setCursor.literalCount() != 0 || this.setCursor.emptyBit() != v {
		this.pushback(0)
		this.setCursor.resetMarker(this.buffer, this.actualSizeInWords, this.actualSizeInWords-1)
		if v {
			this.setCursor.setEmptyBit(v)
		}
	}

	runlen := this.setCursor.emptyCount()
	whatWeCanAdd := int64(math.Min(float64(number), float64(LargestLiteralCount - uint64(runlen))))

	this.setCursor.setEmptyCount(runlen + whatWeCanAdd)
	number -= whatWeCanAdd

	for uint64(number) >= LargestRunningLengthCount {
		this.pushback(0)
		this.setCursor.resetMarker(this.buffer, this.actualSizeInWords, this.actualSizeInWords-1)
		if v {
			this.setCursor.setEmptyBit(v)
		}

		this.setCursor.setEmptyCount(int64(LargestRunningLengthCount))
		number -= int64(LargestRunningLengthCount)
	}

	if number > 0 {
		this.pushback(0)
		this.setCursor.resetMarker(this.buffer, this.actualSizeInWords, this.actualSizeInWords-1)
		if v {
			this.setCursor.setEmptyBit(v)
		}
		this.setCursor.setEmptyCount(number)
	}

	//fmt.Printf("ewah.go/addStreamOfEmptyWords: sizeinbits = %d, actuaSizeInWords = %d\n", this.sizeInBits, this.actualSizeInWords)
}

// fastAddStreamOfEmptyWords adds many zeroes and ones faster. This does not update sizeInBits
func (this *Ewah) fastAddStreamOfEmptyWords(v bool, number int64) {
	if this.setCursor.emptyBit() != v && this.setCursor.size() == 0 {
		this.setCursor.setEmptyBit(v)
	} else if this.setCursor.literalCount() != 0 || this.setCursor.emptyBit() != v {
		this.pushback(0)
		this.setCursor.resetMarker(this.buffer, this.actualSizeInWords, this.actualSizeInWords-1)
		if v {
			this.setCursor.setEmptyBit(v)
		}
	}

	runlen := this.setCursor.emptyCount()
	whatWeCanAdd := int64(math.Min(float64(number), float64(LargestLiteralCount - uint64(runlen))))

	this.setCursor.setEmptyCount(runlen + whatWeCanAdd)
	number -= whatWeCanAdd

	for uint64(number) >= LargestRunningLengthCount {
		this.pushback(0)
		this.setCursor.resetMarker(this.buffer, this.actualSizeInWords, this.actualSizeInWords-1)
		if v {
			this.setCursor.setEmptyBit(v)
		}

		this.setCursor.setEmptyCount(int64(LargestRunningLengthCount))
		number -= int64(LargestRunningLengthCount)
	}

	if number > 0 {
		this.pushback(0)
		this.setCursor.resetMarker(this.buffer, this.actualSizeInWords, this.actualSizeInWords-1)
		if v {
			this.setCursor.setEmptyBit(v)
		}

		this.setCursor.setEmptyCount(number)
	}
}

// addStreamOfNegatedLiteralWords is similar to addStreamOfLiteralWords except the words are negated
func (this *Ewah) addStreamOfNegatedLiteralWords(data []uint64, start, number int32) {
	leftOverNumber := int64(number)

	for leftOverNumber > 0 {
		numberOfLiteralWords := this.setCursor.literalCount()
		whatWeCanAdd := int64(math.Min(float64(leftOverNumber), float64(LargestLiteralCount - uint64(numberOfLiteralWords))))

		this.setCursor.setLiteralCount(numberOfLiteralWords + whatWeCanAdd)
		leftOverNumber -= whatWeCanAdd
		this.negativePushBack(data, start, int32(whatWeCanAdd))
		this.sizeInBits += whatWeCanAdd * wordInBits

		if leftOverNumber > 0 {
			this.pushback(0)
			this.setCursor.resetMarker(this.buffer, this.actualSizeInWords, this.actualSizeInWords-1)
		}
	}
}

func (this *Ewah) negativePushBack(data []uint64, start, number int32) {
	negativeData := make([]uint64, number)

	for i := int32(0); i < number; i++ {
		negativeData[i] = ^data[start + i]
	}

	this.pushbackMultiple(negativeData, 0, number)
}

// pushback adds an element at the end
//
// This is a convenience method that calls push_back_multiple
func (this *Ewah) pushback(data uint64) {
	this.pushbackMultiple([]uint64{data}, 0, 1)
}

// pushback adds multiple element at the end
//
// This is the C++ vector pushback description. Adds a new element at the end of the vector, after its
// current last element. The content of val is copied (or moved) to the new element.
//
// This effectively increases the container size by one, which causes an automatic reallocation of the
// allocated storage space if -and only if- the new vector size surpasses the current vector capacity.
func (this *Ewah) pushbackMultiple(data []uint64, start, number int32) {
	// If the size of the bitmap is the same as the buffer length, that means the buffer is full, so we need
	// to allocate
	nextSize := this.actualSizeInWords + int64(number)
	bufferCap := int64(cap(this.buffer))
	//fmt.Printf("ewah.go/pushbackMultiple: start = %d, number = %d, size = %d, cap = %d\n", start, number, this.actualSizeInWords, bufferCap)
	if nextSize >= bufferCap {
		var newSize int64
		if nextSize < 32768 {
			newSize = nextSize * 2
		} else if nextSize * 3 / 2 < nextSize {
			// overflow
			newSize = math.MaxInt32
		} else {
			newSize = nextSize * 3 / 2
		}
		oldBuffer := this.buffer
		this.buffer = make([]uint64, newSize)
		copy(this.buffer, oldBuffer)
	}
	//fmt.Printf("ewah.go/pushbackMultiple: copy(this.buffer[%d:], data[%d:%d]), cap=%d", this.actualSizeInWords, start, start+number, cap(this.buffer))
	copy(this.buffer[this.actualSizeInWords:], data[start:start+number])
	this.actualSizeInWords += int64(number)

	// Let's do the right thing and update the set and get cursors
	this.setCursor.quickUpdate(this.buffer, this.actualSizeInWords)
	this.getCursor.quickUpdate(this.buffer, this.actualSizeInWords)
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
		this.buffer = make([]uint64, size)
		copy(this.buffer, oldBuffer)
		this.setCursor.reset(this.buffer, this.actualSizeInWords)
	}

	return this
}
