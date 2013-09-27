package main

import (
	"testing"
)

func TestRegister(t *testing.T) {
	c := newRegister8(nil)
	b := newRegister8(&c)
	c.set(0x01)
	b.set(0x02)
	if b.getWord() != 0x0201 {
		t.Error()
	}
}
