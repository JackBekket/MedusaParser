package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/go-rod/rod"
)

/*
type News struct {
	Title       string
	Description string
	// Add more fields as per your requirements
}
*/

type ArticleShort struct {
	Title string
	Link  string
	Date  string
}

type ArticleFull struct {
	Title   string
	Link    string
	Content string
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
			Title: title,
			Link:  link,
		}

		// Add the article to the articles slice
		news = append(news, article)
	})

	return news, nil
}

func ParseMedusaImportantNewsByDate(date string) ([]ArticleShort, error) {
	// Create a rod browser instance
	browser := rod.New().Timeout(1 * time.Minute).MustConnect()

	// Close the browser when done
	defer browser.MustClose()

	// Create a new page and navigate to the Medusa news site
	page := browser.MustPage("https://meduza.io/live/" + date + "/voyna").MustWaitLoad()
	time.Sleep(2 * time.Second)
	// Check if the page is a 404 error
	status := page.MustInfo().Title
	if status == "404 — Meduza" {
		return nil, errors.New("Page not found")
	}
	// Wait for a few seconds to ensure content is loaded
	hasIt, _, _ := page.Has("[data-testid=important-lead]")
	if !hasIt {
		return nil, errors.New("Page not found")
	}
	// Find the div block with the most important news articles
	div := page.MustElement("[data-testid=important-lead]")
	// Extract the news from the div block
	var news []ArticleShort
	hasIt, _, _ = page.Has("li")
	if !hasIt {
		return nil, errors.New("Page not found")
	}
	articles := div.MustElements("li")
	for _, article := range articles {
		// Extract the title of each article
		title := article.MustText()

		// Try to find the link element within the article
		linkElements := article.MustElements("a")
		var link string
		if len(linkElements) > 0 {
			// If at least one link element is found, extract the href property of the first one
			link = linkElements.First().MustProperty("href").String()
		} else {
			// If no link element is found, use a placeholder
			link = ""
		}

		// Create a new article object with the extracted data
		article := ArticleShort{
			Title: title,
			Link:  link,
			Date:  date,
		}

		// Add the article to the articles slice
		news = append(news, article)
	}

	return news, nil
}

func main() {
	start_date := "2022/02/24"
	//ParseMedusaImportantNewsByDate(start_date)
	FastForward(start_date)
	//ParseAllByDate(start_date)
}

func ParseArticle(link string) ([]ArticleFull, error) {
	if !strings.Contains(link, "meduza.io") {
		return nil, errors.New("Bad link")
	}
	browser := rod.New().Context(context.Background())

	// Connect to the browser instance
	page := browser.MustConnect().MustPage(link).MustWaitLoad()

	// Close the browser after scraping
	defer browser.MustClose()
	// Use Rod's HTML parser to manipulate the DOM
	hasIt, _, _ := page.Has("h1[data-testid=simple-title]")
	if !hasIt {
		return nil, errors.New("Page not found")
	}
	el := page.MustElement("h1[data-testid=simple-title]")

	title := el.MustText()

	// Find the div block with additional content
	hasIt, _, _ = page.Has("div.GeneralMaterial-module-body")
	if !hasIt {
		return nil, errors.New("Page not found")
	}
	el = page.MustElement("div.GeneralMaterial-module-body")

	// Find all paragraphs within the content block
	hasIt, _, _ = page.Has("p")
	if !hasIt {
		return nil, errors.New("Page not found")
	}
	paragraphs := el.MustElements("p")

	// Concatenate text content of all paragraphs
	var content strings.Builder
	for _, p := range paragraphs {
		content.WriteString(p.MustText() + "\n")
	}
	// Get the final content as a string
	contentText := strings.TrimSpace(content.String())

	// Create an ArticleFull object with the extracted data
	article := ArticleFull{
		Title:   title,
		Link:    link,
		Content: contentText,
	}
	// Return the article in a slice
	return []ArticleFull{article}, nil

}

// Get all news list and articles parsed and saved by date
func ParseAllByDate(date string) {

	log.Println("Parse all news from medusa by date: ", date)
	news, err := ParseMedusaImportantNewsByDate(date)
	if err != nil {
		fmt.Println("Page not found, skipping...")

	} else {
		date = date + "/"
		f_date, err := formatDate(date)
		if err != nil {
			log.Fatal(err)
		}

		data_dir, err := createDirectory("medusa_dump")
		if err != nil {
			log.Fatal(err)
		}

		directory, err := createDirectory(data_dir + "/" + f_date)
		if err != nil {
			log.Fatal(err)
		}

		filename_n := directory + "/" + "news_list.txt"
		storeNewsList(news, filename_n)

		// Process the collected news
		for _, n := range news {

			if n.Link != "" {
				articles, err := ParseArticle(n.Link)
				if err != nil {
					fmt.Println(err) // or log.Println(err)
					continue         // move to the next element
				}
				for _, a := range articles {
					filename := directory + "/" + a.Title + ".txt"
					storeArticle(a, filename)
				}
			}
		}
	}
}

func formatDate(dateString string) (string, error) {
	//dateString := "2023/03/12"
	layout := "2006/01/02/"

	// Parse the input date string into a time.Time value
	date, err := time.Parse(layout, dateString)
	if err != nil {
		fmt.Println("Error parsing date:", err)
		return "", err
	}

	// Format the time.Time value to the desired output format
	formattedDate := date.Format("2006-01-02")
	return formattedDate, nil
}

func createDirectory(name string) (string, error) {
	err := os.MkdirAll(name, os.ModePerm)
	if err != nil {
		return "", err
	}
	return name, nil
}

func storeArticle(article ArticleFull, filename string) error {
	// Convert the ArticleFull struct to JSON

	data := article.Title + "\n\n" + article.Content

	//title := ArticleFull.Title

	//data := ArticleFull.Title + ArticleFull.Content

	// Create a new file
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	// Write data to the file
	_, err = file.WriteString(data)
	if err != nil {
		return err
	}

	return nil
}

func storeNewsList(article []ArticleShort, filename string) error {
	// Convert the ArticleFull struct to JSON
	/*
	   data, err := json.Marshal(article)
	   if err != nil {
	       return err
	   }
	*/

	// Create a new file
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	// Iterate over each ArticleFull in the slice
	for _, article := range article {
		// Prepare the string to write, including the title and content.
		// Feel free to adjust the formatting to meet your needs
		output := fmt.Sprintf("Title: %s\nContent: %s\n\n", article.Date, article.Title)

		// Write the formatted string to the file
		if _, err := file.WriteString(output); err != nil {
			// If an error occurs, return the error
			return err
		}
	}

	return nil
}

// get and saves all news from start date till now
func FastForward(start_date string) {
	// Set the starting date
	// startDateStr := "2022/02/24/"

	startDateStr := start_date
	startDate, err := time.Parse("2006/01/02", startDateStr)
	if err != nil {
		fmt.Println("Invalid start date format:", err)
		return
	}

	// Get the current date
	currentDate := time.Now()

	// Iterate over the range of dates
	for date := startDate; date.Before(currentDate); date = date.AddDate(0, 0, 1) {
		// Format the date to match the expected format for the Medusa site
		dateStr := date.Format("2006/01/02")

		// Your parsing logic goes here
		// Call the function or perform the action to parse the news for the given date
		ParseAllByDate(dateStr)
	}
}
