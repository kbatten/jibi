package main

import (
	"fmt"
)

type video struct {
	// 0x0000-0x07FF tile set 1 0-127
	// 0x0800-0x0FFF tile set 1 128-255, set 0 (-1)-(-128)
	// 0x1000-0x17FF tile set 0 0-127
	// 0x1800-0x1BFF tile map 0
	// 0x1C00-0x1FFF tile map 1
	ram memoryDevice
	// 0x00-0xA0
	oam memoryDevice

	frameBuff []uint8 // uint2 256x256

	mode *uint8
	t    *uint32

	width  uint8
	height uint8
}

func newVideo() video {
	oam := newRamModule(0xA0, nil)
	width := uint8(160)
	height := uint8(144)
	return video{newRamModule(0x2000, nil), oam, make([]uint8, 65536),
		new(uint8), new(uint32), width, height}
}

func (v video) readByte(addr addressInterface) uint8 {
	return v.ram.readByte(addr)
}

func (v video) writeByte(addr addressInterface, n uint8) {
	v.ram.writeByte(addr, n)
}

// TODO: don't use framebuffer, dynamically build the line at drawtime
func (v video) drawLine() {
	scrollX := v.oam.readByte(address(0x02))
	scrollY := v.oam.readByte(address(0x03))

	curline := v.oam.readByte(address(4))
	line := ""
	yInd := (uint16(scrollY) + uint16(curline)) * uint16(256)
	for x := uint8(0); x < v.width; x++ {
		c := v.frameBuff[uint16(x)+uint16(scrollX)+yInd]
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
	if curline < 120 {
		if curline%2 == 0 {
			fmt.Printf("\x1B[160D%s", line)
		} else {
			fmt.Printf("\x1B[160D%s\n", line)
		}
	}
}

func (v video) paintTile(tileData []uint8, x, y uint8) {
	addr := 0
	// convert tile data into 2bpp bitmap
	for yOff := uint8(0); yOff < 8; yOff++ {
		yInd := (uint16(y) + uint16(yOff)) * uint16(256)
		l := tileData[addr]   //v.ram.readByte(address(addr))
		h := tileData[addr+1] //v.ram.readByte(address(addr + 1))
		addr += 2

		for xOff := uint8(0); xOff < 8; xOff++ {
			px := (((h >> (7 - xOff)) & 0x01) << 1) + (l>>(7-xOff))&0x01
			ind := uint16(x) + uint16(xOff) + yInd
			if uint32(ind) < uint32(len(v.frameBuff)) {
				v.frameBuff[ind] = px
			}
		}
	}
}

func (v video) paint() {
	//tileData = []uint8{
	//	0x7C, 0x7C, 0x00, 0xC6, 0xC6, 0x00, 0x00, 0xFE,
	//	0xC6, 0xC6, 0x00, 0xC6, 0xC6, 0x00, 0x00, 0x00,
	//}
	//v.paintTile(tileData, 16, 16)

	// background
	// tile map 0 0x1800-0x1BFF
	// tile set 1 0x0000

	// tile map 1 0x1C00
	//x = 0
	//y = 0
	//for i := 0x0000; i < 0x0400; i += 16 {
	//	tileInd := v.ram.readByte(address(0x1800 + i))
	//	ind := uint16(tileInd) * 16
	//	for j := uint16(0); j < 16; j++ {
	//		tileData[j] = v.ram.readByte(address(ind + j))
	//	}
	//	v.paintTile(tileData, x, y)
	//	x += 8
	//	if x >= v.width {
	//		x = 0
	//		y += 8
	//	}
	//}
/*
	lcdCtrl := v.oam.readByte(address(0))
	ctrlBackground := lcdCtrl&0x01 == 0x01
	ctrlSprites := lcdCtrl&0x02 == 0x02
	ctrlSpriteSize := lcdCtrl&0x04 == 0x04
	ctrlBgTileMap := lcdCtrl&0x08 == 0x08
	ctrlBgTileSet := lcdCtrl&0x10 == 0x10
	ctrlWindow := lcdCtrl&0x20 == 0x20
	ctrlWindowTileMap := lcdCtrl&0x40 == 0x40
	ctrlDisplay := lcdCtrl&0x80 == 0x80
*/
	// update frameBuffer to handle two verticle pixels per line
	// doesn't work for odd values of scrollY
	// only b&w, not grayscale
	for y := uint16(1); y < 256; y += 2 {
		for x := uint16(0); x < 256; x++ {
			botInd := uint16(y)*uint16(256) + uint16(x)
			upInd := uint16(y-1)*uint16(256) + uint16(x)
			v.frameBuff[botInd] = v.frameBuff[botInd]<<2 + v.frameBuff[upInd]
		}
	}
}

func (v video) blank() {
	// move to 0,0
	fmt.Print("\x1B[H")
	for i := range v.frameBuff {
		v.frameBuff[i] = 0
	}
}

func (v video) step(t uint8) {
	lcdCtrl := v.oam.readByte(address(0))
	// turn on lcd
	lcdCtrl |=0x80
	v.oam.writeByte(address(0), lcdCtrl)

	*v.t += uint32(t)
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
			curline := v.oam.readByte(address(4)) + 1
			v.oam.writeByte(address(4), curline)
			if curline == v.height {
				*v.mode = 1 // end of last line
			} else {
				*v.mode = 2 // end of line
			}
		}
	case 1: // vblank 10 lines
		if *v.t >= 456 {
			*v.t -= 456
			curline := v.oam.readByte(address(4)) - 10
			v.oam.writeByte(address(4), curline)
			if curline >= v.height { //underflow
				v.blank()
				v.paint()
				v.oam.writeByte(address(4), 0)
				*v.mode = 2
			}
		}
	}
}

func (v video) String() string {
	return fmt.Sprintf("<video>")
}
