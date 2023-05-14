package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/nbd-wtf/go-nostr"
)

func GetArgs() (bool, string, string, []string) {
	var sflag = flag.Bool("s", false, "Read note from stdin")
	var kflag = flag.String("k", "keys", "File name to keys")
	var rflag = flag.String("r", "relays", "File name to a list of relays")
	flag.Parse()
	return *sflag, *kflag, *rflag, flag.Args()
}

func GetKeys(kfname string) (string, string, error) {
	var sk string
	var pub string
	filep, ferr := os.Open(kfname)

	if ferr != nil {
		return "", "", ferr
	}
	defer filep.Close()

	scanner := bufio.NewScanner(filep)

	if scanner.Scan() {
		sk = scanner.Text()
	} else {
		if scanner.Err() == nil {
			return "", "", scanner.Err()
		} else {
			return "", "", io.EOF
		}
	}

	scanner.Scan()
	pub = scanner.Text()
	if scanner.Err() != nil {
		return "", "", scanner.Err()
	}

	return sk, pub, nil
}

func GetRelays(rfname string) ([]string, error) {
	filep, ferr := os.Open(rfname)

	if ferr != nil {
		return nil, ferr
	}
	defer filep.Close()

	var relays []string

	scanner := bufio.NewScanner(filep)

	for scanner.Scan() {
		relays = append(relays, scanner.Text())
	}

	if scanner.Err() != nil {
		return nil, scanner.Err()
	} else {
		return relays, nil
	}
}

func ReadNoteFromFIle(filename string) (string, error) {
	filep, err := os.Open(filename)
	if err != nil {
		return "", err
	}
	defer filep.Close()
	var retstr string = ""
	var retcount int = 0
	var ferr error
	buff := make([]byte, 1024)
	for {
		retcount, ferr = filep.Read(buff)
		if ferr == nil {
			retstr = retstr + string(buff[:retcount])
		} else {
			if ferr == io.EOF {
				return retstr, nil
			} else {
				return "", ferr
			}
		}
	}
}

func main() {
	mtags := nostr.Tags{}

	_, kfname, rfname, files := GetArgs()

	sk, pub, kerr := GetKeys(kfname)
	if kerr != nil {
		fmt.Print("Key read error: ")
		fmt.Println(kerr)
		return
	}

	relays, rerr := GetRelays(rfname)
	if rerr != nil {
		fmt.Print("Realay read error: ")
		fmt.Println("rerr")
	}

	var mmsg []string

	for _, fname := range files {
		msg, rerr := ReadNoteFromFIle(fname)
		if rerr != nil {
			fmt.Print("Io Error: ")
			fmt.Println(rerr)
			return
		}
		mmsg = append(mmsg, msg)
	}

	for _, msg := range mmsg {
		mtime := nostr.Now()
		ev := nostr.Event{
			PubKey:    pub,
			CreatedAt: mtime,
			Kind:      1,
			Tags:      mtags,
			Content:   msg,
		}
		ev.Sign(sk)
		for _, url := range relays {
			relay, e := nostr.RelayConnect(context.Background(), url)
			if e != nil {
				fmt.Println(e)
				continue
			}
			st, rer := relay.Publish(context.Background(), ev)
			fmt.Println("published to ", url, st, ev)
			fmt.Println(rer)
		}
	}
}
