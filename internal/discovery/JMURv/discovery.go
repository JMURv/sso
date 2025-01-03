package discovery

import (
	"bytes"
	"context"
	"fmt"
	"github.com/JMURv/sso/internal/discovery"
	"github.com/goccy/go-json"
	"net/http"
)

type findRequest struct {
	Name string `json:"name"`
}

type findResponse struct {
	Address string `json:"address"`
}

type registerRequest struct {
	Name    string `json:"name"`
	Address string `json:"address"`
}

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
	req, err := json.Marshal(registerRequest{Name: d.name, Address: d.addr})
	if err != nil {
		return err
	}

	res, err := http.Post(fmt.Sprintf("%v/register", d.url), "application/json", bytes.NewBuffer(req))
	if err != nil {
		return err
	}

	if res.StatusCode != http.StatusOK {
		return discovery.ErrFailedToRegister
	}

	return nil
}

func (d *Discovery) Deregister() error {
	req, err := json.Marshal(registerRequest{Name: d.name, Address: d.addr})
	if err != nil {
		return err
	}

	res, err := http.Post(fmt.Sprintf("%v/deregister", d.url), "application/json", bytes.NewBuffer(req))
	if err != nil {
		return err
	}

	if res.StatusCode != http.StatusOK {
		return discovery.ErrFailedToDeregister
	}

	return nil
}

func (d *Discovery) FindServiceByName(ctx context.Context, name string) (string, error) {
	req, err := json.Marshal(findRequest{Name: name})
	if err != nil {
		return "", err
	}

	post, err := http.Post(fmt.Sprintf("%v/find", d.url), "application/json", bytes.NewBuffer(req))
	if err != nil {
		return "", err
	}

	if post.StatusCode != http.StatusOK {
		return "", discovery.ErrFailedToFindService
	}

	res := &findResponse{}
	if err = json.NewDecoder(post.Body).Decode(&res); err != nil {
		return "", err
	}

	return res.Address, nil
}
