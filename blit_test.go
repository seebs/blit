package blit

import (
	"fmt"
	"math/rand"
	"testing"
)

var prng = rand.New(rand.NewSource(0))

func weightedRandomBits(density uint64) uint64 {
	var out uint64
	if density == 64 {
		out--
		return out
	}
	if density == 0 {
		return out
	}
	scale := 64
	for scale > 1 {
		if density&1 != 0 {
			out = out | prng.Uint64()
		} else {
			out = out & prng.Uint64()
		}
		density >>= 1
		scale >>= 1
	}
	return out
}

const maxDepth = 10
const sampleSize = (1 << 16)

// existing is the hypothetical source bitmap we'd be modifying
var existing []*Bitmap
var ids [][]uint64
var values [][]uint64

func init() {
	values = make([][]uint64, maxDepth)
	existing = make([]*Bitmap, maxDepth)
	ids = make([][]uint64, 65)
	for i := uint64(1); i < 65; i++ {
		ids[i] = make([]uint64, 0, sampleSize)
		for j := uint64(0); j < sampleSize; j += 64 {
			bits := weightedRandomBits(i)
			for k := uint64(0); k < 64; k++ {
				if (bits>>k)&1 != 0 {
					ids[i] = append(ids[i], j+k)
				}
			}

		}
	}
	for i := uint(1); i < maxDepth; i++ {
		values[i] = make([]uint64, sampleSize)
		maxValue := int64(1 << uint(i))
		for j := 0; j < sampleSize; j++ {
			values[i][j] = uint64(prng.Int63n(maxValue))
		}
		bm := NewBitmap(i, sampleSize)
		existing[i] = bm
		for j := uint(0); j < i; j++ {
			for k := 0; k < sampleSize>>6; k++ {
				bm.Bits[j][k] = prng.Uint64()
			}
		}
		// set "data present" bits
		for k := 0; k < sampleSize>>6; k++ {
			bm.Bits[i][k] = prng.Uint64()
		}
	}
}

func TestCompareResults(t *testing.T) {
	empty := NewBitmap(6, 64)
	idBatch := make([]uint64, 64)
	for i := 0; i < 64; i++ {
		idBatch[i] = uint64(i)
	}
	naive := empty.SplatNaive(idBatch, idBatch)
	fancy := empty.SplatFancy(idBatch, idBatch)
	for i := uint(0); i < naive.Depth; i++ {
		if naive.Bits[i][0] != fancy.Bits[i][0] {
			t.Fatalf("mismatch in output")
		}
	}
}

func BenchmarkSplatNaive(b *testing.B) {
	for d := uint64(1); d < maxDepth; d += 3 {
		title := fmt.Sprintf("depth=%d", d)
		b.Run(title, func(innerB *testing.B) {
			for _, density := range []int{1, 8, 64} {
				title2 := fmt.Sprintf("density=%d", density)
				innerB.Run(title2, func(innerB2 *testing.B) {
					benchmarkSplatNaive(innerB2, d, density)
				})
			}
		})
	}
}

func benchmarkSplatNaive(b *testing.B, depth uint64, density int) {
	src := existing[depth]
	for i := 0; i < b.N; i++ {
		_ = src.SplatNaive(ids[density], values[depth])
	}
}

func BenchmarkSplatFancy(b *testing.B) {
	for d := uint64(1); d < maxDepth; d += 3 {
		title := fmt.Sprintf("depth=%d", d)
		b.Run(title, func(innerB *testing.B) {
			for _, density := range []int{1, 8, 64} {
				title2 := fmt.Sprintf("density=%d", density)
				innerB.Run(title2, func(innerB2 *testing.B) {
					benchmarkSplatFancy(innerB2, d, density)
				})
			}
		})
	}
}

func benchmarkSplatFancy(b *testing.B, depth uint64, density int) {
	src := existing[depth]
	for i := 0; i < b.N; i++ {
		_ = src.SplatFancy(ids[density], values[depth])
	}
}
