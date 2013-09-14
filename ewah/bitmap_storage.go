/*
 * Copyright (c) 2013 Zhen, LLC. http://zhen.io. All rights reserved.
 * Use of this source code is governed by the Apache 2.0 license.
 *
 */

package ewah

type BitmapStorage interface {
	add(uint64)
	addStreamOfLiteralWords([]uint64, int32, int32)
	addStreamOfEmptyWords(bool, int64)
	addStreamOfNegatedLiteralWords([]uint64, int32, int32)
	setSizeInBits(int64) error
}
