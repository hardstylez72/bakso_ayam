package scanner

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
)

func MakeUrl(strs ...string) string {
	return strings.Join(strs, "")
}

func Request[Req any, Res any](ctx context.Context, cli *http.Client, method string, url string, req *Req) (*Res, error) {

	marshal, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	var reqBody io.Reader
	if method != http.MethodGet {
		reqBody = bytes.NewBuffer(marshal)
	} else {
		reqBody = nil
	}

	r, err := http.NewRequestWithContext(ctx, method, url, reqBody)
	if err != nil {
		return nil, err
	}
	r.Header.Set("Content-Type", "application/json")
	res, err := cli.Do(r)
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()

	if res.StatusCode != 200 {
		body, err := io.ReadAll(res.Body)
		if err != nil {
			return nil, err
		}
		return nil, errors.New(string(body))
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	var ser Res
	if err := json.Unmarshal(body, &ser); err != nil {
		return nil, err
	}

	return &ser, nil
}
