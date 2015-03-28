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
	checksum := flag.String("checksum", "md5", "checksum strategy to use. Currently implemented: md5, sha256 ")

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
		_ = DeflateCheck(*fp, *checksum)
	}
}

type Checksum struct {
	Sha256 [sha256.Size]byte
	Md5    string
}

func Base64md5(data []byte) string {
	h := md5.New()
	h.Write(data)
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}

// checksumFromPath returns a struct with the checksum of the file at path using the strategy select with strategy string
// currently implemented sha256 and md5
func checksumFromPath(file string, strategy string) Checksum {
	var localchecksum Checksum
	data, err := ioutil.ReadFile(file)
	if err != nil {
		log.Fatal(err)
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

// ChecksumFromArr returns a struct with the checksum of the byte array passed the strategy select with strategy string
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

// DeflateCheck deflates a file to file.xz, and deletes file if the checksum of the
// original file is the same as that of the byte stream coming form inflating back file.xz.
// Use either stragegy md5 || sha256 to check for the correctness of the deflating process.
// If there is an error, its returned  and the old file is not deleted, BUT
// there is no guarantee that the deflated file ihas been  created.
func DeflateCheck(file string, strategy string) error {
	checksum := checksumFromPath(file, strategy)
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
			} else {
				checksum2 := ChecksumFromArr(data, strategy)
				var err error
				switch strategy {
				default:
					panic("Not implemented")
				case "md5":
					before := checksum.Md5
					after := checksum2.Md5
					if before != after {
						err = errors.New("something went wrong md5 don't match")
						return err
					}
				case "sha256":
					before := checksum.Sha256
					after := checksum2.Sha256
					if before != after {
						err = errors.New("something went wrong sha256 don't match")
						return err
					}
				}
			}
			if err == nil {
				fmt.Printf("Removing old file \n")
				os.Remove(file)
				fmt.Printf("xz removed: %v \n", file)
				return nil
			} else {
				return err
			}

		}

	}
}
