// Package client https://v8.1c.ru/edi/edi_stnd/131/
// https://dev.1c-bitrix.ru/api_help/sale/algorithms/index.php
// https://www.clickon.ru/blog/exchange-1c-with-bitrix/
// https://mrcappuccino.ru/blog/post/1c-exchange
package client

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"

	"golang.org/x/net/html/charset"
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

func unexpectedResponse(method string, res *resty.Response) error {
	errorMsg := res.String()

	contentType := res.Header()["Content-Type"]
	if len(contentType) == 1 {
		// convert cp1251 -> utf8 if possible
		utf8, err := charset.NewReader(bytes.NewBuffer(res.Body()), contentType[0])
		if err == nil {
			buf, err := ioutil.ReadAll(utf8)
			if err == nil {
				errorMsg = string(buf[:])
			}
		}
	}

	return fmt.Errorf("%s: unexpected response: %v", method, errorMsg)
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
		return fmt.Errorf("checkauth: server error")
	}

	cred := strings.Split(res.String(), "\n")
	if len(cred) < 3 {
		return unexpectedResponse("checkauth", res)
	}
	if cred[0] != "success" {
		return unexpectedResponse("checkauth", res)
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
func (c *Client) Init() (bool, int64, error) {
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
		return false, 0, fmt.Errorf("init: server error")
	}

	cred := strings.Split(res.String(), "\n")
	if len(cred) == 2 {
		z := strings.SplitAfterN(cred[0], "zip=", 2)
		if len(z) == 2 {
			f := strings.SplitAfterN(cred[1], "file_limit=", 2)
			if len(f) == 2 {
				if flimit, err := strconv.Atoi(f[1]); err == nil {
					return z[1] == "yes", int64(flimit), nil
				}
			}
		}
	}
	return false, 0, unexpectedResponse("init", res)
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
		return fmt.Errorf("file: server error")
	}

	cred := strings.Split(res.String(), "\n")
	if len(cred) >= 1 {
		if cred[0] == "success" {
			return nil
		}
	}

	return unexpectedResponse("file", res)
}

// Import process uploaded files
// returns (true, nil) if in progress, (false, nil) if complete
func (c *Client) Import(fname string) (bool, error) {
	query := map[string]string{
		"type":     c._type,
		"mode":     "import",
		"filename": fname,
	}
	if len(c.sessID) > 0 {
		query["sessid"] = c.sessID
	}

	res, err := c.R().
		SetQueryParams(query).
		Get("/")

	if err != nil {
		return false, err
	}

	if res.StatusCode() != http.StatusOK {
		return false, fmt.Errorf("import: server error")
	}

	cred := strings.Split(res.String(), "\n")
	if len(cred) >= 1 {
		switch cred[0] {
		case "success":
			return false, nil
		case "progress":
			return true, nil
		}
	}

	return false, unexpectedResponse("import", res)
}

// Deactivate - деактивация незагруженных товаров и всей информации по ним (при полной выгрузке)
func (c *Client) Deactivate() error {
	if len(c.sessID) == 0 {
		return fmt.Errorf("deactivate: can`t execute without sessid")
	}

	if len(c.timestamp) == 0 {
		return fmt.Errorf("deactivate: can`t execute without timestamp")
	}

	res, err := c.R().
		SetQueryParams(map[string]string{
			"type":      c._type,
			"mode":      "deactivate",
			"sessid":    c.sessID,
			"timestamp": c.timestamp,
		}).
		Get("/")

	if err != nil {
		return err
	}

	if res.StatusCode() != http.StatusOK {
		return fmt.Errorf("deactivate: server error")
	}

	return nil
}

// Complete - %)
// "особо важного тут ничего не делается, а только бросается событие, что импорт успешно завершен"
func (c *Client) Complete() error {
	if len(c.sessID) == 0 {
		return fmt.Errorf("complete: can`t execute without sessid")
	}

	res, err := c.R().
		SetQueryParams(map[string]string{
			"type":   c._type,
			"mode":   "complete",
			"sessid": c.sessID,
		}).
		Get("/")

	if err != nil {
		return err
	}

	if res.StatusCode() != http.StatusOK {
		return fmt.Errorf("complete: server error")
	}

	return nil
}
