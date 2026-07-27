package twitter2

import (
	"bytes"
	"context"
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

func (c *Client) getJSON(ctx context.Context, v any, u *url.URL) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.bearerToken))
	req.Header.Set("User-Agent", "sabadisambiguator")

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
