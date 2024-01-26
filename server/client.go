package server

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/tclairet/merklestore/merkletree"
)

type Client struct {
	url string
}

func NewClient(url string) Client {
	return Client{
		url: url,
	}
}

func (c Client) Upload(root string, index, total int, file io.Reader) error {
	content, err := io.ReadAll(file)
	if err != nil {
		return err
	}
	b, err := json.Marshal(UploadRequest{
		Root:    root,
		Index:   index,
		Total:   total,
		Content: content,
	})
	if err != nil {
		return err
	}
	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s%s", c.url, uploadRoute), bytes.NewBuffer(b))
	if err != nil {
		return err
	}
	response, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	if response.StatusCode != http.StatusOK {
		var message JSONError
		if err := json.NewDecoder(response.Body).Decode(&message); err != nil {
			return err
		}
		return fmt.Errorf("invalid server response %d error '%s'", response.StatusCode, message.Error)
	}
	return nil
}

func (c Client) Request(root string, index int) (io.Reader, *merkletree.Proof, error) {
	b, err := json.Marshal(RequestRequest{
		Root:  root,
		Index: index,
	})
	if err != nil {
		return nil, nil, err
	}
	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s%s", c.url, requestRoute), bytes.NewBuffer(b))
	if err != nil {
		return nil, nil, err
	}
	response, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, nil, err
	}
	if response.StatusCode != http.StatusOK {
		var message JSONError
		if err := json.NewDecoder(response.Body).Decode(&message); err != nil {
			return nil, nil, err
		}
		return nil, nil, fmt.Errorf("invalid server response %d error '%s'", response.StatusCode, message.Error)
	}
	var requestResponse RequestResponse
	if err := json.NewDecoder(response.Body).Decode(&requestResponse); err != nil {
		return nil, nil, err
	}
	return bytes.NewBuffer(requestResponse.Content), merkletree.NewProof(sha256.New, requestResponse.Proof), nil
}
