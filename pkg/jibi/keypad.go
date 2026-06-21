package jibi

import (
	"os"
	"time"
)

// quit   0x1b <esc>
// up     0x77 w
// down   0x73 s
// left   0x61 a
// right  0x64 d
// b      0x2E .
// a      0x2F /
// select 0x5C \
// start  0x0A <enter>

// A Key is one of the 9 buttons (8 Game Boy + Quit).
type Key uint8

// List of 8 buttons plus quit.
const (
	KeyQuit Key = iota
	KeyUp
	KeyDown
	KeyLeft
	KeyRight
	KeyB
	KeyA
	KeySelect
	KeyStart
)

func (k Key) String() string {
	switch k {
	case KeyQuit:
		return "quit"
	case KeyUp:
		return "up"
	case KeyDown:
		return "down"
	case KeyLeft:
		return "left"
	case KeyRight:
		return "right"
	case KeyB:
		return "b"
	case KeyA:
		return "a"
	case KeySelect:
		return "select"
	case KeyStart:
		return "start"
	}
	return "UNKNOWN"
}

// keyMap holds the mapping from raw byte (ASCII) to Game Boy Key.
var keyMap = map[byte]Key{
	'\x1b':  KeyQuit,
	'w':     KeyUp,
	's':     KeyDown,
	'a':     KeyLeft,
	'd':     KeyRight,
	'.':     KeyB,
	'/':     KeyA,
	'\\':    KeySelect,
	'\n':    KeyStart,
	'\r':    KeyStart,
}

type valueChan struct {
	v Byte
	c chan bool
}

// A Keypad manages reading the actual key input, and the button states.
type Keypad struct {
	CommanderInterface

	mmu     Mmu
	mmuKeys AddressKeys

	p1013low bool

	keys map[Key]valueChan

	ttyConfig string
	runSetup  bool

	// quit is closed when the escape key is pressed.
	quit chan struct{}
}

// Quit returns a channel that is closed when the escape key is pressed.
func (k *Keypad) Quit() chan struct{} {
	return k.quit
}

// NewKeypad returns a new Keypad object and starts up a goroutine.
func NewKeypad(mmu Mmu, runSetup bool) *Keypad {
	commander := NewCommander("keypad")
	keys := map[Key]valueChan{
		KeyQuit:     valueChan{1, make(chan bool, 1)},
		KeyUp:       valueChan{1, make(chan bool, 1)},
		KeyDown:     valueChan{1, make(chan bool, 1)},
		KeyLeft:     valueChan{1, make(chan bool, 1)},
		KeyRight:    valueChan{1, make(chan bool, 1)},
		KeyB:        valueChan{1, make(chan bool, 1)},
		KeyA:        valueChan{1, make(chan bool, 1)},
		KeySelect:   valueChan{1, make(chan bool, 1)},
		KeyStart:    valueChan{1, make(chan bool, 1)},
	}
	mmuKeys := AddressKeys(0)
	mmuKeys = mmu.LockAddr(AddrP1, mmuKeys)
	kp := &Keypad{
		CommanderInterface: commander,
		mmu:                mmu,
		mmuKeys:            mmuKeys,
		keys:               keys,
		runSetup:           runSetup,
		quit:               make(chan struct{}),
	}
	cmdHandlers := map[Command]CommandFn{
		CmdKeyDown:  kp.cmdKeyDown,
		CmdKeyUp:    kp.cmdKeyUp,
		CmdString:   kp.cmdString,
		CmdKeyCheck: kp.cmdKeyCheck,
	}
	// no state functions so cmds are synchronous
	commander.start(nil, cmdHandlers, nil)
	go kp.loopKeyboard()
	mmu.SetKeypad(kp)
	return kp
}

func (k *Keypad) String() string {
	resp := make(chan string)
	k.RunCommand(CmdString, resp)
	return <-resp
}

func (k *Keypad) cmdString(resp interface{}) {
	if resp, ok := resp.(chan string); !ok {
		panic("invalid command response type")
	} else {
		resp <- k.str()
	}
}

func (k *Keypad) str() string {
	s := ""
	for key, vc := range k.keys {
		if vc.v == 1 {
			s += "  " + key.String() + "  "
		} else {
			s += " [" + key.String() + "] "
		}
	}
	return s
}

func (k *Keypad) cmdKeyDown(data interface{}) {
	if key, ok := data.(Key); !ok {
		panic("invalid command response type")
	} else {
		if k.keys[key].v == 1 { // inputs are pulled high
			k.keys[key] = valueChan{0, k.keys[key].c}
			c := k.keys[key].c
			go func() {
				// clear channel
				for loop := true; loop; {
					select {
					case <-c:
					default:
						loop = false
					}
				}
				// loop while we get at least one keypress
				for gotOne := true; gotOne; {
					timeout := time.After(200 * time.Millisecond)
					gotOne = false
					for loop := true; loop; {
						select {
						case <-c:
							gotOne = true
						case <-timeout:
							loop = false
						}
					}
				}
				k.RunCommand(CmdKeyUp, data)
			}()
			k.mmu.SetInterrupt(InterruptKeypad, k.mmuKeys)
		} else {
			// this chan has a buffer of 1, so even though the write is
			// non-blocking one keypress can be queued.
			select {
			case k.keys[key].c <- true:
			default:
			}
		}
	}
}

func (k *Keypad) cmdKeyUp(data interface{}) {
	if key, ok := data.(Key); !ok {
		panic("invalid command response type")
	} else {
		k.keys[key] = valueChan{1, k.keys[key].c}
	}
}

func (k *Keypad) cmdKeyCheck(data interface{}) {
	b, _ := k.mmu.ReadIoByte(AddrP1, k.mmuKeys)
	p15 := (b & 0x20) >> 5
	p14 := (b & 0x10) >> 4

	p13 := (p14 | k.keys[KeyRight].v) & (p15 | k.keys[KeyA].v)
	p12 := (p14 | k.keys[KeyLeft].v) & (p15 | k.keys[KeyB].v)
	p11 := (p14 | k.keys[KeyUp].v) & (p15 | k.keys[KeySelect].v)
	p10 := (p14 | k.keys[KeyDown].v) & (p15 | k.keys[KeyStart].v)

	p1310 := p10 | (p11 << 1) | (p12 << 2) | (p13 << 3)

	k.writeByte(AddrP1, p1310)
}

func (kp *Keypad) readByte(addr Word) Byte {
	return kp.mmu.ReadByteAt(addr, kp.mmuKeys)
}

func (kp *Keypad) writeByte(addr Word, b Byte) {
	kp.mmu.WriteByteAt(addr, b, kp.mmuKeys)
}

func (kp *Keypad) loopKeyboard() {
	b := make([]byte, 1)
	for {
		os.Stdin.Read(b)
		if b[0] == '\x1b' {
			// Escape closes the quit channel to signal the main loop to exit.
			close(kp.quit)
			return
		} else if key, ok := keyMap[b[0]]; ok {
			kp.RunCommand(CmdKeyDown, key)
		}
	}
}
