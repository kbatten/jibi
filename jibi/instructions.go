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

// z reset
// n reset
// h and c set or reset according to operation
func (c *Cpu) addWordR(a Word, b Byte) Word {
	sum := a + Word(b)
	// check half carry
	if (a&0xFF)+Word(b) > 0xFF {
		c.f.setFlag(flagH)
	}
	// check carry
	if sum < a {
		c.f.setFlag(flagC)
	}
	// reset other flags
	c.f.resetFlag(flagZ)
	c.f.resetFlag(flagN)
	return sum
}

func (c *Cpu) bit(b uint8, n Byte) {
	set := 1<<b&n == 1<<b
	if !set {
		c.f.setFlag(flagZ)
	} else {
		c.f.resetFlag(flagZ)
	}
	c.f.resetFlag(flagN)
	c.f.setFlag(flagH)
}

func (c *Cpu) xor(a, b Byte) Byte {
	r := a ^ b
	c.f.reset()
	if r == 0 {
		c.f.setFlag(flagZ)
	}
	return Byte(r)
}

func (c *Cpu) and(a, b Byte) Byte {
	fmt.Println(c.str())
	panic("untested")
	r := a & b
	c.f.reset()
	if r == 0 {
		c.f.setFlag(flagZ)
	}
	c.f.setFlag(flagH)
	return Byte(r)
}

func (c *Cpu) or(a, b Byte) Byte {
	r := a | b
	c.f.reset()
	if r == 0 {
		c.f.setFlag(flagZ)
	}
	return Byte(r)
}

func (c *Cpu) inc(a Byte) Byte {
	r := a + 1
	if r == 0 {
		c.f.setFlag(flagZ)
	} else {
		c.f.resetFlag(flagZ)
	}
	c.f.resetFlag(flagN)
	if a&0x0F == 0x0F {
		c.f.setFlag(flagH)
	} else {
		c.f.resetFlag(flagH)
	}
	return Byte(r)
}

func (c *Cpu) dec(a Byte) Byte {
	r := a - 1
	if r == 0 {
		c.f.setFlag(flagZ)
	} else {
		c.f.resetFlag(flagZ)
	}
	c.f.setFlag(flagN)
	if a&0x0F == 0x00 {
		c.f.setFlag(flagH)
	} else {
		c.f.resetFlag(flagH)
	}
	return Byte(r)
}

func (c *Cpu) sbc(a, b Byte) Byte {
	fmt.Println(c.str())
	panic("inst")
	carry := Byte(0)
	if c.f.getFlag(flagC) {
		carry = 1
	}
	r := a - (b + carry)
	c.f.reset()
	if r == 0 {
		c.f.setFlag(flagZ)
	}
	c.f.setFlag(flagN)
	if a&0x0F < (b&0x0F + carry) {
		c.f.setFlag(flagH)
	}
	if a < b+carry {
		c.f.setFlag(flagC)
	}
	return Byte(r)
}

func (c *Cpu) sub(a, b Byte) Byte {
	r := a - b
	c.f.reset()
	if r == 0 {
		c.f.setFlag(flagZ)
	}
	c.f.setFlag(flagN)
	if a&0x0F < b&0x0F {
		c.f.setFlag(flagH)
	}
	if a < b {
		c.f.setFlag(flagC)
	}
	return Byte(r)
}

func (c *Cpu) adc(a, b Byte) Byte {
	carry := Byte(0)
	if c.f.getFlag(flagC) {
		carry = 1
	}
	r := a + b + carry
	c.f.reset()
	if r == 0 {
		c.f.setFlag(flagZ)
	}
	if a&0x0F+b&0x0F+carry > 0x0F {
		c.f.setFlag(flagH)
	}
	if uint16(a)+uint16(b)+uint16(carry) > 0xFF {
		c.f.setFlag(flagC)
	}
	return Byte(r)
}

func (c *Cpu) add(a, b Byte) Byte {
	r := a + b
	c.f.reset()
	if r == 0 {
		c.f.setFlag(flagZ)
	}
	if a&0x0F+b&0x0F > 0x0F {
		c.f.setFlag(flagH)
	}
	if uint16(a)+uint16(b) > 0xFF {
		c.f.setFlag(flagC)
	}
	return Byte(r)
}

// rotate right through carry (yes, naming is odd)
func (c *Cpu) rr(n Byte) Byte {
	r := n >> 1
	if c.f.getFlag(flagC) { // old carry is bit 7
		r += 1 << 7
	}
	c.f.reset()
	if r == 0 {
		c.f.setFlag(flagZ)
	}
	if n&0x01 == 0x01 { // carry is old bit 0
		c.f.setFlag(flagC)
	}
	return Byte(r)
}

// rotate left through carry
func (c *Cpu) rl(n Byte) Byte {
	r := n << 1
	if c.f.getFlag(flagC) { // old carry is bit 0
		r += 1
	}
	c.f.reset()
	if r == 0 {
		c.f.setFlag(flagZ)
	}
	if n&0x80 == 0x80 { // carry is old bit 7
		c.f.setFlag(flagC)
	}
	return Byte(r)
}

// rotate left, old bit 7 to carry
func (c *Cpu) rlc(n Byte) Byte {
	r := n>>7 | n<<1
	c.f.reset()
	if r == 0 {
		c.f.setFlag(flagZ)
	}
	if n&0x80 == 0x80 { // carry is old bit 7
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

func (c *Cpu) jp(addr Word) {
	c.pc = register16(addr)
}

func (c *Cpu) callF(f Byte, addr Word) {
	fmt.Println(c.str())
	panic("untested")
	if c.f.getFlag(f) == true {
		c.call(addr)
	}
}

func (c *Cpu) call(addr Word) {
	c.push(c.pc)
	c.jp(addr)
}

func (c *Cpu) pop() Word {
	c.sp += 2
	return c.readWord(c.sp - 2)
}

func (c *Cpu) push(w Word) {
	c.sp -= 2
	c.writeWord(c.sp, w)
}
