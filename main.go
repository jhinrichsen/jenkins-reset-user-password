package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"
)

const (
	defaultUserdir = "${JENKINS_HOME}/users"
)

// return codes:
// 1 wrong usage
// 2 missing JENKINS_HOME environment variable
func main() {
	not := flag.Bool("not", false,
		"invert regular expressions (match -> no match)")
	userdir := flag.String("userdir", defaultUserdir,
		"Jenkins user directory")
	flag.Usage = Usage
	flag.Parse()
	if flag.NArg() == 0 {
		flag.Usage()
		os.Exit(1)
	}
	res := rexes(flag.Args())
	users := list(*userdir)
	users = filter(users, res, *not)
	for _, u := range users {
		log.Printf("user %s matches\n", u)
	}
}

func Usage() {
	fmt.Fprintf(os.Stderr, "Usage: %s <userID>...\n", os.Args[0])
	desc := "Resets password for matching user IDs to 'test'\n" +
		"user IDs may contain regexp\n"
	fmt.Fprintf(os.Stderr, desc)
	flag.PrintDefaults()
}

func die(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func list(filename string) []string {
	// optionally resolve JENKINS_HOME
	if filename == defaultUserdir {
		jh := os.Getenv("JENKINS_HOME")
		if len(jh) == 0 {
			fmt.Fprintf(os.Stderr, "missing JENKINS_HOME\n")
			os.Exit(2)
		}
		filename = filepath.Join(jh, "users")
	}
	fis, err := ioutil.ReadDir(filename)
	die(err)
	var ss []string
	for _, fi := range fis {
		ss = append(ss, fi.Name())
	}
	return ss
}

func filter(users []string, res []regexp.Regexp, inverse bool) (ss []string) {
	for _, id := range flag.Args() {
		matchesAll := true
		for _, r := range res {
			match := r.MatchString(id)
			if inverse {
				match = !match
			}
			if match {
				matchesAll = false
				break
			}
		}
		if matchesAll {
			ss = append(ss, id)
		}
	}
	return
}

func rexes(ss []string) (rs []regexp.Regexp) {
	for i, s := range ss {
		r, err := regexp.Compile(s)
		if err != nil {
			msg := "illegal regular expression at position %d: %v"
			log.Fatal(fmt.Errorf(msg, i, err))
		}
		rs = append(rs, *r)
	}
	return
}
