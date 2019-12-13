package vftpd

import (
	"fmt"
	"io"
	"math/rand"
	"net"
	"os"
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"
)

type section struct {
	username string
	password string
	localIP  string
	fileName string
	dataConn net.Conn
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
		write(w, 230, "Login successful.")
	case "pwd":
		write(w, 257, `"/" is the current directory.`)
	case "pasv":
		//s.mode = "passive"
		port := 50000 + rand.Intn(51009-50000)
		//TODO: try again if port is in use
		//TODO: add ipv6 support?
		addr, err := net.ResolveTCPAddr("tcp", "0.0.0.0:"+fmt.Sprintf("%d", port))
		if err != nil {
			log.Errorln(err)
			write(w, 550, "resolve TCP error: "+err.Error())
			return nil
		}
		li, err := net.ListenTCP("tcp", addr)
		if err != nil {
			log.Errorln(err)
			write(w, 550, "listen TCP error: "+err.Error())
			return nil
		}
		ip := strings.Join(strings.Split(s.localIP, "."), ",")
		h := strconv.Itoa(port >> 8)
		l := strconv.Itoa(port % 256)
		msg := "Entering passive mode (" + ip + "," + h + "," + l + ")"
		log.Println(msg)
		write(w, 227, msg)

		/*
		   227 Entering Passive Mode (127,0,0,1,133,62).
		   150 Here comes the directory listing.
		   drwxr-xr-x    8 1000     1000         4096 Dec 10 15:27 Projects
		   drwx------    2 1000     1000         4096 Dec 09 14:48 Teste
		   drwxr-xr-x    2 1000     1000         4096 Nov 17 11:50 bin
		   -rw-r--r--    1 1000     1000         1201 Nov 17 11:27 crg.eti.br
		   drwxr-xr-x    5 1000     1000         4096 Nov 16 12:30 go
		   226 Directory send OK.
		*/

		go (func() {
			s.dataConn, err = li.Accept()
			if err != nil {
				// TODO: handle error
				log.Errorln(err)
			}
			// TODO: handle data
			//closer(dataConn)
		})()
	case "type":
		write(w, 200, "set type to:"+p[1])

	case "list":
		write(w, 150, "Here comes the directory listing.")
		// TODO: implement list and remove fake files
		s.dataConn.Write([]byte(`drwxr-xr-x    8 1000     1000         4096 Dec 10 15:27 Projects
drwx------    2 1000     1000         4096 Dec 09 14:48 Teste
drwxr-xr-x    2 1000     1000         4096 Nov 17 11:50 bin
-rw-r--r--    1 1000     1000         1201 Nov 17 11:27 crg.eti.br
drwxr-xr-x    5 1000     1000         4096 Nov 16 12:30 go
`))
		write(w, 226, "Directory send OK.")
		closer(s.dataConn)
		//case "eprt":
		//case "port":
	case "stor":
		s.fileName = p[1]
		log.Println("storing file", s.fileName)
		defer closer(s.dataConn)

		write(w, 125, "Data connection already opened; transfer starting.")

		wo, err := os.Create("file_" + s.fileName)
		if err != nil {
			log.Errorln("error store file", s.fileName, err)
			break
		}
		defer closer(wo)
		n, err := io.Copy(wo, s.dataConn)
		if err != nil {
			log.Errorln("error store file", s.fileName, err)
		}
		log.Printf("copied %v bytes\n", n)
		err = wo.Sync()
		if err != nil {
			log.Errorln("error store file", s.fileName, err)
		}

		write(w, 250, "Requested file action okay, completed.")
	default:
		write(w, 500, "not supported")
	}
	return nil
}

func doService(conn net.Conn) {
	defer closer(conn)
	s := &section{}
	write(conn, 220, "vftpd")
	buf := make([]byte, 512)
	localAddr := conn.LocalAddr()
	log.Println(">>", localAddr.String())
	a := strings.Split(localAddr.String(), ":")
	s.localIP = a[0]
	for {
		n, err := conn.Read(buf)
		if err != nil {
			if err == io.EOF {
				log.Println("connection closed")
				return
			}
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
