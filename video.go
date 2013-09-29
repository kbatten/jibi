package main

import (
	"fmt"
)

type video struct {
	ram memoryDevice

	oam memoryDevice

	frameBuff []uint8 // uint2

	mode *uint8
	line *int16
	t    *uint32
}

func newVideo() video {
	oam := newRamModule(0xA0, nil)
	return video{newRamModule(0x2000, nil), oam, make([]uint8, 160*144),
		new(uint8), new(int16), new(uint32)}
}

func (v video) readByte(addr addressInterface) uint8 {
	//fmt.Println("read:", addr, v.ram.readByte(addr))
	return v.ram.readByte(addr)
}

func (v video) writeByte(addr addressInterface, n uint8) {
	//fmt.Printf("write: 0x%04X 0x%02X\n", 0x8000+addr.Uint16(), n)
	v.ram.writeByte(addr, n)
}

func (v video) drawLine() {
	line := ""
	for x := int16(0); x < 160; x++ {
		y := *v.line * 160
		c := v.frameBuff[y+x]
		o := " "
		if c == 1 {
			o = "."
		} else if c == 2 {
			o = ":"
		} else if c == 3 {
			o = "#"
		}
		line += o
	}
	if *v.line %2 == 0 {
		fmt.Println(line)
	}
}

func (v video) paintTile(palette uint8, tile uint8, x, y uint8) {
	addr := uint16(0x0000)
	if palette == 0 {
		addr = 0x1C00
		addr += uint16(tile) * 16
	} else if palette == 1 {
		addr = 0x180
		if int(tile) < 0 {
			addr -= uint16(-int(tile)) * 16
		} else {
			addr += uint16(int(tile)) * 16
		}
	}
	for line := uint8(0); line < 8; line++ {
		a := v.ram.readByte(address(addr))
		b := v.ram.readByte(address(addr + 1))

		xOff := uint8(0)
		yOff := line * 160
		v.frameBuff[x+y+xOff+yOff] = a >> 6
		xOff++
		v.frameBuff[x+y+xOff+yOff] = a >> 4 & 0x04
		xOff++
		v.frameBuff[x+y+xOff+yOff] = a >> 2 & 0x04
		xOff++
		v.frameBuff[x+y+xOff+yOff] = a & 0x04
		xOff++
		v.frameBuff[x+y+xOff+yOff] = b >> 6
		xOff++
		v.frameBuff[x+y+xOff+yOff] = b >> 4 & 0x04
		xOff++
		v.frameBuff[x+y+xOff+yOff] = b >> 2 & 0x04
		xOff++
		v.frameBuff[x+y+xOff+yOff] = b & 0x04
		xOff++
		addr += 2
	}
}

func (v video) blank() {
	// move to 0,0
	fmt.Print("\x1B[H")

	// paint background to frame buffer

	// paint sprites to frame buffer
	for i := uint8(0); i < 40; i++ {
		so := []uint8{
			v.oam.readByte(address(i * 4)),
			v.oam.readByte(address(i*4 + 1)),
			v.oam.readByte(address(i*4 + 2)),
			v.oam.readByte(address(i*4 + 3)),
		}
		y := so[0] + 16
		x := so[1] + 8
		tile := so[2]
		palette := (so[3] & 0x10) >> 4
		v.paintTile(palette, tile, x, y)
	}
}

func (v video) step(t uint8) {
	*v.t += uint32(t)
	//	fmt.Println(*v.mode, *v.t, t)
	switch *v.mode {
	case 2: // scanline
		if *v.t >= 80 {
			*v.t -= 80
			*v.mode = 3
		}
	case 3: // scanline
		if *v.t >= 172 {
			*v.t -= 172
			*v.mode = 0

			// draw line
			v.drawLine()
		}
	case 0: // hblank
		if *v.t >= 204 {
			*v.t -= 204
			*v.line++
			if *v.line == 143 {
				*v.mode = 1 // end of last line
			} else {
				*v.mode = 2 // end of line
			}
		}
	case 1: // vblank 10 lines
		if *v.t >= 456 {
			*v.t -= 456
			*v.line -= 10
			if *v.line <= 0 {
				v.blank()
				*v.line = 0
				*v.mode = 2
			}
		}
	}
}

func (v video) String() string {
	return fmt.Sprintf("<video>")
}
