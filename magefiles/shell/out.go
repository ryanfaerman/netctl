package shell

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
)

func Exec(cmd string, args ...string) (string, error) {
	c := exec.Command(cmd, args...)
	c.Env = os.Environ()
	// c.Stderr = os.Stderr
	b, err := c.Output()
	if err != nil {
		return "", errors.Join(err, fmt.Errorf(`failed to run %v %q`, cmd, args))
	}
	return string(b), nil
}
