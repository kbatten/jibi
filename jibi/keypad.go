package jibi

import (
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

	p14  Byte
	p15  Byte
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
		KeyUp:     valueChan{0, make(chan bool)},
		KeyDown:   valueChan{0, make(chan bool)},
		KeyLeft:   valueChan{0, make(chan bool)},
		KeyRight:  valueChan{0, make(chan bool)},
		KeyB:      valueChan{0, make(chan bool)},
		KeyA:      valueChan{0, make(chan bool)},
		KeySelect: valueChan{0, make(chan bool)},
		KeyStart:  valueChan{0, make(chan bool)},
	}
	kp := &Keypad{
		CommanderInterface: commander,
		keys:               keys,
	}
	cmdHandlers := map[Command]CommandFn{
		CmdReadByteAt:  kp.cmdReadByteAt,
		CmdWriteByteAt: kp.cmdWriteByteAt,
		CmdKeyDown:     kp.cmdKeyDown,
		CmdKeyUp:       kp.cmdKeyUp,
		CmdString:      kp.cmdString,
	}
	// no state functions so cmds are synchronous
	commander.start(nil, cmdHandlers, nil)
	go loopKeyboard(kp)

	mmu.HandleMemory(AddrP1, AddrP1, kp)
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
		if vc.v == 0 {
			s += "  " + key.String() + "  "
		} else {
			s += " [" + key.String() + "] "
		}
	}
	return s
}

// ReadByteAt reads a single byte from the keypad at the specified address.
func (k *Keypad) ReadByteAt(addr Worder, b chan Byte) {
	req := ReadByteAtReq{addr.Word(), b}
	k.RunCommand(CmdReadByteAt, req)
}

// WriteByteAt writes a single byte to the keypad at the specified address.
func (k *Keypad) WriteByteAt(addr Worder, b Byter) {
	req := WriteByteAtReq{addr.Word(), b.Byte()}
	k.RunCommand(CmdWriteByteAt, req)
}

func (k *Keypad) cmdReadByteAt(resp interface{}) {
	if req, ok := resp.(ReadByteAtReq); !ok {
		panic("invalid command response type")
	} else {
		req.b <- k.readByte(req.addr)
	}
}

func (k *Keypad) cmdWriteByteAt(resp interface{}) {
	if req, ok := resp.(WriteByteAtReq); !ok {
		panic("invalid command response type")
	} else {
		k.writeByte(req.addr, req.b)
	}
}

func (k *Keypad) cmdKeyDown(data interface{}) {
	if key, ok := data.(Key); !ok {
		panic("invalid command response type")
	} else {
		if k.keys[key].v == 0 {
			k.keys[key] = valueChan{1, k.keys[key].c}
			c := k.keys[key].c
			go func() {
				for gotOne := true; gotOne; {
					timeout := time.After(500 * time.Millisecond)
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
		} else {
			k.keys[key].c <- true
		}
	}
}

func (k *Keypad) cmdKeyUp(data interface{}) {
	if key, ok := data.(Key); !ok {
		panic("invalid command response type")
	} else {
		k.keys[key] = valueChan{0, k.keys[key].c}
	}
}

func loopKeyboard(kp *Keypad) {
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

func (k *Keypad) readByte(addr Worder) Byte {
	return k.p14&k.keys[KeyRight].v | k.p15&k.keys[KeyA].v +
		(k.p14*k.keys[KeyLeft].v|k.p15&k.keys[KeyB].v)<<1 +
		(k.p14*k.keys[KeyUp].v|k.p15&k.keys[KeySelect].v)<<2 +
		(k.p14*k.keys[KeyDown].v|k.p15&k.keys[KeyStart].v)<<3
}

func (k *Keypad) writeByte(addr Worder, b Byter) {
	k.p15 = b.Byte() & 0x20
	k.p14 = b.Byte() & 0x10
}
