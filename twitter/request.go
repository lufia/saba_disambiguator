package twitter2

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

func getJSON(v any, u *url.URL, header http.Header) error {
	req := &http.Request{
		Method: "GET",
		Header: header,
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
