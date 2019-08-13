package jibi

type TestMmu struct {
	ram []Byte
}

func newTestMmu() Mmu {
	return TestMmu{make([]Byte, 0x10000)}
}

func (tm TestMmu) ReadByteAt(addr Word) Byte {
	return tm.ram[addr]
}

func (tm TestMmu) WriteByteAt(addr Word, b Byte) {
	tm.ram[addr] = b
}

func (tm TestMmu) WriteElevatedByteAt(addr Word, b Byte) {
	tm.ram[addr] = b
}

func (tm TestMmu) ReadWordAt(addr Word) Word {
	return BytesToWord(tm.ReadByteAt(addr+1), tm.ReadByteAt(addr))
}

func (tm TestMmu) WriteWordAt(addr, w Word) {
	tm.WriteByteAt(addr+1, w.High())
	tm.WriteByteAt(addr, w.Low())
}

func (tm TestMmu) ReadIoByte(addr Word) Byte {
	return tm.ram[addr]
}

func (tm TestMmu) SetGpu(gpu *Gpu) {
}

func (tm TestMmu) SetKeypad(kp *Keypad) {
}

func (tm TestMmu) SetInterrupt(in Interrupt) {
}

func (tm TestMmu) ResetInterrupt(in Interrupt) {
}
