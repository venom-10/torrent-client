package bitfield

type Bitfield []byte

func (bt Bitfield) HasPiece(index int) bool {
	byteIndex := index / 8
	offset := index % 8
	if index < 0 || byteIndex >= len(bt) {
		return false
	}
	return (bt[byteIndex] & (1 << (7 - offset))) != 0
}

func (bt Bitfield) SetPiece(index int) {
	byteIndex := index / 8
	offset := index % 8
	if index < 0 || byteIndex >= len(bt) {
		return
	}
	bt[byteIndex] |= (1 << (7 - offset))
}
