package flags

import (
	"fmt"
	"strings"
)

func ParseAddr(flagArgs string, value *string) error {
	val := strings.Split(flagArgs, ":")

	if len(val) != 2 {
		return fmt.Errorf("addr must be ip:port")
	}

	*value = flagArgs

	return nil
}
