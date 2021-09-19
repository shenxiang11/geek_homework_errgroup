package main

import (
	"context"
	"fmt"
	"golang.org/x/sync/errgroup"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	eg, _ := errgroup.WithContext(context.Background())

	var server1 http.Server
	var server2 http.Server

	quit := make(chan os.Signal)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL)

	eg.Go(func() error {
		mux := http.NewServeMux()
		mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("server 1"))
		})
		server1 = http.Server{
			Addr:    ":20001",
			Handler: mux,
		}
		return server1.ListenAndServe()
	})

	eg.Go(func() error {
		mux := http.NewServeMux()
		mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("server 2"))
		})
		mux.HandleFunc("/shutdown", func(w http.ResponseWriter, r *http.Request) {
			quit <- syscall.SIGKILL
		})
		server2 = http.Server{
			Addr:    ":20002",
			Handler: mux,
		}
		return server2.ListenAndServe()
	})

	for s := range quit {
		switch s {
		case syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL:
			fmt.Println("退出")
			server1.Shutdown(context.Background())
			server2.Shutdown(context.Background())
			fmt.Println("退出完毕")
			os.Exit(0)
		}
	}

	if err := eg.Wait(); err != nil {
		fmt.Println("Error occured:", err)
	}

}
