package main

import (
    "fmt"
	"os"
	"io"
	"log"
	"bufio"
	"strings"
	"sync"
	"net/http"
	// "io/ioutil"

	"github.com/PuerkitoBio/goquery"
)

var picUrlKey = "var mhurl="
var picHost = "http://p1.xiaoshidi.net/"
var baseFolder = "C:/Users/i311688/Desktop/MyTemp/manga/hzw/"

func main() {
	if len(os.Args) < 2 {
		log.Fatal("file path of list required")
	}
	path := os.Args[1]
	fmt.Printf("file path is: %s\n", path)

	file, err := os.Open(path)
    if err != nil {
        log.Fatal(err)
    }
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var wg sync.WaitGroup
	var urls []string
    for scanner.Scan() {
		url := scanner.Text()
		url = strings.Trim(url, " ")
		if url != "" {
			urls = append(urls, url)
		}
	}
	wg.Add(len(urls))
	for _, url := range urls{
		go download(url, &wg)
	}
	wg.Wait()
}

func download(url string, wg *sync.WaitGroup) {
	defer wg.Done()	
	initFolder := false
	folder := ""
	picIdx := 0
	title := ""
	baseUrl := url
	for {
		hasNext := false
		fmt.Printf("download %s\n", url)
		resp, err := http.Get(url)
		if err != nil {
			fmt.Printf("%v\n", err)
			return
		}
		
		doc, err := goquery.NewDocumentFromReader(resp.Body)
		if err != nil {
			fmt.Printf("download %s failed: %v\n", url, err)
			resp.Body.Close()
			return
		}

		if initFolder == false {
			doc.Find("title").Each(func(i int, s *goquery.Selection) {
				title = strings.Trim(s.Text(), " ")
				folder = baseFolder + title + "/"
				os.MkdirAll(folder, 0777)
				initFolder = true
			})
		}

		// find the pic url
		keyLen := len(picUrlKey)
		doc.Find("body script").Each(func(i int, s *goquery.Selection) {
			text := s.Text()
			idx := strings.Index(text, picUrlKey)
			if idx >= 0 {
				text = text[idx:]
				idx2 := strings.Index(text, ";")
				// fmt.Printf("script is: %d, %d, %s\n", keyLen, idx2, text)
				picUrl := text[keyLen + 1 : idx2 - 1]
				downloadPic(fmt.Sprintf("%s%s", picHost, picUrl), folder, picIdx)
			}
		})
		
		doc.Find("a.pure-button.pure-button-primary").Each(func(i int, s *goquery.Selection) {
			if s.Text() == "下一页" {
				href, _ := s.Attr("href")
				url = baseUrl + href
				// fmt.Printf("next is %s\n", url)
				hasNext = true
			}
		})
		resp.Body.Close()
		
		picIdx++
		if hasNext == false {
			fmt.Printf("%s downloading finished: %d\n", title, picIdx)
			break
		}
	}
}

func downloadPic (picUrl string, folder string, picIdx int) {
	fmt.Printf("download %s\n", picUrl)
	filePath := folder + fmt.Sprintf("%04d", picIdx) + ".jpg"
	if _, e := os.Stat(filePath); os.IsNotExist(e) == false {
		fmt.Printf("%s is already downloaded\n", picUrl)
		return
	}

	resp, err := http.Get(picUrl)
	if err != nil {
		fmt.Printf("download pic %s failed: %v\n", picUrl, err)
		return
	}
	defer resp.Body.Close()
	img, _ := os.Create(filePath)
    defer img.Close()
	_, err = io.Copy(img, resp.Body)
	if err == nil {
		fmt.Println("done")
	} else {
		fmt.Println("error")
	}
}