package flags

import (
	"fmt"
	"strings"
)

// func ParseAddr(flagArgs string) (addr string, err error) {
// 	val := strings.Split(flagArgs, ":")
// 	if len(val) != 2 {
// 		return "", fmt.Errorf("addr must be ip:port")
// 	}

// 	addr = flagArgs

// 	return addr, nil
// }

func ParseAddr(flagArgs string, value *string) error {
	val := strings.Split(flagArgs, ":")
	if len(val) != 2 {
		return fmt.Errorf("addr must be ip:port")
	}

	*value = flagArgs

	return nil
}
