package jibi

// A list of all the special memory addresses.
const (
	AddrIF             Word = 0xFF0F
	AddrLCDC           Word = 0xFF40
	AddrStat           Word = 0xFF41
	AddrSCY            Word = 0xFF42
	AddrSCX            Word = 0xFF43
	AddrLY             Word = 0xFF44
	AddrLYC            Word = 0xFF45
	AddrBGP            Word = 0xFF47
	AddrSpritePalette0 Word = 0xFF48
	AddrSpritePalette1 Word = 0xFF49
	AddrWY             Word = 0xFF4A
	AddrWX             Word = 0xFF4B
	AddrIE             Word = 0xFFFF
)
