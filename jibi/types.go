package jibi

// A Bit is a single bit.
type Bit uint8

// A Byte is an 8 bit byte.
type Byte uint8

// A Word is a 16 bit word.
type Word uint16

// High returns the high Byte.
func (w Word) High() Byte {
	return Byte(w >> 8)
}

// Low returns the low Byte.
func (w Word) Low() Byte {
	return Byte(w)
}
