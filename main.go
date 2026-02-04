package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
)

const BufSize = 4096

func receiveFromPipe(pipename string, comms chan string) error {
	pipe, err := os.Open(pipename)
	if err != nil {
		return err
	}
	buf := make([]byte, BufSize)
	for {
		n, err := pipe.Read(buf)
		if err != nil && err != io.EOF {
			return err
		}

		if n > 0 {
			comms <- string(buf[:n])
		}
	}
}

func receiveFromStdin(comms chan string) error {
	input := os.Stdin
	buf := make([]byte, BufSize)
	for {
		n, err := input.Read(buf)
		if err != nil {
			return err
		}

		if n > 0 {
			comms <- string(buf[:n])
		}
	}
}

func listen(comms chan string, isClosed chan bool, stdin bool, inputs ...string) {
	for _, i := range inputs {
		input := i
		go func() {
			err := receiveFromPipe(input, comms)
			if err != io.EOF {
				log.Println(err)
			}
		}()
	}
	if stdin {
		go func() {
			err := receiveFromStdin(comms)
			if err != io.EOF {
				log.Println(err)
			}
			isClosed <- true
		}()
	}
}

// func cleanup()

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "%s allows to replace the input of a program with multiple pipes and stdin\n", os.Args[0])
		fmt.Println("\nUsage:")
		flag.PrintDefaults()
	}
	stdin := flag.Bool("s", false, "receive from stdin")
	flag.Parse()

	args := flag.Args()
	inputs := []string{}
	for _, a := range args {
		inputs = append(inputs, a)
	}

	comms := make(chan string)
	isClosed := make(chan bool)

	listen(comms, isClosed, *stdin, inputs...)

	for {
		select {
		case line := <-comms:
			_, err := os.Stdout.WriteString(line)
			if err != nil {
				log.Println(err)
			}
		case <-isClosed:
			log.Println("Stdin is closed")
			os.Exit(0)
		}
	}
}
