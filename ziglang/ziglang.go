package ziglang

import (
	"archive/tar"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
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
	fmt.Printf("Zig: downlaoding %v\n", e.version)
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
	fmt.Println("Zig: getting master")
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

func Install() error {
	entry, err := getEntry()
	if err != nil {
		return err
	}
	err = entry.download()
	if err != nil {
		return err
	}
	err = entry.extract()
	if err != nil {
		return err
	}
	return nil
}
func Update() error {
	fmt.Println("zig update")
	return nil
}
