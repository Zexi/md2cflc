package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"

	"github.com/seppestas/go-confluence"
)

var (
	username      = flag.String("u", "", "Confluence username")
	passwd        = flag.String("p", "", "Confluence password")
	pageId        = flag.String("pageid", "", "Confluence page ID to update")
	confluenceURL = flag.String("wiki", "", "Confluence wiki http URL")
)

func optionParse() {
	flag.Parse()
}

func markdownFile() string {
	files := flag.Args()
	if len(files) == 0 {
		fmt.Printf("Please specify markdown file.\n")
		os.Exit(1)
	}
	f := files[0]
	if _, err := os.Stat(f); err != nil {
		log.Fatalf("markdownFile: %v", err)
	}
	return f
}

func Markdown2ConfluenceWiki(file string) (string, error) {
	helper := "markdown2confluence"
	cmd := exec.Command(helper, file)
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return string(out), nil
}

func updateCheck() (err error) {
	if *username == "" || *passwd == "" {
		err = fmt.Errorf("'username' and 'password' must provided")
		return
	}
	if *pageId == "" {
		err = fmt.Errorf("'pageId' not input")
		return
	}
	if *confluenceURL == "" {
		err = fmt.Errorf("confluence wiki URL not input")
		return
	}
	return
}

func newPageByOldPage(oldPage *confluence.Content, content string) *confluence.Content {
	newPage := *oldPage
	newPage.Version.Number = oldPage.Version.Number + 1
	newPage.Body.Storage.Value = content
	newPage.Body.Storage.Representation = "wiki"
	return &newPage
}

func doUpdate(url, username, passwd, pageId, content string) (err error) {
	auth := confluence.BasicAuth(username, passwd)
	wiki, err := confluence.NewWiki(url, auth)
	if err != nil {
		return
	}

	oldPage, err := wiki.GetContent(pageId, []string{"version"})
	if err != nil {
		return
	}

	newPage := newPageByOldPage(oldPage, content)
	ret, err := wiki.UpdateContent(newPage)
	if err != nil {
		return
	}
	return
}

func UpdateContent(content string) error {
	err := updateCheck()
	if err != nil {
		return err
	}

	return doUpdate(*confluenceURL, *username, *passwd, *pageId, content)
}

func main() {
	optionParse()

	wikiContent, err := Markdown2ConfluenceWiki(markdownFile())
	if err != nil {
		log.Fatalf("Convert markdown to wiki: %v", err)
	}

	if len(*confluenceURL) == 0 {
		fmt.Printf("%s", wikiContent)
		os.Exit(0)
	}

	err = UpdateContent(wikiContent)
	if err != nil {
		log.Fatalf("Update err: %v", err)
	}
}
