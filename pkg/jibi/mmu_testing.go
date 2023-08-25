package jibi

type TestMmu struct {
	ram []Byte
}

func newTestMmu() Mmu {
	return TestMmu{make([]Byte, 0x10000)}
}

func (tm TestMmu) LockAddr(addr Word, ak AddressKeys) AddressKeys {
	return AddressKeys(0)
}

func (tm TestMmu) UnlockAddr(addr Word, ak AddressKeys) AddressKeys {
	return AddressKeys(0)
}

func (tm TestMmu) ReadByteAt(addr Word, ak AddressKeys) Byte {
	return tm.ram[addr]
}

func (tm TestMmu) WriteByteAt(addr Word, b Byte, ak AddressKeys) {
	tm.ram[addr] = b
}

func (tm TestMmu) ReadIoByte(addr Word, ak AddressKeys) (Byte, bool) {
	return tm.ram[addr], true
}

func (tm TestMmu) SetGpu(gpu *Gpu) {
}

func (tm TestMmu) SetKeypad(kp *Keypad) {
}

func (tm TestMmu) SetInterrupt(in Interrupt, ak AddressKeys) {
}
