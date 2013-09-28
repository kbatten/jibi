package main

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

func (r register8) setWord(u uint16) {
	if r.lrp == nil {
		panic("lower register is nil")
	}
	r.set(uint8(u >> 8))
	r.lrp.set(uint8(u & 0xFF))
}

func (r register8) getWord() uint16 {
	return uint16(r.get())<<8 + uint16(r.lrp.get())
}

func (r register8) get() uint8 {
	return *r.vp & r.mask
}

func (r register8) set(v uint8) {
	*r.vp = v
}

type register16 uint16

func (r register16) Uint16() uint16 {
	return uint16(r)
}

func (r register16) String() string {
	return fmt.Sprintf("0x%04X", uint16(r))
}
