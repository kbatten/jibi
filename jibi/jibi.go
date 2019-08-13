package jibi

import "os"

// Options holds various options.
type Options struct {
	Status   bool
	MaxTicks int
	LogInst  bool
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
	lcd := NewLcd()
	gpu := NewGpu(mmu, lcd, cpu.AttachClock())
	kp := NewKeypad(mmu)

	return Jibi{options, mmu, cpu, lcd, gpu, cart, kp}
}

// Run starts the Jibi and waits till it ends before returning.
func (j Jibi) Run() {
	// init other hardware
	j.kp.Init()
	defer j.kp.Close()

	j.lcd.Init()
	defer j.lcd.Close()

	// MaxTicks
	var totalTicksClk chan ClockType
	if j.O.MaxTicks > 0 {
		totalTicksClk = j.cpu.AttachClock()
	}
	totalTicks := int(0)

	// LogInst
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
}
