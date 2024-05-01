package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
	"time"
	"unicode"

	"github.com/go-rod/rod"
)

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

func main() {
	start_date := "2022/02/24"
	//ParseAllByDate(start_date)
	//ParseOlderHTML(start_date)
	FastForward(start_date)

}

func FastForward(start_date string) {
	startDateStr := start_date
	startDate, err := time.Parse("2006/01/02", startDateStr)
	if err != nil {
		fmt.Println("Invalid start date format:", err)
		return
	}
	currentDate := time.Now()
	for date := startDate; date.Before(currentDate); date = date.AddDate(0, 0, 1) {
		dateStr := date.Format("2006/01/02")

		date_str := dateStr + "/"
		f_date, err := formatDate(date_str)
		if err != nil {
			log.Fatal(err)
		}

		data_dir, err := createDirectory("medusa_dump")
		if err != nil {
			log.Fatal(err)
		}

		directory, err := createDirectory(data_dir + "/" + f_date)
		if err != nil {
			if os.IsExist(err) {
				log.Println("skipping date :", directory)
				continue // skip this iteration if the directory already exists
			} else {
				log.Fatal(err)
			}
		}

		ParseAllByDate(dateStr)
	}
}

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

		/* filename_n := directory + "/" + "news_list.txt"
		storeNewsList(news, filename_n) */

		hasNonEmptyLink := false
		for _, n := range news {
			if n.Link != "" {
				hasNonEmptyLink = true
				break
			}
		}

		if hasNonEmptyLink {
			for _, n := range news {

				if n.Link != "" {
					articles, err := ParseArticle(n.Link)
					if err != nil {
						fmt.Println(err) // or log.Println(err)
						continue         // move to the next element
					}
					for _, a := range articles {
						filename := directory + "/" + getFirstSentence(a.Title) + ".txt"
						storeArticle(a, filename)
					}
				}
			}
		}
		date = strings.TrimSuffix(date, "/")
		articles := GetArticlesFromOlderHTML(date)
		for i, a := range articles {
			if i == 0 {
				continue
			}
			a.Link = "https://meduza.io/live/" + date + "/voyna"
			filename := directory + "/" + getFirstSentence(a.Title) + ".txt"
			storeArticle(a, filename)
		}
	}
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
		return nil, errors.New("Page not found...")
	}
	// Wait for a few seconds to ensure content is loaded
	hasIt, _, _ := page.Has("[data-testid=important-lead]")
	if !hasIt {
		return nil, errors.New("Leads not found...")
	}
	// Find the div block with the most important news articles
	div := page.MustElement("[data-testid=important-lead]")
	// Extract the news from the div block
	var news []ArticleShort
	hasIt, _, _ = page.Has("li")
	if !hasIt {
		return nil, errors.New("Elements not found...")
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
		return nil, errors.New("Title not found, skipping article...")
	}
	el := page.MustElement("h1[data-testid=simple-title]")

	title := el.MustText()

	// Find the div block with additional content
	hasIt, _, _ = page.Has("div.GeneralMaterial-module-body")
	if !hasIt {
		return nil, errors.New("Content not found, skipping article...")
	}
	el = page.MustElement("div.GeneralMaterial-module-body")

	// Find all paragraphs within the content block
	hasIt, _, _ = page.Has("p")
	if !hasIt {
		return nil, errors.New("Paragraphs not found, skipping article...")
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

	return []ArticleFull{article}, nil

}

func storeArticle(article ArticleFull, filename string) error {

	filename = getFirstSentence(filename)
	data := article.Title + "\n\n" + article.Content
	if containsMoreThanTwoParagraphs(data) {
		file, err := os.Create(filename)
		if err != nil {
			return err
		}
		defer file.Close()
		_, err = file.WriteString(data)
		if err != nil {
			return err
		}
		return nil
	}
	return nil
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
	if _, err := os.Stat(name); os.IsNotExist(err) {

		err := os.MkdirAll(name, os.ModePerm)
		if err != nil {
			return "", err
		}
	} else if err != nil {
		return "", err
	}
	return name, nil
}

func storeNewsList(article []ArticleShort, filename string) error {
	file, _ := os.Create(filename)
	defer file.Close()

	for _, article := range article {
		output := fmt.Sprintf("Title: %s\nContent: %s\n\n", article.Date, article.Title)
		if _, err := file.WriteString(output); err != nil {
			return err
		}
	}

	return nil
}

func ParseOlderHTML(date string) ([]ArticleFull, error) {
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

	hasIt, _, _ := page.Has(".Slide-module-isInLive")
	if !hasIt {
		return nil, errors.New("Elements not found, skipping this article...")
	}
	elements := page.MustElements(".Slide-module-isInLive")

	var articles []ArticleFull

	// Iterate through each element
	for _, element := range elements {
		var title string
		hasIt, _, _ := element.Has("h4")
		if hasIt {
			titleElement := element.MustElement("h4")

			// Get the text of the h4 element (which is the title)
			title, _ = titleElement.Text()

		}

		var contentElements rod.Elements

		hasIt, _, _ = page.Has("p")
		if hasIt {
			contentElements = element.MustElements("p")
		}

		hasLink, _, _ := element.Has("a")
		if hasLink {

			aElement := element.MustElement("a")
			link, _ := aElement.Attribute("href")
			if link != nil && strings.Contains(*link, "/feature") && !strings.Contains(*link, "meduza.io") {

				linkFull := "https://meduza.io" + *link
				article, err := ParseFeature(linkFull)
				if err == nil {
					articles = append(articles, article)
				} else {
					fmt.Println(err)
				}
			}
		}

		// Concatenate text from all p elements to form the content
		var content string
		for i, contentElement := range contentElements {

			text, err := contentElement.Text()
			if title == "" && i == 0 {
				title = text
			}
			if err != nil {
				// Handle error if unable to get text
				return nil, err
			}
			content += text + "\n" // Add newline between paragraphs

		}

		// Append the extracted article to the articles slice
		articles = append(articles, ArticleFull{
			Title:   title,
			Content: content,
		})
	}

	return articles, nil
}

func PrintArticles(articles []ArticleFull) {
	for _, article := range articles {
		fmt.Println("Title: " + article.Title)
		fmt.Println("Content: " + article.Content)
	}
}

func removeDuplicates(articles []ArticleFull) []ArticleFull {
	// Map to store unique titles or contents
	uniqueMap := make(map[string]bool)
	var uniqueArticles []ArticleFull

	for _, article := range articles {
		// Check if both title and content are non-empty
		if article.Title != "" || article.Content != "" {
			// Generate a unique key based on non-empty fields
			key := article.Title + "||" + article.Content
			// Check if the key exists in the map
			if !uniqueMap[key] {
				// If it doesn't exist, add it to the map and append to the uniqueArticles slice
				uniqueMap[key] = true
				uniqueArticles = append(uniqueArticles, article)
			}
		}
	}

	return uniqueArticles
}

func GetArticlesFromOlderHTML(date string) []ArticleFull {
	art, _ := ParseOlderHTML(date)
	art = removeDuplicates(art)
	return art
}

func ParseFeature(link string) (ArticleFull, error) {
	var article ArticleFull
	browser := rod.New().Context(context.Background())
	if !strings.Contains(link, "meduza.io/feature") {
		return article, errors.New("Bad link, skipping that article...")
	}
	page := browser.MustConnect().MustPage(link).MustWaitLoad()
	defer browser.MustClose()
	hasIt, _, _ := page.Has("data-testid=rich-title")
	if hasIt {
		el := page.MustElement("data-testid=rich-title")
		article.Title = el.MustText()
	}

	// Find the div block with additional content
	hasIt, _, _ = page.Has("div.GeneralMaterial-module-article")
	if !hasIt {
		return article, errors.New("Can't recognise the article, skipping...")
	}
	el := page.MustElement("div.GeneralMaterial-module-article")

	// Find all paragraphs within the content block
	hasIt, _, _ = page.Has("p")
	if !hasIt {
		return article, errors.New("Paragraphs in article not found, skipping...")
	}
	paragraphs := el.MustElements("p")

	// Concatenate text content of all paragraphs
	var content strings.Builder
	for _, p := range paragraphs {
		content.WriteString(p.MustText() + "\n")
	}

	if article.Title == "" {
		article.Title = paragraphs[0].MustText()
	}

	contentText := strings.TrimSpace(content.String())

	article.Link = link
	article.Content = contentText

	return article, nil

}

func getFirstSentence(text string) string {
	// Split the text by ".", "!", "?" to get individual sentences
	sentences := strings.FieldsFunc(text, func(r rune) bool {
		return r == '.' || r == '!' || r == '?'
	})
	if len(sentences) > 0 {
		words := strings.Fields(sentences[0])
		if len(words) > 6 {
			return strings.Join(words[:6], " ")
		}
		return sentences[0]
	}
	return text
}

func containsMoreThanTwoParagraphs(article string) bool {
	paragraphs := strings.Split(article, "\n\n")
	if len(paragraphs) < 1 {
		return false
	}
	secondParagraphWords := strings.FieldsFunc(paragraphs[1], func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsNumber(r)
	})
	return len(secondParagraphWords) > 15
}
