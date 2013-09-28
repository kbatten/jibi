package main

func bytesToWord(msb, lsb uint8) uint16 {
	return uint16(msb)<<8 + uint16(lsb)
}

func wordToBytes(w uint16) (lsb uint8, msb uint8) {
	return uint8(w >> 8), uint8(w & 0xFF)
}

func bytesToAddress(msb, lsb uint8) address {
	return address(uint16(msb)<<8 + uint16(lsb))
}
