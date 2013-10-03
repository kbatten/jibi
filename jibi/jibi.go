package jibi

type Options struct {
	Skipbios bool
}

type Jibi struct {
	O Options
	Q chan error // quit channel

	m *Mmu
	c *Cpu
}

// TODO: move to mmu.go
// The MMU is the generic to<->from addressing interface that all modules will
// use. The MMU maps back into each module as needed. So a module can access
// its own memory locally, it can also access it through the MMU. A module
// must use the MMU for all external memory access.

func New(rom []Byte, options Options) Jibi {
	c := NewCpu()
	m := NewMmu()
	c.ConnectMmu(m)
	quit := make(chan error, 1)

	return Jibi{options, quit, m, c}
}

func (j Jibi) Run() error {
	j.Play()
	return <-j.Q
}

func (j Jibi) Play() {
	go func() {
		for i := 0; i < 5; i++ {
			j.c.Step() // TODO: have the cpu have its own go func
		}
		j.Q <- nil
	}()
}
