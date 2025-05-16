package main

import (
	"flag"
	"log"
	"os"
	"path"

	"github.com/ReneFernandoOM/barebones-http-go/server"
)

func main() {
	tmpDir := flag.String("directory", "/tmp/", "Set a directory to store tmp files")
	addr := flag.String("addr", "0.0.0.0:4221", "Server address")
	flag.Parse()

	ser := server.NewServer(*addr, *tmpDir, "gzip")

	ser.RegisterEndpoint("GET", "/", func(req *server.Request) server.Response {
		return server.TextResponse(200, "")
	})

	ser.RegisterEndpoint("GET", "/echo/{msg}/", func(req *server.Request) server.Response {
		echoMsg := req.Params["msg"]

		return server.TextResponse(200, echoMsg)
	})

	ser.RegisterEndpoint("GET", "/user-agent/", func(req *server.Request) server.Response {
		return server.TextResponse(200, req.Headers["user-agent"])
	})

	ser.RegisterEndpoint("GET", "/files/{fileName}/", func(req *server.Request) server.Response {
		fileName := req.Params["fileName"]
		data, err := os.ReadFile(path.Join(ser.TmpDir, fileName))
		if err != nil {
			return server.TextResponse(404, "File not found")
		}

		return server.FileResponse(200, data)
	})

	ser.RegisterEndpoint("POST", "/files/{fileName}/", func(req *server.Request) server.Response {
		fileName := req.Params["fileName"]
		absPath := path.Join(ser.TmpDir, fileName)
		err := os.WriteFile(absPath, req.Body, 0644)
		if err != nil {
			return server.TextResponse(500, "")
		}

		return server.TextResponse(201, "")
	})

	if err := ser.ListenAndServe(); err != nil {
		log.Fatalf("Server error: %v:", err)
	}
}
