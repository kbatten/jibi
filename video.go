package main

import (
	"fmt"
)

type video struct {
	// 0x0000-0x07FF tile set 1 0-127
	// 0x0800-0x0FFF tile set 1 128-255, set 0 (-1)-(-128)
	// 0x1000-0x17FF tile set 0 0-127
	// 0x1800-0x1BFF tile map 0
	// 0x1C00-0x1FFF tile map 2
	ram memoryDevice
	// 0x00-0xA0
	oam memoryDevice

	frameBuff []uint8 // uint2

	mode *uint8
	line *uint8 // TODO: use oam memory
	t    *uint32

	width  uint8
	height uint8
}

func newVideo() video {
	oam := newRamModule(0xA0, nil)
	width := uint8(160)
	height := uint8(144)
	return video{newRamModule(0x2000, nil), oam, make([]uint8,
		uint16(width)*uint16(height)),
		new(uint8), new(uint8), new(uint32), width, height}
}

func (v video) readByte(addr addressInterface) uint8 {
	//fmt.Println("read:", addr, v.ram.readByte(addr))
	return v.ram.readByte(addr)
}

func (v video) writeByte(addr addressInterface, n uint8) {
	//fmt.Printf("write: 0x%04X 0x%02X\n", 0x8000+addr.Uint16(), n)
	v.ram.writeByte(addr, n)
}

// TODO: don't use framebuffer, dynamically build the line at drawtime
func (v video) drawLine() {
	line := ""
	yInd := uint16(*v.line) * uint16(v.width)
	for x := uint8(0); x < v.width; x++ {
		c := v.frameBuff[uint16(x)+yInd]
		// half height pixes don't use grayscale
		o := " "
		if c == 1 {
			o = "'" // 0001
		} else if c == 2 {
			o = "'" // 0010
		} else if c == 3 {
			o = "'" // 0011
		} else if c == 4 {
			o = "." // 0100
		} else if c == 5 {
			o = ":" // 0101
		} else if c == 6 {
			o = ":" //0110
		} else if c == 7 {
			o = ":" // 0111
		} else if c == 8 {
			o = "." // 1000
		} else if c == 9 {
			o = ":" // 1001
		} else if c == 10 {
			o = ":" // 1010
		} else if c == 11 {
			o = ":" // 1011
		} else if c == 12 {
			o = "." // 1100
		} else if c == 13 {
			o = ":" // 1101
		} else if c == 14 {
			o = ":" // 1110
		} else if c == 15 {
			o = ":" // 1111
		}

		line += o
	}
	if *v.line < 120 {
		if *v.line%2 == 0 {
			fmt.Print("\x1B[160D", line)
		} else {
			fmt.Println("\x1B[160D", line)
		}
	}
}

func (v video) paintTile(tileData []uint8, x, y uint8) {
	addr := 0
	// convert tile data into 2bpp bitmap
	for yOff := uint8(0); yOff < 8; yOff++ {
		yInd := (uint16(y) + uint16(yOff)) * uint16(v.width)
		l := tileData[addr]   //v.ram.readByte(address(addr))
		h := tileData[addr+1] //v.ram.readByte(address(addr + 1))
		addr += 2

		for xOff := uint8(0); xOff < 8; xOff++ {
			px := (((h >> (7 - xOff)) & 0x01) << 1) + (l>>(7-xOff))&0x01
			ind := uint16(x) + uint16(xOff) + yInd
			if ind < uint16(len(v.frameBuff)) {
				v.frameBuff[ind] = px
			}
		}
	}
}

func (v video) blank() {
	// move to 0,0
	fmt.Print("\x1B[H")

	//v.paintTile([]uint8{0x7C, 0x7C, 0x00, 0xC6, 0xC6, 0x00, 0x00, 0xFE,
	//	0xC6, 0xC6, 0x00, 0xC6, 0xC6, 0x00, 0x00, 0x00}, 4, 4)

	tileData := make([]uint8, 16)
	for buffer := 0; buffer < 3; buffer++ {
		addr := uint16(0x0000)
		if buffer == 1 {
			addr = 0x0800
		} else if buffer == 2 {
			addr = 0x1000
		}
		for tile := uint16(0); tile < 6; tile++ {
			for i := 0; i < 16; i++ {
				tileData[i] = v.ram.readByte(address(addr + tile*16))
			}
		}
	}

	// update frameBuffer to handle two verticle pixels per line
	for y := uint8(1); y < v.height; y += 2 {
		for x := uint8(0); x < v.width; x++ {
			botInd := uint16(y)*uint16(v.width) + uint16(x)
			upInd := uint16(y-1)*uint16(v.width) + uint16(x)
			v.frameBuff[botInd] = v.frameBuff[botInd]<<2 + v.frameBuff[upInd]
		}
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
			if *v.line == v.height {
				*v.mode = 1 // end of last line
			} else {
				*v.mode = 2 // end of line
			}
		}
	case 1: // vblank 10 lines
		if *v.t >= 456 {
			*v.t -= 456
			*v.line -= 10
			if *v.line >= v.height { //underflow
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
