bitmap
======

The bitmap package implements the Enhanced Word-Aligned Hybrid (EWAH) bitmap compression algorithms, for now. The setup is so that multiple bitmap compressions can be implemented under the same [bitmap interface](https://github.com/reducedb/bitmap/blob/master/bitmap.go).

For more details please refer to the [blog post](http://zhen.org/blog/bitmap-compression-using-ewah-in-go/).

Please see [ewah_test.go](https://github.com/reducedb/bitmap/blob/master/ewah/ewah_test.go) for examples of how to use.

