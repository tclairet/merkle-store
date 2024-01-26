package tests

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"testing"
	"time"

	"github.com/tclairet/merklestore/client"
	"github.com/tclairet/merklestore/files"
	"github.com/tclairet/merklestore/server"
)

func TestE2E(t *testing.T) {
	t.Cleanup(cleanUp)

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
	tests := []struct {
		nbInputs int
	}{
		{1},
		{5},
		{50},
	}
	for _, tt := range tests {
		t.Run(fmt.Sprintf("%d", tt.nbInputs), func(t *testing.T) {
			var inputs []string
			for i := 0; i < tt.nbInputs; i++ {
				if err := fileHandler.Save(strconv.Itoa(i), bytes.NewBuffer([]byte(strconv.Itoa(i)))); err != nil {
					t.Fatal(err)
				}
				inputs = append(inputs, strconv.Itoa(i))
			}

			root, err := uploader.Upload(inputs)
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() {
				os.RemoveAll(root)
			})

			for i := 0; i < tt.nbInputs; i++ {
				if err := uploader.Download(root, i); err != nil {
					t.Fatal(err)
				}
			}
		})
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
	os.RemoveAll("backup.json")
	os.RemoveAll("root.json")
}
