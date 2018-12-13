package utils

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"time"
)

var netTransport = &http.Transport{
	Dial: (&net.Dialer{
		Timeout: 5 * time.Second,
	}).Dial,
	TLSHandshakeTimeout:   5 * time.Second,
	ResponseHeaderTimeout: 5 * time.Second,
}

var HTTPClient = &http.Client{
	Transport: netTransport,
	Timeout:   2 * time.Second,
}

type LimitedReader struct {
	R io.ReadCloser // underlying reader/closer
	N int64         // max bytes remaining
}

func (l *LimitedReader) Read(p []byte) (n int, err error) {
	if l.N <= 0 {
		return 0, io.EOF
	}
	if int64(len(p)) > l.N {
		p = p[0:l.N]
	}
	n, err = l.R.Read(p)
	l.N -= int64(n)
	return
}

func (l *LimitedReader) Close() error {
	return l.R.Close()
}

// LimitBody is similar to io.LimitReader but it takes an io.ReadCloser
// and accepts the limit in megabytes
func LimitBody(rc io.ReadCloser, megaBytes int64) io.ReadCloser {
	return &LimitedReader{R: rc, N: megaBytes * 1024 * 1024}
}

func GetJSON(url string, data interface{}) error {
	response, err := HTTPClient.Get(url)
	if err != nil {
		return err
	}
	err = json.NewDecoder(response.Body).Decode(data)
	if err != nil {
		return err
	}
	io.Copy(ioutil.Discard, response.Body)
	response.Body.Close()
	return nil
}

func FakeBrowser(req *http.Request) {
	req.Header.Set("Pragma", "no-cache")
	req.Header.Set("Origin", "https://www.hackterms.com")
	req.Header.Set("Accept-Encoding", "gzip, deflate, br")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")
	req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/64.0.3282.119 Safari/537.36")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")
	req.Header.Set("Accept", "*/*")
	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("X-Requested-With", "XMLHttpRequest")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Referer", "https://www.hackterms.com/")
}

func FakePOST(url string, body io.Reader) ([]byte, error) {
	request, err := http.NewRequest("POST", url, body)
	if err != nil {
		return []byte{}, err
	}
	FakeBrowser(request)
	response, err := HTTPClient.Do(request)
	if err != nil {
		return []byte{}, err
	}
	data, err := ioutil.ReadAll(response.Body)
	response.Body.Close()
	return data, err
}
