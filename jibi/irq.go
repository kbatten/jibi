package jibi

type Irq struct {
	mmu *Mmu
}

func NewIrq(mmu *Mmu) *Irq {
	return &Irq{mmu}
}

type Interrupt uint8

const (
	InterruptVblank Interrupt = 0x01 << iota
	InterruptLCDC
	InterruptTimer
	InterruptSerial
	InterruptKeypad
)

func (i Interrupt) Uint8() uint8 {
	return uint8(i)
}

func (i Interrupt) Address() Word {
	switch i {
	case InterruptVblank:
		return Word(0x0040)
	case InterruptLCDC:
		return Word(0x0048)
	case InterruptTimer:
		return Word(0x0050)
	case InterruptSerial:
		return Word(0x0058)
	case InterruptKeypad:
		return Word(0x0060)
	default:
		return Word(0)
	}
}

func (i Interrupt) String() string {
	switch i {
	case InterruptVblank:
		return "Vblank"
	case InterruptLCDC:
		return "LCDC"
	case InterruptTimer:
		return "Timer"
	case InterruptSerial:
		return "Serial"
	case InterruptKeypad:
		return "Keypad"
	default:
		return "UNKNOWN"
	}
}

func (irq *Irq) Interrupt(in Interrupt) {
}

// GetInterrupt returns the highest priority enabled interrupt.
func (irq *Irq) GetInterrupt() Interrupt {
	// TODO: handle these memory addresses locally
	iereg := irq.mmu.ReadByteAt(AddrInterruptEnable)
	iflag := irq.mmu.ReadByteAt(AddrIF)
	if Byte(InterruptVblank)&iereg&iflag != 0 {
		return InterruptVblank
	} else if Byte(InterruptLCDC)&iereg&iflag != 0 {
		return InterruptLCDC
	} else if Byte(InterruptTimer)&iereg&iflag != 0 {
		return InterruptTimer
	} else if Byte(InterruptSerial)&iereg&iflag != 0 {
		return InterruptSerial
	} else if Byte(InterruptKeypad)&iereg&iflag != 0 {
		return InterruptKeypad
	}
	return 0
}

func (irq *Irq) ResetInterrupt(i Interrupt) {
	iflag := irq.mmu.ReadByteAt(AddrIF)
	iflag &= (Byte(i) ^ 0xFF)
	irq.mmu.WriteByteAt(AddrIF, iflag)
}
