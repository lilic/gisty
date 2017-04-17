package gist

import (
	"bytes"
	"encoding/json"
	"net/http"
	"time"
)

const base = "https://api.github.com/gists"

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
type Request struct {
	method string
	url    string
	token  string
	body   *Gist
}

type Response struct {
	resp *http.Response
	err  error
}

func newRequest(method string, url string) *Request {
	return &Request{
		method: method,
		url:    url,
	}
}

func (r *Request) Token(tkn string) *Request {
	r.token = tkn
	return r
}

func (r *Request) Body(g *Gist) *Request {
	r.body = g
	return r
}

func (r *Request) Do() *Response {
	body := bytes.NewBuffer(nil)
	if r.body != nil {
		err := json.NewEncoder(body).Encode(r.body)
		if err != nil {
			return &Response{resp: nil, err: err}
		}
	}
	req, err := http.NewRequest(r.method, r.url, body)
	if err != nil {
		return &Response{resp: nil, err: err}
	}
	if r.token != "" {
		req.Header.Add("Authorization", "Token "+r.token)
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return &Response{resp: nil, err: err}
	}
	return &Response{resp: resp, err: nil}
}

func (r *Response) Handle(input interface{}) error {
	if r.err != nil {
		return r.err
	}
	defer r.resp.Body.Close()
	return json.NewDecoder(r.resp.Body).Decode(input)
}

func Create(token string, requestGist *Gist) (*Gist, error) {
	gist := &Gist{}
	err := newRequest("POST", base).Token(token).Body(requestGist).Do().Handle(gist)
	if err != nil {
		return nil, err
	}
	return gist, nil
}

func Show(token string, id string) (*Gist, error) {
	gist := &Gist{}
	url := base + "/" + id
	err := newRequest("GET", url).Token(token).Do().Handle(gist)
	if err != nil {
		return nil, err
	}
	return gist, nil
}

func Update(token string, id string, requestGist *Gist) (*Gist, error) {
	url := base + "/" + id
	gist := &Gist{}
	err := newRequest("PATCH", url).Token(token).Body(requestGist).Do().Handle(gist)
	if err != nil {
		return nil, err
	}
	return gist, nil
}

func List(token string) ([]*Gist, error) {
	gists := []*Gist{}
	err := newRequest("GET", base).Token(token).Do().Handle(&gists)
	if err != nil {
		return nil, err
	}
	return gists, nil
}
