package blit

// Bitmap holds bitplanes
type Bitmap struct {
	// Bits[0] = low-order bit
	// topmost bit is "value present"
	Bits  [][]uint64
	data  []uint64 // single internal allocation
	words uint64   // number of 64-bit words to use
	Depth uint
	Size  uint64
}

// NewBitmap yields a bitmap
func NewBitmap(depth uint, size uint64) *Bitmap {
	b := Bitmap{}
	b.Bits = make([][]uint64, depth+1)
	b.Depth = depth
	b.Size = size
	b.words = (b.Size + 63) >> 6
	b.data = make([]uint64, uint64(depth+1)*b.words)
	for i := uint(0); i < depth+1; i++ {
		b.Bits[i] = b.data[uint64(i)*b.words : uint64(i+1)*b.words]
	}
	return &b
}

// Clone duplicates a bitmap.
func (b *Bitmap) Clone() *Bitmap {
	nb := &Bitmap{Bits: b.Bits, Depth: b.Depth, Size: b.Size, words: b.words}
	nb.Bits = make([][]uint64, nb.Depth+1)
	nb.data = make([]uint64, len(b.data))
	copy(nb.data, b.data)
	for i := uint(0); i < b.Depth+1; i++ {
		nb.Bits[i] = nb.data[uint64(i)*nb.words : uint64(i+1)*nb.words]
	}
	return nb
}

// SplatNaive does the easy/obvious thing.
func (b *Bitmap) SplatNaive(ids []uint64, values []uint64) *Bitmap {
	nb := b.Clone()
	for idx, id := range ids {
		word := id >> 6
		bit := uint64(id & 0x3f)
		mask := uint64(1 << bit)
		value := values[idx]
		for j := uint(0); j < nb.Depth; j++ {
			if value&(1<<uint64(j)) != 0 {
				nb.Bits[j][word] |= mask
			} else {
				nb.Bits[j][word] &^= mask
			}
		}
		nb.Bits[nb.Depth][word] |= mask
	}
	return nb
}

// Apply just dumps in the bits from 'set' and 'mask' and zeroes out set.
func (b *Bitmap) Apply(word uint64, mask uint64, set []uint64) {
	for i := uint(0); i < b.Depth; i++ {
		b.Bits[i][word] = (b.Bits[i][word] &^ mask) | set[i]
		set[i] = 0
	}
	b.Bits[b.Depth][word] |= mask
}

// SplatFancy tries to do this in a fancier way
func (b *Bitmap) SplatFancy(ids []uint64, values []uint64) *Bitmap {
	nb := b.Clone()
	var mask uint64
	set := make([]uint64, b.Depth)
	var previousWord uint64
	for idx, id := range ids {
		value := values[idx]
		word := uint64(id) >> 6
		bit := uint64(id & 0x3f)
		if word != previousWord {
			nb.Apply(word, mask, set)
			mask = 0
			previousWord = word
		}
		for i := uint(0); i < b.Depth; i++ {
			set[i] |= ((value >> i) & 1) << bit
		}
		mask |= (1 << bit)
	}
	nb.Apply(previousWord, mask, set)
	return nb
}
