package merklestore

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/tclairet/merklestore/merkletree"
)

var _ Server = Client{}

type Server interface {
	Upload(name string, file io.Reader) error
	Request(name string) (io.Reader, *merkletree.Proof, error)
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

func (c Client) Request(name string) (io.Reader, *merkletree.Proof, error) {
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/%s", c.url, name), nil)
	if err != nil {
		return nil, nil, err
	}
	response, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, nil, err
	}
	if response.StatusCode != http.StatusOK {
		return nil, nil, fmt.Errorf("invalid server response %d", response.StatusCode)
	}
	var proofResponse ProofResponse
	if err := json.NewDecoder(response.Body).Decode(&proofResponse); err != nil {
		return nil, nil, err
	}
	return bytes.NewBuffer(proofResponse.File), merkletree.NewProof(sha256.New, proofResponse.Proof), nil
}

type ProofResponse struct {
	File  []byte   `json:"file"`
	Proof [][]byte `json:"proof"`
}
