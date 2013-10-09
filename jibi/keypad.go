package jibi

import (
//	"fmt"
	"os"
	"os/exec"
	"time"
)

// up     0x77 w
// down   0x73 s
// left   0x61 a
// right  0x64 d
// b      0x2E .
// a      0x2F /
// select 0x5C \
// start  0x0A <enter>

// A Key is one of the 8 buttons.
type Key uint8

// List of 8 buttons.
const (
	KeyUp Key = iota
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

type valueChan struct {
	v Byte
	c chan bool
}

// A Keypad manages reading the actual key input, and the button states.
type Keypad struct {
	CommanderInterface

	mmu     *Mmu
	mmuKeys AddressKeys

	p1013low bool

	keys map[Key]valueChan
}

func setupInput() {
	// disable input buffering
	exec.Command("stty", "-F", "/dev/tty", "cbreak", "min", "1").Run()
	// do not display entered characters on the screen
	exec.Command("stty", "-F", "/dev/tty", "-echo").Run()
}

// NewKeypad returns a new Keypad object and starts up a goroutine.
func NewKeypad(mmu *Mmu, runSetup bool) *Keypad {
	if runSetup {
		setupInput()
	}
	commander := NewCommander("keypad")
	keys := map[Key]valueChan{
		KeyUp:     valueChan{1, make(chan bool)},
		KeyDown:   valueChan{1, make(chan bool)},
		KeyLeft:   valueChan{1, make(chan bool)},
		KeyRight:  valueChan{1, make(chan bool)},
		KeyB:      valueChan{1, make(chan bool)},
		KeyA:      valueChan{1, make(chan bool)},
		KeySelect: valueChan{1, make(chan bool)},
		KeyStart:  valueChan{1, make(chan bool)},
	}
	mmuKeys := AddressKeys(0)
	mmuKeys = mmu.LockAddr(AddrP1, mmuKeys)
	kp := &Keypad{
		CommanderInterface: commander,
		mmu:                mmu,
		mmuKeys:            mmuKeys,
		keys:               keys,
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
		if k.keys[key].v == 0 {
			k.keys[key] = valueChan{0, k.keys[key].c}
			c := k.keys[key].c
			go func() {
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
			//k.mmu.SetInterrupt(InterruptKeypad, k.mmuKeys)
		} else {
			k.keys[key].c <- true
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
/*
	b := k.mmu.ReadIoByte(AddrP1, k.mmuKeys)
	p15 := (b & 0x20) >> 5
	p14 := (b & 0x10) >> 4

	p1310 := p14&k.keys[KeyRight].v | p15&k.keys[KeyA].v +
		(p14&k.keys[KeyLeft].v|p15&k.keys[KeyB].v)<<1 +
		(p14&k.keys[KeyUp].v|p15&k.keys[KeySelect].v)<<2 +
		(p14&k.keys[KeyDown].v|p15&k.keys[KeyStart].v)<<3

	fmt.Println(p1310)
*/
	//k.writeByte(AddrP1, p1310)
}

func (kp *Keypad) readByte(addr Worder) Byte {
	return kp.mmu.ReadByteAt(addr, kp.mmuKeys)
}

func (kp *Keypad) writeByte(addr Worder, b Byter) {
	kp.mmu.WriteByteAt(addr, b, kp.mmuKeys)
}

func (kp *Keypad) loopKeyboard() {
	b := make([]byte, 1)
	for {
		os.Stdin.Read(b)
		switch b[0] {
		case 0x77: // w
			kp.RunCommand(CmdKeyDown, KeyUp)
		case 0x73: // s
			kp.RunCommand(CmdKeyDown, KeyDown)
		case 0x61: // a
			kp.RunCommand(CmdKeyDown, KeyLeft)
		case 0x64: // d
			kp.RunCommand(CmdKeyDown, KeyRight)
		case 0x2E: // .
			kp.RunCommand(CmdKeyDown, KeyB)
		case 0x2F: // /
			kp.RunCommand(CmdKeyDown, KeyA)
		case 0x5C: // \
			kp.RunCommand(CmdKeyDown, KeySelect)
		case 0x0A: // <enter>
			kp.RunCommand(CmdKeyDown, KeyStart)
		case 0x70: // p
			panic("KeyPanic")
		}
	}
}
