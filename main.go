// package main
package main

import (
	"crypto/sha256"
	"encoding/hex"
	"flag"
	"fmt"
	"io"
	"os"
)

// main
func main() {
	conf := parseFlags() // TODO

	//fmt.Println(force)
	//fmt.Println(quiet)

	if conf.Source == conf.Target && conf.Source != "" {
		fmt.Println("Error: source must not be the same as target")
		os.Exit(1)
	} else if conf.Source == "" {
		flag.Usage()
		os.Exit(1)
	}

	if err := run(conf); err != nil {
		fmt.Println(err)
		os.Exit(2)
	}
}

type configuration struct {
	Source  string
	Target  string
	Force   bool
	NoCheck bool
	Quiet   bool
	BufSize int
}

// parseFlags interprets the command line flags.
func parseFlags() configuration {
	conf := configuration{}
	//flag.Usage = func() { // TODO
	//	fmt.Fprintf(os.Stderr, "Usage ofAARRGH %s:\n", os.Args[0])
	//	flag.PrintDefaults()
	// doc package here
	//}
	nyi := "\n(not yet implemented)"
	sourcePtr := flag.String("source", "", "the source image file or "+
		"device")
	targetPtr := flag.String("target", "", "the target device or image"+
		" file")
	forcePtr := flag.Bool("force", false, "force the operation without"+
		" confirmation and summary of pending actions"+nyi)
	quietPtr := flag.Bool("quiet", false, "don't show progress"+nyi)
	noCheckPtr := flag.Bool("nocheck", false, "skip checksum creation "+
		"and comparison")
	bufSizePtr := flag.Int("buffersize", 1024, "buffer size in kilobytes")
	flag.Parse()

	conf.Source = *sourcePtr
	conf.Target = *targetPtr
	conf.Force = *forcePtr
	conf.Quiet = *quietPtr
	conf.NoCheck = *noCheckPtr
	conf.BufSize = *bufSizePtr

	return conf
}

// run wraps the process together.
func run(c configuration) error {
	var bs = 1024 * c.BufSize
	var output, input *os.File
	var br, bw int64
	var err error

	if input, err = os.Open(c.Source); err != nil {
		return err
	}

	if _, err := os.Stat(c.Target); os.IsNotExist(err) {
		if output, err = os.Create(c.Target); err != nil {
			return err
		}
	} else {
		if output, err = os.OpenFile(c.Target, os.O_WRONLY,
			os.ModeAppend); err != nil {
			return err
		}
	}

	if br, bw, err = doWrite(input, output, bs); err != nil {
		return err
	}

	if err := input.Close(); err != nil {
		return err
	}

	if err := output.Close(); err != nil {
		return err
	}

	fmt.Println(br, bw)
	if !c.NoCheck { // skip this if nocheck is true
		if err = compare(c.Source, c.Target, br, bw); err != nil {
			return err
		}
	}

	return nil
}

// doWrite performs the read-write process. input and output can be devices,
// partitions or image files containing raw device data.
func doWrite(input, output *os.File, bs int) (int64, int64, error) {
	var bytesRead, bytesWritten int64
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

		bytesRead += int64(br)
		bytesWritten += int64(bw)
	}

	return bytesRead, bytesWritten, err
}

// compare the sha256 checksums of input and output. Note that those can
// never match when using a cryptographically sound random source to generate
// the input.
func compare(source, target string, br, bw int64) error {
	queue := make(chan string, 2)

	go func(f string) {
		s, err := sha256sumFile(f, br)
		if err != nil {
			fmt.Println(err)
			os.Exit(4)
		}
		queue <- s
	}(source)

	go func(f string) {
		s, err := sha256sumFile(f, bw)
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

	if s1 == s2 {
		fmt.Println("checksums match")
	} else {
		fmt.Println("checksums do NOT match")
		fmt.Println("This means that something went wrong," +
			"\nor that your input has changed since the" +
			"\nstart of this program." +
			"\nFor more detailed information, please call this" +
			"\nprogram again with the -h flag to read the" +
			"\nbuilt-in documentation.")
	}

	return nil
}

// wrapShaFile adds the filename and checksum to a string slice. This
// makes the hashes identifiable, e.g. in case of a bad backup.
//func wrapShaFile(file string, bytes int64) (string, error) {
//
//}

// sha256sumFile takes the pathname of a file to generate a sha256 hash from.
// bytes tells the function how many bytes should be read. In case of USB
// media, block devices or partitions the medium will often be larger than the
// flashed image file and creating a checksum of the whole device would result
// in a different checksum.
// The checksum is returned as a human-readable hexadecimal string.
func sha256sumFile(file string, bytes int64) (string, error) {
	var f *os.File
	var s string
	var err error

	if f, err = os.Open(file); err != nil {
		return s, err
	}

	defer f.Close()

	h := sha256.New()
	if _, err := io.CopyN(h, f, bytes); err != nil {
		f.Close()
		return s, err
	}

	s = hex.EncodeToString(h.Sum(nil))

	return s, nil
}
