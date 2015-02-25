package power

import (
"github.com/codeskyblue/go-sh"
"os"
)

func init() {
	_, err := os.Stat("/usr/bin/cpower")
	if os.IsNotExist(err) {
	      panic("File /usr/bin/cpower does not exist")
	}
}

// NOTE: Because on BBB any of libraries for gpio have problem
// with reading from sysfs (bad file descriptor or always 0)
// I decided to use simple bashscript for handling power
// It is not as bad as it looks :D
func Switch(command string, machines string) {
	sh.Command("/usr/bin/cpower", command, machines).Run()
}
