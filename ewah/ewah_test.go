/*
 * Copyright (c) 2013 Zhen, LLC. http://zhen.io. All rights reserved.
 * Use of this source code is governed by the Apache 2.0 license.
 *
 */

package ewah

import (
	"testing"
	"math/rand"
)

const (
	c1 uint32 = 0xcc9e2d51
	c2 uint32 = 0x1b873593
)

var (
	nums, nums10 []int64
	bm, bm10 *Ewah
	count int = 10000
)

func init() {
	nums = make([]int64, count)
	nums10 = make([]int64, count)

	bit := int64(0)
	rand.Seed(int64(c1))
	for i := 0; i < count; i++ {
		bit += int64(rand.Intn(1000)+1)
		nums[i] = bit
	}

	bit = int64(0)
	rand.Seed(int64(c2))
	for i := 0; i < count; i++ {
		bit += int64(rand.Intn(1000)+1)
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
		if !bm10.Set(nums10[i]).Get(nums10[i]) {
			t.Fatalf("Problem setting bm10[%d] with number %d\n", i, nums10[i])
		}
	}
	for i := 0; i < count; i++ {
		if ! bm.Get(nums[i]) {
			t.Fatalf("Check(%d) at %d failed\n", nums[i], i)
		}
	}
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

func TestGet2(t *testing.T) {
	for i := 0; i < count; i++ {
		if ! bm.Get2(nums[i]) {
			t.Fatalf("Get2(%d) at %d failed\n", nums[i], i)
		}
	}
	//bm.PrintStats(false)
}

func TestGet3(t *testing.T) {
	for i := 0; i < count; i++ {
		if ! bm.Get3(nums[i]) {
			t.Fatalf("Get3(%d) at %d failed\n", nums[i], i)
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

func BenchmarkGet2(b *testing.B) {
	//fmt.Printf("BenchmarkSetAndGet2 %d bits\n", b.N)
	failed := 0
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		if ! bm.Get2(nums[i%count]) {
			failed += 1
		}
	}

	b.StopTimer()
	if failed > 0 {
		b.Fatal("Test failed with", failed, "bits")
	}
}

func BenchmarkGet3(b *testing.B) {
	//fmt.Printf("BenchmarkSetAndGet2 %d bits\n", b.N)
	failed := 0
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		if ! bm.Get3(nums[i%count]) {
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

func BenchmarkCardinality2(b *testing.B) {
	for i := 0; i < b.N; i++ {
		bm.Cardinality2()
	}
}

func BenchmarkCardinality3(b *testing.B) {
	for i := 0; i < b.N; i++ {
		bm.Cardinality3()
	}
}

func BenchmarkCardinality4(b *testing.B) {
	for i := 0; i < b.N; i++ {
		bm.Cardinality4()
	}
}

func BenchmarkAnd(b *testing.B) {
	for i := 0; i < b.N; i++ {
		if bm.And(bm10) == nil {
			b.Fatal("BenchmarkAnd: Problem with And() at i =", i)
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

func BenchmarkAndNot(b *testing.B) {
	for i := 0; i < b.N; i++ {
		if bm.AndNot(bm10) == nil {
			b.Fatal("BenchmarkAnd: Problem with AndNot() at i =", i)
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
