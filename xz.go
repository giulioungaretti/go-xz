// xz is a wrapper around a xz executable that must be avaialbe at path
package xz

import (
	"bytes"
	"crypto/md5"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
)

// Checksum contains a Sha256 checksum as a byte array
// and a md5 check sum as string.
type Checksum struct {
	Sha256 [sha256.Size]byte
	Md5    string
}

// Base64md5 converts a md5 checksum as []byte to a base 16 encoded string
func Base64md5(data []byte) string {
	h := md5.New()
	h.Write(data)
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}

// ChecksumFromPath returns a struct with the checksum of the file at path using the strategy selected with strategy string
// currently implemented sha256 and md5
func ChecksumFromPath(file string, strategy string) Checksum {
	var localchecksum Checksum
	data, err := ioutil.ReadFile(file)
	if err != nil {
		panic(err)
	}
	switch strategy {
	default:
		panic("Not implemented")
	case "md5":
		localchecksum.Md5 = Base64md5(data)
	case "sha256":
		localchecksum.Sha256 = sha256.Sum256(data)
	}
	return localchecksum
}

// ChecksumFromArr returns a struct with the checksum of the byte array passed the strategy selected with strategy string
// currently implemented sha256 and md5
func ChecksumFromArr(data []byte, strategy string) Checksum {
	var localchecksum Checksum
	switch strategy {
	default:
		panic("Not implemented")
	case "md5":
		localchecksum.Md5 = Base64md5(data)
	case "sha256":
		localchecksum.Sha256 = sha256.Sum256(data)
	}
	return localchecksum

}

// Reader inflates the file (named file).
// if stdout is true the inflated file is returned  as io.ReadCloser
// else it's written to disk.
func Reader(file string, stdout bool) (io.ReadCloser, error) {
	f, err := os.Open(file)
	if err != nil {
		panic(err)
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

// Writer deflates the file (named file) to disk
// if keep is  true the original file is kept on disk else
// is blindly removed
func Writer(file string, keep bool) error {
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
	return err
}

// DeflateCheck deflates a file to file.xz, and deletes file if  all went good.
// integrity check is done internally by xz. see: http://www.freebsd.org/cgi/man.cgi?query=xz&sektion=1&manpath=FreeBSD+8.3-RELEASE
// If there is an error, its returned  and the old file is not deleted, BUT
// there is no guarantee that the deflated file has been  created.
func DeflateCheck(file string, strategy string) error {
	keep := true
	err := XzWriter(file, keep)
	if err != nil {
		fmt.Printf("Err: %v \n", err)
		return err
	} else {
		if err == nil {
			fmt.Printf("Removing old file \n")
			os.Remove(file)
			fmt.Printf("Removed: %v \n", file)
			return nil
		} else {
			return err
		}
	}
}
