package jibi

import (
	"fmt"
)

type Command int

const (
	CmdNil Command = iota

	CmdHandleMemory
	CmdUnloadBios
	CmdNotifyUnhandledMemory
	cmdMMU

	CmdClock
	CmdNotifyInstruction
	cmdCPU

	cmdIRQ

	cmdLCD

	CmdNotifyFrame
	cmdGPU

	cmdCART

	CmdKeyDown
	CmdKeyUp
	cmdKEYPAD

	CmdReadByteAt
	CmdWriteByteAt
	CmdString
	CmdPlay
	CmdPause
	CmdStop
	cmdALL

	cmdCPUGPU
)

func (c Command) String() string {
	switch c {
	case CmdNil:
		return "CmdNil"
	case CmdHandleMemory:
		return "CmdHandleMemory"
	case CmdUnloadBios:
		return "CmdUnloadBios"
	case CmdNotifyUnhandledMemory:
		return "CmdNotifyUnhandledMemory"
	case cmdMMU:
		return "cmdMMU"
	case CmdClock:
		return "CmdClock"
	case CmdNotifyInstruction:
		return "CmdNotifyInstruction"
	case cmdCPU:
		return "cmdCPU"
	case cmdIRQ:
		return "cmdIRQ"
	case cmdLCD:
		return "cmdLCD"
	case CmdNotifyFrame:
		return "CmdNotifyFrame"
	case cmdGPU:
		return "cmdGPU"
	case cmdCART:
		return "cmdCART"
	case CmdKeyDown:
		return "CmdKeyDown"
	case CmdKeyUp:
		return "CmdKeyUp"
	case cmdKEYPAD:
		return "cmdKEYPAD"
	case CmdReadByteAt:
		return "CmdReadByteAt"
	case CmdWriteByteAt:
		return "CmdWriteByteAt"
	case CmdString:
		return "CmdString"
	case CmdPlay:
		return "CmdPlay"
	case CmdPause:
		return "CmdPause"
	case CmdStop:
		return "CmdStop"
	case cmdALL:
		return "cmdALL"
	case cmdCPUGPU:
		return "cmdCPUGPU"
	}
	return "CmdUNKNOWN"
}

type CommandResponse struct {
	cmd  Command
	resp interface{}
}

type CommanderStateFn func(bool, uint32) (CommanderStateFn, bool, uint32, uint32)

type CommanderInterface interface {
	RunCommand(Command, interface{})
	Start(CommanderStateFn, map[Command]CommandFn, chan uint8)
}

type Commander struct {
	name string
	c    chan CommandResponse
}

func NewCommander(name string) Commander {
	c := Commander{name, make(chan CommandResponse)}
	return c
}

func (c Commander) Start(state CommanderStateFn, handlerFns map[Command]CommandFn, clk chan uint8) {
	go loopCommander(c, state, handlerFns, clk)
}

func (c Commander) RunCommand(cmd Command, resp interface{}) {
	c.c <- CommandResponse{cmd, resp}
}

func (c Commander) String() string {
	return c.name
}

type MemoryCommander interface {
	ReadByteAt(Worder) Byte
	WriteByteAt(Worder, Byter)
	CommanderInterface
}

type CommandFn func(interface{})

func loopCommander(c Commander, state CommanderStateFn, handlerFns map[Command]CommandFn, clk chan uint8) {
	playing := false
	first := true
	t := uint32(0)
	tnext := uint32(0) // time needed to run next state
	for running := true; running; {
		var cmdr CommandResponse
		if !playing || state == nil {
			//if c.String() != "mmu" { fmt.Println("A", c) }
			cmdr = <-c.c
			//if c.String() != "mmu" {fmt.Println("A", c, cmdr.cmd) }
		} else if t <= tnext {
			// we have enough cycles to run the next state without waiting for the clock
			select {
			case cmdr = <-c.c:
			default:
			}
		} else if t < tnext {
			if clk == nil {
				panic(fmt.Sprintf("Commander %s requires a clock", c))
			}
			select {
			case cmdr = <-c.c:
			case to := <-clk:
				t += (uint32(to))
			}
		}
		if cmdr.cmd != CmdNil {
			if cmdr.cmd == CmdPlay {
				playing = true
			} else if cmdr.cmd == CmdPause {
				playing = false
			} else {
				if _, ok := handlerFns[cmdr.cmd]; !ok {
					if cmdr.cmd != CmdStop {
						panic(fmt.Sprintf("Commander %s requires handler for %s", c, cmdr.cmd))
					}
				} else {
					handlerFns[cmdr.cmd](cmdr.resp)
				}
				if cmdr.cmd == CmdStop {
					running = false
					playing = false
				}
			}
		} else {
			// consume all commands before running next state
			if state != nil && playing && (tnext >= t || first) {
				state, first, t, tnext = state(first, t)
			} else if !playing {
				t = 0
			}
		}
	}
}
