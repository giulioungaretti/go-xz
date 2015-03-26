package main

import (
	"bytes"
	"crypto/sha256"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
)

func main() {
	fp := flag.String("file", "", "name of file")
	inflate := flag.Bool("inflate", false, "inflate archive")
	deflate := flag.Bool("deflate", false, "deflate archive")
	deflatecheck := flag.Bool("check", false, "deflaten anche check checksums")
	stdout := flag.Bool("stdout", false, "write to file or to stdout when inflating")
	keep := flag.Bool("keep", false, "keep original file after deflating")

	flag.Parse()
	if *deflate {
		err := xzWriter(*fp, *keep)
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
	if *deflatecheck {
		deflateCheck(*fp)
	}
}

func checksumFromPath(file string) [sha256.Size]byte {
	data, err := ioutil.ReadFile(file)
	if err != nil {
		log.Fatal(err)
	}
	checkSum := sha256.Sum256(data)
	return checkSum
}

func checksumFromArr(data []byte) [sha256.Size]byte {
	checkSum := sha256.Sum256(data)
	return checkSum
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

func xzWriter(file string, keep bool) error {
	// Create an *exec.Cmd
	var cmd *exec.Cmd
	if keep {
		cmd = exec.Command("xz", "--keep", file)
	} else {
		cmd = exec.Command("xz", file)
	}
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
	// checksum
	return err
}

func deflateCheck(file string) error {
	checksum := checksumFromPath(file)
	keep := true
	err := xzWriter(file, keep)
	if err != nil {
		log.Printf("Err: %v", err)
		return err
	} else {
		stdout := true
		xzfile := fmt.Sprintf("%v.xz", file)
		r, err := xzReader(xzfile, stdout)
		if err != nil {
			log.Printf("Err: %v", err)
			return err
		} else {
			data, err := ioutil.ReadAll(r)
			if err != nil {
				log.Printf("Err: %v", err)
				return err
				checksum2 := checksumFromArr(data)
				if checksum != checksum2 {
					err := errors.New("something went wrong sha256 don't match")
					return err
				} else {
					log.Printf("Removing old file")
					os.Remove(file)
				}
			}
		}
	}
	return nil
}
