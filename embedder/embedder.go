package main

import (
	"fmt"
    "io/ioutil"
    "os"
    "path/filepath"
    "strings"
	"github.com/tmc/langchaingo/embeddings"
)

type FileData struct {
    Directory string
    Filename  string
    Content   string
    Date      string
}

func GetDocsShemaByFile(fileData []FileData) []schema.Document {
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

            fileData := FileData{
                Directory: directory,
                Filename:  filename,
                Content:   string(content),
            }

            fmt.Println("Directory:", directory)
            fmt.Println("Filename:", filename)
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

func main() {
    _, err := parseFiles("/../medusa_dump")
    if err != nil {
        fmt.Println(err)
    }
}