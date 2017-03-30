package download

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

// Run downloads the specified thread into output directory
func Run(thread, output string) {
	resp, err := http.Get(fmt.Sprintf("http://boards.4chan.org%v", thread))
	if err != nil {
		log.Fatal(err)
	}
	pageBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	page := string(pageBytes)
	reg := regexp.MustCompile("i\\.4cdn\\.org/([a-z])+/([0-9]+)\\.([a-z0-9])+")
	images := reg.FindAllString(page, -1)
	imageMap := make(map[string]bool)
	for _, img := range images {
		imageMap[img] = true
	}
	images = make([]string, len(imageMap))
	i := 0
	for img := range imageMap {
		images[i] = img
		i++
	}
	sort.Strings(images)
	err = os.MkdirAll(output, 0700)
	if err != nil {
		log.Fatal(err)
	}
	if len(images) == 0 {
		log.Print("no images found")
		return
	}
	prefixLen := strings.LastIndex(images[0], "/")
	ready := make(chan string, len(images))
	for _, img := range images {
		go func(img string) {
			if _, err := os.Stat(filepath.Join(output, img[prefixLen:])); err == nil {
				log.Printf("skipping %v", img)
				ready <- img
				return
			}
			log.Printf("downloading %v...", img)
			resp, err := http.Get(fmt.Sprintf("http://%v", img))
			if err != nil {
				log.Print("ERROR: ", err)
			}
			f, err := os.Create(filepath.Join(output, img[prefixLen:]))
			if err != nil {
				log.Print("ERROR: ", err)
			}
			_, err = io.Copy(f, resp.Body)
			if err != nil {
				log.Print("ERROR: ", err)
			}
			f.Close()
			log.Printf("downloaded %v", img)
			ready <- img
		}(img)
	}
	for i := 0; i < len(images); i++ {
		img := <-ready
		log.Printf("finished %v (%02v / %02v)", img, i+1, len(images))
	}
}
