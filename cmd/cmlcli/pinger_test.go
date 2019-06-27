package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
	gock "gopkg.in/h2non/gock.v1"
)

func TestNewPinger(t *testing.T) {
	p, err := newPinger("https://username:password@localhost/1c.php")
	assert.Nil(t, err)
	assert.Equal(t, evCheckAuth, p.next)

	p, err = newPinger("https://localhost")
	assert.Nil(t, err)
	assert.Equal(t, evCheckAuth, p.next)
}

func TestCheckAuth(t *testing.T) {
	defer gock.Off()

	gock.New("https://localhost").
		Get("/").
		Reply(200).
		BodyString("success\ncookiename\ncookievalue")

	p, _ := newPinger("https://localhost")
	err := p.fsm.Do(evCheckAuth)
	assert.Nil(t, err)
	assert.Equal(t, evInit, p.next)
	assert.Equal(t, stAuth, p.fsm.State)

	gock.New("https://localhost").
		Get("/").
		Reply(200).
		BodyString("failure")

	p, _ = newPinger("https://localhost")
	err = p.fsm.Do(evCheckAuth)
	assert.NotNil(t, err)
	assert.Equal(t, evFail, p.next)
	assert.Equal(t, stNew, p.fsm.State)
}
