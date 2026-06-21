package jibi

import (
	"os/exec"
)

func (k *Keypad) Init() {
	if k.runSetup == true {
		// save tty config
		out, err := exec.Command("stty", "-F", "/dev/tty", "-g").CombinedOutput()
		if err != nil {
			panic("stty")
		}
		// trim the trailing newline
		k.ttyConfig = string(out[:len(out)-1])

		// disable input buffering
		exec.Command("stty", "-F", "/dev/tty", "cbreak", "min", "1").Run()
		// do not display entered characters on the screen
		exec.Command("stty", "-F", "/dev/tty", "-echo").Run()
	}
}

func (k *Keypad) Close() {
	if k.runSetup == true {
		// restore tty config
		exec.Command("stty", "-F", "/dev/tty", k.ttyConfig).Run()
	}
}
