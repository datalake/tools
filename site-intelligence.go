// Copyright 2016 The Data Lake Authors. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"flag"
	"os"
	"fmt"
	"github.com/barakmich/glog"
	"runtime"
	"github.com/datalake/tools/internal/config"
	"github.com/datalake/tools/internal/db"
)

var (
	configFile         = flag.String("config", "", "Path to an explicit configuration file.")
	urlFile            = flag.String("urlFile", "", "URL file to use for list of base URLs to scan.")
	initOption         = flag.Bool("init", false, "Initialize the database before using it.")
	urlDepth           = flag.Int("depth", 1, "Depth of URL redirection to explore")
)

// Filled in by `go build -ldflags "-X main.Version=0.0.1 main.BuildDate=$(date -u '+.%Y%m%d.%H%M%S')"`.
var (
	BuildDate string = "today"
	Version   string
)

func usage() {
	fmt.Fprintln(os.Stderr, `
Usage:
  site-intelligence COMMAND [flags]

Commands:
  init      Create an empty database.
  version   Version information.

Flags:`)
	flag.PrintDefaults()
}

func init() {
	flag.Usage = usage
}

func configFrom(file string) *config.Config {
	// Find the file...
	if file != "" {
		if _, err := os.Stat(file); os.IsNotExist(err) {
			glog.Fatalln("Cannot find specified configuration file", file, ", aborting.")
		}
	} else if _, err := os.Stat(os.Getenv("DL_SITE_INTELLIGENCE_CFG")); err == nil {
		file = os.Getenv("DL_SITE_INTELLIGENCE_CFG")
	} else if _, err := os.Stat("/etc/dl_site_intelligence.cfg"); err == nil {
		file = "/etc/dl_site_intelligence.cfg"
	}
	if file == "" {
		glog.Infoln("Couldn't find a config file in either $DL_SITE_INTELLIGENCE_CFG or /etc/dl_site_intelligence.cfg. Going by flag defaults only.")
	}
	cfg, err := config.Load(file)
	if err != nil {
		glog.Fatalln(err)
	}

	if cfg.UrlFile == "" {
		cfg.UrlFile = *urlFile
	}

	if cfg.UrlDepth == 0 {
		cfg.UrlDepth = *urlDepth
	}

	return cfg
}

func main() {
	// No command? Print command usage to educate.
	if len(os.Args) == 1 {
		fmt.Fprintln(os.Stderr, "site-intelligence is a discovery tool for the deep web.")
		usage()
		os.Exit(1)
	}

	cmd := os.Args[1]
	os.Args = append(os.Args[:1], os.Args[2:]...)
	flag.Parse()

	var buildString string
	if Version != "" {
		buildString = fmt.Sprint("site-intelligence ", Version, " built ", BuildDate)
		glog.Infoln(buildString)
	}

	cfg := configFrom(*configFile)

	if os.Getenv("GOMAXPROCS") == "" {
		runtime.GOMAXPROCS(runtime.NumCPU())
		glog.Infoln("Setting GOMAXPROCS to", runtime.NumCPU())
	} else {
		glog.Infoln("GOMAXPROCS currently", os.Getenv("GOMAXPROCS"), " -- not adjusting")
	}

	var (
		err    error
	)
	switch cmd {
	case "version":
		if Version != "" {
			fmt.Println(buildString)
		} else {
			fmt.Println("Site-intelligence snapshot")
		}
		glog.Infoln("Exiting")
		os.Exit(0)
	case "init":
		err = db.Init(cfg)
		if err != nil {
			break
		}

	default:
		fmt.Println("No command", cmd)
		usage()
	}
	if err != nil {
		glog.Errorln(err)
	}
}
