package jibi

import ()

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

func (cpu *Cpu) cmdSetInterrupt(resp interface{}) {
	if in, ok := resp.(Interrupt); !ok {
		panic("invalid command response type")
	} else {
		cpu.setInterrupt(in)
	}
}

func (cpu *Cpu) setInterrupt(in Interrupt) {
	if cpu.ime == 1 {
		if cpu.ie&Byte(in) == Byte(in) {
			cpu.iflags |= Byte(in)
		}
	}
}

// getInterrupt returns the highest priority enabled interrupt.
func (cpu *Cpu) getInterrupt() Interrupt {
	iereg := cpu.ie
	iflag := cpu.iflags
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

// resetInterrupt resets the specific interrupt.
func (cpu *Cpu) resetInterrupt(i Interrupt) {
	iflag := cpu.readByte(AddrIF)
	iflag &= (Byte(i) ^ 0xFF)
	cpu.writeByte(AddrIF, iflag)
}
