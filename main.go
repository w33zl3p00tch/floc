package main

import (
	"crypto/sha256"
	"encoding/hex"
	"flag"
	"fmt"
	"io"
	"os"
)

func main() {
	var source, target string
	var force, quiet bool
	source, target, force, quiet = parseFlags()

	fmt.Println(force)
	fmt.Println(quiet)

	if source == target {
		fmt.Println("Error: source must not be the same as target")
		os.Exit(1)
	}

	if err := run(source, target, force, quiet); err != nil {
		fmt.Println(err)
		os.Exit(2)
	}
}

func parseFlags() (string, string, bool, bool) {
	sourcePtr := flag.String("source", "", "the source image file or "+
		"device")
	targetPtr := flag.String("target", "", "the target device or image file")
	forcePtr := flag.Bool("force", false, "force the operation without"+
		" confirmation and summary of pending actions")
	quietPtr := flag.Bool("quiet", false, "don't show progress")
	flag.Parse()

	source := *sourcePtr
	target := *targetPtr
	force := *forcePtr
	quiet := *quietPtr

	return source, target, force, quiet
}

func run(source, target string, f, q bool) error {
	const bs = 1024 * 1024 // 1 megabyte buffer size
	var output, input *os.File
	var br, bw uint64 = 0, 0
	var err error

	if input, err = os.Open(source); err != nil {
		return err
	}

	if _, err := os.Stat(target); os.IsNotExist(err) {
		if output, err = os.Create(target); err != nil {
			return err
		}
	} else {
		if output, err = os.OpenFile(target, os.O_WRONLY,
			os.ModeAppend); err != nil {
			return err
		}
	}

	if br, bw, err = doWrite(input, output, bs); err != nil {
		return err
	}

	if err := input.Close(); err != nil {
		return (err)
	}

	if err := output.Close(); err != nil {
		return (err)
	}

	fmt.Println(br, bw)
	compare(source, target)

	return nil
}

func doWrite(input, output *os.File, bs int) (uint64, uint64, error) {
	var bytesRead, bytesWritten uint64
	var eof bool = false
	var err error

	for {
		buffer := make([]byte, bs)
		br, err := input.Read(buffer)
		if err == io.EOF {
			eof = true
		} else if err != nil {
			return bytesRead, bytesWritten, err
		}

		if eof && br == 0 {
			fmt.Println("copied")
			break
		}

		bw, err := output.Write(buffer[:br])
		if err != nil {
			return bytesRead, bytesWritten, err
		}

		bytesRead += uint64(br)
		bytesWritten += uint64(bw)
	}

	return bytesRead, bytesWritten, err
}

func compare(source, target string) error {
	queue := make(chan string, 2)

	go func(f string) {
		s, err := sha256sumFile(f)
		if err != nil {
			fmt.Println(err)
			os.Exit(4)
		}
		queue <- s
	}(source)

	go func(f string) {
		s, err := sha256sumFile(f)
		if err != nil {
			fmt.Println(err)
			os.Exit(4)
		}
		queue <- s
	}(target)

	s1 := <-queue
	s2 := <-queue

	fmt.Println(s1)
	fmt.Println(s2)

	if s1 == s1 {
		fmt.Println("checksums match")
	}

	return nil
}

func sha256sumFile(file string) (string, error) {
	var f *os.File
	var s string
	var err error

	if f, err = os.Open(file); err != nil {
		return s, err
	}

	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		f.Close()
		return s, err
	}

	s = hex.EncodeToString(h.Sum(nil))

	return s, nil
}
