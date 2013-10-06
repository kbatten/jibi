package jibi

// An Irq is an interrupt request handler. It provides helper functions for
// dealing with interrutps.
type Irq struct {
	mmu MemoryDevice
}

// NewIrq returns a new Irq object.
func NewIrq(mmu MemoryDevice) *Irq {
	return &Irq{mmu}
}

// An Interrupt is what it is.
type Interrupt uint8

// A list of all five interrupts.
const (
	InterruptVblank Interrupt = 0x01 << iota
	InterruptLCDC
	InterruptTimer
	InterruptSerial
	InterruptKeypad
)

// Address returns the restart address associated with the interrupt.
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

// SetInterrupt triggers the specific interrupt if that interrupt is enabled.
// TODO: figure out what to do with ime here.
func (irq *Irq) SetInterrupt(in Interrupt) {
	iereg := irq.mmu.ReadByteAt(AddrInterruptEnable)
	iflag := irq.mmu.ReadByteAt(AddrIF) | Byte(in)
	if iereg&Byte(in) == Byte(in) {
		irq.mmu.WriteByteAt(AddrIF, iflag)
	}
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

// ResetInterrupt resets the specific interrupt.
func (irq *Irq) ResetInterrupt(i Interrupt) {
	iflag := irq.mmu.ReadByteAt(AddrIF)
	iflag &= (Byte(i) ^ 0xFF)
	irq.mmu.WriteByteAt(AddrIF, iflag)
}
