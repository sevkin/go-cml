package main

import (
	"fmt"
	"log"
	"net/url"

	"github.com/sevkin/cml/client"
	"github.com/sevkin/fsm"
)

const (
	stNew fsm.State = iota
	stAuth
	stInit
	stFile
	stImport
	stDone
)
const (
	evCheckAuth fsm.Input = iota
	evInit
	evFile
	evImport
	evFail
	evDone
)

type (
	pinger struct {
		*client.Client
		username, password string
		fsm                *fsm.FSM
		next               fsm.Input
	}
)

func newPinger(rawURL string) (*pinger, error) {
	url, err := url.Parse(rawURL)
	if err != nil {
		return nil, err
	}

	var username, password string

	if url.User != nil {
		username = url.User.Username()
		password, _ = url.User.Password()
		url.User = nil
	}

	endpoint := url.String()

	if endpoint == "" {
		return nil, fmt.Errorf("bad endpoint for url: %s", rawURL)
	}

	p := &pinger{
		Client:   client.New(endpoint, client.Catalog),
		username: username,
		password: password,
		next:     evCheckAuth,
	}

	fsm := fsm.New(stNew).
		On(evCheckAuth, stNew, stAuth, p._checkauth).
		On(evInit, stAuth, stInit, p._init).
		On(evFile, stInit, stFile, p._file).
		On(evFile, stFile, stFile, p._file).
		On(evImport, stInit, stImport, p._import).
		On(evImport, stFile, stImport, p._import).
		On(evImport, stImport, stImport, p._import).
		// TODO log fail
		On(evFail, stNew, stDone).
		On(evFail, stAuth, stDone).
		On(evFail, stInit, stDone).
		On(evFail, stFile, stDone).
		On(evFail, stImport, stDone).
		On(evDone, stNew, stDone).
		On(evDone, stAuth, stDone).
		On(evDone, stInit, stDone).
		On(evDone, stFile, stDone).
		On(evDone, stImport, stDone)

	p.fsm = fsm
	return p, nil
}

func (p *pinger) _checkauth() error {
	err := p.Auth(p.username, p.password)
	if err == nil {
		p.next = evInit
		return nil
	}
	p.next = evFail
	return err
}

func (p *pinger) _init() error {
	// TODO if.. p.next = evImport else
	p.next = evFile
	return nil
}

func (p *pinger) _file() error {
	// TODO files and parts
	p.next = evImport
	return nil
}

func (p *pinger) _import() error {
	// TODO repeat until progress
	p.next = evDone
	return nil
}

func (p *pinger) ping(term chan struct{}) {
	for p.fsm.State != stDone {
		log.Printf("%#v", p.fsm.State)

		select {
		case <-term:
			return
		default:
			// TODO log analyze error
			err := p.fsm.Do(p.next)
			if err != nil {
				log.Printf("%#v", err)
			}
		}
	}
}
