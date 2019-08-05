package jibi

import "github.com/nsf/termbox-go"

// An LcdTermbox outputs characters to the terminal.
type LcdTermbox struct {
	disableRender       bool
	prevLine            []Byte
	lineIndex           uint8 // lcd output line index
	prevScreenLineIndex int   // previous screen output line index
}

func NewLcdTermbox() Lcd {
	return &LcdTermbox{}
}

func (lcd *LcdTermbox) Init() {
	termbox.Init()
	termbox.HideCursor()
	lcd.Blank()
}

func (lcd *LcdTermbox) Close() {
	termbox.Close()
}

// DrawLine draws the Byte Slice to the current line index, then advances the
// index.
func (lcd *LcdTermbox) DrawLine(line []Byte) {
	outputLine := make([]Byte, len(line))
	copy(outputLine, line)

	// calculate the screen Y
	screenLineIndex := int(float64(lcd.lineIndex) * 50.0 / float64(lcdHeight))
	if screenLineIndex == lcd.prevScreenLineIndex && lcd.lineIndex != 0 {
		// compress previous line and this line into one
		for i := range line {
			outputLine[i] = outputLine[i] | (lcd.prevLine[i]&0xC0 | (lcd.prevLine[i]&0x03)<<2)
		}
	}

	const runeNone rune = ' '
	const runeLow rune = '.'
	const runeHigh rune = '\''
	const runeLowHigh rune = ':'

	if lcd.disableRender == false {
		for i, c := range outputLine {
			if c == 1 { // 0001
				termbox.SetCell(i, screenLineIndex, runeNone, termbox.ColorWhite, termbox.ColorBlack)
			} else if c == 2 { // 0010
				termbox.SetCell(i, screenLineIndex, runeLow, termbox.ColorWhite, termbox.ColorBlack)
			} else if c == 3 { // 0011
				termbox.SetCell(i, screenLineIndex, runeLow, termbox.ColorWhite, termbox.ColorBlack)
			} else if c == 4 { // 0100
				termbox.SetCell(i, screenLineIndex, runeNone, termbox.ColorWhite, termbox.ColorBlack)
			} else if c == 5 { // 0101
				termbox.SetCell(i, screenLineIndex, runeNone, termbox.ColorWhite, termbox.ColorBlack)
			} else if c == 6 { // 0110
				termbox.SetCell(i, screenLineIndex, runeLow, termbox.ColorWhite, termbox.ColorBlack)
			} else if c == 7 { // 0111
				termbox.SetCell(i, screenLineIndex, runeLow, termbox.ColorWhite, termbox.ColorBlack)
			} else if c == 8 { // 1000
				termbox.SetCell(i, screenLineIndex, runeHigh, termbox.ColorWhite, termbox.ColorBlack)
			} else if c == 9 { // 1001
				termbox.SetCell(i, screenLineIndex, runeHigh, termbox.ColorWhite, termbox.ColorBlack)
			} else if c == 10 { // 1010
				termbox.SetCell(i, screenLineIndex, runeLowHigh, termbox.ColorWhite, termbox.ColorBlack)
			} else if c == 11 { // 1011
				termbox.SetCell(i, screenLineIndex, runeLowHigh, termbox.ColorWhite, termbox.ColorBlack)
			} else if c == 12 { // 1100
				termbox.SetCell(i, screenLineIndex, runeHigh, termbox.ColorWhite, termbox.ColorBlack)
			} else if c == 13 { // 1101
				termbox.SetCell(i, screenLineIndex, runeHigh, termbox.ColorWhite, termbox.ColorBlack)
			} else if c == 14 { // 1110
				termbox.SetCell(i, screenLineIndex, runeLowHigh, termbox.ColorWhite, termbox.ColorBlack)
			} else if c == 15 { // 1111
				termbox.SetCell(i, screenLineIndex, runeLowHigh, termbox.ColorWhite, termbox.ColorBlack)
			}
		}
	}

	termbox.Flush()

	lcd.prevScreenLineIndex = screenLineIndex
	lcd.prevLine = line
	lcd.lineIndex++
}

// Blank moves the cursor to the upper left.
func (lcd *LcdTermbox) Blank() {
	if lcd.disableRender == false {
		termbox.Clear(termbox.ColorWhite, termbox.ColorBlack)
	}
	lcd.lineIndex = 0
}

// DisableRender turns off rendering of lines. Only use while Paused.
func (lcd *LcdTermbox) DisableRender() {
	lcd.disableRender = true
}
