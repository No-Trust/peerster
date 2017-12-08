package common

import (
  "log"
  "os"
)

func CheckError(err error) {
	if err != nil {
		log.Fatal("Error : ", err)
		os.Exit(-1)
	}
}

func CheckRead(err error) bool {
	if err != nil {
		log.Print("Read Error ", err)
		return true
	}
	return false
}
