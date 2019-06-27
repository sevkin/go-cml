package client

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	gock "gopkg.in/h2non/gock.v1"
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
		BodyString("success\ncookiename\ncookievalue\nsessid=session\ntimestamp=time\n")

	assert.Nil(t, client.Auth("username", "password"))
	assert.Equal(t, 1, len(client.Cookies))
	assert.Equal(t, "cookiename", client.Cookies[0].Name)
	assert.Equal(t, "cookievalue", client.Cookies[0].Value)
	assert.Equal(t, "session", client.sessID)
	assert.Equal(t, "time", client.timestamp)
}

func TestClientReAuth(t *testing.T) {
	client := New("http://localhost/1c.php", Catalog)

	defer gock.Off()

	gock.New("http://localhost").
		Get("/1c.php").
		Reply(200).
		BodyString("success\ncookiename\ncookievalue\nsessid=session\ntimestamp=time\n")

	assert.Nil(t, client.Auth("username", "password"))
	assert.Equal(t, 1, len(client.Cookies))

	gock.New("http://localhost").
		Get("/1c.php").
		Reply(200).
		BodyString("success\ncookiename2\ncookievalue2\nsessid=session2\ntimestamp=time2\n")

	assert.Nil(t, client.Auth("username", "password"))
	assert.Equal(t, 1, len(client.Cookies))

	assert.Equal(t, "cookiename2", client.Cookies[0].Name)
	assert.Equal(t, "cookievalue2", client.Cookies[0].Value)
	assert.Equal(t, "session2", client.sessID)
	assert.Equal(t, "time2", client.timestamp)

	gock.New("http://localhost").
		Get("/1c.php").
		Reply(200).
		BodyString("success\ncookiename3\ncookievalue3\n")

	assert.Nil(t, client.Auth("username", "password"))
	assert.Equal(t, "", client.sessID)
	assert.Equal(t, "", client.timestamp)
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

// ////////////////////////////////////////////////////////////////////////////

func TestClientInitSuccess(t *testing.T) {
	client := New("http://localhost/1c.php", Catalog)

	defer gock.Off()

	gock.New("http://localhost").
		Get("/1c.php").
		MatchHeader("Cookie", "cookiename=cookievalue").
		MatchParams(map[string]string{"mode": "init", "type": client._type}).
		Reply(200).
		BodyString("zip=yes\nfile_limit=1024\n")

	assert.Equal(t, 0, len(client.Cookies))
	client.SetCookie(&http.Cookie{Name: "cookiename", Value: "cookievalue"})

	zip, flimit, err := client.Init()
	assert.Nil(t, err)
	assert.Equal(t, 1024, flimit)
	assert.True(t, zip)
}

func TestClientInitSessID(t *testing.T) {
	client := New("http://localhost/1c.php", Catalog)

	defer gock.Off()

	gock.New("http://localhost").
		Get("/1c.php").
		MatchParams(map[string]string{"mode": "init", "type": client._type, "sessid": "session"}).
		Reply(200).
		BodyString("zip=yes\nfile_limit=1024\n")

	client.sessID = "session"
	_, _, err := client.Init()
	assert.Nil(t, err)
}

func TestClientInitFailed(t *testing.T) {
	client := New("http://localhost/1c.php", Catalog)

	defer gock.Off()

	gock.New("http://localhost").
		Get("/1c.php").
		Reply(200).
		BodyString("zip=yes\nfile_limit=zero\n")

	_, _, err := client.Init()
	assert.NotNil(t, err)
}

func TestClientInitServerError(t *testing.T) {
	client := New("http://localhost/1c.php", Catalog)

	defer gock.Off()

	gock.New("http://localhost").
		Get("/1c.php").
		Reply(500).
		BodyString("zip=yes\nfile_limit=1024\n")

	_, _, err := client.Init()
	assert.NotNil(t, err)
}
