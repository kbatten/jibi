package main

import (
	"fmt"
)

type cartridge struct {
	rom []uint8

	eram memoryDevice

	// rom info
	name    string
	color   bool
	super   bool
	ct      cartridgeType
	romSize cartridgeRomSize
	ramSize cartridgeRamSize
}

func newCartridge(rom []uint8) cartridge {
	romBanks := make([]uint8, 0x8000)
	copy(romBanks, rom)

	eram := newRamModule(0x2000, nil)

	name := ""
	for _, c := range rom[0x0134 : 0x0142+1] {
		if c == 0 {
			break
		}
		name += string(c)
	}
	color := rom[0x0143] == 0x80
	super := rom[0x0146] == 0x03
	ct := cartridgeType(rom[0x0147])
	romSize := cartridgeRomSize(rom[0x0148])
	ramSize := cartridgeRamSize(rom[0x0149])
	return cartridge{romBanks, eram, name, color, super, ct, romSize, ramSize}
}

func (c cartridge) readByte(addr addressInterface) uint8 {
	return c.rom[addr.Uint16()]
}

func (c cartridge) writeByte(addr addressInterface, n uint8) {
	// rom only
}

func (c cartridge) String() string {
	return fmt.Sprintf(`name: %s
romSize: %s (%d)
ramSize: %s
color: %v
super: %v
type: %s`, c.name, c.romSize, len(c.rom), c.ramSize, c.color, c.super, c.ct)
}

type cartridgeType uint8

func (ct cartridgeType) String() string {
	switch ct {
	case 0x00:
		return "00-ROM"
	case 0x01:
		return "01-ROM+MBC1"
	case 0x02:
		return "02-ROM+MBC1+RAM"
	case 0x03:
		return "03-ROM+MBC1+RAM+BATT"
	case 0x05:
		return "05-ROM+MBC2"
	case 0x06:
		return "06-ROM+MBC2+BATT"
	case 0x08:
		return "08-ROM+RAM"
	case 0x09:
		return "09-ROM+RAM+BATT"
	case 0x0B:
		return "0B-ROM+MMMO1"
	case 0x0C:
		return "0C-ROM+MMMO1+SRAM"
	case 0x0D:
		return "0D-ROM+MMMO1+SRAM+BATT"
	case 0x0F:
		return "0F-ROM+MBC3+TIMER+BATT"
	case 0x10:
		return "10-ROM+MBC3+TIMER+RAM+BATT"
	case 0x11:
		return "11-ROM+MBC3"
	case 0x12:
		return "12-ROM+MBC3+RAM"
	case 0x13:
		return "13-ROM+MBC3+RAM+BATT"
	case 0x19:
		return "19-ROM+MBC5"
	case 0x1A:
		return "1A-ROM+MBC5+RAM"
	case 0x1B:
		return "1B-ROM+MBC5+RAM+BATT"
	case 0x1C:
		return "1C-ROM+MBC5+RUMBLE"
	case 0x1D:
		return "1D-ROM+MBC5+RUMBLE+SRAM"
	case 0x1E:
		return "1E-ROM+MBC5+RUMBLE+SRAM+BATT"
	case 0x1F:
		return "1F-PocketCamera"
	case 0xFD:
		return "FD-BandaiTAMA5"
	case 0xFE:
		return "FE-HudsonHuC_3"
	default:
		return fmt.Sprintf("%0X-UNKNOWN", ct)
	}
}

type cartridgeRomSize uint8

func (cs cartridgeRomSize) String() string {
	switch cs {
	case 0x00:
		return "00-256Kbit,32KByte,2banks"
	case 0x01:
		return "01-512Kbit,64KByte,4banks"
	case 0x02:
		return "02-1Mbit,128KByte,8banks"
	case 0x03:
		return "03-2Mbit,256KByte,16banks"
	case 0x04:
		return "04-4Mbit,512KByte,32banks"
	case 0x05:
		return "05-8Mbit,1MByte,64banks"
	case 0x06:
		return "06-16Mbit,2MByte,128banks"
	case 0x52:
		return "52-9Mbit,1.1MByte,72banks"
	case 0x53:
		return "53-10Mbit,1.2MByte,80banks"
	case 0x54:
		return "54-12Mbit,1.5MByte,96banks"
	default:
		return fmt.Sprintf("%0X-UNKNOWN")
	}
}

type cartridgeRamSize uint8

func (cs cartridgeRamSize) String() string {
	switch cs {
	case 0x00:
		return "00-None"
	case 0x01:
		return "01-16kBit,2kByte,1bank"
	case 0x02:
		return "02-64kBit,8kByte,1bank"
	case 0x03:
		return "03-256kBit,32kByte,4banks"
	case 0x04:
		return "04-1MBit,128kByte,16banks"
	default:
		return fmt.Sprintf("%0X-UNKNOWN")
	}
}
