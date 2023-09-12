package twitter2

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

type Client struct {
	bearerToken string
}

func NewClient(bearerToken string) *Client {
	return &Client{bearerToken: bearerToken}
}

func (c *Client) newHeader() http.Header {
	p := http.Header{}
	p.Set("Authorization", fmt.Sprintf("Bearer %s", c.bearerToken))
	p.Set("User-Agent", "sabadisambiguator")
	return p
}

func (c *Client) getJSON(v any, u *url.URL) error {
	req := &http.Request{
		Method: "GET",
		Header: c.newHeader(),
		URL:    u,
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer func() {
		io.Copy(io.Discard, resp.Body)
		resp.Body.Close()
	}()

	if resp.StatusCode >= 400 {
		errmsg, _ := io.ReadAll(resp.Body)                          // ここに到達した時点でエラー扱いなので、ここのエラーは無視する。
		errmsg = bytes.ReplaceAll(errmsg, []byte("\n"), []byte("")) // ログを考慮して改行を消す。おそらく JSON なので、消して問題ない。
		return fmt.Errorf("getJSON: failed with status %s: %s", resp.Status, errmsg)
	}

	err = json.NewDecoder(resp.Body).Decode(&v)
	return err
}
