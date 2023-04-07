package main

import (
	"log"
	"vftpd"
)

func main() {
	err := vftpd.ListenAndServe("0.0.0.0", 9000)
	if err != nil {
		log.Println(err)
	}
}
