package jibi

// A clock replicates ticks from a source (cpu) to other things that need it

// A ClockType is simply the type used for all clocks
type ClockType uint8

type Clock struct {
	dests    []chan ClockType
	attacher chan chan ClockType
}

// NewClock creates a new clock that will read from source.
func NewClock() *Clock {
	// can attach 10 destinations before AddCycles, bad
	return &Clock{attacher: make(chan chan ClockType, 10)}
}

func (c *Clock) Attach() chan ClockType {
	dest := make(chan ClockType)
	c.attacher <- dest
	return dest
}

// Send cycle count to all destinations
func (c *Clock) AddCycles(cycles ClockType) {
	// attach all pending
	for attaching := true; attaching; {
		select {
		case d := <-c.attacher:
			c.dests = append(c.dests, d)
		default:
			attaching = false
		}
	}
	// broadcast to all destinations
	for _, d := range c.dests {
		d <- cycles
	}
}
