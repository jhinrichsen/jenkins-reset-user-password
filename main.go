package main

import (
	"encoding/xml"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
)

const (
	configFilename = "config.xml"
	defaultUserdir = "${JENKINS_HOME}/users"
	indentation    = "  "
	// 'test'
	defaultPassword = "#jbcrypt:$2a$10$tj0D.U.XvHRW41qwpFvSq.ivWTpBK" +
		"DzjBBNeSF3V.oDi0/E0K4B7a"
)

// return codes:
// 1 wrong usage
// 2 missing JENKINS_HOME environment variable
func main() {
	dryrun := flag.Bool("dryrun", false, "reporting only, no changes")
	not := flag.Bool("not", false,
		"invert regular expressions (match -> no match)")
	/*
		password := flag.String("password", defaultPassword,
			"new encrypted password ('test')")
	*/
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

	for _, u := range users {
		filename := filepath.Join(*userdir, u, configFilename)
		if *dryrun {
			log.Printf("dryrun: skipping %s\n", filename)
		} else {
			log.Printf("processing %s\n", filename)
			f2 := filename + ".0"
			cp(filename, f2)
			resetPassword(f2, filename)
		}
	}
}

func Usage() {
	fmt.Fprintf(os.Stderr, "Usage: %s <userID>...\n", os.Args[0])
	desc := "Resets password for matching user IDs to 'test'\n" +
		"user IDs may contain regexp\n"
	fmt.Fprintf(os.Stderr, desc)
	flag.PrintDefaults()
}

func cp(src, dst string) error {
	s, err := os.Open(src)
	if err != nil {
		return err
	}
	defer s.Close()

	d, err := os.Create(dst)
	if err != nil {
		return err
	}
	if _, err := io.Copy(d, s); err != nil {
		d.Close()
		return err
	}
	return d.Close()
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

func resetPassword(fromFilename, intoFilename string) {
	fromfile, err := os.Open(fromFilename)
	die(err)
	defer fromfile.Close()
	dec := xml.NewDecoder(fromfile)

	intofile, err := os.Create(intoFilename)
	die(err)
	defer intofile.Close()
	enc := xml.NewEncoder(intofile)
	enc.Indent("", indentation)

	triggered := false
	for {
		// Token() returns nil, io.EOF at end of input stream
		tok, err := dec.Token()
		// end of stream?
		if tok == nil && err == io.EOF {
			break
		}
		die(err)

		if triggered {
			switch tok.(type) {
			case xml.CharData:
				tok = xml.CharData(defaultPassword)
				// log.Printf("replaced with token: %+v\n",
				//	defaultPassword)
			}
			triggered = false
		}
		switch typ := tok.(type) {
		case xml.StartElement:
			if typ.Name.Local == "passwordHash" {
				// next node is our text
				triggered = true
			}
		}
		// log.Printf("writing %+v\n", tok)
		err = enc.EncodeToken(tok)
		die(err)
	}
	enc.Flush()
}
