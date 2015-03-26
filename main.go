package main

import (
	"bytes"
	"errors"
	"flag"
	"io"
	"log"
	"os"
	"os/exec"
)

func main() {
	fp := flag.String("file", "", "name of file")
	inflate := flag.Bool("inflate", false, "inflate archive")
	deflate := flag.Bool("deflate", false, "deflate archive")
	stdout := flag.Bool("stdout", false, "write to file or to stdout when inflating")
	flag.Parse()
	// path
	if *deflate {
		err := xzWriter(*fp)
		if err != nil {
			log.Printf("Err: %v", err)
		}
	}
	if *inflate {
		if *stdout {
			r, err := xzReader(*fp, *stdout)
			n, err := io.Copy(os.Stdout, r)
			if err != nil {
				log.Printf("copied %d bytes with err: %v", n, err)
			} else {
				log.Printf("copied %d bytes", n)
			}
		} else {
			_, err := xzReader(*fp, *stdout)
			if err != nil {
				log.Printf("Err: %v", err)
			}
		}

	}
}

func xzReader(file string, stdout bool) (io.ReadCloser, error) {
	f, err := os.Open(file)
	if err != nil {
		log.Fatal(err)
	}
	if stdout {
		rpipe, wpipe := io.Pipe()
		cmd := exec.Command("xz", "--decompress", "--stdout")
		cmd.Stdin = f
		cmd.Stdout = wpipe
		// print the error
		cmd.Stderr = os.Stderr
		go func() {
			err := cmd.Run()
			wpipe.CloseWithError(err)
			defer f.Close()
		}()
		return rpipe, err
	} else {
		// Create an *exec.Cmd
		cmd := exec.Command("xz", "--decompress", file)
		// Stdout buffer
		cmdOutput := &bytes.Buffer{}
		// Attach buffer to command
		cmd.Stdout = cmdOutput
		// Stderr buffer
		cmderror := &bytes.Buffer{}
		// Attach buffer to command
		cmd.Stderr = cmderror
		// Execute command
		err := cmd.Run() // will wait for command to return
		if err != nil {
			errstr := string(cmderror.Bytes())
			err = errors.New(errstr)
		}
		return nil, err
	}
}

func xzWriter(file string) error {
	// Create an *exec.Cmd
	cmd := exec.Command("xz", file)
	//  buffer
	cmdOutput := &bytes.Buffer{}
	// Attach buffer to command
	cmd.Stdout = cmdOutput
	// Stderr buffer
	cmderror := &bytes.Buffer{}
	// Attach buffer to command
	cmd.Stderr = cmderror
	// Execute command
	err := cmd.Run() // will wait for command to return
	if err != nil {
		errstr := string(cmderror.Bytes())
		err = errors.New(errstr)
	}
	return err
}
