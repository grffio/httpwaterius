/*
Copyright (c) grffio.

This source code is licensed under the MIT license found in the
LICENSE file in the root directory of this source tree.
*/

package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/grffio/httpwaterius/internal/httpwaterius"
	"github.com/grffio/httpwaterius/internal/stat"

	"github.com/i-core/rlog"
	"github.com/i-core/routegroup"
	"github.com/kelseyhightower/envconfig"
	"go.uber.org/zap"
)

// version will be filled at compile time.
var version = ""

type config struct {
	Debug     bool                       `envconfig:"debug" default:"false" desc:"a debug mode"`
	Listen    string                     `envconfig:"listen" default:":8080" desc:"a host and port to listen on (<host>:<port>)"`
	IndexFile string                     `envconfig:"index_file" required:"true" desc:"a path to an external index.html file (</opt/index.html>)"`
	Cert      string                     `envconfig:"cert" desc:"path to SSL Certificate file (./server.crt)"`
	Key       string                     `envconfig:"key" desc:"path to SSL Certificzte key file (./server.key)"`
	Service   httpwaterius.ServiceConfig `envconfig:"service"`
}

func main() {
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage of %s:\n", os.Args[0])
		flag.PrintDefaults()
		fmt.Fprintf(flag.CommandLine.Output(), "\n")
		if err := envconfig.Usagef("httpwaterius", &config{}, flag.CommandLine.Output(), envconfig.DefaultListFormat); err != nil {
			panic(err)
		}
	}
	verFlag := flag.Bool("version", false, "print a version")
	flag.Parse()
	if *verFlag {
		fmt.Println("httpwaterius", version)
		os.Exit(0)
	}

	var cnf config
	if err := envconfig.Process("httpwaterius", &cnf); err != nil {
		log.Fatalf("Invalid configuration: %s", err)
	}

	_, err := os.Stat(cnf.IndexFile)
	if os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "Invalid configuration: %s", err)
		os.Exit(1)
	}

	var tlsEnabled bool
	if cnf.Cert != "" {
		if cnf.Key == "" {
			fmt.Fprintf(os.Stderr, "Invalid configuration: the %q file is devined but the key file is missing", cnf.Cert)
			os.Exit(1)
		}
		tlsEnabled = true
	}

	logFunc := zap.NewProduction
	if cnf.Debug {
		logFunc = zap.NewDevelopment
	}
	log, err := logFunc()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create the logger: %s\n", err)
		os.Exit(1)
	}

	router := routegroup.NewRouter(rlog.NewMiddleware(log))
	handler, err := httpwaterius.NewHandler(cnf.IndexFile, cnf.Service)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create the httpwaterius handler: %s\n", err)
		os.Exit(1)
	}
	router.AddRoutes(handler, "/")
	router.AddRoutes(stat.NewHandler(version), "/stat")

	log = log.Named("main")
	log.Info("httpwaterius started", zap.Any("config", cnf), zap.String("version", version))

	if tlsEnabled {
		log.Fatal("httpwaterius finished", zap.Error(http.ListenAndServeTLS(cnf.Listen, cnf.Cert, cnf.Key, router)))
	}
	log.Fatal("httpwaterius finished", zap.Error(http.ListenAndServe(cnf.Listen, router)))
}
