// Package client https://v8.1c.ru/edi/edi_stnd/131/
// https://dev.1c-bitrix.ru/api_help/sale/algorithms/index.php
// https://www.clickon.ru/blog/exchange-1c-with-bitrix/
// https://mrcappuccino.ru/blog/post/1c-exchange
package client

import (
	"fmt"
	"net/http"
	"strings"

	"gopkg.in/resty.v1"
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
