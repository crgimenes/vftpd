package vftpd

import (
	"fmt"
	"net"

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

func doService(conn net.Conn) {
	defer conn.Close()
	w(conn, 220, "vftpd")
	buf := make([]byte, 4096)

	for {
		n, err := conn.Read(buf)
		if err != nil {
			log.Errorln("error reading from client", err.Error())
			return
		}
		log.Printf(">> %q\n", string(buf[:n]))
	}
}
