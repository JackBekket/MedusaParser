package main

import (
	"fmt"
	"log"

	"github.com/PuerkitoBio/goquery"
)

type News struct {
	Title       string
	Description string
	// Add more fields as per your requirements
}

func ParseMedusaNews() ([]News, error) {
	var news []News

	// Fetch the news site's page
	doc, err := goquery.NewDocument("https://meduza.io/")
	if err != nil {
		return nil, err
	}

	// Parse and extract news articles
	doc.Find(".news-section").Each(func(i int, s *goquery.Selection) {
		title := s.Find(".news-text .Title").Text()
		description := s.Find(".news-text .Description").Text()

		news = append(news, News{
			Title:       title,
			Description: description,
			// Set other field values if required
		})
	})

	return news, nil
}

func main() {
	news, err := ParseMedusaNews()
	if err != nil {
		log.Fatal(err)
	}

	// Process the collected news
	for _, n := range news {
		fmt.Println("Title:", n.Title)
		fmt.Println("Description:", n.Description)
		fmt.Println()
	}
}

