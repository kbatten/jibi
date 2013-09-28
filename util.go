package main

func bytesToWord(msb, lsb uint8) uint16 {
	return uint16(msb)<<8 + uint16(lsb)
}
