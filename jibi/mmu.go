package jibi

type Mmu struct {
	// communications
	read  chan mmuRead
	write chan mmuWrite
	cmd   chan mmuCommand

	// memory banks

	// internal state
	biosFinished bool
}

func NewMmu() *Mmu {
	m := &Mmu{make(chan mmuRead), make(chan mmuWrite), make(chan mmuCommand), false}
	go loopMmu(m) //m.loop()
	return m
}

// TODO: rename
func (m *Mmu) rb(addr Worder) byte {
	return 0
}

// TODO: rename
func (m *Mmu) wb(addr Worder, b byte) {
}

func loopMmu(m *Mmu) {
	//func (m Mmu) loop() {
	for {
		select {
		case mr := <-m.read:
			mr.resp <- m.rb(mr.addr)
		case mr := <-m.write:
			m.wb(mr.addr, mr.data)
		case mr := <-m.cmd:
			switch mr {
			case mmuCmdStop:
				break
			case mmuCmdUnloadBios:
				m.biosFinished = true
				break
			default:
				panic("unknown mmu command")
			}
		}
	}
}

type mmuWrite struct {
	addr Word
	data byte
}

type mmuRead struct {
	addr Word
	resp chan byte
}

type mmuCommand uint8

const (
	mmuCmdStop mmuCommand = iota
	mmuCmdUnloadBios
)

type MmuConnection struct {
	read  chan mmuRead
	write chan mmuWrite
	cmd   chan mmuCommand
	resp  chan byte
}

func (m *Mmu) Connect() MmuConnection {
	c := MmuConnection{m.read, m.write, m.cmd, make(chan byte)}
	return c
}

func (m MmuConnection) writeByte(addr Worder, b Byter) {
	m.write <- mmuWrite{Word(addr.Uint16()), b.Uint8()}
}

func (m MmuConnection) readByte(addr Worder) Byte {
	m.read <- mmuRead{Word(addr.Uint16()), m.resp}
	return Byte(<-m.resp)
}

// write low bytes first
func (m MmuConnection) writeWord(addr Worder, w Worder) {
	m.writeByte(addr, w.Low())
	m.writeByte(addr.Inc(), w.High())
}

func (m MmuConnection) readWord(addr Worder) Word {
	l := m.readByte(addr)
	h := m.readByte(addr.Inc())
	return bytesToWord(h, l)
}

func (m MmuConnection) unloadBios() {
	m.cmd <- mmuCmdUnloadBios
}
