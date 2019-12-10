package main

import (
	"os"

	"vftpd"

	"github.com/shiena/ansicolor"
	log "github.com/sirupsen/logrus"
)

func main() {
	log.SetFormatter(&log.TextFormatter{ForceColors: true})
	log.SetOutput(ansicolor.NewAnsiColorWriter(os.Stdout))
	err := vftpd.ListenAndServe("0.0.0.0", 9090)
	if err != nil {
		log.Errorln(err)
	}
}
