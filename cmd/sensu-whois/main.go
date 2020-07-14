package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"
	"time"

	"github.com/danielb42/whiteflag"
)

var (
	stringToLookFor = "alive"
)

func main() {
	log.SetOutput(os.Stderr)

	whiteflag.Alias("s", "server", "sets the whois-server used to perform the check")
	whiteflag.ParseCommandLine()
	whoisServer := whiteflag.GetString("server") + ":43"

	conn, err := net.DialTimeout("tcp", whoisServer, 10*time.Second)
	if err != nil {
		log.Printf("ERROR: could not connect to %s: %s\n\n", whoisServer, err.Error())
		fmt.Printf("%s %d %d\n", "sensu.whois.available", 0, time.Now().Unix())
		fmt.Printf("%s %d %d\n", "sensu.whois.duration", 0, time.Now().Unix())

		if conn != nil {
			conn.Close()
		}

		os.Exit(2)
	}

	timeBegin := time.Now()

	_, err = conn.Write([]byte("alive@whois" + "\r\n"))
	if err != nil {
		log.Printf("ERROR: could not send data to whois: %s\n\n", err.Error())
		fmt.Printf("%s %d %d\n", "sensu.whois.available", 0, timeBegin.Unix())
		fmt.Printf("%s %d %d\n", "sensu.whois.duration", 0, timeBegin.Unix())
		_ = conn.Close()
		os.Exit(2)
	}

	buf, err := ioutil.ReadAll(conn)
	if err != nil {
		log.Printf("ERROR: could not read data from whois: %s\n\n", err.Error())
		fmt.Printf("%s %d %d\n", "sensu.whois.available", 0, timeBegin.Unix())
		fmt.Printf("%s %d %d\n", "sensu.whois.duration", 0, timeBegin.Unix())
		_ = conn.Close()
		os.Exit(2)
	}

	whoisResponseTime := time.Since(timeBegin).Milliseconds()

	if bytes.Contains(buf, []byte(stringToLookFor)) {
		log.Printf("OK: whois replied 'alive'\n\n")
		fmt.Printf("%s %d %d\n", "sensu.whois.available", 1, timeBegin.Unix())
		fmt.Printf("%s %d %d\n", "sensu.whois.duration", whoisResponseTime, timeBegin.Unix())
		_ = conn.Close()
		os.Exit(0)
	} else {
		log.Printf("ERROR: whois did not reply 'alive'\n\n")
		fmt.Printf("%s %d %d\n", "sensu.whois.available", 0, timeBegin.Unix())
		fmt.Printf("%s %d %d\n", "sensu.whois.duration", whoisResponseTime, timeBegin.Unix())
		_ = conn.Close()
		os.Exit(2)
	}
}
