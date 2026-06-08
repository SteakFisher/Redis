package config

import (
	"errors"
	"os"
	"strings"
)

type Config struct {
	dir            string
	appendonly     string
	appenddirname  string
	appendfilename string
	appendfsync    string
}

var config Config

func Default() *Config {
	if config == (Config{}) {
		dir, _ := os.Getwd()
		config = Config{
			dir:            dir,
			appendonly:     "no",
			appenddirname:  "appendonlydir",
			appendfilename: "appendonly.aof",
			appendfsync:    "everysec",
		}
	}

	return &config
}

func (c *Config) Init() {
	if c.appendonly == "yes" {
		if (c.dir != "") && (c.appenddirname != "") {
			dirName := c.dir + "/" + c.appenddirname

			stat, err := os.Stat(dirName)

			if errors.Is(err, os.ErrNotExist) || !stat.IsDir() {
				os.Mkdir(dirName, 0755)
			}
		}

		if c.appendfilename != "" {
			file, _ := os.Create(c.dir + "/" + c.appenddirname + "/" + c.appendfilename + ".1.incr.aof")
			defer file.Close()
		}
	}
}

func (c *Config) Get(option string) string {
	switch strings.ToLower(option) {
	case "dir":
		return c.dir

	case "appendonly":
		return c.appendonly

	case "appenddirname":
		return c.appenddirname

	case "appendfilename":
		return c.appendfilename

	case "appendfsync":
		return c.appendfsync

	default:
		return ""
	}
}

func (c *Config) Set(option string, val *string) {
	if *val == "" {
		return
	}

	switch strings.ToLower(option) {
	case "dir":
		c.dir = *val

	case "appendonly":
		c.appendonly = *val

	case "appenddirname":
		c.appenddirname = *val

	case "appendfilename":
		c.appendfilename = *val

	case "appendfsync":
		c.appendfsync = *val

	default:
	}
}
