package vftpd

import (
	"errors"
	"fmt"
	"io"
	"net"
	"strings"

	log "github.com/sirupsen/logrus"
)

func ListenAndServe(ip string, port int) error {
	log.Println("vftpd listen and serve")
	l, err := net.Listen("tcp", fmt.Sprintf("%s:%d", ip, port))
	if err != nil {
		return err
	}
	for {
		conn, err := l.Accept()
		if err != nil {
			log.Errorln("error accepting", err.Error())
			continue
		}
		log.Println(
			"connection request from",
			conn.RemoteAddr().String())
		go doService(conn)
	}
}

func w(conn net.Conn, code int, message string) {
	s := fmt.Sprintf("%d %s\r\n", code, message)
	conn.Write([]byte(s))
}

func closer(c io.Closer) {
	err := c.Close()
	if err != nil {
		log.Errorln(err)
	}
}

func run(cmd string) error {
	log.Printf(">> %q\n", s)
	p := strings.Split(s, " ")
	if len(p) == 0 {
		return errors.New("error parsing command")
	}

}

func doService(conn net.Conn) {
	defer closer(conn)
	w(conn, 220, "vftpd")
	buf := make([]byte, 512)
	for {
		n, err := conn.Read(buf)
		if err != nil {
			log.Errorln("error reading from client", err.Error())
			return
		}
		cmd := string(buf[:n])
		err = run(cmd)
		if err != nil {
			log.Errorln(err)
			return
		}
	}
}
