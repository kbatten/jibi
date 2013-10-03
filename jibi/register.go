package jibi

import (
	"fmt"
)

type register8 struct {
	vp   *uint8
	lrp  *register8 // lsb of register pair
	mask uint8
}

func newFlagsRegister8() register8 {
	return register8{new(uint8), nil, 0xF0}
}

func newRegister8(lrp *register8) register8 {
	return register8{new(uint8), lrp, 0xFF}
}

func (r register8) String() string {
	return fmt.Sprintf("0x%02X", *r.vp)
}

func (r register8) Uint16() uint16 {
	return uint16(bytesToWord(r.High(), r.Low()))
}

func (r register8) High() Byte {
	return Byte(r.Uint8())
}

func (r register8) Low() Byte {
	if r.lrp == nil {
		panic("lower register is nil")
	}
	return Byte(r.lrp.Uint8())
}

func (r register8) Inc() Word {
	return Word(r.Uint16() + 1)
}

func (r register8) Uint8() uint8 {
	return *r.vp & r.mask
}

func (r register8) set(v Byter) {
	*r.vp = v.Uint8()
}

func (r register8) reset() {
	*r.vp = 0
}

func (r register8) setWord(w Word) {
	if r.lrp == nil {
		panic("lower register is nil")
	}
	r.set(w.High())
	r.lrp.set(w.Low())
}

// flags
const (
	flagZ Byte = 0x80 >> iota
	flagN
	flagH
	flagC
)

func (r register8) flagsString() string {
	fZ := 0
	fN := 0
	fH := 0
	fC := 0
	if r.getFlag(flagZ) {
		fZ = 1
	}
	if r.getFlag(flagN) {
		fN = 1
	}
	if r.getFlag(flagH) {
		fH = 1
	}
	if r.getFlag(flagC) {
		fC = 1
	}
	return fmt.Sprintf("zero:%d sub:%d half:%d carry:%d", fZ, fN, fH, fC)
}

//func (r register8) getWord() Word {
//	return bytesToWord(r.get(), r.lrp.get())
//}

//func (r register8) get() uint8 {
//	return *r.vp & r.mask
//}

func (r register8) setFlag(f Byter) {
	*r.vp |= f.Uint8()
}

func (r register8) resetFlag(f Byter) {
	*r.vp &= (f.Uint8() ^ 0xFF)
}

func (r register8) getFlag(f Byter) bool {
	if *r.vp&f.Uint8() == f.Uint8() {
		return true
	}
	return false
}

type register16 uint16

func (r register16) Uint16() uint16 {
	return uint16(r)
}

func (r register16) High() Byte {
	return Byte(r >> 8)
}

func (r register16) Low() Byte {
	return Byte(r)
}

func (r register16) Inc() Word {
	return Word(r.Uint16() + 1)
}

func (r register16) String() string {
	return fmt.Sprintf("0x%04X", uint16(r))
}
