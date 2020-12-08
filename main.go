package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"math"
	"os"
	"os/exec"
	"strings"
)

const xfconfBin = "xfconf-query"

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}

func run() error {
	in := flag.String("file", "", "json config file")
	bash := flag.Bool("bash", false, "generate a bash script")
	flag.Parse()

	if !*bash && !commandExists(xfconfBin) {
		return fmt.Errorf("%s is not found on your system", xfconfBin)
	}

	if *in == "" {
		return errors.New("file parameter not provided")
	}

	f, err := os.Open(*in)
	if err != nil {
		return err
	}

	conf, err := parseConfig(f)
	if err != nil {
		return errors.New("cannot parse configuration")
	}

	if *bash {
		return conf.toBash()
	}

	return conf.apply()
}

type channel map[string]interface{}
type config map[string]channel

func parseConfig(r io.Reader) (*config, error) {
	var conf config

	dec := json.NewDecoder(r)

	err := dec.Decode(&conf)
	if err != nil {
		return nil, err
	}

	return &conf, nil
}

func (c *config) toBash() error {
	var b strings.Builder

	if len(*c) > 0 {
		b.WriteString("#!/usr/bin/env bash")
	}

	for channel, props := range *c {
		for prop, v := range props {
			set := fmt.Sprintf("--type %q --set \"%v\"", xfconfType(v), v)

			// if array, add multiple times --type and --set for each value
			arr, isArr := v.([]interface{})
			if isArr {
				var setb strings.Builder
				for _, vv := range arr {
					fmt.Fprintf(&setb, "--type %q --set \"%v\" ", xfconfType(vv), vv)
				}
				set = strings.TrimSpace(setb.String())
			}

			// make sure property starts with slash
			prop = fmt.Sprintf("/%s", strings.Trim(prop, "/"))

			fmt.Fprintf(&b, "\n%s --channel %q --property %q --create %s",
				xfconfBin, channel, prop, set)
		}
	}

	fmt.Println(b.String())

	return nil
}

func (c *config) apply() error {
	return nil
}

func commandExists(cmd string) bool {
	_, err := exec.LookPath(cmd)
	return err == nil
}

// xfconfType returns a string containing the
// xfconf type corresponding to a go type.
//
// The list of xfconf types is: int, uint, bool,
// float, double, string.
//
// Only int, bool, double and string are used here.
//
// xfconf also has arrays, but they are not handled
// by this function.
func xfconfType(v interface{}) string {
	switch v.(type) {
	case int:
		return "int"
	case float64:
		// detect as int if it is a round number
		if v.(float64) == math.Trunc(v.(float64)) {
			return "int"
		}
		return "double"
	case string:
		return "string"
	case bool:
		return "bool"
	default:
		return ""
	}
}
