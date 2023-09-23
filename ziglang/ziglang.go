package ziglang

import (
	"archive/tar"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"

	"github.com/ulikunitz/xz"
)

type entry struct {
	version  string
	checksum string
	tarball  string
	file     *os.File
}

func (e *entry) download() error {
	fmt.Printf("Zig: downloading %v\n", e.version)
	var err error
	e.file, err = os.CreateTemp("", fmt.Sprintf("zig%v", e.version))
	if err != nil {
		return err
	}
	res, err := http.Get(e.tarball)
	if err != nil {
		return err
	}
	if _, err := io.Copy(e.file, res.Body); err != nil {
		return err
	}
	_, err = e.file.Seek(0, 0)
	return err

}

func (e *entry) valid() (bool, error) {
	fmt.Println("Zig: validating checksum")
	hash := sha256.New()
	if _, err := io.Copy(hash, e.file); err != nil {
		log.Fatal(err)
	}
	_, err := e.file.Seek(0, 0)
	return fmt.Sprintf("%x", hash.Sum(nil)) == e.checksum, err
}

func (e *entry) extract() error {
	fmt.Println("Zig: extracting")
	xz, err := xz.NewReader(e.file)
	if err != nil {
		return err
	}
	tr := tar.NewReader(xz)

	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	prefix := fmt.Sprintf("%v/.tn/", home)
	os.RemoveAll(prefix + "zig")
	for {
		header, err := tr.Next()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			} else {
				return err
			}
		}
		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.Mkdir(prefix+header.Name, os.FileMode(header.Mode)); err != nil {
				return err
			}
		case tar.TypeReg:
			{
				f, err := os.OpenFile(prefix+header.Name, os.O_CREATE|os.O_RDWR, os.FileMode(header.Mode))
				if err != nil {
					return err
				}
				defer f.Close()

				if _, err := io.Copy(f, tr); err != nil {
					return err
				}
			}

		default:
			return fmt.Errorf("unknown tar type flag: %v", header.Typeflag)
		}

	}
	entries, err := os.ReadDir(prefix)
	if err != nil {
		return err
	}
	for _, e := range entries {
		if e.IsDir() && strings.HasPrefix(e.Name(), "zig") {
			return os.Rename(prefix+e.Name(), prefix+"zig")

		}
	}
	return nil
}

func getEntry() (*entry, error) {
	fmt.Println("Zig: fetching master")
	res, err := http.Get("https://ziglang.org/download/index.json")
	if err != nil {
		return nil, err
	}
	buf, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	var dst map[string]any
	err = json.Unmarshal(buf, &dst)
	if err != nil {
		return nil, err
	}
	var entry entry
	entry.version = dst["master"].(map[string]any)["version"].(string)
	e := dst["master"].(map[string]any)["x86_64-linux"].(map[string]any)
	entry.checksum = e["shasum"].(string)
	entry.tarball = e["tarball"].(string)
	return &entry, nil
}

func version() (string, error) {
	fmt.Println("Zig: checking version")
	cmd := exec.Command("zig", "version")
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return string(out[:len(out)-1]), err
}

func Install() error {
	entry, err := getEntry()
	if err != nil {
		return err
	}
	err = entry.download()
	if err != nil {
		return err
	}
	valid, err := entry.valid()
	if err != nil {
		return err
	}
	if valid {
		err = entry.extract()
		if err != nil {
			return err
		}
	} else {
		fmt.Println("Zig: bad checksum")
	}
	return nil
}

func Update() error {
	version, err := version()
	if err != nil {
		return err
	}
	entry, err := getEntry()
	if err != nil {
		return err
	}
	if entry.version != version {
		fmt.Printf("Zig: version mismatch latest: %v, installed: %v\n", entry.version, version)
		err = entry.download()
		if err != nil {
			return err
		}
		valid, err := entry.valid()
		if err != nil {
			return err
		}
		if valid {
			err = entry.extract()
			if err != nil {
				return err
			}
		} else {
			fmt.Println("Zig: bad checksum")
		}
	} else {
		fmt.Printf("Zig: up to date: %v\n", version)
	}
	return nil
}
