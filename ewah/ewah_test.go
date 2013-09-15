/*
 * Copyright (c) 2013 Zhen, LLC. http://zhen.io. All rights reserved.
 * Use of this source code is governed by the Apache 2.0 license.
 *
 */

package ewah

import (
	"testing"
	"math/rand"
	"fmt"
	"github.com/zhenjl/bitmap"
)

const (
	c1 uint32 = 0xcc9e2d51
	c2 uint32 = 0x1b873593

	count int = 10000
)

var (
	nums, nums10 []int64
	bm, bm10 *Ewah
)

func init() {
	nums = make([]int64, count)
	nums10 = make([]int64, count)

	bit := int64(0)
	rand.Seed(int64(c1))
	for i := 0; i < count; i++ {
		bit += int64(rand.Intn(10000)+1)
		nums[i] = bit
	}

	bit = int64(0)
	rand.Seed(int64(c2))
	for i := 0; i < count; i++ {
		bit += int64(rand.Intn(10000)+1)
		nums10[i] = bit
	}

	bm = New().(*Ewah)
	bm10 = New().(*Ewah)
}

func TestSet(t *testing.T) {
	for i := 0; i < count; i++ {
		if !bm.Set(nums[i]).Get(nums[i]) {
			t.Fatalf("Problem setting bm[%d] with number %d\n", i, nums[i])
		}
	}
	for i := 0; i < count; i++ {
		if !bm10.Set(nums10[i]).Get(nums10[i]) {
			t.Fatalf("Problem setting bm10[%d] with number %d\n", i, nums10[i])
		}
	}
	//bm.PrintStats(false)
	//bm10.PrintStats(false)
}

func TestSet2(t *testing.T) {
	rs := []int64{10, 100, 1000, 10000, 100000}
	bm2 := New().(*Ewah)

	for r := range rs {
		nums2 := make([]int64, count)

		bit := int64(0)
		rand.Seed(int64(c1))
		for i := 0; i < count; i++ {
			bit += int64(rand.Intn(int(rs[r]))+1)
			nums2[i] = bit
		}

		for i := 0; i < count; i++ {
			if bm2.Set(nums2[i]) == nil {
				t.Fatalf("Problem setting bm[%d] with number %d\n", i, nums2[i])
			}
		}

		for i := 0; i < count; i++ {
			if !bm2.Get(nums2[i]) {
				t.Fatalf("Problem checking bm[%d]: should be set%d\n", i, nums2[i])
			}
		}

		bm2.Reset()
		if bm2.Cardinality() != 0 {
			t.Fatal("Problem resetting bm2")
		}
	}
}

func TestGet(t *testing.T) {
	for i := 0; i < count; i++ {
		if ! bm.Get(nums[i]) {
			t.Fatalf("Check(%d) at %d failed\n", nums[i], i)
		}
	}
	//bm.PrintStats(false)
}

func TestSwap(t *testing.T) {
	bm2 := New().(*Ewah)
	bm3 := New().(*Ewah)

	bm2.Set(10)
	bm2.Set(70)
	bm2.Set(100)
	bm2.Set(150)
	bm2.Set(15000)
	bm3.Set(11)
	bm3.Set(13)
	bm3.Set(100)
	bm3.Set(15000)

	bm2.Swap(bm3)

	c2 := bm2.Cardinality()
	if c2 != 4 {
		t.Fatalf("Cardinality of bm2 %d != 4", c2)
	}

	c3 := bm3.Cardinality()
	if c3 != 5 {
		t.Fatalf("Cardinality of bm2 %d != 5", c3)
	}

	nums2 := []int64{11, 13, 100, 15000}
	nums3 := []int64{10, 70, 100, 150, 15000}

	for i := range nums2 {
		if !bm2.Get(nums2[i]) {
			t.Fatalf("Get(%d) failed, should be set\n", nums2[i])
		}
	}

	for i := range nums3 {
		if !bm3.Get(nums3[i]) {
			t.Fatalf("Get(%d) failed, should be set\n", nums3[i])
		}
	}
}

func TestClone(t *testing.T) {
	bm2 := bm.Clone()

	for i := 0; i < count; i++ {
		if ! bm2.Get(nums[i]) {
			t.Fatalf("Check(%d) at %d failed\n", nums[i], i)
		}
	}
	//bm.PrintStats(false)
}

func TestCopy(t *testing.T) {
	bm2 := New().(*Ewah)
	bm2.Copy(bm)

	for i := 0; i < count; i++ {
		if ! bm2.Get(nums[i]) {
			t.Fatalf("Check(%d) at %d failed\n", nums[i], i)
		}
	}
	//bm.PrintStats(false)
}

func TestAnd(t *testing.T) {
	bm2 := New().(*Ewah)
	bm3 := New().(*Ewah)

	bm2.Set(10)
	bm2.Set(70)
	bm2.Set(100)
	bm3.Set(100)
	bm3.Set(15000)

	bm4 := bm2.And(bm3)

	if bm4.Cardinality() != 1 {
		t.Fatal("Cardinality != 1")
	}


	if bm4.Get(10) {
		t.Fatalf("Get(%d) failed, should NOT be set\n", 10)
	}

	if bm4.Get(70) {
		t.Fatalf("Get(%d) failed, should NOT be set\n", 70)
	}

	if !bm4.Get(100) {
		t.Fatalf("Get(%d) failed, should be set\n", 100)
	}

	if bm4.Get(15000) {
		t.Fatalf("Get(%d) failed, should NOT be set\n", 150)
	}
}

func TestAndNot(t *testing.T) {
	bm2 := New().(*Ewah)
	bm3 := New().(*Ewah)

	bm2.Set(10)
	bm2.Set(70)
	bm2.Set(100)
	bm2.Set(150)
	bm2.Set(15000)
	bm3.Set(11)
	bm3.Set(13)
	bm3.Set(100)
	bm3.Set(15000)

	bm4 := bm2.AndNot(bm3)

	if bm4.Cardinality() != 3 {
		bm2.PrintStats(true)
		bm3.PrintStats(true)
		bm4.(*Ewah).PrintStats(true)
		t.Fatal("Cardinality != 3")
	}

	if !bm4.Get(10) {
		t.Fatalf("Get(%d) failed, should be set\n", 10)
	}

	if !bm4.Get(70) {
		t.Fatalf("Get(%d) failed, should be set\n", 70)
	}

	if bm4.Get(100) {
		t.Fatalf("Get(%d) failed, should NOT be set\n", 100)
	}

	if !bm4.Get(150) {
		t.Fatalf("Get(%d) failed, should be set\n", 150)
	}

	if bm4.Get(15000) {
		t.Fatalf("Get(%d) failed, should NOT be set\n", 15000)
	}
}

func TestOr(t *testing.T) {
	bm2 := New().(*Ewah)
	bm3 := New().(*Ewah)

	bm2.Set(10)
	bm2.Set(70)
	bm2.Set(100)
	bm2.Set(150)
	bm2.Set(15000)
	bm3.Set(11)
	bm3.Set(13)
	bm3.Set(100)
	bm3.Set(15000)

	bm4 := bm2.Or(bm3)

	if bm4.Cardinality() != 7 {
		t.Fatal("Cardinality != 7")
	}

	nums2 := []int64{10, 70, 100, 150, 15000, 11, 13}
	for i := range nums2 {
		if !bm4.Get(nums2[i]) {
			t.Fatalf("Get(%d) failed, should be set\n", nums2[i])
		}
	}
}

func TestNot(t *testing.T) {
	bm2 := New().(*Ewah)

	bm2.Set(10)
	bm2.Set(100)
	bm2.Set(10000)

	c1 := bm2.Cardinality()
	size := bm2.sizeInBits
	bm2.Not()
	c2 := bm2.Cardinality()

	nums2 := []int64{10, 100, 10000}
	for i := range nums2 {
		if bm2.Get(nums2[i]) {
			t.Fatalf("Get(%d) failed, should NOT be set\n", nums2[i])
		}
	}

	if c1 != size - c2 {
		t.Fatalf("c1 (%d) != size (%d) - c2 (%d)", c1, size, c2)
	}
}

func TestXor(t *testing.T) {
	bm2 := New().(*Ewah)
	bm3 := New().(*Ewah)

	bm2.Set(10)
	bm2.Set(70)
	bm2.Set(100)
	bm2.Set(150)
	bm2.Set(15000)
	bm3.Set(11)
	bm3.Set(13)
	bm3.Set(100)
	bm3.Set(15000)

	bm4 := bm2.Xor(bm3)

	c := bm4.Cardinality()
	if c != 5 {
		t.Fatalf("Cardinality %d != 2", 5)
	}

	set := []int64{10, 70, 150, 11, 13}
	for i := range set {
		if !bm4.Get(set[i]) {
			t.Fatalf("Get(%d) failed, should be set\n", set[i])
		}
	}

	notset := []int64{100, 15000}
	for i := range notset {
		if bm4.Get(notset[i]) {
			t.Fatalf("Get(%d) failed, should NOT be set\n", notset[i])
		}
	}
}

func BenchmarkGet(b *testing.B) {
	//fmt.Printf("BenchmarkSetAndGet %d bits\n", b.N)
	failed := 0
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		if ! bm.Get(nums[i%count]) {
			failed += 1
		}
	}

	b.StopTimer()
	if failed > 0 {
		b.Fatal("Test failed with", failed, "bits")
	}
}

func BenchmarkCardinality(b *testing.B) {
	for i := 0; i < b.N; i++ {
		bm.Cardinality()
	}
}

func BenchmarkAnd(b *testing.B) {
	for i := 0; i < b.N; i++ {
		if bm.And(bm10) == nil {
			b.Fatal("BenchmarkAnd: Problem with And() at i =", i)
		}
	}
}

func BenchmarkNot(b *testing.B) {
	for i := 0; i < b.N; i++ {
		if bm.Not() == nil {
			b.Fatal("BenchmarkAnd: Problem with Not() at i =", i)
		}
	}
}

func BenchmarkAndNot(b *testing.B) {
	for i := 0; i < b.N; i++ {
		if bm.AndNot(bm10) == nil {
			b.Fatal("BenchmarkAnd: Problem with AndNot() at i =", i)
		}
	}
}

func BenchmarkOr(b *testing.B) {
	for i := 0; i < b.N; i++ {
		if bm.Or(bm10) == nil {
			b.Fatal("BenchmarkAnd: Problem with Or() at i =", i)
		}
	}
}

func BenchmarkXor(b *testing.B) {
	for i := 0; i < b.N; i++ {
		if bm.Xor(bm10) == nil {
			b.Fatal("BenchmarkAnd: Problem with Xor() at i =", i)
		}
	}
}

// f is the function to call, like And, Or, Xor, AndNot
// b1 is the number of bits for the first bitmap
// b2 is the number of bits for the second bitmap
// s1 is the sparsity of the first bitmap
// s2 is the sparsity of the second bitmap
func benchmarkDifferentCombinations(b *testing.B, op string, b1, b2 int, s1, s2 int) {
	m1 := New().(*Ewah)
	m2 := New().(*Ewah)

	bit := int64(0)
	rand.Seed(int64(c1))
	for i := 0; i < b1; i++ {
		bit += int64(rand.Intn(s1)+1)
		m1.Set(bit)
	}

	bit = 0
	rand.Seed(int64(c2))
	for i := 0; i < b2; i++ {
		bit += int64(rand.Intn(s1)+1)
		m2.Set(bit)
	}

	var f func(...bitmap.Bitmap) bitmap.Bitmap
	switch op {
	case "and":
		f = m1.And
	case "or":
		f = m1.Or
	case "andnot":
		f = m1.AndNot
	case "xor":
		f = m1.Xor
	default:
		return
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if f(m2) == nil {
			b.Fatal("Problem with %s benchmark at i =", op, i)
		}
	}
}

func testGenerateData(t *testing.T) {
	is := []int{100, 10000, 1000000}
	js := []int{100, 10000, 1000000}
	ks := []int{3, 30, 300, 3000, 30000}
	ls := []int{3, 30, 300, 3000, 30000}

	m1 := New().(*Ewah)
	m2 := New().(*Ewah)

	for i := range is {
		for j := range js {
			for k := range ks {
				for l := range ls {
					bit := int64(0)
					rand.Seed(int64(c1))
					for a := 0; a < is[i]; a++ {
						bit += int64(rand.Intn(ks[k])+1)
						m1.Set(bit)
					}

					bit = 0
					rand.Seed(int64(c2))
					for b := 0; b < js[j]; b++ {
						bit += int64(rand.Intn(ls[l])+1)
						m2.Set(bit)
					}

					fmt.Printf("%d %d %d %d %d %d %.2f%% %d %d %d %.2f%% %d\n",
						is[i], js[j], ks[k], ls[l],
						m1.Size(), m1.SizeInWords(), (1-float64(m1.SizeInWords()*wordInBits)/float64(m1.Size()))*100, m1.Cardinality(),
						m2.Size(), m2.SizeInWords(), (1-float64(m2.SizeInWords()*wordInBits)/float64(m2.Size()))*100, m2.Cardinality())

					m1.Reset()
					m2.Reset()
				}
			}
		}
	}
}
