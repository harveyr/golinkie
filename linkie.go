package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"regexp"
)

type TestResult struct {
	Url        string
	StatusCode int
}

var colors = map[string]string{
    "header": "\033[95m",
    "blue": "\033[94m",
    "green": "\033[92m",
    "yellow": "\033[93m",
    "red": "\033[91m",
    "bold": "\033[1m",
    "end": "\033[0m",
}

func printColor(s string, color string) {
	formatted := fmt.Sprintf("%s%s%s", colors[color], s, colors["end"])
	fmt.Println(formatted)
}

func fetchHtml(inputUrl string) string {
	parsedUrl, err := url.Parse(inputUrl)
	if err != nil {
		log.Fatalf("Failed to parse url: %s", inputUrl)
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
	if err != nil {
		log.Fatal("Failed to read response body")
	}
	return string(body)
}

func findLinkUrls(inputHtml string, parentUrlRaw string) []string {
	parentUrl, _ := url.Parse(parentUrlRaw)

	regex := regexp.MustCompile(`(?:src|href)="(\S+)"`)
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

func testLinkUrl(linkUrl string, c chan TestResult) {
	resp, err := http.Get(linkUrl)
	var code int
	if err != nil {
		code = -111
	} else {
		code = resp.StatusCode
	}
	result := TestResult{linkUrl, code}
	c <- result
}

func testLinkUrls(linkUrls []string) (chan TestResult) {
	c := make(chan TestResult)
	for _, linkUrl := range linkUrls {
		go testLinkUrl(linkUrl, c)
	}
	return c
}

func main() {
	printColor("--- Linkie Time! ---", "header")
	var inputUrl string
	flag.StringVar(&inputUrl, "url", "", "Provide a valid url")
	flag.Parse()

	responseHtml := fetchHtml(inputUrl)
	printColor(inputUrl, "bold")

	linkUrls := findLinkUrls(responseHtml, inputUrl)
	c := testLinkUrls(linkUrls)
	for i := 0; i < len(linkUrls); i++ {
		result := <-c
		color := "green"
		if result.StatusCode != 200 {
			color = "red"
		}
		printColor(fmt.Sprintf("[%d] %s", result.StatusCode, result.Url), color)
	}
}
