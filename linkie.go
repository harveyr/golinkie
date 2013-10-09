package main

import (
	"fmt"
	"log"
	"regexp"
	"net/http"
	"flag"
	"net/url"
	"io/ioutil"
)

func fetchHtml(inputUrl string) string {
	parsedUrl, err := url.Parse(inputUrl)
	if err != nil {
		log.Fatal(err)
	}
	if !parsedUrl.IsAbs() {
		log.Fatal("Url must be absolute")
	}
	resp, err := http.Get(parsedUrl.String())
	defer resp.Body.Close()
	if err != nil {
		log.Fatalf("Failed to reach %s", parsedUrl)
	}
	if resp.StatusCode != 200 {
		log.Fatalf("Failed to reach %s (%s)", parsedUrl, resp.StatusCode)
	}
	body, err := ioutil.ReadAll(resp.Body)
	return string(body)
}

func findLinkUrls(inputHtml string, parentUrlRaw string) []string {
	parentUrl, _ := url.Parse(parentUrlRaw)

	regex := regexp.MustCompile(`src="(\S+)"`)
	submatches := regex.FindAllStringSubmatch(inputHtml, -1)
	
	linkUrls := make([]string, len(submatches))
	for i, v := range submatches {
		linkUrl, _ := url.Parse(v[1])
		if !linkUrl.IsAbs() {
			linkUrl = parentUrl.ResolveReference(linkUrl)
		}
		linkUrls[i] = linkUrl.String()
	}
	return linkUrls
}

func testLinkUrl(linkUrl string, c chan map[string]int) {
	resp, err := http.Get(linkUrl)
	result := make(map[string]int)
	if err != nil {
		result[linkUrl] = -1
	} else {
		result[linkUrl] = resp.StatusCode
	}
	c <- result
}

func testLinkUrls(linkUrls []string) (chan map[string]int, int) {
	c := make(chan map[string]int)
	for _, linkUrl := range linkUrls {
		go testLinkUrl(linkUrl, c)
	}
	return c, len(linkUrls)
}

func main() {
	var inputUrl string
	flag.StringVar(&inputUrl, "url", "", "Provide a valid url")
	flag.Parse()
	fmt.Println(inputUrl)

	responseHtml := fetchHtml(inputUrl)
	log.Printf("Fetched %d characters", len(responseHtml))

	linkUrls := findLinkUrls(responseHtml, inputUrl)
	c, count := testLinkUrls(linkUrls)
	for i := 0; i < count; i++ {
		result := <- c
		log.Print("result: ", result)
	}
}
