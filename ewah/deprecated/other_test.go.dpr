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
	"time"
	"github.com/zhenjl/bitmap"
)

/*
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
*/

func TestAndCompare(t *testing.T) {
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

func TestAndMultiple(t *testing.T) {
	rs := []int64{10, 100, 1000, 5000, 10000, 100000}

	bms := make([]bitmap.Bitmap, len(rs))

	for i := range rs {
		bit := int64(0)
		rand.Seed(int64(c1) + time.Now().UnixNano())
		bms[i] = New()

		for j := int64(0); j < rs[i]; j++ {
			bit += int64(rand.Intn(int(rs[i]))+1)
			bms[i].(*Ewah).Set(bit)
		}
	}

	bm4 := bms[0].And((bms[1:])...)

	bm5 := bms[0].(*Ewah).And2(bms[1])
	bm6 := bm5.(*Ewah).And2(bms[2])
	bm7 := bm6.(*Ewah).And2(bms[3])
	bm8 := bm7.(*Ewah).And2(bms[4])
	bm9 := bm8.(*Ewah).And2(bms[5])

	if !bm4.(*Ewah).Equal(bm9) {
		fmt.Println("==============> bm4 != bm5")
		//bm2.PrintStats(true)
		//bm3.PrintStats(true)
		t.Fatal("==============> bm4 != bm5")
	}
}

func TestOrMultiple(t *testing.T) {
	rs := []int64{10, 100, 1000, 5000, 10000, 100000}

	bms := make([]bitmap.Bitmap, len(rs))

	for i := range rs {
		bit := int64(0)
		rand.Seed(int64(c1) + time.Now().UnixNano())
		bms[i] = New()

		for j := int64(0); j < rs[i]; j++ {
			bit += int64(rand.Intn(int(rs[i]))+1)
			bms[i].(*Ewah).Set(bit)
		}
	}

	bm4 := bms[0].Or((bms[1:])...)

	bm5 := bms[0].(*Ewah).Or2(bms[1])
	bm6 := bm5.(*Ewah).Or2(bms[2])
	bm7 := bm6.(*Ewah).Or2(bms[3])
	bm8 := bm7.(*Ewah).Or2(bms[4])
	bm9 := bm8.(*Ewah).Or2(bms[5])

	if !bm4.(*Ewah).Equal(bm9) {
		fmt.Println("==============> bm4 != bm5")
		//bm2.PrintStats(true)
		//bm3.PrintStats(true)
		t.Fatal("==============> bm4 != bm5")
	}
}

func TestXorMultiple(t *testing.T) {
	rs := []int64{10, 100, 1000, 5000, 10000, 100000}

	bms := make([]bitmap.Bitmap, len(rs))

	for i := range rs {
		bit := int64(0)
		rand.Seed(int64(c1) + time.Now().UnixNano())
		bms[i] = New()

		for j := int64(0); j < rs[i]; j++ {
			bit += int64(rand.Intn(int(rs[i]))+1)
			bms[i].(*Ewah).Set(bit)
		}
	}

	bm4 := bms[0].Xor((bms[1:])...)

	bm5 := bms[0].(*Ewah).Xor2(bms[1])
	bm6 := bm5.(*Ewah).Xor2(bms[2])
	bm7 := bm6.(*Ewah).Xor2(bms[3])
	bm8 := bm7.(*Ewah).Xor2(bms[4])
	bm9 := bm8.(*Ewah).Xor2(bms[5])

	if !bm4.(*Ewah).Equal(bm9) {
		fmt.Println("==============> bm4 != bm5")
		//bm2.PrintStats(true)
		//bm3.PrintStats(true)
		t.Fatal("==============> bm4 != bm5")
	}
}

func TestAndNotMultiple(t *testing.T) {
	rs := []int64{10, 100, 1000, 5000, 10000, 100000}

	bms := make([]bitmap.Bitmap, len(rs))

	for i := range rs {
		bit := int64(0)
		rand.Seed(int64(c1) + time.Now().UnixNano())
		bms[i] = New()

		for j := int64(0); j < rs[i]; j++ {
			bit += int64(rand.Intn(int(rs[i]))+1)
			bms[i].(*Ewah).Set(bit)
		}
	}

	bm4 := bms[0].AndNot((bms[1:])...)

	bm5 := bms[0].(*Ewah).AndNot2(bms[1])
	bm6 := bm5.(*Ewah).AndNot2(bms[2])
	bm7 := bm6.(*Ewah).AndNot2(bms[3])
	bm8 := bm7.(*Ewah).AndNot2(bms[4])
	bm9 := bm8.(*Ewah).AndNot2(bms[5])

	if !bm4.(*Ewah).Equal(bm9) {
		fmt.Println("==============> bm4 != bm5")
		//bm2.PrintStats(true)
		//bm3.PrintStats(true)
		t.Fatal("==============> bm4 != bm5")
	}
}

func TestOrCompare(t *testing.T) {
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

func TestXorCompare(t *testing.T) {
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

				bm4 := bm2.Xor(bm3)
				bm5 := bm2.Xor(bm3)

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

func TestAndNotCompare(t *testing.T) {
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

/*
func TestNot2(t *testing.T) {
	bit := int64(0)
	rand.Seed(int64(c1))

	bm3 := New().(*Ewah)

	for j := int64(0); j < 100; j++ {
		bit += int64(rand.Intn(int(100)) + 1)
		bm3.Set(bit)
	}

	bm3.Not2()
}
*/

func TestNotCompare(t *testing.T) {
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

/*
func BenchmarkGet1(b *testing.B) {
	//fmt.Printf("BenchmarkSetAndGet %d bits\n", b.N)
	failed := 0
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		if ! bm.Get1(nums[i%count]) {
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

func BenchmarkAnd2(b *testing.B) {
	for i := 0; i < b.N; i++ {
		if bm.And2(bm10) == nil {
			b.Fatal("BenchmarkAnd2: Problem with And() at i =", i)
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

func BenchmarkAndNot2(b *testing.B) {
	for i := 0; i < b.N; i++ {
		if bm.AndNot2(bm10) == nil {
			b.Fatal("BenchmarkAnd: Problem with AndNot2() at i =", i)
		}
	}
}

func BenchmarkOr2(b *testing.B) {
	for i := 0; i < b.N; i++ {
		if bm.Or2(bm10) == nil {
			b.Fatal("BenchmarkAnd: Problem with Or2() at i =", i)
		}
	}
}

func BenchmarkXor2(b *testing.B) {
	for i := 0; i < b.N; i++ {
		if bm.Xor2(bm10) == nil {
			b.Fatal("BenchmarkAnd: Problem with Xor2() at i =", i)
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

func benchmarkDifferentCombinations2(b *testing.B, op string, b1, b2 int, s1, s2 int) {
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

	var f func(bitmap.Bitmap) bitmap.Bitmap
	switch op {
	case "and":
		f = m1.And2
	case "or":
		f = m1.Or2
	case "andnot":
		f = m1.AndNot2
	case "xor":
		f = m1.Xor2
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
*/

