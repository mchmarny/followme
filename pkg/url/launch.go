package url

import (
	"fmt"
	"os/exec"
	"runtime"

	"github.com/pkg/errors"
)

// Open launches passed url in a platform specific way
func Open(url string) error {
	var err error

	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		err = exec.Command("open", url).Start()
	default:
		err = fmt.Errorf("unsupported platform")
	}
	if err != nil {
		return errors.Wrapf(err, "error opening URL: %s", url)
	}
	return nil
}
