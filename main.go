package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
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
	users := list(*userdir)
	var fn func([]string, []string) []string
	if *not {
		fn = minus
	} else {
		fn = intersect
	}
	users = fn(users, flag.Args())
	for _, u := range users {
		log.Printf("user %s matches\n", u)
	}
	log.Printf("total %d match(es)\n", len(users))
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

// result order undefined
func intersect(set1 []string, set2 []string) (ss []string) {
	m1 := make(map[string]bool, len(set1))
	for _, s := range set1 {
		m1[s] = true
	}
	for _, s := range set2 {
		if m1[s] {
			ss = append(ss, s)
		}
	}
	return
}

// set1 - set2, also known as except
// result order undefined
func minus(set1 []string, set2 []string) (ss []string) {
	m1 := make(map[string]bool, len(set1))
	for _, s := range set1 {
		m1[s] = true
	}
	for _, s := range set2 {
		m1[s] = false
	}
	for k, v := range m1 {
		if v {
			ss = append(ss, k)
		}
	}
	return
}
