package fbmuck

type Bitset int

func NewBitset(i ...int) (r Bitset) {
	for _, v := range i {
		r |= 1 << v
	}
	return
}

func (b Bitset) IsFlagged(i ...int) bool {
	return b & NewBitset(i...) != 0
}

func (b Bitset) IsFlaggedAnyOf(i ...int) (r bool) {
	for x := len(i) - 1; !r && x > 0; x--  {
		if b & (1 << i[x]) != 0 {
			r = true
		}
	}
	return
}

func (b *Bitset) FlagAs(i ...int) {
	b |= NewBitset(i...)
}

func (b *Bitset) ClearFlags(i ...int) {
	b &= ^NewBitset(i...)
}