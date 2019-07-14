// Package client https://v8.1c.ru/edi/edi_stnd/131/
// https://dev.1c-bitrix.ru/api_help/sale/algorithms/index.php
// https://www.clickon.ru/blog/exchange-1c-with-bitrix/
// https://mrcappuccino.ru/blog/post/1c-exchange
package client

import (
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	resty "gopkg.in/resty.v1"
)

type (
	// Client of CommerceML site exchange
	Client struct {
		*resty.Client
		_type     string
		sessID    string
		timestamp string
	}

	// Type is Catalog or Sales
	Type int
)

const (
	// Catalog - Выгрузка на сайт
	Catalog Type = iota
	// Sales - Обмен информацией о заказах
	// TODO Sales
)

// New returns Client instance
func New(endpoint string, client Type) *Client {
	return &Client{
		Client: resty.New().SetHostURL(endpoint),
		_type: func() string {
			if client == Catalog {
				return "catalog"
			}
			return "sales"
		}(),
	}
}

// Auth - Начало сеанса
func (c *Client) Auth(username, password string) error {
	res, err := c.R().
		SetQueryParams(map[string]string{
			"type": c._type,
			"mode": "checkauth",
		}).
		SetBasicAuth(username, password).
		Get("/")

	if err != nil {
		return err
	}

	if res.StatusCode() != http.StatusOK {
		return fmt.Errorf("checkauth server error")
	}

	cred := strings.Split(res.String(), "\n")
	if len(cred) < 3 {
		return fmt.Errorf("checkauth unexpected response: %s", res.String())
	}
	if cred[0] != "success" {
		return fmt.Errorf("checkauth unexpected response: %s", res.String())
	}

	c.Cookies = make([]*http.Cookie, 0)
	c.SetCookie(&http.Cookie{
		Name:  cred[1],
		Value: cred[2],
	})

	c.sessID = ""
	if len(cred) >= 4 {
		s := strings.SplitAfterN(cred[3], "sessid=", 2)
		if len(s) == 2 {
			c.sessID = s[1]
		} // TODO else log cred[3]
	}

	c.timestamp = ""
	if len(cred) >= 5 {
		s := strings.SplitAfterN(cred[4], "timestamp=", 2)
		if len(s) == 2 {
			c.timestamp = s[1]
		} // TODO else log cred[4]
	}
	return nil
}

// Init - Запрос параметров от сайта
// returns zip, file_limit, error
func (c *Client) Init() (bool, int, error) {
	query := map[string]string{
		"type": c._type,
		"mode": "init",
	}
	if len(c.sessID) > 0 {
		query["sessid"] = c.sessID
	}

	res, err := c.R().
		SetQueryParams(query).
		Get("/")

	if err != nil {
		return false, 0, err
	}

	if res.StatusCode() != http.StatusOK {
		return false, 0, fmt.Errorf("init server error")
	}

	cred := strings.Split(res.String(), "\n")
	if len(cred) == 2 {
		z := strings.SplitAfterN(cred[0], "zip=", 2)
		if len(z) == 2 {
			f := strings.SplitAfterN(cred[1], "file_limit=", 2)
			if len(f) == 2 {
				if flimit, err := strconv.Atoi(f[1]); err == nil {
					return z[1] == "yes", flimit, nil
				}
			}
		}
	}
	return false, 0, fmt.Errorf("init unexpected response: %s", res.String())
}

// piece returns readers that reads only piece of r with length == n
func piece(r io.Reader, n int64) io.Reader {
	pr, pw := io.Pipe()

	go func() {
		_, err := io.CopyN(pw, r, n)

		if err != nil {
			pw.CloseWithError(err)
		} else {
			pw.Close()
		}
	}()

	return pr
}

// File uploads n bytes from r to server as fname
// returns nil if success
func (c *Client) File(r io.Reader, n int64, fname string) error {
	query := map[string]string{
		"type":     c._type,
		"mode":     "file",
		"filename": fname,
	}
	if len(c.sessID) > 0 {
		query["sessid"] = c.sessID
	}

	res, err := c.R().
		SetQueryParams(query).
		SetBody(piece(r, n)).
		Post("/")

	if err != nil {
		return err
	}

	if res.StatusCode() != http.StatusOK {
		return fmt.Errorf("init server error")
	}

	cred := strings.Split(res.String(), "\n")
	if len(cred) >= 1 {
		if cred[0] == "success" {
			return nil
		}
	}

	return fmt.Errorf("init unexpected response: %s", res.String())
}
