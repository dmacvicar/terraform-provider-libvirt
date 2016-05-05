package libvirt

import (
	"fmt"
)

var diskLetters []rune = []rune("abcdefghijklmnopqrstuvwxyz")

func DiskLetterForIndex(i int) string {

	q := i / len(diskLetters)
	r := i % len(diskLetters)
	letter := diskLetters[r]

	if q == 0 {
		return fmt.Sprintf("%c", letter)
	}

	return fmt.Sprintf("%s%c", DiskLetterForIndex(q - 1), letter)
}

