package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"
)

type Site struct {
	Problem string `json:"problem"`
	URL     string `json:"url"`
}

func main() {

	if len(os.Args) != 2 {
		fmt.Println("There must be a URL given")
		return
	}

	start := time.Now()
	sites := ParseSite(os.Args[1])
	fmt.Println(time.Since(start))

	defer WriteJSON(sites)

}

func WriteJSON(s []*Site) {
	j, _ := json.Marshal(s)
	// file, _ := os.("sites.json", j, os)
	file, _ := os.OpenFile("sites.json", os.O_CREATE|os.O_RDWR, os.ModePerm)
	file.Truncate(0)
	defer file.Close()

	file.Write(j)
}

func ParseSite(pageUrl string) []*Site {
	var parse func(link string)
	var store sync.Map
	var wg sync.WaitGroup
	// rateLimiter := make(chan int, 10)
	deadLinks := make([]*Site, 0)

	// Needed to prevent port exhaustion

	parse = func(url string) {

		if _, ok := store.Load(url); ok {
			return
		}
		store.Store(url, struct{}{})

		resp, err := http.Get(url)

		if err != nil {
			fmt.Println("Error: ", err)
			deadLinks = append(deadLinks, &Site{
				URL:     url,
				Problem: "Error",
			})
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode >= 400 {
			deadLinks = append(deadLinks, &Site{
				URL:     url,
				Problem: "Dead Site",
			})
			fmt.Println("Dead Link: ", resp.Request.URL.String())
			return
		}

		fmt.Println("Parsing Site: ", resp.Request.URL.String())

		body, _ := goquery.NewDocumentFromReader(resp.Body)
		// fmt.Println(body.Url.Path)
		body.Find("a").Each(func(i int, sel *goquery.Selection) {
			href, ok := sel.Attr("href")
			if !ok {
				return
			}

			if _, ok := store.Load(href); ok {
				return
			}

			if _, ok := store.Load(pageUrl + href[1:]); ok {
				return
			}

			if strings.Split(href, "")[0] == "/" { // This will only parse within the site itself, it will not leave the domain.
				wg.Add(1)
				go func() {
					defer wg.Done()
					parse(pageUrl + href[1:])
				}()
			}

			if strings.Contains(href, "http") {
				return
				// rateLimiter <- 1
				// wg.Add(1)
				// time.Sleep(1 * time.Second)
				// go func() {

				// 	defer fmt.Println("Closed")
				// 	defer wg.Done()
				// 	go parse(href)
				// 	<-rateLimiter

				// }()
			}

		})

	}
	parse(pageUrl)
	wg.Wait()
	return deadLinks
}
