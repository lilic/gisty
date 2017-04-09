package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	flag "github.com/spf13/pflag"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"
)

const (
	githubToken = "GITHUB_TOKEN"
	editor      = "EDITOR"
	base        = "https://api.github.com/gists"
)

type Options struct {
	Create  bool
	Public  bool
	Anon    bool
	Desc    string
	Content string
	Show    string
	Edit    string
	List    bool
}

type Gist struct {
	ID          string                    `json:"id,omitempty"`
	Description string                    `json:"description,omitempty"`
	Public      bool                      `json:"public,omitempty"`
	Files       map[GistFilename]GistFile `json:"files,omitempty"`
	HTMLURL     string                    `json:"html_url,omitempty"`
	UpdatedAt   time.Time                 `json:"updated_at,omitempty"`
}

type GistFilename string

type GistFile struct {
	Content string `json:"content,omitempty"`
}

func doRequest(method string, anon bool, tkn string, gistID string, desc string, public bool, filename string, content io.Reader) (*Gist, error) {
	c, err := ioutil.ReadAll(content)
	if err != nil {
		return nil, err
	}

	requestBody := &Gist{
		Public:      public,
		Description: desc,
		Files: map[GistFilename]GistFile{
			GistFilename(filename): GistFile{
				Content: string(c),
			},
		},
	}
	body := bytes.NewBuffer(nil)
	err = json.NewEncoder(body).Encode(requestBody)
	if err != nil {
		return nil, err
	}
	url := base
	if method == "PATCH" {
		url += "/" + gistID
	}
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}
	if anon == false {
		req.Header.Add("Authorization", "Token "+tkn)
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	gist := Gist{}
	json.NewDecoder(resp.Body).Decode(&gist)
	return &gist, nil
}

func getGist(id string) *Gist {
	client := &http.Client{}
	req, err := http.NewRequest("GET", base+"/"+id, nil)
	if err != nil {
		log.Fatal(err)
	}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
		return nil
	}
	defer resp.Body.Close()
	gist := Gist{}
	json.NewDecoder(resp.Body).Decode(&gist)
	return &gist
}

func getGists(tkn string) []*Gist {
	client := &http.Client{}
	req, err := http.NewRequest("GET", base, nil)
	if err != nil {
		log.Fatal(err)
	}
	req.Header.Add("Authorization", "Token "+tkn)
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	gists := []*Gist{}
	json.NewDecoder(resp.Body).Decode(&gists)
	return gists
}

func printGist(gist *Gist) {
	fmt.Printf("ID: %s\n", gist.ID)
	fmt.Printf("URL: %s\n", gist.HTMLURL)
	fmt.Printf("Date: %s\n\n", gist.UpdatedAt)
	if gist.Description != "" {
		fmt.Println(gist.Description)
	}
	for filename, _ := range gist.Files {
		fmt.Println(filename)
	}
	fmt.Println()
}

func runCreate(o Options) int {
	var content io.Reader

	// Content from STDIN.
	info, err := os.Stdin.Stat()
	if err != nil {
		log.Fatal(err)
	}
	if ((info.Mode() & os.ModeCharDevice) != os.ModeCharDevice) && info.Size() > 0 {
		content = bufio.NewReader(os.Stdin)
	}

	// Content from flag.
	if len(o.Content) > 0 {
		content = strings.NewReader(o.Content)
	}

	// Create a user gist.
	token := os.Getenv(githubToken)
	if token == "" && o.Anon {
		fmt.Printf("Please set ENV variable $%s.\n", githubToken)
		return 1
	}
	g, err := doRequest("POST", o.Anon, token, "", o.Desc, o.Public, "text1.txt", content)
	if err != nil {
		log.Fatal(err)
	}
	printGist(g)
	return 0
}

func runShow(o Options) int {
	gist := getGist(o.Show)
	printGist(gist)
	return 0
}

func runEdit(o Options) int {
	token := os.Getenv(githubToken)
	if token == "" {
		fmt.Printf("Please set ENV variable $%s.\n", githubToken)
		return 1
	}
	e := os.Getenv(editor)
	if e == "" {
		e = "vim"
	}

	var content []byte
	var filename string
	gist := getGist(o.Edit)
	for f, gf := range gist.Files {
		content = []byte(gf.Content)
		filename = string(f)
	}
	tmpFile, err := ioutil.TempFile(os.TempDir(), "gisty")
	if err != nil {
		log.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())
	if _, err := tmpFile.Write(content); err != nil {
		log.Fatal(err)
	}
	if err := tmpFile.Close(); err != nil {
		log.Fatal(err)
	}

	cmd := exec.Command(e, tmpFile.Name())
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Start()
	if err != nil {
		log.Fatal(err)
	}
	err = cmd.Wait()
	if err != nil {
		log.Fatal(err)
	}

	file, err := os.OpenFile(tmpFile.Name(), 0, 0)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	c := bufio.NewReader(file)
	g, err := doRequest("PATCH", false, token, o.Edit, "", gist.Public, filename, c)
	if err != nil {
		log.Fatal(err)
	}
	printGist(g)
	return 0
}

func runList(o Options) int {
	token := os.Getenv(githubToken)
	if token == "" && o.Anon {
		fmt.Printf("Please set ENV variable $%s.\n", githubToken)
		return 1
	}
	gists := getGists(token)
	for _, gist := range gists {
		printGist(gist)
	}
	return 0
}

func Main() int {
	options := Options{}
	flags := flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	flags.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
		flags.PrintDefaults()
	}
	flags.BoolVar(&options.Create, "create", false, "create a private gist that will be stored under your profile.")
	flags.BoolVar(&options.Public, "public", false, "create a public gist.")
	flags.BoolVar(&options.Anon, "anon", false, "create an anonymous private gist.")
	flags.StringVar(&options.Desc, "description", "", "specify gist description, if not provided will be left blank.")
	flags.StringVar(&options.Content, "content", "", "specify content of the gist")
	flags.StringVar(&options.Show, "show", "", "pass a gist ID and it displays a gist.")
	flags.StringVar(&options.Edit, "edit", "", "pass a gist ID to be able to edit your gist.")
	flags.BoolVar(&options.List, "list", false, "lists first 30 of your gists.")
	flags.Parse(os.Args[1:])

	if options.Create {
		return runCreate(options)
	}
	if options.Show != "" {
		return runShow(options)
	}
	if options.Edit != "" {
		return runEdit(options)
	}
	if options.List {
		return runList(options)
	}

	flags.Usage()
	return 1
}

func main() {
	os.Exit(Main())
}
