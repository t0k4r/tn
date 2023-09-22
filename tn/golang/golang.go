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

func getEntries() []entry {
	fmt.Println("Go: fetching go versions")
	var entries []entry
	res, err := http.Get("https://go.dev/dl/")
	if err != nil {
		log.Fatal(err)
	}
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		log.Fatal(err)
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
	return entries
}

func (e *entry) download() error {
	fmt.Printf("Go: downloading %v\n", e.name)
	var err error
	e.file, err = os.Create(fmt.Sprintf("/tmp/%v", e.name))
	if err != nil {
		return err
	}
	res, err := http.Get(e.url)
	if err != nil {
		return err
	}
	_, err = io.Copy(e.file, res.Body)
	if err != nil {
		return err
	}
	_, err = e.file.Seek(0, 0)
	return err
}
func (e *entry) valid() (bool, error) {
	fmt.Println("Go: validating go checksum")
	hash := sha256.New()
	_, err := io.Copy(hash, e.file)
	if err != nil {
		log.Fatal(err)
	}
	sum := fmt.Sprintf("%x", hash.Sum(nil))
	_, err = e.file.Seek(0, 0)
	return sum == e.checksum, err
}
func (e *entry) extract() error {
	fmt.Println("Go: extracting")
	gz, err := gzip.NewReader(e.file)
	if err != nil {
		return err
	}
	tr := tar.NewReader(gz)
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
			err := os.Mkdir(header.Name, 0775)
			if err != nil {
				return err
			}
		case tar.TypeReg:
			{
				file, err := os.Create(header.Name)
				if err != nil {
					return err
				}
				defer file.Close()

				_, err = io.Copy(file, tr)
				if err != nil {
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
	name := fmt.Sprintf("%v-%v", runtime.GOOS, runtime.GOARCH)
	entries := getEntries()
	for _, entry := range entries {
		if strings.Contains(entry.name, name) {
			err := entry.download()
			if err != nil {
				return err
			}
			valid, err := entry.valid()
			if err != nil {
				return err
			}
			if valid {
				err := entry.extract()
				if err != nil {
					return err
				}
			}
			return nil
		}
	}
	return nil
}
func Update() error {
	fmt.Println("go update")
	return nil
}
