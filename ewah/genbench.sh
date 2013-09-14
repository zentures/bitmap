#!/bin/bash

cat <<EOF > benchmark_2_autogen_test.go
package ewah

import (
    "testing"
)
EOF

cat <<EOF > benchmark_autogen_test.go
package ewah

import (
    "testing"
)
EOF

# h = operations
# i = # of bits for bitmap 1
# j = # of bits for bitmap 2
# k = random distance between bits for bitmap 1 (0-k) - indicates sparsity
# l = random distance between bits for bitmap 1 (0-l) - indicates sparsity
for h in and or xor andnot
do
    for i in 100 10000 1000000
    do
        for j in 100 10000 1000000
        do
            for k in 3 30 300 3000 30000
            do
                for l in 3 30 300 3000 30000
                do
                    cat <<EOF >> benchmark_autogen_test.go
func Benchmark_${h}_${i}_${j}_${k}_${l}(b *testing.B) {
    benchmarkDifferentCombinations(b, "$h", $i, $j, $k, $l)
}

EOF

                    cat <<EOF >> benchmark_2_autogen_test.go
func Benchmark_2_${h}_${i}_${j}_${k}_${l}(b *testing.B) {
    benchmarkDifferentCombinations2(b, "$h", $i, $j, $k, $l)
}

EOF
                done
            done
        done
    done
done

go test -bench "^Benchmark_2_and_" > 2_and.out
go test -bench "^Benchmark_and_" > 1_and.out
go test -bench "^Benchmark_2_or_" > 2_or.out
go test -bench "^Benchmark_or_" > 1_or.out
go test -bench "^Benchmark_2_xor_" > 2_xor.out
go test -bench "^Benchmark_xor_" > 1_xor.out
go test -bench "^Benchmark_2_andnot_" > 2_andnot.out
go test -bench "^Benchmark_andnot_" > 1_andnot.out
