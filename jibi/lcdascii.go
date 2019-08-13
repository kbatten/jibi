package jibi

import (
	"fmt"
)

const (
	lcdWidth  Byte = 160
	lcdHeight Byte = 144
)

type Lcd interface {
	Init()
	Close()
	DrawLine(bl []Byte)
	Blank()
	DisableRender()
}

// An LcdASCII outputs as ascii characters to the terminal.
type LcdASCII struct {
	dr           bool
	prevLine     []Byte
	lineIndex    uint8
	prevDrawLine uint8
}

func NewLcd() Lcd {
	return &LcdASCII{}
}

func (lcd *LcdASCII) Init() {
	fmt.Printf("\x1B[?25l") // hide the cursor
}

func (lcd *LcdASCII) Close() {
	fmt.Printf("\x1B[?25h") // show the cursor
}

// DrawLine draws the Byte Slice to the current line index, then advances the
// index.
func (lcd *LcdASCII) DrawLine(bl []Byte) {
	blO := make([]Byte, len(bl))
	copy(blO, bl)
	drawLine := uint8(float64(lcd.lineIndex) * 50.0 / float64(lcdHeight))
	if drawLine == lcd.prevDrawLine && lcd.lineIndex != 0 {
		// compress previous line and this line into one
		for i := range bl {
			bl[i] = bl[i] | (lcd.prevLine[i]&0xC0 | (lcd.prevLine[i]&0x03)<<2)
		}
	}

	ls := make([]byte, lcdWidth)
	var o byte
	for i, c := range bl {
		o = ' '
		if c == 1 {
			o = ' ' // 0001
		} else if c == 2 {
			o = '.' // 0010
		} else if c == 3 {
			o = '.' // 0011
		} else if c == 4 {
			o = ' ' // 0100
		} else if c == 5 {
			o = ' ' // 0101
		} else if c == 6 {
			o = '.' //0110
		} else if c == 7 {
			o = '.' // 0111
		} else if c == 8 {
			o = '\'' // 1000
		} else if c == 9 {
			o = '\'' // 1001
		} else if c == 10 {
			o = ':' // 1010
		} else if c == 11 {
			o = ':' // 1011
		} else if c == 12 {
			o = '\'' // 1100
		} else if c == 13 {
			o = '\'' // 1101
		} else if c == 14 {
			o = ':' // 1110
		} else if c == 15 {
			o = ':' // 1111
		}
		ls[i] = o
	}
	if lcd.dr == false {
		fmt.Printf("\x1B[%d;H%s", drawLine, ls)
	}

	lcd.prevDrawLine = drawLine
	lcd.prevLine = blO
	lcd.lineIndex++
}

// Blank moves the cursor to the upper left.
func (lcd *LcdASCII) Blank() {
	/*
		if lcd.dr == false {
			fmt.Print("\x1B[2J")
		}
	*/
	lcd.lineIndex = 0
}

// DisableRender turns off rendering of lines. Only use while Paused.
func (lcd *LcdASCII) DisableRender() {
	lcd.dr = true
}
