package main

import (
//	"fmt"
)

type mmu struct {
	rom  memoryDevice // 32k
	vram memoryDevice // 8k
	eram memoryDevice // 8k
	wram memoryDevice // 8k
	iram memoryDevice // 8k
}

type memoryDevice interface {
	readByte(addressInterface) uint8
	writeByte(addressInterface, uint8)
}

type ramModule []uint8

func newRamModule(size uint16, data []uint8) *ramModule {
	rm := make(ramModule, size)
	copy(rm, data)
	return &rm
}
func (r *ramModule) readByte(addr addressInterface) uint8 {
	a := addr.Uint16()
	if a > uint16(len(*r)) {
		panic("ram read out of range")
	}
	return (*r)[a]
}

func (r *ramModule) writeByte(addr addressInterface, b uint8) {
	a := addr.Uint16()
	if a > uint16(len(*r)) {
		panic("ram write out of range")
	}
	(*r)[a] = b
}

type romModule []uint8

func newRomModule(size uint16, data []uint8) *romModule {
	rm := make(romModule, size)
	copy(rm, data)
	return &rm
}
func (r *romModule) readByte(addr addressInterface) uint8 {
	a := addr.Uint16()
	if a > uint16(len(*r)) {
		panic("rom read out of range")
	}
	return (*r)[a]
}

func (r *romModule) writeByte(addressInterface, uint8) {
	// nop
}

func newMmu(cart cartridge) mmu {
	mc := mmu{
		rom:  cart,
		vram: newRamModule(0x2000, nil),
		eram: newRamModule(0x2000, nil),
		wram: newRamModule(0x2000, nil),
		iram: newRamModule(0x2000, nil)}
	return mc
}

type addressInterface interface {
	Uint16() uint16
}

type address uint16

func (u address) Uint16() uint16 { return uint16(u) }

func (mc mmu) readByte(addr addressInterface) uint8 {
	a := addr.Uint16()
	if a < 0x8000 {
		return mc.rom.readByte(address(a))
	}
	a -= 0x8000
	if a < 0x2000 {
		return mc.vram.readByte(address(a))
	}
	a -= 0x2000
	if a < 0x2000 {
		return mc.eram.readByte(address(a))
	}
	a -= 0x2000
	if a < 0x2000 {
		return mc.wram.readByte(address(a))
	}
	a -= 0x2000
	return mc.iram.readByte(address(a))
}

func (mc mmu) readWord(addr address) uint16 {
	l := mc.readByte(addr)
	h := mc.readByte(address(addr.Uint16() + 1))
	return bytesToWord(h, l)
}

func (mc mmu) writeByte(addr addressInterface, b uint8) {
	a := addr.Uint16()
	if a < 0x8000 {
		mc.rom.writeByte(address(a), b)
		return
	}
	a -= 0x8000
	if a < 0x2000 {
		mc.vram.writeByte(address(a), b)
		return
	}
	a -= 0x2000
	if a < 0x2000 {
		mc.eram.writeByte(address(a), b)
		return
	}
	a -= 0x2000
	if a < 0x2000 {
		mc.wram.writeByte(address(a), b)
		return
	}
	a -= 0x2000
	mc.iram.writeByte(address(a), b)
}

func (mc mmu) writeWord(addr address, w uint16) {
	h, l := wordToBytes(w)
	mc.writeByte(addr, l)
	mc.writeByte(address(addr.Uint16()+1), h)
}
