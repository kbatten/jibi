package jibi

type Bit uint8

type Byte uint8

func (b Byte) Byte() Byte {
	return b
}

type Byter interface {
	Byte() Byte
}

type Word uint16

func (w Word) Word() Word {
	return w
}

func (w Word) High() Byte {
	return Byte(w >> 8)
}

func (w Word) Low() Byte {
	return Byte(w)
}

type Worder interface {
	Word() Word
	High() Byte
	Low() Byte
}
