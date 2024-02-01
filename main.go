package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/PuerkitoBio/goquery"
)

/*
type News struct {
	Title       string
	Description string
	// Add more fields as per your requirements
}
*/


type ArticleShort struct {
	Title       string
	Link        string
   }

   
func ParseMedusaImportantNews() ([]ArticleShort, error) {
	// Make an HTTP GET request to the Medusa news site
	resp, err := http.Get("https://meduza.io/live/2024/02/01/voyna")
	if err != nil {
	 return nil, err
	}
	defer resp.Body.Close()
   
	// Parse the HTML document
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
	 return nil, err
	}
   
	// Find the div block with the most important news articles
	div := doc.Find("div[data-testid=important-lead]")
   
	// Extract the news from the div block
	var news []ArticleShort
	div.Find("li").Each(func(i int, s *goquery.Selection) {
	 // Extract the title, description, and link of each article
	 //title := s.Find("a").Text()
	 title := s.Text()
	 link, _ := s.Find("a").Attr("href")
   
	 // Create a new article object with the extracted data
	 article := ArticleShort{
	  Title:       title,
	  Link:        link,
	 }
   
	 // Add the article to the articles slice
	 news = append(news, article)
	})
   
	return news, nil
   }



   // get all important news by date. date should be in format 2024/02/01 (yyyy/mm/dd)
   func ParseMedusaImportantNewsByDate(date string) ([]ArticleShort, error) {
	// Make an HTTP GET request to the Medusa news site
	resp, err := http.Get("https://meduza.io/live/"+date+"/voyna")
	if err != nil {
	 return nil, err
	}
	defer resp.Body.Close()
   
	// Parse the HTML document
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
	 return nil, err
	}
   
	// Find the div block with the most important news articles
	div := doc.Find("div[data-testid=important-lead]")
   
	// Extract the news from the div block
	var news []ArticleShort
	div.Find("li").Each(func(i int, s *goquery.Selection) {
	 // Extract the title, description, and link of each article
	 //title := s.Find("a").Text()
	 title := s.Text()
	 link, _ := s.Find("a").Attr("href")
   
	 // Create a new article object with the extracted data
	 article := ArticleShort{
	  Title:       title,
	  Link:        link,
	 }
   
	 // Add the article to the articles slice
	 news = append(news, article)
	})
   
	return news, nil
   }
   

func main() {
	news, err := ParseMedusaImportantNewsByDate("2024/02/01")
	if err != nil {
		log.Fatal(err)
	}

	// Process the collected news
	for _, n := range news {
		fmt.Println("Title:", n.Title)
		//fmt.Println("Description:", n.Description)
		fmt.Println("Link: ", n.Link)
	}
}

