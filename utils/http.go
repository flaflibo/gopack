package utils

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

type Request struct {
	username string
	password string
	baseUrl  string
	client   *http.Client
}

func (h *Request) Init(username string, password string, baseUrl string) {
	h.username = username
	h.password = password
	h.baseUrl = baseUrl
	h.client = &http.Client{}
}

func (h *Request) Get(apiUrl string, params map[string]string) ([]byte, error) {

	newUrl, _ := url.Parse(fmt.Sprintf("%s%s", h.baseUrl, apiUrl))
	urlParams := url.Values{}

	for key, value := range params {
		urlParams.Add(key, value)
	}
	newUrl.RawQuery = urlParams.Encode()

	req, err := http.NewRequest("GET", newUrl.String(), nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", "application/json")
	req.SetBasicAuth(h.username, h.password)

	resp, err := h.client.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("get request failed", resp.Status)
	}

	return io.ReadAll(resp.Body)
}

func (h *Request) Post(url string, data []byte) ([]byte, error) {
	newUrl := fmt.Sprintf("%s%s", h.baseUrl, url)

	req, err := http.NewRequest("POST", newUrl, bytes.NewBuffer(data))
	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", "application/json")
	req.SetBasicAuth(h.username, h.password)
	resp, err := h.client.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != 200 && resp.StatusCode != 201 {
		return nil, fmt.Errorf("post request failed", resp.Status)
	}

	return io.ReadAll(resp.Body)
}
