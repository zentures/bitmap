#!/bin/bash

cat <<EOF >> benchmark_test.go
package ewah

import (
    "testing"
)
EOF

for h in and or xor andnot
do
    for i in 10 1000 100000 10000000
    do
        for j in 10 1000 100000 10000000
        do
            for k in 10 30 50
            do
                for l in 10 30 50
                do
                    cat <<EOF >> benchmark_test.go
func Benchmark_${h}_${i}_${j}_${k}_${l}(b *testing.B) {
    benchmarkDifferentCombinations(b, "$h", $i, $j, $k, $l)
}

EOF
                done
            done
        done
    done
done

for h in and or xor andnot
do
    for i in 10 1000 100000 10000000
    do
        for j in 10 1000 100000 10000000
        do
            for k in 10 30 50
            do
                for l in 10 30 50
                do
                    cat <<EOF >> benchmark_test.go
func Benchmark_${h}_${i}_${j}_${k}_${l}_2(b *testing.B) {
    benchmarkDifferentCombinations2(b, "$h", $i, $j, $k, $l)
}

EOF
                done
            done
        done
    done
done
