package jibi

import (
	"fmt"
)

// An LcdASCII outputs as ascii characters to the terminal.
type LcdASCII struct {
	prevLine            []Byte
	lineIndex           uint8 // lcd output line index
	prevScreenLineIndex int   // previous screen output line index
}

func NewLcdASCII() Lcd {
	return &LcdASCII{}
}

func (lcd *LcdASCII) Init() {
}

func (lcd *LcdASCII) Close() {
}

// DrawLine draws the Byte Slice to the current line index, then advances the
// index.
func (lcd *LcdASCII) DrawLine(line []Byte) {
	outputLine := make([]Byte, len(line))
	copy(outputLine, line)

	// calculate the output Y
	screenLineIndex := int(float64(lcd.lineIndex) * 50.0 / float64(lcdHeight))
	if screenLineIndex == lcd.prevScreenLineIndex && lcd.lineIndex != 0 {
		// compress previous line and this line into one
		for i := range line {
			outputLine[i] = outputLine[i] | (lcd.prevLine[i]&0xC0 | (lcd.prevLine[i]&0x03)<<2)
		}
	}

	ls := make([]byte, lcdWidth)
	var o byte
	for i, c := range outputLine {
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
	fmt.Printf("\x1B[%d;H%s", screenLineIndex, ls)

	lcd.prevScreenLineIndex = screenLineIndex
	lcd.prevLine = line
	lcd.lineIndex++
}

// Blank moves the cursor to the upper left.
func (lcd *LcdASCII) Blank() {
	fmt.Print("\x1B[0;0H")
	lcd.lineIndex = 0
}
