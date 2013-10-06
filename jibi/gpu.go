package jibi

import ()

// A Gpu is the graphics processing unit. It handles drawing the background,
// window and sprites. It also triggers interrutps.
type Gpu struct {
	Commander

	// 0x0000-0x07FF tile set 1 0-127
	// 0x0800-0x0FFF tile set 1 128-255, set 0 (-1)-(-128)
	// 0x1000-0x17FF tile set 0 0-127
	// 0x1800-0x1BFF tile map 0
	// 0x1C00-0x1FFF tile map 1

	mmu MemoryDevice
	lcd Lcd
	irq *Irq
	clk chan ClockType

	frameBuffer []Byte

	lcdc Byte
}

// NewGpu creates a Gpu and starts a goroutine.
func NewGpu(mmu MemoryCommander, irq *Irq, lcd Lcd, clk chan ClockType) *Gpu {
	commander := NewCommander("gpu")
	gpu := &Gpu{commander,
		mmu, lcd, irq, clk, make([]Byte, 256*256),
		Byte(0),
	}
	cmdHandlers := map[Command]CommandFn{
		CmdReadByteAt:  gpu.cmdReadByteAt,
		CmdWriteByteAt: gpu.cmdWriteByteAt,
	}
	commander.Start(gpu.stateScanlineOam, cmdHandlers, nil)
	handler := MemoryHandlerRequest{
		AddrLCDC, AddrLCDC, gpu,
	}
	mmu.RunCommand(CmdHandleMemory, handler)
	return gpu
}

func (g *Gpu) cmdReadByteAt(resp interface{}) {
	if req, ok := resp.(ReadByteAtReq); !ok {
		panic("invalid command response type")
	} else {
		req.b <- g.readByte(req.addr)
	}
}

func (g *Gpu) cmdWriteByteAt(resp interface{}) {
	if req, ok := resp.(WriteByteAtReq); !ok {
		panic("invalid command response type")
	} else {
		g.writeByte(req.addr, req.b)
	}
}

// ReadByteAt reads a single byte from the gpu at the specified address.
func (g *Gpu) ReadByteAt(addr Worder) Byte {
	req := ReadByteAtReq{addr.Word(), make(chan Byte)}
	g.RunCommand(CmdReadByteAt, req)
	return <-req.b
}

// WriteByteAt writes a single byte to the gpu at the specified address.
func (g *Gpu) WriteByteAt(addr Worder, b Byter) {
	req := WriteByteAtReq{addr.Word(), b.Byte()}
	g.RunCommand(CmdWriteByteAt, req)
}

func (g *Gpu) readByte(addr Worder) Byte {
	a := addr.Word()
	if AddrLCDC == a {
		return g.lcdc
	}
	return g.mmu.ReadByteAt(addr)
}

func (g *Gpu) writeByte(addr Worder, b Byter) {
	a := addr.Word()
	if AddrLCDC == a {
		g.lcdc = b.Byte()
	} else {
		g.mmu.WriteByteAt(addr, b)
	}
}

func paintTile(frameBuffer []Byte, tileData []Byte, x, y uint8, above, xflip, yflip bool, palette Byte) {
	addr := 0
	// convert tile data into 2bpp bitmap
	for yOff := uint8(0); yOff < 8; yOff++ {
		yInd := (uint16(y) + uint16(yOff)) * uint16(256)
		l := tileData[addr]
		h := tileData[addr+1]
		addr += 2

		for xOff := uint8(0); xOff < 8; xOff++ {
			px := (((h >> (7 - xOff)) & 0x01) << 1) + (l>>(7-xOff))&0x01
			ind := uint16(x) + uint16(xOff) + yInd
			if uint32(ind) < uint32(len(frameBuffer)) {
				frameBuffer[ind] = px
			}
		}
	}
}

func (g *Gpu) generateLine(line Byte) []Byte {
	scrollX := g.readByte(AddrSCX)
	scrollY := g.readByte(AddrSCY)
	offset := uint16(line+scrollY)*256 + uint16(scrollX)
	return g.frameBuffer[offset : offset+uint16(lcdWidth)-1]
}

func (g *Gpu) generateFrame() {
	for i := range g.frameBuffer {
		g.frameBuffer[i] = 0
	}
	// read in map, tileset data
	windowTilemap := (g.lcdc & 0x40) >> 6
	windowDisplay := g.lcdc&0x20 == 0x20
	bgTileset := (g.lcdc & 0x10) >> 4
	bgTilemap := (g.lcdc & 0x08) >> 3
	bgSpriteSize := (g.lcdc & 0x04) >> 2
	objDisplay := g.lcdc&0x02 == 0x02
	bgDisplay := g.lcdc&0x01 == 0x01

	// draw background
	if bgDisplay {
		addrTilemap := Word(0x9800)
		if bgTilemap == 1 {
			addrTilemap = 0x9C00
		}
		addrTileset := Word(0x8800)
		if bgTileset == 1 {
			addrTileset = 0x8000
		}
		x := uint8(0)
		y := uint8(0)
		tileData := make([]Byte, 16)
		for tile := Word(0x0000); tile < 0x0400; tile += 16 {
			ind := Word(0)
			if bgTileset == 0 {
				tileInd := int8(g.readByte(addrTilemap + tile))
				ind = Word(int32(addrTileset) + int32(tileInd)*16)
			} else {
				tileInd := g.readByte(addrTilemap + tile)
				ind = addrTileset + Word(tileInd)*16
			}
			for j := Word(0); j < 16; j++ {
				tileData[j] = g.readByte(ind + j)
			}
			paintTile(g.frameBuffer, tileData, x, y, false, false, false, 2)
			x += 8
			if x == 0 {
				y += 8
			}
		}

		// draw window
		if windowDisplay {
			addrTilemap := Word(0x9800)
			if windowTilemap == 1 {
				addrTilemap = 0x9C00
			}
			addrTileset := Word(0x8800)
			if bgTileset == 1 {
				addrTileset = 0x8000
			}

			x := uint8(0)
			y := uint8(0)
			for tile := Word(0x0000); tile < 0x0400; tile += 16 {
				ind := Word(0)
				if bgTileset == 0 {
					tileInd := int8(g.readByte(addrTilemap + tile))
					ind = Word(int32(addrTileset) + int32(tileInd)*16)
				} else {
					tileInd := g.readByte(addrTilemap + tile)
					ind = addrTileset + Word(tileInd)*16
				}
				for j := Word(0); j < 16; j++ {
					tileData[j] = g.readByte(ind + j)
				}
				paintTile(g.frameBuffer, tileData, x, y, false, false, false, 2)
				x += 8
				if x == 0 {
					y += 8
				}
			}
		}
	}

	// draw sprites (oam)
	if objDisplay {
		addrTileset := Word(0x8000)
		oamaddr := Word(0xFE00)
		spriteSize := Word(16)
		if bgSpriteSize == 1 {
			spriteSize = 32
		}
		tileData := make([]Byte, spriteSize)

		for sprite := Word(0x0000); sprite < 0x00A0; sprite += 4 {
			x := uint8(g.readByte(oamaddr+sprite) - 8)
			y := uint8(g.readByte(oamaddr+sprite+1) - 16)
			tile := g.readByte(oamaddr + sprite + 2)
			if bgSpriteSize == 1 {
				tile = tile & 0xEF
			}
			attr := g.readByte(oamaddr + sprite + 3)
			above := attr&0x80 == 0
			yflip := attr&0x40 == 0x80
			xflip := attr&0x20 == 0x40
			palette := attr & 0x10 >> 1
			ind := addrTileset + Word(tile)*spriteSize
			for j := Word(0); j < spriteSize; j++ {
				tileData[j] = g.readByte(ind + j)
			}
			paintTile(g.frameBuffer, tileData, x, y, above, xflip, yflip, palette)
		}
	}
}

func (g *Gpu) stateScanlineOam(first bool, t uint32) (CommanderStateFn, bool, uint32, uint32) {
	if t >= 80 {
		t -= 80
		return g.stateScanlineVram, true, t, 172
	}
	return g.stateScanlineOam, false, t, 80
}

func (g *Gpu) stateScanlineVram(first bool, t uint32) (CommanderStateFn, bool, uint32, uint32) {
	if first {
		curline := g.readByte(AddrLY)
		g.lcd.DrawLine(g.generateLine(curline))
	}
	if t >= 172 {
		t -= 172
		return g.stateHblank, true, t, 204
	}
	return g.stateScanlineVram, false, t, 172
}

func (g *Gpu) stateHblank(first bool, t uint32) (CommanderStateFn, bool, uint32, uint32) {
	if t >= 204 {
		t -= 204
		curline := g.readByte(AddrLY) + 1
		g.writeByte(AddrLY, curline)
		if curline == lcdHeight-1 {
			return g.stateVblank, true, t, 456
		}
		return g.stateScanlineOam, true, t, 80
	}
	return g.stateHblank, false, t, 204
}

func (g *Gpu) stateVblank(first bool, t uint32) (CommanderStateFn, bool, uint32, uint32) {
	if first {
		g.irq.SetInterrupt(InterruptVblank)
		g.lcd.Blank()
		g.generateFrame()
		panic("vblank")
	}
	if t >= 456 {
		t -= 456
		curline := g.readByte(AddrLY) + 1
		if curline > lcdHeight-1+10 {
			g.writeByte(AddrLY, Byte(0))
			return g.stateScanlineOam, true, t, 80
		}
		g.writeByte(AddrLY, Byte(curline))
	}
	return g.stateVblank, false, t, 456
}
