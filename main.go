package main

import (
	"compress/gzip"
	"io"
	"log"
	"net"
	"os"
)

func main() {
	// Create a listener on a random port.
	listener, err := net.Listen("tcp", "127.0.0.1:")
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Server listening on: " + listener.Addr().String())
	done := make(chan struct{})
	go func() {
		defer func() { done <- struct{}{} }()
		for {
			conn, err := listener.Accept()
			if err != nil {
				log.Println(err)
				return
			}
			go func(c net.Conn) {
				defer func() {
					c.Close()
					done <- struct{}{}
				}()
				buf := make([]byte, 1024)
				for {
					n, err := c.Read(buf)
					if err != nil {
						if err != io.EOF {
							log.Println(err)
						}

						return
					}
					log.Printf("received: %q", buf[:n])
					log.Printf("bytes: %d", n)

				}

			}(conn)
		}
	}()

	conn, err := net.Dial("tcp", listener.Addr().String())
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Connected to server.")

	file, err := os.Open("./file.txt")
	if err != nil {
		log.Fatal(err)
	}

	pr, pw := io.Pipe()
	w, err := gzip.NewWriterLevel(pw, 7)
	if err != nil {
		log.Fatal(err)
	}
	go func() {
		n, err := io.Copy(w, file)
		if err != nil {
			log.Fatal(err)
		}
		w.Close()
		pw.Close()
		log.Printf("copied to piped writer via the compressed writer: %d", n)
	}()

	n, err := io.Copy(conn, pr)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("copied to connection: %d", n)

	conn.Close()
	<-done
	listener.Close()
	<-done
}
