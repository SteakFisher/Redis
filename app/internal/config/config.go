package config

import (
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

func Init() *Config {
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
