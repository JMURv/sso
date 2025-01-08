package discovery

import (
	"bytes"
	"context"
	"fmt"
	"github.com/JMURv/sso/internal/discovery"
	conf "github.com/JMURv/sso/pkg/config"
	"github.com/goccy/go-json"
	"log"
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

func New(url *conf.SrvDiscoveryConfig, name string, addr *conf.ServerConfig) discovery.ServiceDiscovery {
	return &Discovery{
		url:  fmt.Sprintf("%v://%v:%v", url.Scheme, url.Host, url.Port),
		name: name,
		addr: fmt.Sprintf("%v://%v:%v", addr.Scheme, addr.Domain, addr.Port),
	}
}

func (d *Discovery) Close() error {
	return nil
}

func (d *Discovery) Register(_ context.Context) error {
	req, err := json.Marshal(registerRequest{Name: d.name, Address: d.addr})
	if err != nil {
		return err
	}

	res, err := http.Post(fmt.Sprintf("%v/register", d.url), "application/json", bytes.NewBuffer(req))
	if err != nil {
		return err
	}

	if res.StatusCode != http.StatusCreated {
		log.Println(res.StatusCode)
		return discovery.ErrFailedToRegister
	}

	return nil
}

func (d *Discovery) Deregister(_ context.Context) error {
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

func (d *Discovery) FindServiceByName(_ context.Context, name string) (string, error) {
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
