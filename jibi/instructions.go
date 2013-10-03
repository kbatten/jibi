package jibi

import (
	"fmt"
)

// holds the instruction currently being fetched
type instruction struct {
	o opcode
	p []Byte // params
}

func newInstruction(o opcode, ps ...Byte) instruction {
	p := make([]Byte, len(ps))
	copy(p, ps)
	return instruction{o, p}
}

func (i instruction) String() string {
	ps := ""
	for _, v := range i.p {
		ps += fmt.Sprintf("0x%02X ", v)
	}
	return fmt.Sprintf("%s [ 0x%02X %s]", i.o, uint16(i.o), ps)
}

func (c *Cpu) bit(b uint8, n Byter) {
	set := 1<<b&n.Uint8() == 1<<b
	if !set {
		c.f.setFlag(flagZ)
	} else {
		c.f.resetFlag(flagZ)
	}
	c.f.resetFlag(flagN)
	c.f.setFlag(flagH)
}

func (c *Cpu) xor(a, b Byter) Byte {
	r := a.Uint8() ^ b.Uint8()
	c.f.reset()
	if r == 0 {
		c.f.setFlag(flagZ)
	}
	return Byte(r)
}

func (c *Cpu) and(a, b Byter) Byte {
	r := a.Uint8() & b.Uint8()
	c.f.reset()
	if r == 0 {
		c.f.setFlag(flagZ)
	}
	c.f.setFlag(flagH)
	return Byte(r)
}

func (c *Cpu) or(a, b Byter) Byte {
	r := a.Uint8() | b.Uint8()
	c.f.reset()
	if r == 0 {
		c.f.setFlag(flagZ)
	}
	return Byte(r)
}

func (c *Cpu) inc(a Byter) Byte {
	r := a.Uint8() + 1
	if r == 0 {
		c.f.setFlag(flagZ)
	} else {
		c.f.resetFlag(flagZ)
	}
	c.f.resetFlag(flagN)
	if a.Uint8()&0x0F == 0x0F {
		c.f.setFlag(flagH)
	} else {
		c.f.resetFlag(flagH)
	}
	return Byte(r)
}

func (c *Cpu) dec(a Byter) Byte {
	r := a.Uint8() - 1
	if r == 0 {
		c.f.setFlag(flagZ)
	} else {
		c.f.resetFlag(flagZ)
	}
	c.f.setFlag(flagN)
	if a.Uint8()&0x0F != 0x0F {
		c.f.setFlag(flagH)
	} else {
		c.f.resetFlag(flagH)
	}
	return Byte(r)
}

func (c *Cpu) sbc(a, b Byter) Byte {
	carry := uint8(0)
	if c.f.getFlag(flagC) {
		carry = 1
	}
	r := a.Uint8() - (b.Uint8() + carry)
	c.f.reset()
	if r == 0 {
		c.f.setFlag(flagZ)
	}
	c.f.setFlag(flagN)
	if a.Uint8()&0x0F >= (b.Uint8()&0x0F + carry) {
		c.f.setFlag(flagH)
	}
	if a.Uint8() >= b.Uint8()+carry {
		c.f.setFlag(flagC)
	}
	return Byte(r)
}

func (c *Cpu) sub(a, b Byter) Byte {
	r := a.Uint8() - b.Uint8()
	c.f.reset()
	if r == 0 {
		c.f.setFlag(flagZ)
	}
	c.f.setFlag(flagN)
	if a.Uint8()&0x0F >= b.Uint8()&0x0F {
		c.f.setFlag(flagH)
	}
	if a.Uint8() >= b.Uint8() {
		c.f.setFlag(flagC)
	}
	return Byte(r)
}

func (c *Cpu) adc(a, b Byter) Byte {
	carry := uint8(0)
	if c.f.getFlag(flagC) {
		carry = 1
	}
	r := a.Uint8() + b.Uint8() + carry
	c.f.reset()
	if r == 0 {
		c.f.setFlag(flagZ)
	}
	if a.Uint8()&0x0F+b.Uint8()&0x0F+carry > 0x0F {
		c.f.setFlag(flagH)
	}
	if uint16(a.Uint8())+uint16(b.Uint8())+uint16(carry) > 0xFF {
		c.f.setFlag(flagC)
	}
	return Byte(r)
}

func (c *Cpu) add(a, b Byter) Byte {
	r := a.Uint8() + b.Uint8()
	c.f.reset()
	if r == 0 {
		c.f.setFlag(flagZ)
	}
	if a.Uint8()&0x0F+b.Uint8()&0x0F > 0x0F {
		c.f.setFlag(flagH)
	}
	if uint16(a.Uint8())+uint16(b.Uint8()) > 0xFF {
		c.f.setFlag(flagC)
	}
	return Byte(r)
}

// rotate right through carry (yes, naming is odd)
func (c *Cpu) rr(n Byter) Byte {
	r := n.Uint8() >> 1
	if c.f.getFlag(flagC) { // old carry is bit 7
		r += 1 << 7
	}
	c.f.reset()
	if r == 0 {
		c.f.setFlag(flagZ)
	}
	if n.Uint8()&0x01 == 0x01 { // carry is old bit 0
		c.f.setFlag(flagC)
	}
	return Byte(r)
}

// rotate left through carry
func (c *Cpu) rl(n Byter) Byte {
	r := n.Uint8() << 1
	if c.f.getFlag(flagC) { // old carry is bit 0
		r += 1
	}
	c.f.reset()
	if r == 0 {
		c.f.setFlag(flagZ)
	}
	if n.Uint8()&0x80 == 0x80 { // carry is old bit 7
		c.f.setFlag(flagC)
	}
	return Byte(r)
}

func (c *Cpu) jrF(f Byte, n int8) {
	if c.f.getFlag(f) == true {
		c.jr(n)
	}
}

func (c *Cpu) jrNF(f Byte, n int8) {
	if c.f.getFlag(f) == false {
		c.jr(n)
	}
}

func (c *Cpu) jr(n int8) {
	if n < 0 {
		c.pc -= register16(-n)
		return
	}
	c.pc += register16(n)
}

func (c *Cpu) jp(addr Worder) {
	c.pc = register16(addr.Uint16())
}

func (c *Cpu) callF(f Byte, addr Worder) {
	if c.f.getFlag(f) == true {
		c.call(addr)
	}
}

func (c *Cpu) call(addr Worder) {
	c.push(c.pc)
	c.jp(addr)
}

func (c *Cpu) pop() Word {
	r := c.mm.readWord(c.sp)
	c.sp += 2
	return r
}

func (c *Cpu) push(w Worder) {
	c.mm.writeWord(c.sp-2, w)
	c.sp -= 2
}
