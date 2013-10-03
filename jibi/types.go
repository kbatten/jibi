package jibi

type Bit uint8

type Byte uint8

func (b Byte) Uint8() uint8 {
	return uint8(b)
}

type Byter interface {
	Uint8() uint8
}

type Word uint16

func (w Word) Uint16() uint16 {
	return uint16(w)
}

func (w Word) High() Byte {
	return Byte(w >> 8)
}

func (w Word) Low() Byte {
	return Byte(w)
}

func (w Word) Inc() Word {
	return Word(w.Uint16() + 1)
}

type Worder interface {
	Uint16() uint16
	High() Byte
	Low() Byte
	Inc() Word
}
