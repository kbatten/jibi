package jibi

import (
	"fmt"
)

const (
	lcdWidth  Byte = 160
	lcdHeight Byte = 144
)

type Lcd interface {
	DrawLine([]Byte)
	Blank()
	DisableRender()
}

type LcdAscii struct {
	dr         bool
	prevLine   []Byte
	lineIndex  uint8
	prevLineId uint8
}

func NewLcdAscii() *LcdAscii {
	return &LcdAscii{}
}

func (lcd *LcdAscii) DrawLine(bl []Byte) {
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
		lineId := uint8(float64(lcd.lineIndex) * 120.0 / float64(lcdHeight))
		if lcd.prevLineId != lineId || lineId == 0 {
			if lcd.lineIndex%2 == 0 {
				fmt.Printf("\x1B[170D%s", ls)
			} else {
				fmt.Printf("\x1B[170D%s\n", ls)
			}
		}
		lcd.prevLineId = lineId
	}

	lcd.prevLine = bl
	lcd.lineIndex++
}

func (lcd *LcdAscii) Blank() {
	if lcd.dr == false {
		fmt.Print("\x1B[0;0H")
	}
	lcd.lineIndex = 0
}

func (lcd *LcdAscii) DisableRender() {
	lcd.dr = true
}
