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
	bm.PrintStats(false)
	bm10.PrintStats(false)
}
/*
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

func TestAnd2(t *testing.T) {
	bm2 := New().(*Ewah)
	bm3 := New().(*Ewah)

	bm2.Set(10)
	bm2.Set(70)
	bm2.Set(100)
	bm3.Set(100)
	bm3.Set(300)
	bm3.Set(15000)

	bm4 := bm2.And2(bm3)
	//bm4.(*Ewah).PrintStats(true)

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

func TestAndAndAnd2(t *testing.T) {
	bm2 := bm.And(bm10)
	bm3 := bm.And2(bm10)

	c2 := bm2.(*Ewah).Cardinality()
	c3 := bm3.(*Ewah).Cardinality()

	bm2.(*Ewah).PrintStats(true)
	bm3.(*Ewah).PrintStats(true)

	if c2 != c3 {
		t.Fatalf("c2(%d) != c3(%d)\n", c2, c3)
	}
}

func TestAnd2More(t *testing.T) {
	rs := []int64{10, 100, 1000, 5000, 10000, 100000}

	for h := range rs {
		for i := range rs {
			bit := int64(0)
			rand.Seed(int64(c1))

			bm2 := New().(*Ewah)

			for j := int64(0); j < rs[i]; j++ {
				bit += int64(rand.Intn(int(rs[h]))+1)
				bm2.Set(bit)
			}

			for k := range rs {
				bit2 := int64(0)
				rand.Seed(int64(c2))

				bm3 := New().(*Ewah)

				for l := int64(0); l < rs[k]; l++ {
					bit2 += int64(rand.Intn(int(rs[h]))+1)
					bm3.Set(bit2)
				}

				bm4 := bm2.And(bm3)
				bm5 := bm2.And2(bm3)

				if !bm4.(*Ewah).Equal(bm5) {
					fmt.Printf("************* Testing h = %d, i = %d, k = %d\n", rs[h], rs[i], rs[k])
					fmt.Println("==============> bm4 != bm5")
					bm2.PrintStats(true)
					bm3.PrintStats(true)
					bm4.(*Ewah).PrintStats(true)
					bm5.(*Ewah).PrintStats(true)
					t.Fatal("==============> bm4 != bm5")
				}
			}
		}
	}
}
*/

func TestOr2More(t *testing.T) {
	rs := []int64{10, 100, 1000, 5000, 10000, 100000}

	for h := range rs {
		for i := range rs {
			bit := int64(0)
			rand.Seed(int64(c1))

			bm2 := New().(*Ewah)

			for j := int64(0); j < rs[i]; j++ {
				bit += int64(rand.Intn(int(rs[h]))+1)
				bm2.Set(bit)
			}

			for k := range rs {
				bit2 := int64(0)
				rand.Seed(int64(c2))

				bm3 := New().(*Ewah)

				for l := int64(0); l < rs[k]; l++ {
					bit2 += int64(rand.Intn(int(rs[h]))+1)
					bm3.Set(bit2)
				}

				bm4 := bm2.Or(bm3)
				bm5 := bm2.Or(bm3)

				if !bm4.(*Ewah).Equal(bm5) {
					fmt.Printf("************* Testing h = %d, i = %d, k = %d\n", rs[h], rs[i], rs[k])
					fmt.Println("==============> bm4 != bm5")
					bm2.PrintStats(true)
					bm3.PrintStats(true)
					bm4.(*Ewah).PrintStats(true)
					bm5.(*Ewah).PrintStats(true)
					t.Fatal("==============> bm4 != bm5")
				}
			}
		}
	}
}
/*
func TestAndNot2More(t *testing.T) {
	rs := []int64{10, 100, 1000, 5000, 10000, 100000}

	for h := range rs {
		for i := range rs {
			bit := int64(0)
			rand.Seed(int64(c1))

			bm2 := New().(*Ewah)

			for j := int64(0); j < rs[i]; j++ {
				bit += int64(rand.Intn(int(rs[h]))+1)
				bm2.Set(bit)
			}

			for k := range rs {
				bit2 := int64(0)
				rand.Seed(int64(c2))

				bm3 := New().(*Ewah)

				for l := int64(0); l < rs[k]; l++ {
					bit2 += int64(rand.Intn(int(rs[h]))+1)
					bm3.Set(bit2)
				}

				bm4 := bm2.AndNot(bm3)
				bm5 := bm2.AndNot2(bm3)

				if !bm4.(*Ewah).Equal(bm5) {
					fmt.Printf("************* Testing h = %d, i = %d, k = %d\n", rs[h], rs[i], rs[k])
					fmt.Println("==============> bm4 != bm5")
					bm2.PrintStats(true)
					bm3.PrintStats(true)
					bm4.(*Ewah).PrintStats(true)
					bm5.(*Ewah).PrintStats(true)
					t.Fatal("==============> bm4 != bm5")
				}
			}
		}
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

func TestNot2(t *testing.T) {
	bit := int64(0)
	rand.Seed(int64(c1))

	bm3 := New().(*Ewah)

	for j := int64(0); j < 100; j++ {
		bit += int64(rand.Intn(int(100)) + 1)
		bm3.Set(bit)
	}

	fmt.Printf("************* Testing Not - bm3\n")
	bm3.Not2()
	bm3.PrintStats(false)
}

func TestNot2More(t *testing.T) {
	rs := []int64{10, 100, 1000, 5000, 10000, 100000}

	for h := range rs {
		for i := range rs {
			bit := int64(0)
			rand.Seed(int64(c1))

			bm2 := New().(*Ewah)
			bm3 := New().(*Ewah)
			bm4 := New().(*Ewah)

			for j := int64(0); j < rs[i]; j++ {
				bit += int64(rand.Intn(int(rs[h]))+1)
				bm2.Set(bit)
				bm3.Set(bit)
				bm4.Set(bit)
			}

			bm2.Not()
			bm3.Not2()

			if !bm2.Equal(bm3) {
				fmt.Printf("************* Testing Not h = %d, i = %d\n", rs[h], rs[i])
				fmt.Println("==============> bm2 != bm3")
				bm2.PrintStats(true)
				bm3.PrintStats(true)
				t.Fatal("==============> bm2 != bm3")
			}

			bm2.Not()
			bm3.Not2()

			if !bm2.Equal(bm4) {
				fmt.Printf("************* Testing Not Not\n")
				fmt.Println("==============> bm2 != bm4")
				bm2.PrintStats(true)
				bm3.PrintStats(true)
				t.Fatal("==============> bm2 != bm3")
			}

			if !bm3.Equal(bm4) {
				fmt.Printf("************* Testing Not Not\n")
				fmt.Println("==============> bm3 != bm4")
				bm2.PrintStats(true)
				bm3.PrintStats(true)
				t.Fatal("==============> bm3 != bm4")
			}

		}
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

func BenchmarkAnd2(b *testing.B) {
	for i := 0; i < b.N; i++ {
		if bm.And2(bm10) == nil {
			b.Fatal("BenchmarkAnd2: Problem with And() at i =", i)
		}
	}
}

func BenchmarkNot(b *testing.B) {
	for i := 0; i < b.N; i++ {
		if bm.Not2() == nil {
			b.Fatal("BenchmarkAnd: Problem with And() at i =", i)
		}
	}
}

func BenchmarkNot2(b *testing.B) {
	for i := 0; i < b.N; i++ {
		if bm.Not2() == nil {
			b.Fatal("BenchmarkAnd2: Problem with And() at i =", i)
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

func BenchmarkAndNot2(b *testing.B) {
	for i := 0; i < b.N; i++ {
		if bm.AndNot2(bm10) == nil {
			b.Fatal("BenchmarkAnd: Problem with AndNot() at i =", i)
		}
	}
}
*/
func BenchmarkOr(b *testing.B) {
	for i := 0; i < b.N; i++ {
		if bm.Or(bm10) == nil {
			b.Fatal("BenchmarkAnd: Problem with AndNot() at i =", i)
		}
	}
}

func BenchmarkOr2(b *testing.B) {
	for i := 0; i < b.N; i++ {
		if bm.Or2(bm10) == nil {
			b.Fatal("BenchmarkAnd: Problem with AndNot() at i =", i)
		}
	}
}
