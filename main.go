package main

import (
	"io"
	"log"
	"os"
	"os/exec"
)

func main() {
	// path
	fp := "./testdata.xz"
	r := xzReader(fp)
	// copy to stdout
	n, err := io.Copy(os.Stdout, r)
	if err != nil {
		log.Printf("copied %d bytes with err: %v", n, err)
	} else {
		log.Printf("copied %d bytes", n)
	}
}

func xzReader(file string) io.ReadCloser {
	f, err := os.Open(file)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	rpipe, wpipe := io.Pipe()
	cmd := exec.Command("xz", "--decompress", "--stdout")
	cmd.Stdin = f
	cmd.Stdout = wpipe

	go func() {
		err := cmd.Run()
		wpipe.CloseWithError(err)
	}()

	return rpipe
}

func xzCompresser(file string) io.ReadCloser {
	f, err := os.Open(file)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	rpipe, wpipe := io.Pipe()

	cmd := exec.Command("xz", "--stdout")

	cmd.Stdin = f
	cmd.Stdout = wpipe

	go func() {
		err := cmd.Run()
		wpipe.CloseWithError(err)
	}()
	//TODO
	// return the new file
	// and delete the old one if no error
	return rpipe
}
