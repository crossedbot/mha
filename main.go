package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net"
	"net/http"
	"os"
	"strconv"

	"github.com/crossedbot/common/golang/logger"
	"github.com/crossedbot/common/golang/server"
	"github.com/crossedbot/mha/analyzer"
)

type Flags struct {
	ConfigFile string
}

func main() {
	f := flags()
	c := configuration(f.ConfigFile)
	fmt.Println(logger.SetFile(c.Logging.File))
	s := newServer(&c)
	if err := s.Start(); err != nil {
		fatal("failed to start HTTP server: %s", err.Error())
		return
	}
	select {}
}

func fatal(format string, a ...interface{}) {
	logger.Error(fmt.Errorf(format, a...))
	os.Exit(1)
}

func flags() Flags {
	config := flag.String("config-file", "", "path to configuration file")
	flag.Parse()
	return Flags{ConfigFile: *config}
}

func configuration(path string) Config {
	var c Config
	fmt.Println(Load(path, &c))
	return c
}

func newServer(c *Config) server.Server {
	s := server.New(
		net.JoinHostPort(c.Server.Host, strconv.Itoa(c.Server.Port)),
		c.Server.ReadTimeout,
		c.Server.WriteTimeout,
	)
	s.Add(analyze, http.MethodPost, "/")
	return s
}

func analyze(w http.ResponseWriter, r *http.Request, p server.Parameters) {
	type data struct {
		Email string `json:"email"`
	}
	var d data
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&d)
	if err != nil {
		server.JsonResponse(w, "failed to parse request body", http.StatusBadRequest)
		return
	}
	content, err := analyzer.Analyze([]byte(d.Email))
	if err != nil {
		server.JsonResponse(w, "failed to analyze email data", http.StatusUnprocessableEntity)
		return
	}
	server.JsonResponse(w, content, http.StatusOK)
}
