package jibi

const (
	lcdWidth  Byte = 160
	lcdHeight Byte = 144
)

type Lcd interface {
	DrawLine(line []Byte)
	Blank()
	Init()
	Close()
}

func NewLcd(renderer string) Lcd {
	switch renderer {
	case "ascii":
		return NewLcdASCII()
	case "termbox":
		return NewLcdTermbox()
	default:
		panic("unknown renderer")
	}
}
