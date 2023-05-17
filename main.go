package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/nbd-wtf/go-nostr"
	"gopkg.in/yaml.v3"
)

func GetArgs() (bool, string, []string) {
	var sflag = flag.Bool("s", false, "Read note from stdin")
	var cflag = flag.String("c", "config.yaml", "File name to a list of relays")
	flag.Parse()
	return *sflag, *cflag, flag.Args()
}

func GetConfig(ConfigFileName string) (map[string]interface{}, error) {
	fp, ferr := os.Open(ConfigFileName)
	if ferr != nil {
		return nil, ferr
	}
	decoder := yaml.NewDecoder(fp)
	var m map[string]interface{}
	derr := decoder.Decode(&m)
	if derr != nil {
		return nil, derr
	}
	return m, nil
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

	_, cfname, files := GetArgs()

	mcfg, cerr := GetConfig(cfname)
	if cerr != nil {
		fmt.Print("Config file read error: ")
		fmt.Println(cerr)
		return
	}

	sk := mcfg["sk"].(string)
	pub := mcfg["pk"].(string)

	var relays []string
	rls := mcfg["relays"].([]interface{})
	for _, rl := range rls {
		relays = append(relays, rl.(string))
	}

	mtags := nostr.Tags{}
	emjs := mcfg["emojis"].([]interface{})
	for i := range emjs {
		emj := emjs[i].([]interface{})
		mtags = append(mtags, nostr.Tag{"emoji", emj[0].(string), emj[1].(string)})
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
