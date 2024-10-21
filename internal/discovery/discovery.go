package discovery

import (
	"bytes"
	"fmt"
	"github.com/goccy/go-json"
	"net/http"
)

type Discovery struct {
	url  string
	name string
	addr string
}

func New(url, name, addr string) *Discovery {
	return &Discovery{
		url:  url,
		name: name,
		addr: addr,
	}
}

func (d *Discovery) Register() error {
	if req, err := json.Marshal(map[string]string{
		"name":    d.name,
		"address": d.addr,
	}); err != nil {
		return err
	} else {
		post, err := http.Post(fmt.Sprintf("%v/register", d.url), "application/json", bytes.NewBuffer(req))
		if err != nil || post.StatusCode != http.StatusCreated {
			return err
		}
	}

	return nil
}

func (d *Discovery) Deregister() error {
	if req, err := json.Marshal(map[string]string{
		"name":    d.name,
		"address": d.addr,
	}); err != nil {
		return err
	} else {
		post, err := http.Post(fmt.Sprintf("%v/deregister", d.url), "application/json", bytes.NewBuffer(req))
		if err != nil || post.StatusCode != http.StatusOK {
			return err
		}
	}

	return nil
}
