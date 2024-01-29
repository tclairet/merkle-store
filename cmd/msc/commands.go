package main

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "msc",
	Short: "msc (MerkleStoreClient) is an uploader and downloader backed by a Merkletree",
	CompletionOptions: cobra.CompletionOptions{
		DisableDefaultCmd: true,
	},
}

var (
	uploadCmd = &cobra.Command{
		Use:   "upload [FILES]",
		Short: "Upload set of files",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := MerkleStoreClient()
			if err != nil {
				return err
			}
			root, err := client.Upload(args)
			if err != nil {
				return err
			}
			fmt.Println("Files Upload with success")
			fmt.Println("Merkle Root:", root)
			fmt.Println("use it to retrieve your files")
			return nil
		},
	}

	downloadCmd = &cobra.Command{
		Use:   "download ROOT_HASH [FILES_INDEX]",
		Short: "Download the i file",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := MerkleStoreClient()
			if err != nil {
				return err
			}
			if len(args) < 2 {
				return fmt.Errorf("you must provide the root hash and indexes of the files you want to download")
			}
			var indexes []int
			for i := 1; i < len(args); i++ {
				index, err := strconv.ParseInt(args[i], 10, 10)
				if err != nil {
					return fmt.Errorf("index must be a int")
				}
				indexes = append(indexes, int(index))
			}
			if err := client.Download(args[0], indexes...); err != nil {
				return err
			}
			fmt.Println("Files Download with success")
			for i := 0; i < len(args[1:]); i++ {
				fmt.Printf("\t%s/%s\n", args[0], args[i+1])
			}
			return nil
		},
	}
)
