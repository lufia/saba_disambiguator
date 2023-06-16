package twitter2

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

func getJSON[T any](u *url.URL, header http.Header) (v T, err error) {
	req := &http.Request{
		Method: "GET",
		Header: header,
		URL:    u,
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return v, err
	}
	defer func() {
		io.Copy(io.Discard, resp.Body)
		resp.Body.Close()
	}()

	if resp.StatusCode >= 400 {
		return v, fmt.Errorf("getJSON: failed with status: %s", resp.Status)
	}

	err = json.NewDecoder(resp.Body).Decode(&v)
	return v, err
}
