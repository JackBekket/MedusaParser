package main

import (
	"encoding/csv"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	//"github.com/tmc/langchaingo/embeddings"

	localai "github.com/JackBekket/uncensoredgpt_tgbot/lib/embeddings"
	"github.com/tmc/langchaingo/schema"
)

type FileData struct {
	Content string
	Date    string
	Title   string
}

type BioFiles struct {
	Username string
	Content  string
}

func GetDocsShemaByFiles(fileData []FileData) []schema.Document {
	var docs []schema.Document

	for _, data := range fileData {
		doc := schema.Document{
			PageContent: data.Content,
			Metadata: map[string]interface{}{
				"date": data.Date,
			},
		}
		docs = append(docs, doc)
	}

	return docs
}

func parseFiles(path string) ([]FileData, error) {
	var files []FileData

	err := filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() && strings.HasSuffix(info.Name(), ".txt") {
			content, err := ioutil.ReadFile(path)
			if err != nil {
				return err
			}

			directory, filename := filepath.Split(path)

			// Remove "../medusa_dump/" from the directory
			directory = strings.TrimPrefix(directory, "../medusa_dump/")
			// Remove the trailing slash from the directory
			directory = strings.TrimSuffix(directory, "/")

			// Remove ".txt" from the filename
			filename = strings.TrimSuffix(filename, ".txt")

			fileData := FileData{
				Date:    directory,
				Title:   filename,
				Content: string(content),
			}

			fmt.Println("Date:", directory)
			fmt.Println("Title:", filename)
			fmt.Println("Content:", string(content))

			files = append(files, fileData)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return files, nil
}

func CallSemanticSearch(promt string, max_results int) {
	docs, err := localai.SemanticSearch(promt, max_results)
	if err != nil {
		log.Println(err)
	}
	log.Println("Semantic Search results:", docs)
}

func CallRagSearch(promt string, max_results int) {
	result, err := localai.RagSearch(promt, max_results)
	if err != nil {
		log.Println(err)
	}
	log.Println("RAG result:", result)
}

func main() {
	//files := getFiles("../medusa_dump_Test")
	/* if err != nil {
		fmt.Println(err)
	} */

	//docs := GetDocsShemaByFiles(files)
	docs := GetDocsSchemaFromCSV("../mdbio1000.csv")
	localai.LoadDocsToStore(docs)
	//CallSemanticSearch("Python", 5)
	CallRagSearch("What can you tell me about Janith-Sandamal?", 25)
	CallRagSearch("What can you tell me about p-miralles?", 25)

}

func getFiles(dir string) []FileData {
	var filesData []FileData
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() {

			content, err := ioutil.ReadFile(path)
			if err != nil {
				return err
			}
			date := filepath.Base(filepath.Dir(path))
			title := extractTitle(string(content))

			fileData := FileData{
				Content: string(content),
				Date:    date,
				Title:   title,
			}
			filesData = append(filesData, fileData)
		}
		return nil
	})
	if err != nil {
		fmt.Println(err)
	}

	return filesData
}

func extractTitle(content string) string {
	sentences := strings.Split(content, ".")
	if len(sentences) > 0 {
		return sentences[0]
	}
	return ""
}

func GetDocsSchemaFromCSV(filename string) []schema.Document {
	var docs []schema.Document

	// Open the CSV file
	file, err := os.Open(filename)
	if err != nil {
		log.Fatalf("Failed to open CSV file: %s", err)
	}
	defer file.Close()

	// Create a new CSV reader
	reader := csv.NewReader(file)

	// Read the header
	_, err = reader.Read()
	if err != nil {
		log.Fatalf("Failed to read CSV header: %s", err)
	}

	// Read the rest of the rows
	var files []BioFiles
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalf("Failed to read CSV record: %s", err)
		}

		files = append(files, BioFiles{
			Username: record[1],
			Content:  record[2],
		})
	}

	// Populate the docs slice
	for _, data := range files {
		doc := schema.Document{
			PageContent: data.Content,
			Metadata: map[string]interface{}{
				"username": data.Username,
			},
		}
		docs = append(docs, doc)
	}

	return docs
}
