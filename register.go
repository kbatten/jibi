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
	h, l := wordToBytes(u)
	r.set(h)
	r.lrp.set(l)
}

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

func (r register8) getWord() uint16 {
	return bytesToWord(r.get(), r.lrp.get())
}

func (r register8) get() uint8 {
	return *r.vp & r.mask
}

func (r register8) set(v uint8) {
	*r.vp = v
}

func (r register8) setFlag(f uint8) {
	*r.vp |= f
}

func (r register8) resetFlag(f uint8) {
	*r.vp &= (f ^ 0xFF)
}

func (r register8) getFlag(f uint8) bool {
	if *r.vp&f == f {
		return true
	}
	return false
}

type register16 uint16

func (r register16) Uint16() uint16 {
	return uint16(r)
}

func (r register16) String() string {
	return fmt.Sprintf("0x%04X", uint16(r))
}
