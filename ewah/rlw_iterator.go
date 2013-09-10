/*
 * Copyright (c) 2013 Zhen, LLC. http://zhen.io. All rights reserved.
 * Use of this source code is governed by the Apache 2.0 license.
 *
 */

package ewah

type RLWIterator interface {
	next() bool
	getLiteralWordAt(int32) int64
	getNumberOfLiteralWords() int32
	getRunningBit() bool
	size() int64
	getRunningLength() int64
	discardFirstWords(int64)
}
