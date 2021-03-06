package gathering

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/url"
	"time"
)

// Root URL to send data to
var root = ""
var version = ""
var client = &http.Client{Timeout: 600 * time.Second}

// Token set by cli
var Token string

// Upload creates an API request.
func Upload(path string, body interface{}) (*http.Request, error) {
	uri := root + path
	// Gzip data
	// 	buf := new(bytes.Buffer)
	//	gz := gzip.NewWriter(buf)
	//	err := json.NewEncoder(gz).Encode(body)
	json, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	//	gz.Close()
	req, err := http.NewRequest("POST", uri, bytes.NewBuffer(json))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", fmt.Sprintf("gathering/%s", version))
	req.Header.Set("Authorization", "token "+Token)
	return req, nil
}

// UploadFile uploads a file
func UploadFile(path, name string, file io.Reader) (*http.Request, error) {
	uri := root + path
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile(name, name)
	if err != nil {
		return nil, err
	}
	_, err = io.Copy(part, file)
	if err != nil {
		return nil, err
	}
	err = writer.Close()
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", uri, body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("User-Agent", fmt.Sprintf("gathering/%s", version))
	req.Header.Set("Authorization", "token "+Token)
	return req, err
}

// Do sends an API request and returns the API response.
func Do(req *http.Request, v interface{}) (*http.Response, error) {
	resp, err := client.Do(req)
	if err != nil {
		if e, ok := err.(*url.Error); ok {
			if url, err := url.Parse(e.URL); err == nil {
				e.URL = url.String()
				return nil, e
			}
		}
		return nil, err
	}
	defer func() {
		// Drain up to 512 bytes and close the body to let
		// the Transport reuse the connection
		io.CopyN(ioutil.Discard, resp.Body, 512)
		resp.Body.Close()
	}()
	if err != nil {
		// even though there was an error, we still return the response
		// in case the caller wants to inspect it further
		return resp, err
	}
	if v != nil {
		if w, ok := v.(io.Writer); ok {
			io.Copy(w, resp.Body)
		} else {
			err = json.NewDecoder(resp.Body).Decode(v)
			if err == io.EOF {
				err = nil // ignore EOF errors caused by empty response body
			}
		}
	}
	return resp, err
}
