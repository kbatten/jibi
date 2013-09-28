package main

import (
	"testing"
)

func TestRomModule(t *testing.T) {
	r := newRomModule(0x10, []uint8{0xF0, 0xAA})
	if r.readByte(Uint16(0x0000)) != 0xF0 {
		t.Error("incorrect read")
	}
	if r.readByte(Uint16(0x0001)) != 0xAA {
		t.Error("incorrect read")
	}
	r.writeByte(Uint16(0x0002), 0xBB)
	if r.readByte(Uint16(0x0002)) != 0x00 {
		t.Error("incorrect read")
	}
}

func TestRamModule(t *testing.T) {
    r := newRamModule(0x10, []uint8{0xF0, 0xAA})
    if r.readByte(Uint16(0x0000)) != 0xF0 {
        t.Error("incorrect read")
    }
    if r.readByte(Uint16(0x0001)) != 0xAA {
        t.Error("incorrect read")
    }
    r.writeByte(Uint16(0x0002), 0xBB)
    if r.readByte(Uint16(0x0002)) != 0xBB {
        t.Error("incorrect read")
    }
    if r.readByte(Uint16(0x0003)) != 0x00 {
        t.Error("incorrect read")
    }
}
