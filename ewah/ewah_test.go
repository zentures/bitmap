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

var (
	nums []int64
	bm *Ewah
	count int = 10000
)

func init() {
	nums = make([]int64, count)
	bit := int64(0)
	for i := 0; i < count; i++ {
		bit += int64(rand.Intn(10000)+1)
		//fmt.Println(bit)
		nums[i] = bit
	}

	bm = New().(*Ewah)
}

func TestSet(t *testing.T) {

	for i := 0; i < count; i++ {
		if bm.Set(nums[i]) == nil {
			t.Fatal("Problem setting bit i", i, "with number", nums[i])
		}
	}

	//bm.PrintStats(false)
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
