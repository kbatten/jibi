package jibi

import "os"

// Options holds various options.
type Options struct {
	MaxTicks int
	LogInst  bool
	Skipbios bool
	Render   bool
	Keypad   bool
	Squash   bool
	Every    bool
}

// Jibi is the glue that holds everything together.
type Jibi struct {
	O Options

	mmu  Mmu
	cpu  *Cpu
	lcd  Lcd
	gpu  *Gpu
	cart *Cartridge
	kp   *Keypad
}

// New returns a new Jibi in a Paused state.
func New(rom []Byte, options Options) Jibi {
	cart := NewCartridge(rom)
	mmu := NewMmu(cart)
	cpu := NewCpu(mmu, bios)
	lcd := NewLcd(options.Squash)
	gpu := NewGpu(mmu, lcd, cpu.AttachClock())
	kp := NewKeypad(mmu, options.Keypad)

	if options.Skipbios {
		cpu.RunCommand(CmdUnloadBios, nil)
	}
	if !options.Render {
		lcd.DisableRender()
	}

	return Jibi{options, mmu, cpu, lcd, gpu, cart, kp}
}

// RunCommand displatches a command to the correct piece.
func (j Jibi) RunCommand(cmd Command, resp chan string) {
	if cmd < cmdCPU {
		j.cpu.RunCommand(cmd, resp)
	} else if cmd < cmdGPU {
		j.gpu.RunCommand(cmd, resp)
	} else if cmd < cmdKEYPAD {
		j.kp.RunCommand(cmd, resp)
	} else if cmd < cmdALL {
		j.cpu.RunCommand(cmd, resp)
		j.gpu.RunCommand(cmd, resp)
		j.kp.RunCommand(cmd, resp)
	}
}

// Run starts the Jibi and waits till it ends before returning.
func (j Jibi) Run() {
	j.kp.Init()
	defer j.kp.Close()

	j.lcd.Init()
	defer j.lcd.Close()

	var totalTicksClk chan ClockType
	if j.O.MaxTicks > 0 {
		totalTicksClk = j.cpu.AttachClock()
	}
	totalTicks := int(0)

	var instructions chan string
	var logFile *os.File
	if j.O.LogInst == true {
		instructions = j.cpu.AttachInstructions()
		var err error
		logFile, err = os.Create("jibi.log")
		if err != nil {
			panic(err)
		}
	}

	j.Play()

	for running := true; running; {
		select {
		case s := <-instructions:
			logFile.WriteString(s)
			logFile.WriteString("\n")
		case t := <-totalTicksClk:
			totalTicks += int(t)
			if totalTicks > j.O.MaxTicks {
				running = false
			}
		}
	}

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
