package main

import ()

type mmu struct {
	rom  memoryDevice // 0000-7FFF 32k
	vram memoryDevice // 8000-9FFF 8k
	eram memoryDevice // A000-BFFF 8k
	wram memoryDevice // C000-DFFF 8k
	oam  memoryDevice // FE00-FE9F
	io   memoryDevice // FF00-FF7F
	zero memoryDevice // FF80-FFFF
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
}

func newMmu(cart cartridge, vid video) mmu {
	mc := mmu{
		rom:  cart,
		vram: vid,
		eram: cart.eram,
		wram: newRamModule(0x2000, nil),
		oam:  vid.oam,
		io:   newRamModule(0x4D, nil),
		zero: newRamModule(0x80, nil)}
	return mc
}

type addressInterface interface {
	Uint16() uint16
}

type address uint16

func (u address) Uint16() uint16 { return uint16(u) }

func (mc mmu) String() string {
	return "<mmu>"
}

func (mc mmu) readByte(addr addressInterface) uint8 {
	a := addr.Uint16()
	if 0 <= a && a < 0x8000 {
		return mc.rom.readByte(address(a))
	}
	if 0x8000 <= a && a < 0xA000 {
		return mc.vram.readByte(address(a - 0x8000))
	}
	if 0xA000 <= a && a < 0xC000 { // switchable ram
		return mc.eram.readByte(address(a - 0xA000))
	}
	if 0xC000 <= a && a < 0xE000 {
		return mc.wram.readByte(address(a - 0xC000))
	}
	if 0xE000 <= a && a < 0xFE00 { // echo wram
		return mc.wram.readByte(address(a - 0xE000))
	}
	if 0xFE00 <= a && a < 0xFEA0 {
		return mc.oam.readByte(address(a - 0xFE00))
	}
	if 0xFF00 <= a && a < 0xFF4D {
		return mc.io.readByte(address(a - 0xFF00))
	}
	if 0xFF80 <= a && a <= 0xFFFF {
		return mc.zero.readByte(address(a - 0xFF80))
	}
	return 0
}

func (mc mmu) readWord(addr address) uint16 {
	l := mc.readByte(addr)
	h := mc.readByte(address(addr.Uint16() + 1))
	return bytesToWord(h, l)
}

func (mc mmu) writeByte(addr addressInterface, b uint8) {
	a := addr.Uint16()
	if 0 <= a && a < 0x8000 {
		mc.rom.writeByte(address(a), b)
	}
	if 0x8000 <= a && a < 0xA000 {
		mc.vram.writeByte(address(a-0x8000), b)
	}
	if 0xA000 <= a && a < 0xC000 { // switchable ram
		mc.eram.writeByte(address(a-0xA000), b)
	}
	if 0xC000 <= a && a < 0xE000 {
		mc.wram.writeByte(address(a-0xC000), b)
	}
	if 0xE000 <= a && a < 0xFE00 { // echo wram
		mc.wram.writeByte(address(a-0xE000), b)
	}
	if 0xFE00 <= a && a < 0xFEA0 {
		mc.oam.writeByte(address(a-0xFE00), b)
	}
	if 0xFF00 <= a && a < 0xFF4D {
		mc.io.writeByte(address(a-0xFF00), b)
	}
	if 0xFF80 <= a && a <= 0xFFFF {
		mc.zero.writeByte(address(a-0xFF80), b)
	}
}

func (mc mmu) writeWord(addr address, w uint16) {
	h, l := wordToBytes(w)
	mc.writeByte(addr, l)
	mc.writeByte(address(addr.Uint16()+1), h)
}
