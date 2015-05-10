package xz

import (
	"bytes"
	"crypto/md5"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
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
	checksum := flag.String("checksum", "md5", "checksum strategy to use. Currently implemented: md5, sha256 ")

	flag.Parse()
	if *deflate {
		err := XzWriter(*fp, *keep)
		if err != nil {
			fmt.Printf("Err: %v", err)
		}
	}
	if *inflate {
		if *stdout {
			r, err := XzReader(*fp, *stdout)
			n, err := io.Copy(os.Stdout, r)
			if err != nil {
				fmt.Printf("copied %d bytes with err: %v", n, err)
			} else {
				fmt.Printf("copied %d bytes", n)
			}
		} else {
			_, err := XzReader(*fp, *stdout)
			if err != nil {
				fmt.Printf("Err: %v", err)
			}
		}

	}
	if *deflatecheck {
		_ = DeflateCheck(*fp, *checksum)
	}
}

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

// XzReader inflates the file (named file).
// if stdout is true the inflated file is returned  as io.ReadCloser
// else it's written to disk.
func XzReader(file string, stdout bool) (io.ReadCloser, error) {
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

// XzWriter deflates the file (named file) to disk
// if keep is  true the original file is kept on disk else
// is blindly removed
func XzWriter(file string, keep bool) error {
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
