/*
 * Copyright (c) 2013 Zhen, LLC. http://zhen.io. All rights reserved.
 * Use of this source code is governed by the Apache 2.0 license.
 *
 */

package ewah

// Wikipedia: http://en.wikipedia.org/wiki/Hamming_weight
const (
	m1  uint64 = 0x5555555555555555 //binary: 0101...
	m2  uint64 = 0x3333333333333333 //binary: 00110011..
	m4  uint64 = 0x0f0f0f0f0f0f0f0f //binary:  4 zeros,  4 ones ...
	h01 uint64 = 0x0101010101010101 //the sum of 256 to the power of 0,1,2,3...
)

type bitCounter struct {
	oneBits uint64
}

func newBitCounter() BitmapStorage {
	return &bitCounter{}
}

var _ BitmapStorage = (*bitCounter)(nil)

func (this *bitCounter) add(newdata uint64) {
	this.oneBits += popcount_3(newdata)
}

func (this *bitCounter) addStreamOfLiteralWords(data []uint64, start, number int32) {
	for _, v := range data[start : start+number] {
		this.add(v)
	}
}

func (this *bitCounter) addStreamOfEmptyWords(v bool, number int64) {
	if v {
		this.oneBits += uint64(number * wordInBits)
	}
}

func (this *bitCounter) addStreamOfNegatedLiteralWords(data []uint64, start, number int32) {
	for _, v := range data[start : start+number] {
		this.add(^v)
	}
}

func (this *bitCounter) getCount() uint64 {
	return this.oneBits
}

func (this *bitCounter) setSizeInBits(bits int64) error {
	return nil
}

// This is better when most bits in x are 0
// It uses 3 arithmetic operations and one comparison/branch per "1" bit in x.
func popcount_4(x uint64) uint64 {
	count := uint64(0)
	for ; x != 0; count++ {
		x &= x - 1
	}

	return count
}

// Wikipedia: http://en.wikipedia.org/wiki/Hamming_weight, popcount_3
func popcount_3(x uint64) uint64 {
	x -= (x >> 1) & m1             //put count of each 2 bits into those 2 bits
	x = (x & m2) + ((x >> 2) & m2) //put count of each 4 bits into those 4 bits
	x = (x + (x >> 4)) & m4        //put count of each 8 bits into those 8 bits
	return (x * h01) >> 56         //returns left 8 bits of x + (x<<8) + (x<<16) + (x<<24) + ...
}
