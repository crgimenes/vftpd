package vftpd

import (
	"fmt"
	"io"
	"net"
	"strings"

	log "github.com/sirupsen/logrus"
)

type section struct {
	username string
	password string
}

func ListenAndServe(ip string, port int) error {
	log.Println("vftpd listen at port", port)
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

func write(w io.Writer, code int, message string) {
	s := fmt.Sprintf("%d %s\r\n", code, message)
	w.Write([]byte(s))
}

func closer(c io.Closer) {
	err := c.Close()
	if err != nil {
		log.Errorln(err)
	}
}

func run(w io.Writer, cmd string, s *section) error {
	cmd = strings.TrimSpace(cmd)
	log.Printf(">> %q\n", cmd)
	p := strings.Split(cmd, " ")
	c := strings.ToLower(p[0])

	switch c {
	case "", "noop":
		// noop
	case "quit", "bye", ":q":
		write(w, 221, "have a nice day")
		return io.EOF
	case "syst":
		write(w, 215, "UNIX Type: L8.")
	case "user":
		if len(p) != 2 {
			write(w, 550, "format error")
			break
		}
		s.username = p[1]
		write(w, 331, "Username ok, send password.")
	case "pass":
		s.password = cmd[5:]
		//TODO: validate password
		log.Printf(">>%q, %q\n", s.username, s.password)
		write(w, 230, " Login successful.")
		break
	default:
		write(w, 550, "not supported")
	}
	return nil
}

func doService(conn net.Conn) {
	defer closer(conn)
	s := &section{}
	write(conn, 220, "vftpd")
	buf := make([]byte, 512)
	for {
		n, err := conn.Read(buf)
		if err != nil {
			log.Errorln("error reading from client", err.Error())
			return
		}
		cmd := string(buf[:n])
		err = run(conn, cmd, s)
		if err == io.EOF {
			return
		}
		if err != nil {
			log.Errorln(err)
			return
		}
	}
}
