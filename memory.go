package main

import ()

type memoryController struct {
	addr chan uint16
	data chan uint8

	rom  memoryDevice // 32k (only first four banks)
	vram memoryDevice // 8k
	eram memoryDevice // 8k (only first bank)
	wram memoryDevice // 8k
	iram memoryDevice // 8k
}

type memoryDevice interface {
	readByte(addressInterface) uint8
	writeByte(addressInterface, uint8)
}

// TODO: add support for banks
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

// TODO: add support for banks
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

func newMemoryController(rom []uint8) memoryController {
	mc := memoryController{
		rom:  newRomModule(0x8000, rom),
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

func (mc memoryController) readByte(addr addressInterface) uint8 {
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

func (mc memoryController) writeByte(addr addressInterface, b uint8) {
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

func (mc *memoryController) readWord(addr address) uint16 {
	l := mc.readByte(addr)
	h := mc.readByte(address(addr.Uint16() + 1))
	return uint16(h)<<8 + uint16(l)
}
