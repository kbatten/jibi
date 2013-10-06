package jibi

import (
	"fmt"
)

const (
	lcdWidth  Byte = 160
	lcdHeight Byte = 144
)

// An Lcd is an interface the Gpu uses to communicate with the display.
type Lcd interface {
	DrawLine([]Byte)
	Blank()
	DisableRender()
}

// An LcdASCII outputs as ascii characters to the terminal.
type LcdASCII struct {
	dr         bool
	prevLine   []Byte
	lineIndex  uint8
	prevLineID uint8
}

// NewLcdASCII returns an LcdASCII object.
func NewLcdASCII() *LcdASCII {
	return &LcdASCII{}
}

// DrawLine draws the Byte Slice to the current line index, then advances the
// index.
func (lcd *LcdASCII) DrawLine(bl []Byte) {
	// compress every two lines into 1
	if lcd.lineIndex%2 == 1 {
		for i := range bl {
			bl[i] = bl[i]<<2 + lcd.prevLine[i]
		}
	}
	ls := ""
	for _, c := range bl {
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
		ls += o
	}

	if lcd.dr == false {
		lineID := uint8(float64(lcd.lineIndex) * 120.0 / float64(lcdHeight))
		if lcd.prevLineID != lineID || lineID == 0 {
			if lcd.lineIndex%2 == 0 {
				fmt.Printf("\x1B[170D%s", ls)
			} else {
				fmt.Printf("\x1B[170D%s\n", ls)
			}
		}
		lcd.prevLineID = lineID
	}

	lcd.prevLine = bl
	lcd.lineIndex++
}

// Blank moves the cursor to the upper left.
func (lcd *LcdASCII) Blank() {
	if lcd.dr == false {
		fmt.Print("\x1B[0;0H")
	}
	lcd.lineIndex = 0
}

// DisableRender turns off rendering of lines. Only use while Paused.
func (lcd *LcdASCII) DisableRender() {
	lcd.dr = true
}
