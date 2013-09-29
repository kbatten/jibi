package main

import (
	"fmt"
)

type video struct {
	ram memoryDevice
}

func newVideo() video {
	return video{newRamModule(0x2000, nil)}
}

func (v video) readByte(addr addressInterface) uint8 {
	return v.ram.readByte(addr)
}

func (v video) writeByte(addr addressInterface, n uint8) {
	v.ram.writeByte(addr, n)
}

func (v video) String() string {
	return fmt.Sprintf("<video>")
}

