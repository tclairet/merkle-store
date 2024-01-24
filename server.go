package merklestore

import (
	"fmt"
	"io"
	"net/http"
)

var _ Server = Client{}

type Server interface {
	Upload(name string, file io.Reader) error
}

type Client struct {
	url string
}

func (c Client) Upload(name string, file io.Reader) error {
	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/%s", c.url, name), file)
	if err != nil {
		return err
	}
	response, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("invalid server response %d", response.StatusCode)
	}
	return nil
}
