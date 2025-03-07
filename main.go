package main

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"
)

const (
	MAIN = "https://scrape-me.dreamsofcode.io/"
)

func main() {

	if len(os.Args) != 2 {
		fmt.Println("There must be a URL given")
		return
	}

	start := time.Now()
	ParseSite(os.Args[1])
	fmt.Println(time.Since(start))
}

func ParseSite(pageUrl string) []string {
	var parse func(link string)
	var store sync.Map
	var wg sync.WaitGroup
	deadLinks := make([]string, 5)

	// Needed to prevent port exhaustion
	t := &http.Transport{
		IdleConnTimeout: 30 * time.Second,
	}

	client := &http.Client{
		Timeout:   5 * time.Second,
		Transport: t,
	}

	parse = func(url string) {

		resp, err := client.Get(url)

		if err != nil {
			fmt.Println("Error: ", err)
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode >= 400 {
			deadLinks = append(deadLinks, url)
			fmt.Println("Dead Link: ", resp.Request.URL.String())
			return
		}
		if _, ok := store.Load(url); ok {
			return
		}

		fmt.Println("Parsing Site: ", resp.Request.URL.String())

		store.Store(url, struct{}{})

		body, _ := goquery.NewDocumentFromReader(resp.Body)

		body.Find("a").Each(func(i int, sel *goquery.Selection) {
			href, ok := sel.Attr("href")
			if ok {
				if strings.Contains(href, "http") {
					wg.Add(1)
					go func() {
						defer wg.Done()
						parse(href)
					}()
				}

				// if splitUrl[0] == "/" { // This will only parse within the site itself, it will not leave the domain.
				// 	wg.Add(1)
				// 	go func() {
				// 		defer wg.Done()
				// 		parse(href)
				// 	}()
				// }

			}
		})

	}
	parse(pageUrl)
	wg.Wait()
	return deadLinks
}
