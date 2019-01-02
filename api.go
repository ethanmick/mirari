package mirari

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"
)

var root = ""
var client = &http.Client{Timeout: 30 * time.Second}

// Upload creates an API request.
func Upload(path string, body interface{}) (*http.Request, error) {
	uri := root + path
	// Gzip data
	buf := new(bytes.Buffer)
	gz := gzip.NewWriter(buf)
	err := json.NewEncoder(gz).Encode(body)
	if err != nil {
		return nil, err
	}
	gz.Close()
	req, err := http.NewRequest("POST", uri, buf)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Content-Encoding", "gzip")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "mirari/0.0.1")
	req.Header.Set("Authorization", "token")
	return req, nil
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
