package jibi

import (
	"fmt"
	"time"
)

// Options holds various options.
type Options struct {
	Skipbios bool
	Render   bool
	Quick    bool
}

// Jibi is the glue that holds everything together.
type Jibi struct {
	O Options

	mmu  *Mmu
	cpu  *Cpu
	irq  *Irq
	lcd  Lcd
	gpu  *Gpu
	cart *Cartridge
	kp   *Keypad
}

// New returns a new Jibi in a Paused state.
func New(rom []Byte, options Options) Jibi {
	mmu := NewMmu(bios)
	irq := NewIrq(mmu)
	cpu := NewCpu(mmu, irq)
	lcd := NewLcdASCII()
	gpu := NewGpu(mmu, irq, lcd, cpu.Clock())
	cart := NewCartridge(mmu, rom)
	kp := NewKeypad(mmu, options.Quick)

	if options.Skipbios {
		mmu.RunCommand(CmdUnloadBios, nil)
	}
	if !options.Render {
		lcd.DisableRender()
	}

	return Jibi{options, mmu, cpu, irq, lcd, gpu, cart, kp}
}

// RunCommand displatches a command to the correct piece.
func (j Jibi) RunCommand(cmd Command, resp chan string) {
	if cmd < cmdMMU {
		j.mmu.RunCommand(cmd, resp)
	} else if cmd < cmdCPU {
		j.cpu.RunCommand(cmd, resp)
	} else if cmd < cmdGPU {
		j.gpu.RunCommand(cmd, resp)
	} else if cmd < cmdKEYPAD {
		j.kp.RunCommand(cmd, resp)
	} else if cmd < cmdALL {
		j.cpu.RunCommand(cmd, resp)
		j.mmu.RunCommand(cmd, resp)
		j.gpu.RunCommand(cmd, resp)
		j.kp.RunCommand(cmd, resp)
	} else if cmd < cmdCPUGPU {
		j.cpu.RunCommand(cmd, resp)
		j.gpu.RunCommand(cmd, resp)
	}
}

// Run starts the Jibi and waits till it ends before returning.
func (j Jibi) Run() {
	//resp := make(chan string)
	//j.RunCommand(CmdNotifyInstruction, resp)
	//j.RunCommand(CmdNotifyUnhandledMemory, resp)
	//j.RunCommand(CmdNotifyFrame, resp)
	cpuClk := j.cpu.Clock()
	j.Play()
	ticker := time.NewTicker(time.Second)
	for running := true; running; {
		select {
		case <-ticker.C:
			cpuHz := float64(0)
			for loop := true; loop; {
				select {
				case t := <-cpuClk:
					cpuHz += float64(t)
				default:
					loop = false
				}
			}
			if j.O.Render {
				fmt.Printf("\x1B[s\x1B[59;0H\x1B[K\n"+
					"\x1B[K\n"+
					"\x1B[K\n"+
					"\x1B[K\n"+
					"\x1B[K\n"+
					"\x1B[K"+
					"\x1B[59;0H%s\n%s\n"+
					"cpu: %.2fMhz\x1B[u", j.cpu, j.kp, cpuHz/1e6)
			} else {
				fmt.Printf("%s\n%s\ncpu: %.2fMhz\n", j.cpu, j.kp, cpuHz/1e6)
			}
			if j.O.Quick {
				running = false
			}
		}
	}
	ticker.Stop()
	j.Stop()
}

// Play starts the Jibi and returns immediately.
func (j Jibi) Play() {
	j.RunCommand(CmdPlay, nil)
}

// Pause pauses the Jibi and returns immediately.
func (j Jibi) Pause() {
	j.RunCommand(CmdPause, nil)
}

// Stop stops the Jibi and all its goroutines and returns immediately.
func (j Jibi) Stop() {
	j.RunCommand(CmdStop, nil)
}
