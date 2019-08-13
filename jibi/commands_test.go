package jibi

import (
	"fmt"
	"testing"
)

func TestOp01(t *testing.T) {
	cpu := NewCpu(newTestMmu(), []Byte{0x01, 0x32, 0x01})
	defer cpu.RunCommand(CmdStop, nil)

	// LD BC, nn
	cpu.fetch()
	cpu.execute()
	if cpu.b.Word() != Word(0x0132) {
		t.Error()
	}
}

func TestOp02(t *testing.T) {
	cpu := NewCpu(newTestMmu(), []Byte{0x02})
	defer cpu.RunCommand(CmdStop, nil)

	// LD (BC), A
	cpu.a.set(0x55)
	cpu.b.setWord(0xFFFC)
	cpu.fetch()
	cpu.execute()
	if cpu.readByte(0xFFFC) != 0x55 {
		t.Error()
	}
}

func TestOp03(t *testing.T) {
	cpu := NewCpu(newTestMmu(), []Byte{0x03})
	defer cpu.RunCommand(CmdStop, nil)

	// INC BC
	cpu.pc = 0
	cpu.b.setWord(0x1FFE)
	cpu.fetch()
	cpu.execute()
	if cpu.b.Word() != Word(0x1FFF) {
		t.Error()
	}

	// INC BC -- carry
	cpu.pc = 0
	cpu.b.setWord(0x1FFF)
	cpu.fetch()
	cpu.execute()
	if cpu.b.Word() != Word(0x2000) {
		t.Error()
	}

	// INC BC -- overflow
	cpu.pc = 0
	cpu.b.setWord(0xFFFF)
	cpu.fetch()
	cpu.execute()
	if cpu.b.Word() != Word(0x0000) {
		t.Error()
	}
}

func TestOp04(t *testing.T) {
	cpu := NewCpu(newTestMmu(), []Byte{0x04})
	defer cpu.RunCommand(CmdStop, nil)

	// INC B -- NZ, NN, NH
	cpu.pc = 0
	cpu.b.set(0x00)
	cpu.fetch()
	cpu.execute()
	if cpu.b.Byte() != Byte(0x01) {
		t.Error()
	}
	if cpu.f.getFlag(flagZ) != false {
		t.Error()
	}
	if cpu.f.getFlag(flagN) != false {
		t.Error()
	}
	if cpu.f.getFlag(flagH) != false {
		t.Error()
	}
	cpu.f.reset()

	// INC B -- Z, NN, H
	cpu.pc = 0
	cpu.b.set(0xFF)
	cpu.fetch()
	cpu.execute()
	if cpu.b.Byte() != Byte(0x00) {
		t.Error()
	}
	if cpu.f.getFlag(flagZ) != true {
		t.Error()
	}
	if cpu.f.getFlag(flagN) != false {
		t.Error()
	}
	if cpu.f.getFlag(flagH) != true {
		t.Error()
	}
	cpu.f.reset()

	// INC B -- NZ, NN, H
	cpu.pc = 0
	cpu.b.set(0xEF)
	cpu.fetch()
	cpu.execute()
	if cpu.b.Byte() != Byte(0xF0) {
		t.Error()
	}
	if cpu.f.getFlag(flagZ) != false {
		t.Error()
	}
	if cpu.f.getFlag(flagN) != false {
		t.Error()
	}
	if cpu.f.getFlag(flagH) != true {
		t.Error()
	}
	cpu.f.reset()
}

func TestOp05(t *testing.T) {
	cpu := NewCpu(newTestMmu(), []Byte{0x05})
	defer cpu.RunCommand(CmdStop, nil)

	// DEC B -- NZ, NH
	cpu.pc = 0
	cpu.b.set(0x02)
	cpu.fetch()
	cpu.execute()
	if cpu.b.Byte() != Byte(0x01) {
		t.Error()
	}
	if cpu.f.getFlag(flagZ) != false {
		t.Error()
	}
	if cpu.f.getFlag(flagN) != true {
		t.Error()
	}
	if cpu.f.getFlag(flagH) != false {
		t.Error()
	}
	cpu.f.reset()

	// DEC B -- Z, NH
	cpu.pc = 0
	cpu.b.set(0x01)
	cpu.fetch()
	cpu.execute()
	if cpu.b.Byte() != Byte(0x00) {
		t.Error()
	}
	if cpu.f.getFlag(flagZ) != true {
		t.Error()
	}
	if cpu.f.getFlag(flagN) != true {
		t.Error()
	}
	if cpu.f.getFlag(flagH) != false {
		t.Error()
	}
	cpu.f.reset()

	// DEC B -- NZ, H
	cpu.pc = 0
	cpu.b.set(0xF0)
	cpu.fetch()
	cpu.execute()
	if cpu.b.Byte() != Byte(0xEF) {
		t.Error()
	}
	if cpu.f.getFlag(flagZ) != false {
		t.Error()
	}
	if cpu.f.getFlag(flagN) != true {
		t.Error()
	}
	if cpu.f.getFlag(flagH) != true {
		t.Error()
	}
	cpu.f.reset()
}

func TestOp06(t *testing.T) {
	cpu := NewCpu(newTestMmu(), []Byte{0x06, 0x32})
	defer cpu.RunCommand(CmdStop, nil)

	// LD B, #
	cpu.fetch()
	cpu.execute()
	if cpu.b.Byte() != Byte(0x32) {
		t.Error()
	}
}

func TestOp07(t *testing.T) {
	cpu := NewCpu(newTestMmu(), []Byte{0x07})
	defer cpu.RunCommand(CmdStop, nil)

	// RLCA -- bit7 low, NZ, NC
	cpu.pc = 0
	cpu.a.set(Byte(0x7F))
	cpu.fetch()
	cpu.execute()
	if cpu.a.Byte() != Byte(0xFE) {
		t.Error(fmt.Sprintf("0x%02X", cpu.a.Byte()))
	}
	if cpu.f.getFlag(flagZ) != false {
		t.Error()
	}
	if cpu.f.getFlag(flagN) != false {
		t.Error()
	}
	if cpu.f.getFlag(flagH) != false {
		t.Error()
	}
	if cpu.f.getFlag(flagC) != false {
		t.Error()
	}

	// RLCA -- bit7 high, NZ, C
	cpu.pc = 0
	cpu.a.set(Byte(0xFE))
	cpu.fetch()
	cpu.execute()
	if cpu.a.Byte() != Byte(0xFD) {
		t.Error(fmt.Sprintf("0x%02X", cpu.a.Byte()))
	}
	if cpu.f.getFlag(flagZ) != false {
		t.Error()
	}
	if cpu.f.getFlag(flagN) != false {
		t.Error()
	}
	if cpu.f.getFlag(flagH) != false {
		t.Error()
	}
	if cpu.f.getFlag(flagC) != true {
		t.Error()
	}

	// RLCA -- bit7 high, Z, NC
	cpu.pc = 0
	cpu.a.set(Byte(0x00))
	cpu.fetch()
	cpu.execute()
	if cpu.a.Byte() != Byte(0x00) {
		t.Error(fmt.Sprintf("0x%02X", cpu.a.Byte()))
	}
	if cpu.f.getFlag(flagZ) != true {
		t.Error()
	}
	if cpu.f.getFlag(flagN) != false {
		t.Error()
	}
	if cpu.f.getFlag(flagH) != false {
		t.Error()
	}
	if cpu.f.getFlag(flagC) != false {
		t.Error()
	}
}

func TestOp0C(t *testing.T) {
	cpu := NewCpu(newTestMmu(), []Byte{0x0C})
	defer cpu.RunCommand(CmdStop, nil)

	// INC C -- NZ, NH
	cpu.c.set(Byte(0x44))
	cpu.fetch()
	cpu.execute()
	if cpu.c.Byte() != Byte(0x45) {
		t.Error()
	}
	if cpu.f.getFlag(flagZ) != false {
		t.Error()
	}
	if cpu.f.getFlag(flagN) != false {
		t.Error()
	}
	if cpu.f.getFlag(flagH) != false {
		t.Error()
	}

	// INC C -- NZ, H
	cpu.pc = 0
	cpu.f.reset()
	cpu.c.set(Byte(0x1F))
	cpu.fetch()
	cpu.execute()
	if cpu.c.Byte() != Byte(0x20) {
		t.Error()
	}
	if cpu.f.getFlag(flagZ) != false {
		t.Error()
	}
	if cpu.f.getFlag(flagN) != false {
		t.Error()
	}
	if cpu.f.getFlag(flagH) != true {
		t.Error()
	}

	// INC C -- Z, H
	cpu.pc = 0
	cpu.f.reset()
	cpu.c.set(Byte(0xFF))
	cpu.fetch()
	cpu.execute()
	if cpu.c.Byte() != Byte(0x00) {
		t.Error()
	}
	if cpu.f.getFlag(flagZ) != true {
		t.Error()
	}
	if cpu.f.getFlag(flagN) != false {
		t.Error()
	}
	if cpu.f.getFlag(flagH) != true {
		t.Error()
	}
}

func TestOp0E(t *testing.T) {
	cpu := NewCpu(newTestMmu(), []Byte{0x0E, 0x32})
	defer cpu.RunCommand(CmdStop, nil)

	// LD C, #
	cpu.fetch()
	cpu.execute()
	if cpu.c.Byte() != Byte(0x32) {
		t.Error()
	}
}

func TestOp11(t *testing.T) {
	cpu := NewCpu(newTestMmu(), []Byte{0x11, 0x32, 0x01})
	defer cpu.RunCommand(CmdStop, nil)

	// LD DE, nn
	cpu.fetch()
	cpu.execute()
	if cpu.d.Word() != Word(0x0132) {
		t.Error()
	}
}

func TestOp17(t *testing.T) {
	cpu := NewCpu(newTestMmu(), []Byte{0x17})
	defer cpu.RunCommand(CmdStop, nil)

	// RLA -- carry, bit7 low, NC
	cpu.f.setFlag(flagC)
	cpu.a.set(Byte(0x7F))
	cpu.fetch()
	cpu.execute()
	if cpu.a.Byte() != Byte(0xFF) {
		t.Error(fmt.Sprintf("0x%02X", cpu.c.Byte()))
	}
	if cpu.f.getFlag(flagZ) != false {
		t.Error()
	}
	if cpu.f.getFlag(flagN) != false {
		t.Error()
	}
	if cpu.f.getFlag(flagH) != false {
		t.Error()
	}
	if cpu.f.getFlag(flagC) != false {
		t.Error()
	}

	// RLA -- no carry, bit7 low, NC
	cpu.pc = 0
	cpu.f.resetFlag(flagC)
	cpu.a.set(Byte(0x7F))
	cpu.fetch()
	cpu.execute()
	if cpu.a.Byte() != Byte(0xFE) {
		t.Error(fmt.Sprintf("0x%02X", cpu.c.Byte()))
	}
	if cpu.f.getFlag(flagZ) != false {
		t.Error()
	}
	if cpu.f.getFlag(flagN) != false {
		t.Error()
	}
	if cpu.f.getFlag(flagH) != false {
		t.Error()
	}
	if cpu.f.getFlag(flagC) != false {
		t.Error()
	}

	// RLA -- carry, bit7 high, C
	cpu.pc = 0
	cpu.f.setFlag(flagC)
	cpu.a.set(Byte(0xBF))
	cpu.fetch()
	cpu.execute()
	if cpu.a.Byte() != Byte(0x7F) {
		t.Error(fmt.Sprintf("0x%02X", cpu.c.Byte()))
	}
	if cpu.f.getFlag(flagZ) != false {
		t.Error()
	}
	if cpu.f.getFlag(flagN) != false {
		t.Error()
	}
	if cpu.f.getFlag(flagH) != false {
		t.Error()
	}
	if cpu.f.getFlag(flagC) != true {
		t.Error()
	}

	// RLA -- no carry, bit7 high, C
	cpu.pc = 0
	cpu.f.resetFlag(flagC)
	cpu.a.set(Byte(0xBF))
	cpu.fetch()
	cpu.execute()
	if cpu.a.Byte() != Byte(0x7E) {
		t.Error(fmt.Sprintf("0x%02X", cpu.c.Byte()))
	}
	if cpu.f.getFlag(flagZ) != false {
		t.Error()
	}
	if cpu.f.getFlag(flagN) != false {
		t.Error()
	}
	if cpu.f.getFlag(flagH) != false {
		t.Error()
	}
	if cpu.f.getFlag(flagC) != true {
		t.Error()
	}

	// RLA -- no carry, bit7 high, C, Z
	cpu.pc = 0
	cpu.f.resetFlag(flagC)
	cpu.a.set(Byte(0x80))
	cpu.fetch()
	cpu.execute()
	if cpu.a.Byte() != Byte(0x00) {
		t.Error(fmt.Sprintf("0x%02X", cpu.c.Byte()))
	}
	if cpu.f.getFlag(flagZ) != true {
		t.Error()
	}
	if cpu.f.getFlag(flagN) != false {
		t.Error()
	}
	if cpu.f.getFlag(flagH) != false {
		t.Error()
	}
	if cpu.f.getFlag(flagC) != true {
		t.Error()
	}

	// RLA -- no carry, bit7 low, NC, Z
	cpu.pc = 0
	cpu.f.resetFlag(flagC)
	cpu.a.set(Byte(0x00))
	cpu.fetch()
	cpu.execute()
	if cpu.a.Byte() != Byte(0x00) {
		t.Error(fmt.Sprintf("0x%02X", cpu.c.Byte()))
	}
	if cpu.f.getFlag(flagZ) != true {
		t.Error()
	}
	if cpu.f.getFlag(flagN) != false {
		t.Error()
	}
	if cpu.f.getFlag(flagH) != false {
		t.Error()
	}
	if cpu.f.getFlag(flagC) != false {
		t.Error()
	}
}

func TestOp13(t *testing.T) {
	cpu := NewCpu(newTestMmu(), []Byte{0x13})
	defer cpu.RunCommand(CmdStop, nil)

	// INC DE
	cpu.pc = 0
	cpu.d.setWord(0x1FFE)
	cpu.fetch()
	cpu.execute()
	if cpu.d.Word() != Word(0x1FFF) {
		t.Error()
	}

	// INC DE -- carry
	cpu.pc = 0
	cpu.d.setWord(0x1FFF)
	cpu.fetch()
	cpu.execute()
	if cpu.d.Word() != Word(0x2000) {
		t.Error()
	}

	// INC DE -- overflow
	cpu.pc = 0
	cpu.d.setWord(0xFFFF)
	cpu.fetch()
	cpu.execute()
	if cpu.d.Word() != Word(0x0000) {
		t.Error()
	}
}

func TestOp18(t *testing.T) {
	cpu := NewCpu(newTestMmu(), []Byte{0x00, 0x00, 0x00, 0x00, 0x18, 0x05, 0x18, 0xFC})
	defer cpu.RunCommand(CmdStop, nil)

	// JR * -- positive offset
	cpu.pc = register16(0x04)
	cpu.fetch()
	cpu.execute()
	if cpu.pc != Word(0x04+0x02+0x05) {
		t.Error()
	}

	// JR * -- negative offset
	cpu.pc = register16(0x06)
	cpu.fetch()
	cpu.execute()
	if cpu.pc != Word(0x06+0x02-0x04) {
		t.Error()
	}
}

func TestOp1A(t *testing.T) {
	cpu := NewCpu(newTestMmu(), []Byte{0x1A})
	defer cpu.RunCommand(CmdStop, nil)

	// LD A, (DE)
	cpu.d.setWord(Word(0xFF80))
	cpu.writeWord(0xFF80, 0x05)
	cpu.fetch()
	cpu.execute()
	if cpu.a.Byte() != Byte(0x05) {
		t.Error()
	}
}

func TestOp20(t *testing.T) {
	cpu := NewCpu(newTestMmu(), []Byte{0x20, 0x05, 0x00, 0x00, 0x00, 0x00, 0x00, 0x20, 0xFC})
	defer cpu.RunCommand(CmdStop, nil)

	// JR NZ, 05 -- Z
	cpu.pc = 0
	cpu.f.setFlag(flagZ)
	cpu.fetch()
	cpu.execute()
	if cpu.pc != Word(0x02) {
		t.Error()
	}

	// JR NZ, 05 -- NZ, positive offset
	cpu.pc = 0
	cpu.f.resetFlag(flagZ)
	cpu.fetch()
	cpu.execute()
	if cpu.pc != Word(0x07) {
		t.Error()
	}

	// JR NZ, FC -- NZ, negative offset
	cpu.f.resetFlag(flagZ)
	cpu.fetch()
	cpu.execute()
	if cpu.pc != Word(0x05) {
		t.Error()
	}
}

func TestOp21(t *testing.T) {
	cpu := NewCpu(newTestMmu(), []Byte{0x21, 0x32, 0x01})
	defer cpu.RunCommand(CmdStop, nil)

	// LD HL, nn
	cpu.fetch()
	cpu.execute()
	if cpu.h.Word() != Word(0x0132) {
		t.Error()
	}
}

func TestOp22(t *testing.T) {
	cpu := NewCpu(newTestMmu(), []Byte{0x22})
	defer cpu.RunCommand(CmdStop, nil)

	// LDI (HL), A
	cpu.a.set(0x89)
	cpu.h.setWord(0xFF80)
	cpu.fetch()
	cpu.execute()
	b := cpu.readByte(Word(0xFF80))
	if b != Byte(0x89) {
		t.Error()
	}
	if cpu.h.Word() != Word(0xFF81) {
		t.Error()
	}
}

func TestOp23(t *testing.T) {
	cpu := NewCpu(newTestMmu(), []Byte{0x23})
	defer cpu.RunCommand(CmdStop, nil)

	// INC HL
	cpu.pc = 0
	cpu.h.setWord(0x1FFE)
	cpu.fetch()
	cpu.execute()
	if cpu.h.Word() != Word(0x1FFF) {
		t.Error()
	}

	// INC HL -- carry
	cpu.pc = 0
	cpu.h.setWord(0x1FFF)
	cpu.fetch()
	cpu.execute()
	if cpu.h.Word() != Word(0x2000) {
		t.Error()
	}

	// INC HL -- overflow
	cpu.pc = 0
	cpu.h.setWord(0xFFFF)
	cpu.fetch()
	cpu.execute()
	if cpu.h.Word() != Word(0x0000) {
		t.Error()
	}
}

func TestOp28(t *testing.T) {
	cpu := NewCpu(newTestMmu(), []Byte{0x00, 0x00, 0x00, 0x00, 0x28, 0x05, 0x28, 0xFC})
	defer cpu.RunCommand(CmdStop, nil)

	// JR Z * -- Z, positive offset
	cpu.pc = register16(0x04)
	cpu.f.setFlag(flagZ)
	cpu.fetch()
	cpu.execute()
	if cpu.pc != Word(0x04+0x02+0x05) {
		t.Error()
	}

	// JR Z * -- Z, negative offset
	cpu.pc = register16(0x06)
	cpu.f.setFlag(flagZ)
	cpu.fetch()
	cpu.execute()
	if cpu.pc != Word(0x06+0x02-0x04) {
		t.Error()
	}

	// JR Z * -- NZ, positive offset
	cpu.pc = register16(0x04)
	cpu.f.resetFlag(flagZ)
	cpu.fetch()
	cpu.execute()
	if cpu.pc != Word(0x04+0x02) {
		t.Error()
	}

	// JR Z * -- NZ, negative offset
	cpu.pc = register16(0x06)
	cpu.f.resetFlag(flagZ)
	cpu.fetch()
	cpu.execute()
	if cpu.pc != Word(0x06+0x02) {
		t.Error()
	}
}

func TestOp31(t *testing.T) {
	cpu := NewCpu(newTestMmu(), []Byte{0x31, 0x32, 0x01})
	defer cpu.RunCommand(CmdStop, nil)

	// LD SP, nn
	cpu.fetch()
	cpu.execute()
	if cpu.sp != Word(0x0132) {
		t.Error()
	}
}

func TestOp32(t *testing.T) {
	cpu := NewCpu(newTestMmu(), []Byte{0x32})
	defer cpu.RunCommand(CmdStop, nil)

	// LDD (HL), A
	cpu.a.set(0x89)
	cpu.h.setWord(0xFF80)
	cpu.fetch()
	cpu.execute()
	b := cpu.readByte(Word(0xFF80))
	if b != Byte(0x89) {
		t.Error()
	}
	if cpu.h.Word() != Word(0xFF7F) {
		t.Error()
	}
}

func TestOp3E(t *testing.T) {
	cpu := NewCpu(newTestMmu(), []Byte{0x3E, 0x32})
	defer cpu.RunCommand(CmdStop, nil)

	// LD A, #
	cpu.fetch()
	cpu.execute()
	if cpu.a.Byte() != Byte(0x32) {
		t.Error()
	}
}

func TestOp4F(t *testing.T) {
	cpu := NewCpu(newTestMmu(), []Byte{0x4F})
	defer cpu.RunCommand(CmdStop, nil)

	// LD C, A
	cpu.a.set(0x05)
	cpu.fetch()
	cpu.execute()
	if cpu.c.Byte() != Byte(0x05) {
		t.Error()
	}
}

func TestOp77(t *testing.T) {
	cpu := NewCpu(newTestMmu(), []Byte{0x77})
	defer cpu.RunCommand(CmdStop, nil)

	// LD (HL), A
	cpu.a.set(0x89)
	cpu.h.setWord(0xFF80)
	cpu.fetch()
	cpu.execute()
	b := cpu.readByte(Word(0xFF80))
	if b != Byte(0x89) {
		t.Error()
	}
}

func TestOpA8(t *testing.T) {
	cpu := NewCpu(newTestMmu(), []Byte{0xA8})
	defer cpu.RunCommand(CmdStop, nil)

	// XOR B -- NZ
	cpu.pc = 0
	cpu.f.set(Byte(0xFF))
	cpu.a.set(Byte(0x0F))
	cpu.b.set(Byte(0xFF))
	cpu.fetch()
	cpu.execute()
	if cpu.a.Byte() != Byte(0xF0) {
		t.Error()
	}
	if cpu.f.getFlag(flagZ) != false {
		t.Error()
	}
	if cpu.f.getFlag(flagN) != false {
		t.Error()
	}
	if cpu.f.getFlag(flagH) != false {
		t.Error()
	}
	if cpu.f.getFlag(flagC) != false {
		t.Error()
	}
	cpu.f.reset()

	// XOR B -- Z
	cpu.pc = 0
	cpu.f.set(Byte(0xFF))
	cpu.a.set(Byte(0x0F))
	cpu.b.set(Byte(0x0F))
	cpu.fetch()
	cpu.execute()
	if cpu.a.Byte() != Byte(0x00) {
		t.Error()
	}
	if cpu.f.getFlag(flagZ) != true {
		t.Error()
	}
	if cpu.f.getFlag(flagN) != false {
		t.Error()
	}
	if cpu.f.getFlag(flagH) != false {
		t.Error()
	}
	if cpu.f.getFlag(flagC) != false {
		t.Error()
	}
	cpu.f.reset()
}

func TestOpAF(t *testing.T) {
	cpu := NewCpu(newTestMmu(), []Byte{0xAF})
	defer cpu.RunCommand(CmdStop, nil)

	// XOR A -- Z
	cpu.pc = 0
	cpu.f.set(Byte(0xFF))
	cpu.a.set(Byte(0x0F))
	cpu.fetch()
	cpu.execute()
	if cpu.a.Byte() != Byte(0x00) {
		t.Error()
	}
	if cpu.f.getFlag(flagZ) != true {
		t.Error()
	}
	if cpu.f.getFlag(flagN) != false {
		t.Error()
	}
	if cpu.f.getFlag(flagH) != false {
		t.Error()
	}
	if cpu.f.getFlag(flagC) != false {
		t.Error()
	}
	cpu.f.reset()
}

func TestOpB1(t *testing.T) {
	cpu := NewCpu(newTestMmu(), []Byte{0xB1})
	defer cpu.RunCommand(CmdStop, nil)

	// OR C -- NZ
	cpu.pc = 0
	cpu.f.set(Byte(0xFF))
	cpu.a.set(Byte(0x0F))
	cpu.c.set(Byte(0xF0))
	cpu.fetch()
	cpu.execute()
	if cpu.a.Byte() != Byte(0xFF) {
		t.Error()
	}
	if cpu.f.getFlag(flagZ) != false {
		t.Error()
	}
	if cpu.f.getFlag(flagN) != false {
		t.Error()
	}
	if cpu.f.getFlag(flagH) != false {
		t.Error()
	}
	if cpu.f.getFlag(flagC) != false {
		t.Error()
	}
	cpu.f.reset()

	// OR C -- Z
	cpu.pc = 0
	cpu.f.set(Byte(0xFF))
	cpu.a.set(Byte(0x00))
	cpu.c.set(Byte(0x00))
	cpu.fetch()
	cpu.execute()
	if cpu.a.Byte() != Byte(0x00) {
		t.Error()
	}
	if cpu.f.getFlag(flagZ) != true {
		t.Error()
	}
	if cpu.f.getFlag(flagN) != false {
		t.Error()
	}
	if cpu.f.getFlag(flagH) != false {
		t.Error()
	}
	if cpu.f.getFlag(flagC) != false {
		t.Error()
	}
	cpu.f.reset()
}

func TestC1(t *testing.T) {
	cpu := NewCpu(newTestMmu(), []Byte{0xC1, 0x00, 0x00, 0x00, 0x03, 0x20})
	defer cpu.RunCommand(CmdStop, nil)

	cpu.sp = register16(0x04)

	// POP BC
	cpu.fetch()
	cpu.execute()
	if cpu.sp != Word(0x06) {
		t.Error()
	}
	if cpu.b.Word() != Word(0x2003) {
		t.Errorf(fmt.Sprintf("0x%04X", cpu.b.Word()))
	}
}

func TestOpC5(t *testing.T) {
	cpu := NewCpu(newTestMmu(), []Byte{0xC5})
	defer cpu.RunCommand(CmdStop, nil)

	cpu.sp = register16(0xFFFE)

	// PUSH BC
	cpu.b.setWord(0x6004)
	cpu.fetch()
	cpu.execute()
	if cpu.sp != Word(0xFFFC) {
		t.Error()
	}
	w := BytesToWord(cpu.readByte(cpu.sp+1), cpu.readByte(cpu.sp))
	if w != Word(0x6004) {
		t.Errorf(fmt.Sprintf("0x%04X", w))
	}
}

func TestOpC9(t *testing.T) {
	cpu := NewCpu(newTestMmu(), []Byte{0xC9})
	defer cpu.RunCommand(CmdStop, nil)

	// RET
	cpu.sp = register16(0xFFFC)
	cpu.writeWord(cpu.sp, 0x0F01)
	cpu.fetch()
	cpu.execute()
	if cpu.pc != Word(0x0F01) {
		t.Error()
	}
	if cpu.sp != Word(0xFFFE) {
		t.Error()
	}
}

func TestOpCB11(t *testing.T) {
	cpu := NewCpu(newTestMmu(), []Byte{0xCB, 0x11})
	defer cpu.RunCommand(CmdStop, nil)

	// RL C -- carry, bit7 low, NC
	cpu.f.setFlag(flagC)
	cpu.c.set(Byte(0x7F))
	cpu.fetch()
	cpu.execute()
	if cpu.c.Byte() != Byte(0xFF) {
		t.Error(fmt.Sprintf("0x%02X", cpu.c.Byte()))
	}
	if cpu.f.getFlag(flagZ) != false {
		t.Error()
	}
	if cpu.f.getFlag(flagN) != false {
		t.Error()
	}
	if cpu.f.getFlag(flagH) != false {
		t.Error()
	}
	if cpu.f.getFlag(flagC) != false {
		t.Error()
	}

	// RL C -- no carry, bit7 low, NC
	cpu.pc = 0
	cpu.f.resetFlag(flagC)
	cpu.c.set(Byte(0x7F))
	cpu.fetch()
	cpu.execute()
	if cpu.c.Byte() != Byte(0xFE) {
		t.Error(fmt.Sprintf("0x%02X", cpu.c.Byte()))
	}
	if cpu.f.getFlag(flagZ) != false {
		t.Error()
	}
	if cpu.f.getFlag(flagN) != false {
		t.Error()
	}
	if cpu.f.getFlag(flagH) != false {
		t.Error()
	}
	if cpu.f.getFlag(flagC) != false {
		t.Error()
	}

	// RL C -- carry, bit7 high, C
	cpu.pc = 0
	cpu.f.setFlag(flagC)
	cpu.c.set(Byte(0xBF))
	cpu.fetch()
	cpu.execute()
	if cpu.c.Byte() != Byte(0x7F) {
		t.Error(fmt.Sprintf("0x%02X", cpu.c.Byte()))
	}
	if cpu.f.getFlag(flagZ) != false {
		t.Error()
	}
	if cpu.f.getFlag(flagN) != false {
		t.Error()
	}
	if cpu.f.getFlag(flagH) != false {
		t.Error()
	}
	if cpu.f.getFlag(flagC) != true {
		t.Error()
	}

	// RL C -- no carry, bit7 high, C
	cpu.pc = 0
	cpu.f.resetFlag(flagC)
	cpu.c.set(Byte(0xBF))
	cpu.fetch()
	cpu.execute()
	if cpu.c.Byte() != Byte(0x7E) {
		t.Error(fmt.Sprintf("0x%02X", cpu.c.Byte()))
	}
	if cpu.f.getFlag(flagZ) != false {
		t.Error()
	}
	if cpu.f.getFlag(flagN) != false {
		t.Error()
	}
	if cpu.f.getFlag(flagH) != false {
		t.Error()
	}
	if cpu.f.getFlag(flagC) != true {
		t.Error()
	}

	// RL C -- no carry, bit7 high, C, Z
	cpu.pc = 0
	cpu.f.resetFlag(flagC)
	cpu.c.set(Byte(0x80))
	cpu.fetch()
	cpu.execute()
	if cpu.c.Byte() != Byte(0x00) {
		t.Error(fmt.Sprintf("0x%02X", cpu.c.Byte()))
	}
	if cpu.f.getFlag(flagZ) != true {
		t.Error()
	}
	if cpu.f.getFlag(flagN) != false {
		t.Error()
	}
	if cpu.f.getFlag(flagH) != false {
		t.Error()
	}
	if cpu.f.getFlag(flagC) != true {
		t.Error()
	}

	// RL C -- no carry, bit7 low, NC, Z
	cpu.pc = 0
	cpu.f.resetFlag(flagC)
	cpu.c.set(Byte(0x00))
	cpu.fetch()
	cpu.execute()
	if cpu.c.Byte() != Byte(0x00) {
		t.Error(fmt.Sprintf("0x%02X", cpu.c.Byte()))
	}
	if cpu.f.getFlag(flagZ) != true {
		t.Error()
	}
	if cpu.f.getFlag(flagN) != false {
		t.Error()
	}
	if cpu.f.getFlag(flagH) != false {
		t.Error()
	}
	if cpu.f.getFlag(flagC) != false {
		t.Error()
	}
}

func TestOpCB7C(t *testing.T) {
	cpu := NewCpu(newTestMmu(), []Byte{0xCB, 0x7C})
	defer cpu.RunCommand(CmdStop, nil)

	// BIT 7, H -- NZ
	cpu.pc = 0
	cpu.h.set(Byte(0x80))
	cpu.fetch()
	cpu.execute()
	if cpu.f.getFlag(flagZ) != false {
		t.Error()
	}
	if cpu.f.getFlag(flagN) != false {
		t.Error()
	}
	if cpu.f.getFlag(flagH) != true {
		t.Error()
	}
	cpu.f.reset()

	// BIT 7, H -- Z
	cpu.pc = 0
	cpu.h.set(Byte(0x7F))
	cpu.fetch()
	cpu.execute()
	if cpu.f.getFlag(flagZ) != true {
		t.Error()
	}
	if cpu.f.getFlag(flagN) != false {
		t.Error()
	}
	if cpu.f.getFlag(flagH) != true {
		t.Error()
	}
	cpu.f.reset()
}

func TestOpCB87(t *testing.T) {
	cpu := NewCpu(newTestMmu(), []Byte{0xCB, 0x87})
	defer cpu.RunCommand(CmdStop, nil)

	// RES 0, A -- already set
	cpu.pc = 0
	cpu.a.set(0xFF)
	cpu.fetch()
	cpu.execute()
	if cpu.a.Byte() != Byte(0xFE) {
		t.Error()
	}

	// RES 0, A -- already reset
	cpu.pc = 0
	cpu.a.set(0xFE)
	cpu.fetch()
	cpu.execute()
	if cpu.a.Byte() != Byte(0xFE) {
		t.Error()
	}
}

func TestOpCD(t *testing.T) {
	cpu := NewCpu(newTestMmu(), []Byte{0xCD, 0x40, 0x01})
	defer cpu.RunCommand(CmdStop, nil)

	cpu.sp = register16(0xFFFE)

	// CALL nn
	cpu.fetch()
	cpu.execute()
	if cpu.pc != Word(0x0140) {
		t.Error()
	}
	if cpu.sp != Word(0xFFFC) {
		t.Error()
	}
	w := BytesToWord(cpu.readByte(cpu.sp+1), cpu.readByte(cpu.sp))
	if w != Word(0x0003) {
		t.Errorf(fmt.Sprintf("0x%04X", w))
	}
}

func TestOpE0(t *testing.T) {
	cpu := NewCpu(newTestMmu(), []Byte{0xE0, 0x05})
	defer cpu.RunCommand(CmdStop, nil)

	// LDH (n), A
	cpu.a.set(0x89)
	cpu.fetch()
	cpu.execute()
	b := cpu.readByte(Word(0xFF05))
	if b != Byte(0x89) {
		t.Error()
	}
}

func TestOpE2(t *testing.T) {
	cpu := NewCpu(newTestMmu(), []Byte{0xE2})
	defer cpu.RunCommand(CmdStop, nil)

	// LD (C), A
	cpu.a.set(0xF5)
	cpu.c.set(0x05)
	cpu.fetch()
	cpu.execute()
	if cpu.readByte(0xFF05) != Byte(0xF5) {
		t.Error()
	}
}

func TestOpE6(t *testing.T) {
	cpu := NewCpu(newTestMmu(), []Byte{0xE6, 0xF0})
	defer cpu.RunCommand(CmdStop, nil)

	// AND # -- NZ
	cpu.pc = 0
	cpu.f.set(0xFF)
	cpu.a.set(0xFF)
	cpu.fetch()
	cpu.execute()
	if cpu.a.Byte() != Byte(0xF0) {
		t.Error()
	}
	if cpu.f.getFlag(flagZ) != false {
		t.Error()
	}
	if cpu.f.getFlag(flagN) != false {
		t.Error()
	}
	if cpu.f.getFlag(flagH) != true {
		t.Error()
	}
	if cpu.f.getFlag(flagC) != false {
		t.Error()
	}

	// AND # -- Z
	cpu.pc = 0
	cpu.f.set(0xFF)
	cpu.a.set(0x0F)
	cpu.fetch()
	cpu.execute()
	if cpu.a.Byte() != Byte(0x00) {
		t.Error()
	}
	if cpu.f.getFlag(flagZ) != true {
		t.Error()
	}
	if cpu.f.getFlag(flagN) != false {
		t.Error()
	}
	if cpu.f.getFlag(flagH) != true {
		t.Error()
	}
	if cpu.f.getFlag(flagC) != false {
		t.Error()
	}
}

func TestOpF0(t *testing.T) {
	cpu := NewCpu(newTestMmu(), []Byte{0xF0, 0x05})
	defer cpu.RunCommand(CmdStop, nil)

	// LDH A, (n)
	cpu.writeByte(0xFF05, 0x89)
	cpu.fetch()
	cpu.execute()
	if cpu.a.Byte() != Byte(0x89) {
		t.Error()
	}
}

func TestOpF3(t *testing.T) {
	cpu := NewCpu(newTestMmu(), []Byte{0xF3, 0xF0, 0x05})
	defer cpu.RunCommand(CmdStop, nil)

	// DI
	cpu.fetch()
	cpu.execute()
	if cpu.ime != Bit(1) {
		t.Error()
	}

	// next instruction
	cpu.fetch()
	cpu.execute()
	if cpu.ime != Bit(0) {
		t.Error()
	}
}

func TestOpFE(t *testing.T) {
	cpu := NewCpu(newTestMmu(), []Byte{0xFE, 0x01, 0xFE, 0x01, 0xFE, 0x0F, 0xFE, 0xF0, 0xFE, 0xF1})
	defer cpu.RunCommand(CmdStop, nil)

	// CP # -- NZ, H, C
	cpu.a.set(0xFF)
	cpu.fetch()
	cpu.execute()
	if cpu.f.getFlag(flagZ) != false {
		t.Error()
	}
	if cpu.f.getFlag(flagN) != true {
		t.Error()
	}
	if cpu.f.getFlag(flagH) != false {
		t.Error()
	}
	if cpu.f.getFlag(flagC) != false {
		t.Error()
	}

	// CP # -- Z, NH, NC
	cpu.a.set(0x01)
	cpu.fetch()
	cpu.execute()
	if cpu.f.getFlag(flagZ) != true {
		t.Error()
	}
	if cpu.f.getFlag(flagN) != true {
		t.Error()
	}
	if cpu.f.getFlag(flagH) != false {
		t.Error()
	}
	if cpu.f.getFlag(flagC) != false {
		t.Error()
	}

	// CP # -- NZ, H, NC
	cpu.a.set(0xF0)
	cpu.fetch()
	cpu.execute()
	if cpu.f.getFlag(flagZ) != false {
		t.Error()
	}
	if cpu.f.getFlag(flagN) != true {
		t.Error()
	}
	if cpu.f.getFlag(flagH) != true {
		t.Error()
	}
	if cpu.f.getFlag(flagC) != false {
		t.Error()
	}

	// CP # -- NZ, NH, C
	cpu.a.set(0xE0)
	cpu.fetch()
	cpu.execute()
	if cpu.f.getFlag(flagZ) != false {
		t.Error()
	}
	if cpu.f.getFlag(flagN) != true {
		t.Error()
	}
	if cpu.f.getFlag(flagH) != false {
		t.Error()
	}
	if cpu.f.getFlag(flagC) != true {
		t.Error()
	}

	// CP # -- NZ, H, C
	cpu.a.set(0xE0)
	cpu.fetch()
	cpu.execute()
	if cpu.f.getFlag(flagZ) != false {
		t.Error()
	}
	if cpu.f.getFlag(flagN) != true {
		t.Error()
	}
	if cpu.f.getFlag(flagH) != true {
		t.Error()
	}
	if cpu.f.getFlag(flagC) != true {
		t.Error()
	}
}

/*
func TestBit(t *testing.T) {
	cpu := NewCpu(newTestMmu(), []Byte{0xCB, 0x7C, 0xCB, 0x7C})
	defer cpu.RunCommand(CmdStop, nil)

	// BIT 7, H -- NZ
	cpu.h.set(Byte(0x80))
	cpu.fetch()
	cpu.execute()
	if cpu.f.getFlag(flagZ) != false {
		t.Error()
	}
	cpu.f.reset()

	// BIT 7, H -- Z
	cpu.h.set(Byte(0x7F))
	cpu.fetch()
	cpu.execute()
	if cpu.f.getFlag(flagZ) != true {
		t.Error()
	}
	cpu.f.reset()
}

func TestJp(t *testing.T) {
	cpu := NewCpu(newTestMmu(), []Byte{0xC3, 0x21, 0x67})
	defer cpu.RunCommand(CmdStop, nil)

	// JP nn
	cpu.fetch()
	cpu.execute()
	if cpu.pc != Word(0x6721) {
		t.Error()
	}
}

func TestJrNF(t *testing.T) {
	cpu := NewCpu(newTestMmu(), []Byte{0x00, 0x00, 0x00, 0x00, 0x20, 0x05, 0x20, 0xFC})
	defer cpu.RunCommand(CmdStop, nil)

	// JR NZ, 05 -- Z
	cpu.pc = register16(0x04)
	cpu.f.setFlag(flagZ)
	cpu.fetch()
	cpu.execute()
	if cpu.pc != Word(0x04+0x02) {
		t.Error()
	}

	// JR NZ, * -- NZ, positive offset
	cpu.pc = register16(0x04)
	cpu.f.resetFlag(flagZ)
	cpu.fetch()
	cpu.execute()
	if cpu.pc != Word(0x04+0x02+0x05) {
		t.Error()
	}

	// JR NZ, * -- NZ, negative offset
	cpu.pc = register16(0x06)
	cpu.f.resetFlag(flagZ)
	cpu.fetch()
	cpu.execute()
	if cpu.pc != Word(0x06+0x02-0x04) {
		t.Error()
	}
}

func TestJrF(t *testing.T) {
	cpu := NewCpu(newTestMmu(), []Byte{0x00, 0x00, 0x00, 0x00, 0x28, 0x05, 0x28, 0xFC})
	defer cpu.RunCommand(CmdStop, nil)

	// JR Z, 05 -- NZ
	cpu.pc = register16(0x04)
	cpu.f.resetFlag(flagZ)
	cpu.fetch()
	cpu.execute()
	if cpu.pc != Word(0x04+0x02) {
		t.Error()
	}

	// JR Z, * -- Z, positive offset
	cpu.pc = register16(0x04)
	cpu.f.setFlag(flagZ)
	cpu.fetch()
	cpu.execute()
	if cpu.pc != Word(0x04+0x02+0x05) {
		t.Error()
	}

	// JR Z, * -- Z, negative offset
	cpu.pc = register16(0x06)
	cpu.f.setFlag(flagZ)
	cpu.fetch()
	cpu.execute()
	if cpu.pc != Word(0x06+0x02-0x04) {
		t.Error()
	}
}

func TestCall(t *testing.T) {
	cpu := NewCpu(newTestMmu(), []Byte{0xCD, 0x40, 0x01})
	defer cpu.RunCommand(CmdStop, nil)

	cpu.sp = register16(0xFFFE)

	// CALL nn
	cpu.fetch()
	cpu.execute()
	if cpu.pc != Word(0x0140) {
		t.Error()
	}
	if cpu.sp != Word(0xFFFC) {
		t.Error()
	}
	w := BytesToWord(cpu.readByte(cpu.sp+1), cpu.readByte(cpu.sp))
	if w != Word(0x0003) {
		t.Errorf(fmt.Sprintf("0x%04X", w))
	}
}

func TestRet(t *testing.T) {
	cpu := NewCpu(newTestMmu(), []Byte{0xC9, 0x40, 0x01})
	defer cpu.RunCommand(CmdStop, nil)

	cpu.sp = register16(0x01)

	// RET
	cpu.fetch()
	cpu.execute()
	if cpu.sp != Word(0x03) {
		t.Error()
	}
	if cpu.pc != Word(0x0140) {
		t.Error()
	}
}

func TestPush(t *testing.T) {
	cpu := NewCpu(newTestMmu(), []Byte{0xC5})
	defer cpu.RunCommand(CmdStop, nil)

	cpu.sp = register16(0xFFFE)

	// PUSH BC
	cpu.b.setWord(0x6004)
	cpu.fetch()
	cpu.execute()
	if cpu.sp != Word(0xFFFC) {
		t.Error()
	}
	w := BytesToWord(cpu.readByte(cpu.sp+1), cpu.readByte(cpu.sp))
	if w != Word(0x6004) {
		t.Errorf(fmt.Sprintf("0x%04X", w))
	}
}

func TestPop(t *testing.T) {
	cpu := NewCpu(newTestMmu(), []Byte{0xC1, 0x00, 0x00, 0x00, 0x03, 0x20})
	defer cpu.RunCommand(CmdStop, nil)

	cpu.sp = register16(0x04)

	// POP BC
	cpu.fetch()
	cpu.execute()
	if cpu.sp != Word(0x06) {
		t.Error()
	}
	if cpu.b.Word() != Word(0x2003) {
		t.Errorf(fmt.Sprintf("0x%04X", cpu.b.Word()))
	}
}

func TestRr(t *testing.T) {
	cpu := NewCpu(newTestMmu(), []Byte{0x1F})
	defer cpu.RunCommand(CmdStop, nil)

	// RLA -- C, bit0 low
	cpu.f.setFlag(flagC)
	cpu.a.set(Byte(0xFE))
	cpu.fetch()
	cpu.execute()
	if cpu.a.Byte() != Byte(0xFF) {
		t.Error(fmt.Sprintf("0x%02X", cpu.a.Byte()))
	}
	if cpu.f.getFlag(flagC) != false {
		t.Error()
	}

	// RLA -- NC, bit0 low
	cpu.pc = 0
	cpu.f.resetFlag(flagC)
	cpu.a.set(Byte(0xFE))
	cpu.fetch()
	cpu.execute()
	if cpu.a.Byte() != Byte(0x7F) {
		t.Error(fmt.Sprintf("0x%02X", cpu.a.Byte()))
	}
	if cpu.f.getFlag(flagC) != false {
		t.Error()
	}

	// RLA -- C, bit0 high
	cpu.pc = 0
	cpu.f.setFlag(flagC)
	cpu.a.set(Byte(0x7F))
	cpu.fetch()
	cpu.execute()
	if cpu.a.Byte() != Byte(0xBF) {
		t.Error(fmt.Sprintf("0x%02X", cpu.a.Byte()))
	}
	if cpu.f.getFlag(flagC) != true {
		t.Error()
	}

	// RLA -- NC, bit0 high
	cpu.pc = 0
	cpu.f.resetFlag(flagC)
	cpu.a.set(Byte(0x7F))
	cpu.fetch()
	cpu.execute()
	if cpu.a.Byte() != Byte(0x3F) {
		t.Error(fmt.Sprintf("0x%02X", cpu.a.Byte()))
	}
	if cpu.f.getFlag(flagC) != true {
		t.Error()
	}
}

func TestRl(t *testing.T) {
	cpu := NewCpu(newTestMmu(), []Byte{0xCB, 0x11})
	defer cpu.RunCommand(CmdStop, nil)

	// RL C -- C, bit7 low
	cpu.f.setFlag(flagC)
	cpu.c.set(Byte(0x7F))
	cpu.fetch()
	cpu.execute()
	if cpu.c.Byte() != Byte(0xFF) {
		t.Error(fmt.Sprintf("0x%02X", cpu.c.Byte()))
	}
	if cpu.f.getFlag(flagC) != false {
		t.Error()
	}

	// RL C -- NC, bit7 low
	cpu.pc = 0
	cpu.f.resetFlag(flagC)
	cpu.c.set(Byte(0x7F))
	cpu.fetch()
	cpu.execute()
	if cpu.c.Byte() != Byte(0xFE) {
		t.Error(fmt.Sprintf("0x%02X", cpu.c.Byte()))
	}
	if cpu.f.getFlag(flagC) != false {
		t.Error()
	}

	// RL C -- C, bit7 high
	cpu.pc = 0
	cpu.f.setFlag(flagC)
	cpu.c.set(Byte(0xBF))
	cpu.fetch()
	cpu.execute()
	if cpu.c.Byte() != Byte(0x7F) {
		t.Error(fmt.Sprintf("0x%02X", cpu.c.Byte()))
	}
	if cpu.f.getFlag(flagC) != true {
		t.Error()
	}

	// RL C -- NC, bit7 high
	cpu.pc = 0
	cpu.f.resetFlag(flagC)
	cpu.c.set(Byte(0xBF))
	cpu.fetch()
	cpu.execute()
	if cpu.c.Byte() != Byte(0x7E) {
		t.Error(fmt.Sprintf("0x%02X", cpu.c.Byte()))
	}
	if cpu.f.getFlag(flagC) != true {
		t.Error()
	}
}

func TestRlc(t *testing.T) {
	cpu := NewCpu(newTestMmu(), []Byte{0xCB, 0x01})
	defer cpu.RunCommand(CmdStop, nil)

	// RLC C -- bit7 high
	cpu.pc = 0
	cpu.c.set(Byte(0xFE))
	cpu.fetch()
	cpu.execute()
	if cpu.c.Byte() != Byte(0xFD) {
		t.Error(fmt.Sprintf("0x%02X", cpu.c.Byte()))
	}
	if cpu.f.getFlag(flagC) != true {
		t.Error()
	}

	// RLC C -- bit7 high
	cpu.pc = 0
	cpu.c.set(Byte(0xFE))
	cpu.fetch()
	cpu.execute()
	if cpu.c.Byte() != Byte(0xFD) {
		t.Error(fmt.Sprintf("0x%02X", cpu.c.Byte()))
	}
	if cpu.f.getFlag(flagC) != true {
		t.Error()
	}
}

func TestDec(t *testing.T) {
	cpu := NewCpu(newTestMmu(), []Byte{0x05})
	defer cpu.RunCommand(CmdStop, nil)

	// DEC B -- NZ, NH
	cpu.b.set(Byte(0x44))
	cpu.fetch()
	cpu.execute()
	if cpu.b.Byte() != Byte(0x43) {
		t.Error()
	}
	if cpu.f.getFlag(flagZ) != false {
		t.Error()
	}
	if cpu.f.getFlag(flagH) != false {
		t.Error()
	}

	// DEC B -- NZ, H
	cpu.pc = 0
	cpu.f.reset()
	cpu.b.set(Byte(0x10))
	cpu.fetch()
	cpu.execute()
	if cpu.b.Byte() != Byte(0x0F) {
		t.Error()
	}
	if cpu.f.getFlag(flagZ) != false {
		t.Error()
	}
	if cpu.f.getFlag(flagH) != true {
		t.Error()
	}

	cpu.pc = 0
	cpu.f.reset()
	cpu.b.set(Byte(0x00))
	cpu.fetch()
	cpu.execute()
	if cpu.b.Byte() != Byte(0xFF) {
		t.Error()
	}
	if cpu.f.getFlag(flagZ) != false {
		t.Error()
	}
	if cpu.f.getFlag(flagH) != true {
		t.Error()
	}

	// DEC B -- Z, NH
	cpu.pc = 0
	cpu.f.reset()
	cpu.b.set(Byte(0x01))
	cpu.fetch()
	cpu.execute()
	if cpu.b.Byte() != Byte(0x00) {
		t.Error()
	}
	if cpu.f.getFlag(flagZ) != true {
		t.Error()
	}
	if cpu.f.getFlag(flagH) != false {
		t.Error()
	}
}

func TestInc(t *testing.T) {
	cpu := NewCpu(newTestMmu(), []Byte{0x04})
	defer cpu.RunCommand(CmdStop, nil)

	// INC B -- NZ, NH
	cpu.b.set(Byte(0x44))
	cpu.fetch()
	cpu.execute()
	if cpu.b.Byte() != Byte(0x45) {
		t.Error()
	}
	if cpu.f.getFlag(flagZ) != false {
		t.Error()
	}
	if cpu.f.getFlag(flagH) != false {
		t.Error()
	}

	// INC B -- NZ, H
	cpu.pc = 0
	cpu.f.reset()
	cpu.b.set(Byte(0x1F))
	cpu.fetch()
	cpu.execute()
	if cpu.b.Byte() != Byte(0x20) {
		t.Error()
	}
	if cpu.f.getFlag(flagZ) != false {
		t.Error()
	}
	if cpu.f.getFlag(flagH) != true {
		t.Error()
	}

	// INC B -- Z, H
	cpu.pc = 0
	cpu.f.reset()
	cpu.b.set(Byte(0xFF))
	cpu.fetch()
	cpu.execute()
	if cpu.b.Byte() != Byte(0x00) {
		t.Error()
	}
	if cpu.f.getFlag(flagZ) != true {
		t.Error()
	}
	if cpu.f.getFlag(flagH) != true {
		t.Error()
	}
}

func TestAddWordR(t *testing.T) {
	cpu := NewCpu(newTestMmu(), []Byte{0xF8, 0x0F})
	defer cpu.RunCommand(CmdStop, nil)

	// LDHL SP, n -- H, C
	cpu.pc = 0
	cpu.sp = register16(0xFFFE)
	cpu.f.reset()
	cpu.fetch()
	cpu.execute()
	if cpu.h.Word() != Word(0x000D) {
		t.Error()
	}
	if cpu.f.getFlag(flagZ) != false {
		t.Error()
	}
	if cpu.f.getFlag(flagN) != false {
		t.Error()
	}
	if cpu.f.getFlag(flagH) != true {
		t.Error()
	}
	if cpu.f.getFlag(flagC) != true {
		t.Error()
	}

	// LDHL SP, n -- NH, NC
	cpu.pc = 0
	cpu.sp = register16(0xFFF0)
	cpu.f.reset()
	cpu.fetch()
	cpu.execute()
	if cpu.h.Word() != Word(0xFFFF) {
		t.Error()
	}
	if cpu.f.getFlag(flagZ) != false {
		t.Error()
	}
	if cpu.f.getFlag(flagN) != false {
		t.Error()
	}
	if cpu.f.getFlag(flagH) != false {
		t.Error()
	}
	if cpu.f.getFlag(flagC) != false {
		t.Error()
	}

	// LDHL SP, n -- H, NC
	cpu.pc = 0
	cpu.sp = register16(0xFEFE)
	cpu.f.reset()
	cpu.fetch()
	cpu.execute()
	if cpu.h.Word() != Word(0xFF0D) {
		t.Error()
	}
	if cpu.f.getFlag(flagZ) != false {
		t.Error()
	}
	if cpu.f.getFlag(flagN) != false {
		t.Error()
	}
	if cpu.f.getFlag(flagH) != true {
		t.Error()
	}
	if cpu.f.getFlag(flagC) != false {
		t.Error()
	}
}

func TestAdd(t *testing.T) {
	cpu := NewCpu(newTestMmu(), []Byte{0x86})
	defer cpu.RunCommand(CmdStop, nil)

	// ADD A, (HL) -- NZ, NH, NC
	cpu.pc = 0
	cpu.f.reset()
	cpu.a.set(Byte(0x24))
	cpu.h.setWord(Word(0xFF80))
	cpu.writeByte(Word(0xFF80), Byte(0x40))
	cpu.fetch()
	cpu.execute()
	if cpu.a.Byte() != Byte(0x64) {
		t.Error()
	}
	if cpu.f.getFlag(flagZ) != false {
		t.Error()
	}
	if cpu.f.getFlag(flagH) != false {
		t.Error()
	}
	if cpu.f.getFlag(flagC) != false {
		t.Error()
	}

	// ADD A, (HL) -- NZ, NH, C
	cpu.pc = 0
	cpu.f.reset()
	cpu.a.set(Byte(0x24))
	cpu.h.setWord(Word(0xFF80))
	cpu.writeByte(Word(0xFF80), Byte(0xF0))
	cpu.fetch()
	cpu.execute()
	if cpu.a.Byte() != Byte(0x14) {
		t.Error()
	}
	if cpu.f.getFlag(flagZ) != false {
		t.Error()
	}
	if cpu.f.getFlag(flagH) != false {
		t.Error()
	}
	if cpu.f.getFlag(flagC) != true {
		t.Error()
	}

	// ADD A, (HL) -- NZ, H, NC
	cpu.pc = 0
	cpu.f.reset()
	cpu.a.set(Byte(0x24))
	cpu.h.setWord(Word(0xFF80))
	cpu.writeByte(Word(0xFF80), Byte(0x3C))
	cpu.fetch()
	cpu.execute()
	if cpu.a.Byte() != Byte(0x60) {
		t.Error()
	}
	if cpu.f.getFlag(flagZ) != false {
		t.Error()
	}
	if cpu.f.getFlag(flagH) != true {
		t.Error()
	}
	if cpu.f.getFlag(flagC) != false {
		t.Error()
	}

	// ADD A, (HL) -- NZ, H, C
	cpu.pc = 0
	cpu.f.reset()
	cpu.a.set(Byte(0x24))
	cpu.h.setWord(Word(0xFF80))
	cpu.writeByte(Word(0xFF80), Byte(0xEC))
	cpu.fetch()
	cpu.execute()
	if cpu.a.Byte() != Byte(0x10) {
		t.Error()
	}
	if cpu.f.getFlag(flagZ) != false {
		t.Error()
	}
	if cpu.f.getFlag(flagH) != true {
		t.Error()
	}
	if cpu.f.getFlag(flagC) != true {
		t.Error()
	}

	// ADD A, (HL) -- Z, NH, NC
	cpu.pc = 0
	cpu.f.reset()
	cpu.a.set(Byte(0x00))
	cpu.h.setWord(Word(0xFF80))
	cpu.writeByte(Word(0xFF80), Byte(0x00))
	cpu.fetch()
	cpu.execute()
	if cpu.a.Byte() != Byte(0x00) {
		t.Error()
	}
	if cpu.f.getFlag(flagZ) != true {
		t.Error()
	}
	if cpu.f.getFlag(flagH) != false {
		t.Error()
	}
	if cpu.f.getFlag(flagC) != false {
		t.Error()
	}

	// ADD A, (HL) -- Z, NH, C
	cpu.pc = 0
	cpu.f.reset()
	cpu.a.set(Byte(0x20))
	cpu.h.setWord(Word(0xFF80))
	cpu.writeByte(Word(0xFF80), Byte(0xE0))
	cpu.fetch()
	cpu.execute()
	if cpu.a.Byte() != Byte(0x00) {
		t.Error()
	}
	if cpu.f.getFlag(flagZ) != true {
		t.Error()
	}
	if cpu.f.getFlag(flagH) != false {
		t.Error()
	}
	if cpu.f.getFlag(flagC) != true {
		t.Error()
	}

	// ADD A, (HL) -- Z, H, C
	cpu.pc = 0
	cpu.f.reset()
	cpu.a.set(Byte(0xFF))
	cpu.h.setWord(Word(0xFF80))
	cpu.writeByte(Word(0xFF80), Byte(0x01))
	cpu.fetch()
	cpu.execute()
	if cpu.a.Byte() != Byte(0x00) {
		t.Error()
	}
	if cpu.f.getFlag(flagZ) != true {
		t.Error()
	}
	if cpu.f.getFlag(flagH) != true {
		t.Error()
	}
	if cpu.f.getFlag(flagC) != true {
		t.Error()
	}
}

func TestAdc(t *testing.T) {
	cpu := NewCpu(newTestMmu(), []Byte{0x8C})
	defer cpu.RunCommand(CmdStop, nil)

	// ADC A, H -- C, NZ, NH, NC
	cpu.pc = 0
	cpu.f.reset()
	cpu.f.setFlag(flagC)
	cpu.a.set(Byte(0x24))
	cpu.h.set(Byte(0x12))
	cpu.fetch()
	cpu.execute()
	if cpu.a.Byte() != Byte(0x37) {
		t.Error()
	}
	if cpu.f.getFlag(flagZ) != false {
		t.Error()
	}
	if cpu.f.getFlag(flagH) != false {
		t.Error()
	}
	if cpu.f.getFlag(flagC) != false {
		t.Error()
	}

	// ADC A, H -- C, NZ, NH, C
	cpu.pc = 0
	cpu.f.reset()
	cpu.f.setFlag(flagC)
	cpu.a.set(Byte(0x24))
	cpu.h.set(Byte(0xE2))
	cpu.fetch()
	cpu.execute()
	if cpu.a.Byte() != Byte(0x07) {
		t.Error()
	}
	if cpu.f.getFlag(flagZ) != false {
		t.Error()
	}
	if cpu.f.getFlag(flagH) != false {
		t.Error()
	}
	if cpu.f.getFlag(flagC) != true {
		t.Error()
	}

	// ADC A, H -- NC, NZ, H, NC
	cpu.pc = 0
	cpu.f.reset()
	cpu.f.resetFlag(flagC)
	cpu.a.set(Byte(0x24))
	cpu.h.set(Byte(0x1E))
	cpu.fetch()
	cpu.execute()
	if cpu.a.Byte() != Byte(0x42) {
		t.Error()
	}
	if cpu.f.getFlag(flagZ) != false {
		t.Error()
	}
	if cpu.f.getFlag(flagH) != true {
		t.Error()
	}
	if cpu.f.getFlag(flagC) != false {
		t.Error()
	}

	// ADC A, H -- NC, NZ, H, C
	cpu.pc = 0
	cpu.f.reset()
	cpu.f.resetFlag(flagC)
	cpu.a.set(Byte(0x24))
	cpu.h.set(Byte(0xEE))
	cpu.fetch()
	cpu.execute()
	if cpu.a.Byte() != Byte(0x12) {
		t.Error()
	}
	if cpu.f.getFlag(flagZ) != false {
		t.Error()
	}
	if cpu.f.getFlag(flagH) != true {
		t.Error()
	}
	if cpu.f.getFlag(flagC) != true {
		t.Error()
	}

	// ADC A, H -- NC, Z, NH, NC
	cpu.pc = 0
	cpu.f.reset()
	cpu.f.resetFlag(flagC)
	cpu.a.set(Byte(0x00))
	cpu.h.set(Byte(0x00))
	cpu.fetch()
	cpu.execute()
	if cpu.a.Byte() != Byte(0x00) {
		t.Error()
	}
	if cpu.f.getFlag(flagZ) != true {
		t.Error()
	}
	if cpu.f.getFlag(flagH) != false {
		t.Error()
	}
	if cpu.f.getFlag(flagC) != false {
		t.Error()
	}

	// ADC A, H -- NC, Z, H, C
	cpu.pc = 0
	cpu.f.reset()
	cpu.f.resetFlag(flagC)
	cpu.a.set(Byte(0x01))
	cpu.h.set(Byte(0xFF))
	cpu.fetch()
	cpu.execute()
	if cpu.a.Byte() != Byte(0x00) {
		t.Error()
	}
	if cpu.f.getFlag(flagZ) != true {
		t.Error()
	}
	if cpu.f.getFlag(flagH) != true {
		t.Error()
	}
	if cpu.f.getFlag(flagC) != true {
		t.Error()
	}

	// ADC A, H -- C, Z, H, C
	cpu.pc = 0
	cpu.f.reset()
	cpu.f.setFlag(flagC)
	cpu.a.set(Byte(0x01))
	cpu.h.set(Byte(0xFE))
	cpu.fetch()
	cpu.execute()
	if cpu.a.Byte() != Byte(0x00) {
		t.Error()
	}
	if cpu.f.getFlag(flagZ) != true {
		t.Error()
	}
	if cpu.f.getFlag(flagH) != true {
		t.Error()
	}
	if cpu.f.getFlag(flagC) != true {
		t.Error()
	}
}

func TestSub(t *testing.T) {
	cpu := NewCpu(newTestMmu(), []Byte{0x90})
	defer cpu.RunCommand(CmdStop, nil)

	// SUB B -- NZ, NH, NC
	cpu.pc = 0
	cpu.f.reset()
	cpu.a.set(Byte(0x24))
	cpu.b.set(Byte(0x12))
	cpu.fetch()
	cpu.execute()
	if cpu.a.Byte() != Byte(0x12) {
		t.Error()
	}
	if cpu.f.getFlag(flagZ) != false {
		t.Error()
	}
	if cpu.f.getFlag(flagH) != false {
		t.Error()
	}
	if cpu.f.getFlag(flagC) != false {
		t.Error()
	}

	// SUB B -- NZ, NH, C
	cpu.pc = 0
	cpu.f.reset()
	cpu.a.set(Byte(0x24))
	cpu.b.set(Byte(0x32))
	cpu.fetch()
	cpu.execute()
	if cpu.a.Byte() != Byte(0xF2) {
		t.Error()
	}
	if cpu.f.getFlag(flagZ) != false {
		t.Error()
	}
	if cpu.f.getFlag(flagH) != false {
		t.Error()
	}
	if cpu.f.getFlag(flagC) != true {
		t.Error()
	}

	// SUB B -- NZ, H, NC
	cpu.pc = 0
	cpu.f.reset()
	cpu.a.set(Byte(0x32))
	cpu.b.set(Byte(0x24))
	cpu.fetch()
	cpu.execute()
	if cpu.a.Byte() != Byte(0x0E) {
		t.Error()
	}
	if cpu.f.getFlag(flagZ) != false {
		t.Error()
	}
	if cpu.f.getFlag(flagH) != true {
		t.Error()
	}
	if cpu.f.getFlag(flagC) != false {
		t.Error()
	}

	// SUB B -- NZ, H, C
	cpu.pc = 0
	cpu.f.reset()
	cpu.a.set(Byte(0x23))
	cpu.b.set(Byte(0x24))
	cpu.fetch()
	cpu.execute()
	if cpu.a.Byte() != Byte(0xFF) {
		t.Error()
	}
	if cpu.f.getFlag(flagZ) != false {
		t.Error()
	}
	if cpu.f.getFlag(flagH) != true {
		t.Error()
	}
	if cpu.f.getFlag(flagC) != true {
		t.Error()
	}

	// SUB B -- Z, NH, NC
	cpu.pc = 0
	cpu.f.reset()
	cpu.a.set(Byte(0x32))
	cpu.b.set(Byte(0x32))
	cpu.fetch()
	cpu.execute()
	if cpu.a.Byte() != Byte(0x00) {
		t.Error()
	}
	if cpu.f.getFlag(flagZ) != true {
		t.Error()
	}
	if cpu.f.getFlag(flagH) != false {
		t.Error()
	}
	if cpu.f.getFlag(flagC) != false {
		t.Error()
	}
}

func TestLdd(t *testing.T) {
	cpu := NewCpu(newTestMmu(), []Byte{0x3A, 0x32, 0x00, 0x89})
	defer cpu.RunCommand(CmdStop, nil)

	// LDD A, (HL)
	cpu.h.setWord(0x0003)
	cpu.fetch()
	cpu.execute()
	if cpu.a.Byte() != Byte(0x89) {
		t.Error()
	}
	if cpu.h.Word() != Word(0x0002) {
		t.Error()
	}

	// LDD (HL), A
	cpu.h.setWord(0xFF80)
	cpu.fetch()
	cpu.execute()
	b := cpu.readByte(Word(0xFF80))
	if b != Byte(0x89) {
		t.Error()
	}
	if cpu.h.Word() != Word(0xFF7F) {
		t.Error()
	}
}

func TestLdi(t *testing.T) {
	cpu := NewCpu(newTestMmu(), []Byte{0x2A, 0x22, 0x00, 0x89})
	defer cpu.RunCommand(CmdStop, nil)

	// LDI A, (HL)
	cpu.h.setWord(0x0003)
	cpu.fetch()
	cpu.execute()
	if cpu.a.Byte() != Byte(0x89) {
		t.Error()
	}
	if cpu.h.Word() != Word(0x0004) {
		t.Error()
	}

	// LDI (HL), A
	cpu.h.setWord(0xFF80)
	cpu.fetch()
	cpu.execute()
	b := cpu.readByte(Word(0xFF80))
	if b != Byte(0x89) {
		t.Error()
	}
	if cpu.h.Word() != Word(0xFF81) {
		t.Error()
	}
}

func TestInc16(t *testing.T) {
	cpu := NewCpu(newTestMmu(), []Byte{0x23, 0x23})
	defer cpu.RunCommand(CmdStop, nil)

	// INC HL
	cpu.h.setWord(0x1FFF)
	cpu.fetch()
	cpu.execute()
	if cpu.h.Word() != Word(0x2000) {
		t.Error()
	}

	// INC HL -- overflow
	cpu.h.setWord(0xFFFF)
	cpu.fetch()
	cpu.execute()
	if cpu.h.Word() != Word(0x0000) {
		t.Error()
	}
}

func TestDec16(t *testing.T) {
	cpu := NewCpu(newTestMmu(), []Byte{0x0B})
	defer cpu.RunCommand(CmdStop, nil)

	// DEC BC
	cpu.b.setWord(0x2000)
	cpu.fetch()
	cpu.execute()
	if cpu.b.Word() != Word(0x1FFF) {
		t.Error()
	}
}
*/
