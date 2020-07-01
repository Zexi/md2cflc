package main

import (
	"bufio"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"strconv"
	"strings"

	"md2cflc/confluence"
	"md2cflc/render"
)

var (
	username      = flag.String("u", "", "Confluence username")
	passwd        = flag.String("p", "", "Confluence password")
	pageId        = flag.String("pageid", "", "Confluence page ID to update")
	confluenceURL = flag.String("wiki", "", "Confluence wiki http URL")
	parentID      = flag.Int("parentid", 0, "parent id of a page")
	title         = flag.String("title", "", "title of a new page")
	space         = flag.String("space", "", "page Space in the wiki")
	verbose       = flag.Bool("verbose", false, "enable debug mode")
	noEscape      = flag.Bool("no-escape", false, "not escape curly brackets")
)

func Debug(data []byte, err error) {
	if err == nil {
		fmt.Printf("%s\n\n", data)
	} else {
		fmt.Printf("%s\n\n", err)
	}
}

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
	return escapeCurlyBrackets(f)
}

func escapeCurlyBrackets(filename string) string {
	fin, _ := os.Open(filename)

	reader := bufio.NewReader(fin)
	ctt, _ := ioutil.ReadAll(reader)

	content := string(ctt)
	if !*noEscape {
		content := strings.Replace(content, "{", `\\{`, -1)
		content = strings.Replace(content, "}", `\\}`, -1)
	}

	bn := path.Base(filename)
	tmpFileName := "/tmp/" + bn + ".tmp"
	out, err := os.OpenFile(tmpFileName, os.O_RDWR|os.O_CREATE, 0755)
	defer out.Close()
	if err != nil {
		log.Fatal(err)
	}

	w := bufio.NewWriter(out)
	w.WriteString(content)
	w.Flush()
	return tmpFileName
}

func Markdown2ConfluenceWiki(file string) (string, error) {
	if strings.HasSuffix(file, ".tmp") {
		defer os.Remove(file)
	}
	markdownContents, err := ioutil.ReadFile(file)
	if err != nil {
		return "", fmt.Errorf("read file %s error: %v", file, err)
	}
	output := render.Run(markdownContents)
	return string(output), nil
}

func updateCheck() (err error) {
	if *username == "" || *passwd == "" {
		err = fmt.Errorf("'username' and 'password' must provided")
		return
	}
	if *pageId == "" && *parentID == 0 {
		err = fmt.Errorf("'pageId' and 'parentID' can not be both empty")
		return
	}

	if *pageId != "" && *parentID > 0 {
		err = fmt.Errorf("please provide pageId OR parentID. Do not provid BOTH")
		return
	}

	if *parentID > 0 && *space == "" {
		err = fmt.Errorf("please provide SPACE key when you provided parentID")
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

func newPage4Create(oldPage *confluence.Content) *confluence.ContentCreate {
	newPage := &confluence.ContentCreate{
		Space:     confluence.Space{""},
		Ancestors: make([]confluence.Ancestor, 0),
		Content:   *oldPage,
	}
	return newPage
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
	_, err = wiki.UpdateContent(newPage, *verbose)
	if err != nil {
		return
	}
	return
}

func doCreate(url, username, passwd, title, content, space string, parentID int) (err error) {

	auth := confluence.BasicAuth(username, passwd)
	wiki, err := confluence.NewWiki(url, auth)
	if err != nil {
		return
	}

	oldPage, err := wiki.GetContent(strconv.Itoa(parentID), []string{"version"})
	if err != nil {
		return
	}

	_newPage := newPageByOldPage(oldPage, content)
	newPage := newPage4Create(_newPage)
	newPage.Version.Number = 1
	newPage.Title = title
	// newPage.Ancestors = append(newPage.Ancestors, confluence.Ancestor{parentID})

	ans := confluence.Ancestor{parentID}
	newPage.Ancestors = append(newPage.Ancestors, ans)
	newPage.Space.Key = space

	_, err = wiki.CreateContent(newPage, *verbose)
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
	if *pageId != "" {
		fmt.Println("do update")
		return doUpdate(*confluenceURL, *username, *passwd, *pageId, content)
	} else if *parentID > 0 && *space != "" {
		fmt.Println("do create")
		return doCreate(*confluenceURL, *username, *passwd, *title, content, *space, *parentID)
	}
	return nil
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

	if err := UpdateContent(wikiContent); err != nil {
		log.Fatalf("Update err: %v", err)
	}
}
