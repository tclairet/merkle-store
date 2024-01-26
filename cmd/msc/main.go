package main

import (
	"fmt"
	"os"

	"github.com/tclairet/merklestore/client"
	"github.com/tclairet/merklestore/files"
	"github.com/tclairet/merklestore/server"
)

var (
	envMerkleStoreServer = os.Getenv("MERKLE_STORE_SERVER")

	merkleStoreServerEnvFlag string
)

func init() {
	rootCmd.PersistentFlags().StringVar(&merkleStoreServerEnvFlag, "server", envMerkleStoreServer, "MerkleStoreServer url")

	rootCmd.AddCommand(uploadCmd, downloadCmd)
}

func MerkleStoreClient() (*client.Uploader, error) {
	fileHandler := files.OS{}
	if merkleStoreServerEnvFlag == "" {
		return nil, fmt.Errorf("--server not provided or MERKLE_STORE_SERVER env variable not set")
	}
	serverClient := server.NewClient(merkleStoreServerEnvFlag)
	return client.NewUploader(fileHandler, serverClient), nil
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
