/*
 * Copyright (c) 2013 Zhen, LLC. http://zhen.io. All rights reserved.
 * Use of this source code is governed by the Apache 2.0 license.
 *
 */

package bitset

import (
	"github.com/willf/bitset"
	"github.com/reducedb/bitmap"
)

type Bitset struct {
	b *bitset.BitSet
}

var _ bitmap.Bitmap = (*Bitset)(nil)

func New() bitmap.Bitmap {
	return &Bitset{
		b: bitset.New(4),
	}
}

func (this *Bitset) Set(i int64) bitmap.Bitmap {
	this.b.Set(uint(i))
	return this
}

func (this *Bitset) Get(i int64) bool {
	return this.b.Test(uint(i))
}

func (this *Bitset) Size() int64 {
	return int64(this.b.Len())
}

func (this *Bitset) Reset() {
	this.b.ClearAll()
}

func (this *Bitset) Clone() bitmap.Bitmap {
	return &Bitset{
		b: this.b.Clone(),
	}
}

func (this *Bitset) Copy(other bitmap.Bitmap) bitmap.Bitmap {
	this.b.Copy(other.(*Bitset).b)
	return this
}

func (this *Bitset) Equal(other bitmap.Bitmap) bool {
	return this.b.Equal(other.(*Bitset).b)
}

func (this *Bitset) Cardinality() int64 {
	return int64(this.b.Count())
}

func (this *Bitset) And(a ...bitmap.Bitmap) bitmap.Bitmap {
	n := len(a)
	bm, ok := a[0].(*Bitset)
	if !ok {
		return nil
	}

	ans := New().(*Bitset)
	ans.b = this.b.Intersection(bm.b)

	for i := 1; i < n; i++ {
		bm, ok := a[i].(*Bitset)
		if !ok {
			return nil
		}

		ans.b = ans.b.Intersection(bm.b)
	}

	return ans
}

func (this *Bitset) Or(a ...bitmap.Bitmap) bitmap.Bitmap {
	n := len(a)
	bm, ok := a[0].(*Bitset)
	if !ok {
		return nil
	}

	ans := New().(*Bitset)
	ans.b = this.b.Union(bm.b)

	for i := 1; i < n; i++ {
		bm, ok := a[i].(*Bitset)
		if !ok {
			return nil
		}

		ans.b = ans.b.Union(bm.b)
	}

	return ans
}

func (this *Bitset) AndNot(a ...bitmap.Bitmap) bitmap.Bitmap {
	n := len(a)
	bm, ok := a[0].(*Bitset)
	if !ok {
		return nil
	}

	ans := New().(*Bitset)
	ans.b = this.b.Difference(bm.b)

	for i := 1; i < n; i++ {
		bm, ok := a[i].(*Bitset)
		if !ok {
			return nil
		}

		ans.b = ans.b.Difference(bm.b)
	}

	return ans
}

func (this *Bitset) Xor(a ...bitmap.Bitmap) bitmap.Bitmap {
	n := len(a)
	bm, ok := a[0].(*Bitset)
	if !ok {
		return nil
	}

	ans := New().(*Bitset)
	ans.b = this.b.SymmetricDifference(bm.b)

	for i := 1; i < n; i++ {
		bm, ok := a[i].(*Bitset)
		if !ok {
			return nil
		}

		ans.b = ans.b.SymmetricDifference(bm.b)
	}

	return ans
}

func (this *Bitset) Not() bitmap.Bitmap {
	this.b = this.b.Complement()
	return this
}

