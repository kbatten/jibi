package jibi

import ()

// A Gpu is the graphics processing unit. It handles drawing the background,
// window and sprites. It also triggers interrutps.
type Gpu struct {
	CommanderInterface

	// 0x0000-0x07FF tile set 1 0-127
	// 0x0800-0x0FFF tile set 1 128-255, set 0 (-1)-(-128)
	// 0x1000-0x17FF tile set 0 0-127
	// 0x1800-0x1BFF tile map 0
	// 0x1C00-0x1FFF tile map 1

	mmu MemoryDevice
	cpu MemoryCommander
	lcd Lcd
	clk chan ClockType

	bgBuffer []Byte // 256x256 background 2bit bitmap buffer
	fgBuffer []Byte // 144x160 foreground 2bit bitmap buffer

	// memory map
	charRam    []Byte
	bgTilemap0 []Byte
	bgTilemap1 []Byte
	oam        []Byte
	lcdc       Byte
	stat       Byte
	scy        Byte
	scx        Byte
	ly         Byte
	bgp        Byte
	obp0       Byte
	obp1       Byte
	wy         Byte
	wx         Byte

	// communication
	rwChan chan Byte

	// metrics
	frameCounters []*Clock
}

// NewGpu creates a Gpu and starts a goroutine.
func NewGpu(mmu *Mmu, cpu MemoryCommander, lcd Lcd, clk chan ClockType) *Gpu {
	commander := NewCommander("gpu")
	gpu := &Gpu{CommanderInterface: commander,
		mmu: mmu, cpu: cpu, lcd: lcd, clk: clk,
		bgBuffer:   make([]Byte, 256*256),
		fgBuffer:   make([]Byte, int(lcdWidth)*int(lcdHeight)),
		charRam:    make([]Byte, 0x1800),
		bgTilemap0: make([]Byte, 0x400),
		bgTilemap1: make([]Byte, 0x400),
		oam:        make([]Byte, 0xA0),
		rwChan:     make(chan Byte),
	}
	cmdHandlers := map[Command]CommandFn{
		CmdReadByteAt:   gpu.cmdReadByteAt,
		CmdWriteByteAt:  gpu.cmdWriteByteAt,
		CmdFrameCounter: gpu.cmdFrameCounter,
	}
	commander.start(gpu.stateScanlineOam, cmdHandlers, clk)
	mmu.HandleMemory(0x8000, 0x97FF, gpu)
	mmu.HandleMemory(0x9800, 0x9BFF, gpu)
	mmu.HandleMemory(0x9C00, 0x9FFF, gpu)
	mmu.HandleMemory(0xFE00, 0xFE9F, gpu)
	mmu.HandleMemory(AddrLCDC, AddrLCDC, gpu)
	mmu.HandleMemory(AddrSTAT, AddrSTAT, gpu)
	mmu.HandleMemory(AddrSCY, AddrSCY, gpu)
	mmu.HandleMemory(AddrSCX, AddrSCX, gpu)
	mmu.HandleMemory(AddrLY, AddrLY, gpu)
	mmu.HandleMemory(AddrBGP, AddrBGP, gpu)
	mmu.HandleMemory(AddrOBP0, AddrOBP0, gpu)
	mmu.HandleMemory(AddrOBP1, AddrOBP1, gpu)
	mmu.HandleMemory(AddrWY, AddrWY, gpu)
	mmu.HandleMemory(AddrWX, AddrWX, gpu)
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

func (g *Gpu) cmdFrameCounter(resp interface{}) {
	if resp, ok := resp.(chan chan ClockType); !ok {
		panic("invalid command response type")
	} else {
		clk := make(chan ClockType, 1)
		g.frameCounters = append(g.frameCounters, NewClock(clk))
		resp <- clk
	}
}

// ReadByteAt reads a single byte from the gpu at the specified address.
func (g *Gpu) ReadByteAt(addr Worder, b chan Byte) {
	req := ReadByteAtReq{addr.Word(), b}
	g.RunCommand(CmdReadByteAt, req)
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
	} else if 0x8000 <= a && a <= 0x97FF {
		return g.charRam[a-0x8000]
	} else if 0x9800 <= a && a <= 0x9BFF {
		return g.bgTilemap0[a-0x9800]
	} else if 0x9C00 <= a && a <= 0x9FFF {
		return g.bgTilemap1[a-0x9C00]
	} else if 0xFE00 <= a && a <= 0xFE9F {
		return g.oam[a-0xFE00]
	} else if AddrSTAT == a {
		return g.stat
	} else if AddrSCY == a {
		return g.scy
	} else if AddrSCX == a {
		return g.scx
	} else if AddrLY == a {
		return g.ly
	} else if AddrBGP == a {
		return g.bgp
	} else if AddrOBP0 == a {
		return g.obp0
	} else if AddrOBP1 == a {
		return g.obp1
	} else if AddrWY == a {
		return g.wy
	} else if AddrWX == a {
		return g.wx
	}
	g.yield()
	g.mmu.ReadByteAt(a, g.rwChan)
	return <-g.rwChan
}

func (g *Gpu) writeByte(addr Worder, b Byter) {
	a := addr.Word()
	if AddrLCDC == a {
		// if bit 7 is reset, pause the gpu
		// if bit 7 is set, play the gpu
		g.lcdc = b.Byte()
		if g.lcdc&0x80 == 0x80 {
			g.play()
		} else {
			g.pause()
			g.ly = 0
		}
	} else if 0x8000 <= a && a <= 0x97FF {
		g.charRam[a-0x8000] = b.Byte()
	} else if 0x9800 <= a && a <= 0x9BFF {
		g.bgTilemap0[a-0x9800] = b.Byte()
	} else if 0x9C00 <= a && a <= 0x9FFF {
		g.bgTilemap1[a-0x9C00] = b.Byte()
	} else if 0xFE00 <= a && a <= 0xFE9F {
		g.oam[a-0xFE00] = b.Byte()
	} else if AddrSTAT == a {
		g.stat = b.Byte()
	} else if AddrSCY == a {
		g.scy = b.Byte()
	} else if AddrSCX == a {
		g.scx = b.Byte()
	} else if AddrLY == a {
		g.ly = b.Byte()
	} else if AddrBGP == a {
		g.bgp = b.Byte()
	} else if AddrOBP0 == a {
		g.obp0 = b.Byte()
	} else if AddrOBP1 == a {
		g.obp1 = b.Byte()
	} else if AddrWY == a {
		g.wy = b.Byte()
	} else if AddrWX == a {
		g.wx = b.Byte()
	} else {
		g.yield()
		g.mmu.WriteByteAt(addr, b)
	}
}

/*
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
*/
func (g *Gpu) generateLine(line Byte) []Byte {
	// get background
	// TODO: bg wraps to the same X, not to X+1, same with Y
	offset := uint16(line+g.scy)*256 + uint16(g.scx)
	lbs := g.bgBuffer[offset : offset+uint16(lcdWidth)-1]
	// TODO: draw up to 10 sprites
	return lbs
}

type tile struct {
	bitmap []Byte // 2bpp bitmap
}

func newTile(tileData []Byte, palette []Byte) tile {
	bitmap := []Byte{}
	// 8x8 tiles
	// convert tile data into 2bpp bitmap
	addr := 0
	xMax := uint8(len(tileData) / 2)
	for yOff := uint8(0); yOff < 8; yOff++ {
		l := tileData[addr]
		h := tileData[addr+1]
		addr += 2

		for xOff := uint8(0); xOff < xMax; xOff++ {
			px := (((h >> (7 - xOff)) & 0x01) << 1) + (l>>(7-xOff))&0x01
			bitmap = append(bitmap, palette[px])
		}
	}
	return tile{bitmap}
}

func (t tile) Paint(buffer []Byte, x, y uint8) {
	width := uint16(0)
	if len(buffer) == 65536 {
		width = uint16(256)
	} else if len(buffer) == int(lcdWidth)*int(lcdHeight) {
		width = uint16(lcdWidth)
	} else {
		panic("unknown buffer type")
	}
	// TODO: sprite flags
	xMax := uint16(8)
	if len(t.bitmap) == 128 {
		xMax = 16
	}
	addr := 0
	for yOff := uint16(0); yOff < 8; yOff++ {
		for xOff := uint16(0); xOff < xMax; xOff++ {
			px := t.bitmap[addr]
			addr++
			buffer[uint16(x)+xOff+(uint16(y)+yOff)*width] = px
		}
	}
}

func (g *Gpu) getWinTiles(tilemap, tileset Byte, palette []Byte) []tile {
	addrTilemap := Word(0x9800)
	if tilemap == 1 {
		addrTilemap = 0x9C00
	}
	addrTileset := Word(0x8800)
	if tileset == 1 {
		addrTileset = 0x8000
	}

	tiles := []tile{}
	for t := Word(0x0000); t < 0x0400; t++ {
		tileData := make([]Byte, 16)
		tileInd := g.readByte(addrTilemap + t)
		addrTile := Word(0)
		if tileset == 0 {
			addrTile = addrTileset + Word(Byte(tileInd+0x80))*16
		} else {
			addrTile = addrTileset + Word(tileInd)*16
		}
		for i := Word(0); i < 16; i++ {
			tileData[i] = g.readByte(addrTile + i)
		}
		tiles = append(tiles, newTile(tileData, palette))
	}

	return tiles
}

func (g *Gpu) getBgTiles(tilemap, tileset Byte, palette []Byte) []tile {
	addrTilemap := Word(0x9800)
	if tilemap == 1 {
		addrTilemap = 0x9C00
	}
	addrTileset := Word(0x8800)
	if tileset == 1 {
		addrTileset = 0x8000
	}

	tiles := []tile{}
	for t := Word(0x0000); t < 0x0400; t++ {
		tileData := make([]Byte, 16)
		tileInd := g.readByte(addrTilemap + t)
		addrTile := Word(0)
		if tileset == 0 {
			addrTile = addrTileset + Word(Byte(tileInd+0x80))*16
		} else {
			addrTile = addrTileset + Word(tileInd)*16
		}
		for i := Word(0); i < 16; i++ {
			tileData[i] = g.readByte(addrTile + i)
		}
		tiles = append(tiles, newTile(tileData, palette))
	}

	return tiles
}

func byteToPalette(p Byte) []Byte {
	return []Byte{p & 0x03, p & 0x0C >> 2, p & 0x30 >> 4, p & 0xC0 >> 6}
}

func (g *Gpu) generateFrame() {
	// read in map, tileset data
	windowTilemap := (g.lcdc & 0x40) >> 6
	windowDisplay := g.lcdc&0x20 == 0x20
	bgTileset := (g.lcdc & 0x10) >> 4
	bgTilemap := (g.lcdc & 0x08) >> 3
	//objSpriteSize := (g.lcdc & 0x04) >> 2
	//objDisplay := g.lcdc&0x02 == 0x02
	bgWinDisplay := g.lcdc&0x01 == 0x01

	// draw background
	if bgWinDisplay {
		x := uint8(0)
		y := uint8(0)
		palette := byteToPalette(g.bgp)
		for _, bgtile := range g.getBgTiles(bgTilemap, bgTileset, palette) {
			bgtile.Paint(g.bgBuffer, x, y)
			x += 8
			if x == 0 {
				y += 8
			}
		}
	}
	if bgWinDisplay && windowDisplay {
		// TODO: this has to be handled line by line
		// wx is read on screen redraw and after a scan line interrupt
		// wy is read on screen redraw
		x := g.wx - 7
		y := g.wy
		palette := byteToPalette(g.bgp)
		for _, wintile := range g.getWinTiles(windowTilemap, bgTileset, palette) {
			wintile.Paint(g.fgBuffer, uint8(x), uint8(y))
			x += 8
			if x >= lcdWidth {
				y += 8
				x = 0
			}
		}
	}
	/*
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
	*/
}

func (g *Gpu) stateScanlineOam(first bool, t uint32) (CommanderStateFn, bool, uint32, uint32) {
	if first {
		g.stat = g.stat&0x7C | 0x2
	}
	if t >= 80 {
		t -= 80
		return g.stateScanlineVram, true, t, 172
	}
	return g.stateScanlineOam, false, t, 80
}

func (g *Gpu) stateScanlineVram(first bool, t uint32) (CommanderStateFn, bool, uint32, uint32) {
	if first {
		g.stat = g.stat&0x7C | 0x3
		g.lcd.DrawLine(g.generateLine(g.ly))
	}
	if t >= 172 {
		t -= 172
		return g.stateHblank, true, t, 204
	}
	if !first {
		panic("wasted gpu cycle")
	}
	return g.stateScanlineVram, false, t, 172
}

func (g *Gpu) stateHblank(first bool, t uint32) (CommanderStateFn, bool, uint32, uint32) {
	if first {
		g.stat = g.stat&0x7C | 0x1
	}
	if t >= 204 {
		t -= 204
		g.ly++
		if g.ly == lcdHeight-1 {
			return g.stateVblank, true, t, 456
		}
		return g.stateScanlineOam, true, t, 80
	}
	if !first {
		panic("wasted gpu cycle")
	}
	return g.stateHblank, false, t, 204
}

func (g *Gpu) stateVblank(first bool, t uint32) (CommanderStateFn, bool, uint32, uint32) {
	if first {
		g.stat = g.stat&0x7C | 0x0
		g.cpu.RunCommand(CmdSetInterrupt, InterruptVblank)
		g.yield()
		g.lcd.Blank()
		g.generateFrame()
		for _, clk := range g.frameCounters {
			clk.AddCycles(1)
		}
	}
	if t >= 456 {
		t -= 456
		g.ly++
		if g.ly > lcdHeight-1+10 {
			g.ly = 0
			return g.stateScanlineOam, true, t, 80
		}
		return g.stateVblank, false, t, 456
	}
	if !first {
		panic("wasted gpu cycle")
	}
	return g.stateVblank, false, t, 456
}
