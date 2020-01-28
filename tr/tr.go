package tr

import (
	"log"
)

func IsOK(err error) bool {
	if err == nil {
		return true
	}
	log.Println(err)
	return false
}
