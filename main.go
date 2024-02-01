package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

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
	Date		string
   }

type ArticleFull struct {
	Title       string
	Link        string
	Content		string
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
	  Date:		   date,
	 }
   
	 // Add the article to the articles slice
	 news = append(news, article)
	})
   
	return news, nil
   }



func ParseArticle(link string) ([]ArticleFull, error) {
		// Make an HTTP GET request to the Medusa news site
		//https://meduza.io/news/2023/03/12/rossiyskie-vlasti-potrebovali-ogranichit-v-roditelskih-pravah-ottsa-shkolnitsy-narisovavshey-antivoennyy-risunok-delo-rassmotryat-15-marta
		resp, err := http.Get(link)
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
		//div := doc.Find("h1[data-testid=simple-title]")
	   
		// Extract the news from the div block
		var news []ArticleFull
		doc.Find("h1[data-testid=simple-title]").Each(func(i int, s *goquery.Selection) {
		 // Extract the title, description, and link of each article
		 //title := s.Find("a").Text()
		 title := s.Text()
		 // link, _ := s.Find("a").Attr("href")
	   
		 // Create a new article object with the extracted data
		 article := ArticleFull{
		  Title:       title,
		  Link:        link,
		 }

		// Find the div block with additional content
        doc.Find("div.GeneralMaterial-module-body").Each(func(i int, s *goquery.Selection) {
            content := strings.TrimSpace(s.Text())
            article.Content = content
        })
	   
		 // Add the article to the articles slice
		 news = append(news, article)
		})

		
	   
		return news, nil
}
   

func main() {
	date := "2023/03/12/"
	news, err := ParseMedusaImportantNewsByDate(date)
	if err != nil {
		log.Fatal(err)
	}

	f_date, err := formatDate(date)
	if err != nil {
		log.Fatal(err)
	}

	directory,err := createDirectory(f_date)
	if err != nil {
		log.Fatal(err)
	}

	filename_n := directory + "/" + "news_list.json"
	storeNewsList(news,filename_n)

	// Process the collected news
	for _, n := range news {
		fmt.Println("Title:", n.Title)



		fmt.Println("Link: ", n.Link)
		if (n.Link != "") {
			articles,err := ParseArticle(n.Link)
			if err != nil {
				log.Fatal(err)
			}

			for _,a := range articles {
				fmt.Println("TitleFull: ", a.Title)
				fmt.Println("Full Content: ", a.Content)
				filename := directory + "/" + a.Title + ".json"
				storeArticle(a,filename)
			}
		}
	}
}




func formatDate(dateString string) (string,error) {
	//dateString := "2023/03/12"
	layout := "2006/01/02/"
   
	// Parse the input date string into a time.Time value
	date, err := time.Parse(layout, dateString)
	if err != nil {
	 fmt.Println("Error parsing date:", err)
	 return "",err
	}
   
	// Format the time.Time value to the desired output format
	formattedDate := date.Format("2006-01-02")
   
	fmt.Println("Formatted date:", formattedDate)
	return formattedDate,nil
}

func createDirectory(name string) (string,error) {
	err := os.MkdirAll(name, os.ModePerm)
    if err != nil {
        return "", err
    }
	return name,nil
}


func storeArticle(article ArticleFull, filename string) error {
    // Convert the ArticleFull struct to JSON
    data, err := json.Marshal(article)
    if err != nil {
        return err
    }

    // Create a new file
    file, err := os.Create(filename)
    if err != nil {
        return err
    }
    defer file.Close()

    // Write the JSON data to the file
    _, err = file.Write(data)
    if err != nil {
        return err
    }

    return nil
}

func storeNewsList(article []ArticleShort, filename string) error {
    // Convert the ArticleFull struct to JSON
    data, err := json.Marshal(article)
    if err != nil {
        return err
    }

    // Create a new file
    file, err := os.Create(filename)
    if err != nil {
        return err
    }
    defer file.Close()

    // Write the JSON data to the file
    _, err = file.Write(data)
    if err != nil {
        return err
    }

    return nil
}
