package golang

import (
	"archive/tar"
	"compress/gzip"
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

type entry struct {
	name     string
	checksum string
	url      string
	file     *os.File
}

func getEntries() ([]entry, error) {
	fmt.Println("Go: fetching versions")
	var entries []entry
	res, err := http.Get("https://go.dev/dl/")
	if err != nil {
		return entries, err
	}
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return entries, err
	}
	doc.Find(".downloadtable tr").Each(func(i int, s *goquery.Selection) {
		var entry entry
		entry.name = s.Find(".filename").Text()
		entry.url = fmt.Sprintf("https://go.dev%v", s.Find("a").AttrOr("href", ""))
		entry.checksum = s.Find("tt").Text()
		if entry.name != "" {
			entries = append(entries, entry)
		}
	})
	return entries, nil
}

func (e *entry) download() (*os.File, error) {
	fmt.Printf("Go: downloading %v\n", e.name)
	var err error
	file, err := os.CreateTemp("", e.name)
	if err != nil {
		return file, err
	}
	res, err := http.Get(e.url)
	if err != nil {
		return file, err
	}
	if _, err = io.Copy(e.file, res.Body); err != nil {
		return file, err
	}
	_, err = e.file.Seek(0, 0)
	return file, err
}
func (e *entry) valid(file *os.File) (bool, error) {
	fmt.Println("Go: validating checksum")
	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		log.Fatal(err)
	}
	_, err := e.file.Seek(0, 0)
	return fmt.Sprintf("%x", hash.Sum(nil)) == e.checksum, err
}
func extract(file *os.File) error {
	fmt.Println("Go: extracting")
	gz, err := gzip.NewReader(file)
	if err != nil {
		return err
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	tr := tar.NewReader(gz)
	prefix := fmt.Sprintf("%v/.tn/", home)
	os.RemoveAll(prefix + "go")
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
	return nil
}

func Install() error {
	entries, err := getEntries()
	if err != nil {
		return err
	}
	return install(entries)
}

func install(entries []entry) error {
	name := fmt.Sprintf("%v-%v", runtime.GOOS, runtime.GOARCH)
	for _, entry := range entries {
		if strings.Contains(entry.name, name) {
			file, err := entry.download()
			if err != nil {
				return err
			}
			defer file.Close()

			if valid, err := entry.valid(file); err != nil || !valid {
				if err != nil {
					return err
				}
				fmt.Println("Go: bad checksum")
			} else {
				return extract(file)
			}
			break
		}

	}
	return nil
}

func version() (string, error) {
	fmt.Println("Go: checking version")
	cmd := exec.Command("go", "version")
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.Split(string(out), " ")[2], err
}
func Update() error {
	ver, err := version()
	if err != nil {
		return err
	}
	entries, err := getEntries()
	if err != nil {
		return err
	}
	newVer := strings.Split(entries[0].name, ".src")[0]
	if newVer != ver {
		fmt.Printf("Go: version mismatch latest: %v, installed: %v\n", newVer, ver)
		install(entries)
	} else {
		fmt.Printf("Go: up to date: %v\n", ver)
	}
	return nil
}
