package tests

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"testing"
	"time"

	"github.com/tclairet/merklestore/client"
	"github.com/tclairet/merklestore/files"
	"github.com/tclairet/merklestore/server"
)

func TestE2E(t *testing.T) {
	defer cleanUp()

	fileHandler := files.OS{}
	store, err := server.NewJsonStore(fileHandler)
	if err != nil {
		panic(err)
	}
	s, err := server.New(fileHandler, store)
	if err != nil {
		panic(err)
	}
	api := server.NewAPI(s)

	go makeServer(api.Routes())

	serverClient := server.NewClient("http://0.0.0.0:3333")
	uploader := client.NewUploader(fileHandler, serverClient)

	root, err := uploader.Upload([]string{"0", "1", "2", "3"})
	if err != nil {
		t.Fatal(err)
	}

	if err := uploader.Download(root, 0); err != nil {
		t.Fatal(err)
	}
}

func makeServer(handler http.Handler) {
	server := &http.Server{Addr: "0.0.0.0:3333", Handler: handler}
	serverCtx, serverStopCtx := context.WithCancel(context.Background())

	// Listen for syscall signals for process to interrupt/quit
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	go func() {
		<-sig

		// Shutdown signal with grace period of 30 seconds
		shutdownCtx, _ := context.WithTimeout(serverCtx, 30*time.Second)

		go func() {
			<-shutdownCtx.Done()
			if shutdownCtx.Err() == context.DeadlineExceeded {
				log.Fatal("graceful shutdown timed out.. forcing exit.")
			}
		}()

		// Trigger graceful shutdown
		err := server.Shutdown(shutdownCtx)
		if err != nil {
			log.Fatal(err)
		}
		serverStopCtx()
	}()

	// Run the server
	err := server.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}

	// Wait for server context to be stopped
	<-serverCtx.Done()
}

func cleanUp() {
	os.Rename("c478fead0c89b79540638f844c8819d9a4281763af9272c7f3968776b6052345/0", "0")
	os.Rename("c478fead0c89b79540638f844c8819d9a4281763af9272c7f3968776b6052345/1", "1")
	os.Rename("c478fead0c89b79540638f844c8819d9a4281763af9272c7f3968776b6052345/2", "2")
	os.Rename("c478fead0c89b79540638f844c8819d9a4281763af9272c7f3968776b6052345/3", "3")
	os.RemoveAll("c478fead0c89b79540638f844c8819d9a4281763af9272c7f3968776b6052345")
	os.RemoveAll("backup.json")
	os.RemoveAll("root")
}
