package client

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/h2non/gock.v1"
)

func TestClientAuthSuccess(t *testing.T) {
	client := New("http://localhost/1c.php", Catalog)
	assert.Equal(t, "catalog", client._type)

	defer gock.Off()

	gock.New("http://localhost").
		Get("/1c.php").
		// Basic base64(username:password)
		MatchHeader("Authorization", "Basic dXNlcm5hbWU6cGFzc3dvcmQ=").
		MatchParams(map[string]string{"mode": "checkauth", "type": client._type}).
		Reply(200).
		BodyString("success\ncookiename\ncookievalue\ntail\n")

	assert.Nil(t, client.Auth("username", "password"))
	assert.Equal(t, 1, len(client.Cookies))
	assert.Equal(t, "cookiename", client.Cookies[0].Name)
	assert.Equal(t, "cookievalue", client.Cookies[0].Value)
}

func TestClientReAuth(t *testing.T) {
	client := New("http://localhost/1c.php", Catalog)

	defer gock.Off()

	gock.New("http://localhost").
		Get("/1c.php").
		Reply(200).
		BodyString("success\ncookiename\ncookievalue\ntail\n")

	assert.Nil(t, client.Auth("username", "password"))
	assert.Equal(t, 1, len(client.Cookies))

	gock.New("http://localhost").
		Get("/1c.php").
		Reply(200).
		BodyString("success\ncookiename2\ncookievalue2\ntail\n")

	assert.Nil(t, client.Auth("username", "password"))
	assert.Equal(t, 1, len(client.Cookies))

	assert.Equal(t, "cookiename2", client.Cookies[0].Name)
	assert.Equal(t, "cookievalue2", client.Cookies[0].Value)
}

func TestClientAuthFailure(t *testing.T) {
	client := New("http://localhost/1c.php", Catalog)

	defer gock.Off()

	gock.New("http://localhost").
		Get("/1c.php").
		Reply(200).
		BodyString("failure\ncookiename\ncookievalue\n")

	assert.NotNil(t, client.Auth("username", "password"))
	assert.Equal(t, 0, len(client.Cookies))
}

func TestClientAuthResponseTooShort(t *testing.T) {
	client := New("http://localhost/1c.php", Catalog)

	defer gock.Off()

	gock.New("http://localhost").
		Get("/1c.php").
		Reply(200).
		BodyString("success\ncookiename\n")

	assert.NotNil(t, client.Auth("username", "password"))
	assert.Equal(t, 0, len(client.Cookies))
}

func TestClientAuthServerError(t *testing.T) {
	client := New("http://localhost/1c.php", Catalog)

	defer gock.Off()

	gock.New("http://localhost").
		Get("/1c.php").
		Reply(500).
		BodyString("success\ncookiename\ncookievalue\n")

	assert.NotNil(t, client.Auth("username", "password"))
	assert.Equal(t, 0, len(client.Cookies))
}
