/*
 * Copyright (c) 2013 Zhen, LLC. http://zhen.io. All rights reserved.
 * Use of this source code is governed by the Apache 2.0 license.
 *
 */

package bitmap

type Bitmap interface {
	Set(int64) Bitmap
	Get(int64) bool
	Size() int64
	Reset()
	Clone() Bitmap
	Copy(Bitmap) Bitmap
	Equal(Bitmap) bool

	Cardinality() int64

	And(Bitmap) Bitmap
	Or(Bitmap) Bitmap
	AndNot(Bitmap) Bitmap
	Xor(Bitmap) Bitmap
	Not() Bitmap
}
