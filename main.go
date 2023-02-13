package main

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/gocolly/colly"
)

const site = "https://***********.ru/"

func main() {
	fName := "data.csv"
	file, err := os.Create(fName)
	if err != nil {
		log.Fatalf("Could not create file, err: %q", err)
		return
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	writer.Write([]string{
		"section",
		"subsection",
		"article",
		"name",
		"price",
	})

	//

	c := colly.NewCollector(colly.UserAgent("yz"), colly.AllowURLRevisit())
	d, dd, ddd := c.Clone(), c.Clone(), c.Clone()
	onError(c, d, dd, ddd)

	c.OnHTML("div.container", func(e *colly.HTMLElement) {
		e.ForEachWithBreak("div.product-body", func(i int, h *colly.HTMLElement) bool {
			// if i == 0 {
			section := h.ChildText("h5.product-title")
			sectionUrl := e.Request.AbsoluteURL(h.ChildAttr("a", "href"))
			fmt.Printf("\n----------\n\n")
			fmt.Println(i, section)
			fmt.Println(i, sectionUrl)

			d.OnRequest(func(r *colly.Request) {
				r.Ctx.Put("section", section)
			})
			d.Visit(sectionUrl)

			return true
			// }

			// return false
		})
	})

	d.OnHTML(".section>.container", func(h *colly.HTMLElement) {
		println()
		section := h.Request.Ctx.Get("section")
		fmt.Printf("section = %s\n\n", section)

		h.ForEachWithBreak(".product-modern-body", func(i int, e *colly.HTMLElement) bool {
			// if i == 0 {
			subsection := e.ChildText(".product-modern-title")
			subsectionUrl := h.Request.AbsoluteURL(e.ChildAttr("a", "href"))
			fmt.Println("subsection =", subsection)
			fmt.Println(subsectionUrl)

			dd.OnRequest(func(r *colly.Request) {
				r.Ctx.Put("section", section)
				r.Ctx.Put("subsection", subsection)
			})
			dd.Visit(subsectionUrl)

			return true
			// }

			// return false
		})
	})

	dd.OnHTML(".table-custom-responsive+.pagination-wrap ul.pagination", func(h *colly.HTMLElement) {
		println()
		section := h.Request.Ctx.Get("section")
		subsection := h.Request.Ctx.Get("subsection")
		fmt.Printf("section = %s\n", section)
		fmt.Printf("subsection = %s\n", subsection)

		pagesStr := h.ChildText("li:nth-last-of-type(2)")
		pages, err := strconv.Atoi(pagesStr)

		dddVisit := func(ddd *colly.Collector, url string) {
			ddd.OnRequest(func(r *colly.Request) {
				r.Ctx.Put("section", section)
				r.Ctx.Put("subsection", subsection)
			})
			ddd.Visit(url)
		}

		if err != nil {
			fmt.Println("no pages")
			dddVisit(ddd, h.Request.URL.String())
		} else {
			for i := 1; i <= pages; i++ {
				fmt.Println("page", i)
				url := h.Request.URL.String() + "?page=" + strconv.Itoa(i)
				fmt.Println(url)
				dddVisit(ddd, url)

				// break
			}
		}
	})

	ddd.OnHTML("html", func(h *colly.HTMLElement) {
		section := h.Request.Ctx.Get("section")
		subsection := h.Request.Ctx.Get("subsection")

		h.DOM.Find("tbody").First().Find("tr").Each(func(i int, s *goquery.Selection) {
			name := s.Find("td:nth-of-type(1)").Text()
			splits := strings.SplitN(name, " ", 2)
			article, name := splits[0], splits[1]
			price := s.Find("td:nth-of-type(2)").Text()
			price = strings.ReplaceAll(price, "у.е.", "")

			fmt.Println("item =", article, name, "; price =", price)

			writer.Write([]string{
				section,
				subsection,
				article,
				name,
				price,
			})
		})
	})

	c.Visit(site)
	println()
}

func onError(collectors ...*colly.Collector) {
	for _, c := range collectors {
		c.OnError(func(r *colly.Response, err error) {
			fmt.Println("Request URL:", r.Request.URL, "failed with response:", r, "\nError:", err)
		})
	}
}
