package main

import (
	"bufio"
	"flag"
	"log"
	"net/http"
	"net/http/httputil"
	"os"
	"os/signal"
	"regexp"
	"strconv"

	"github.com/labstack/gommon/color"
	"github.com/mattn/go-colorable"
)

type (
	interact struct {
		bufScan *bufio.Scanner
	}
)

var (
	port   string
	target string
)

func init() {
	signal.Ignore()
	flag.StringVar(&port, "p", "", "port to listen on")
	flag.StringVar(&target, "t", "", "target host to forward")
	cOut := colorable.NewColorableStdout()
	color.SetOutput(cOut)
	log.SetOutput(color.Output())
}

func main() {
	flag.Parse()

	go runPrxy()
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	status := <-quit
	color.Println()
	color.Println(color.RedBg(" KILLED ", color.Blk), color.Red("proxy server killed by "+status.String()+" signal."))
}

func (itr *interact) tryScan(a *string, inid bool, msg string, errMsg string, validators ...func(string) bool) {
	deli := color.Cyan(">>")
	if inid {
		goto validate
	}
again:
	color.Print(msg, deli, " ")
	if !itr.bufScan.Scan() {
		log.Fatal("failed to scan input.")
		return
	}
	*a = itr.bufScan.Text()
validate:
	for _, v := range validators {
		if !v(*a) {
			goto invalidInput
		}
	}
	return // success
invalidInput:
	color.Println(color.Red(errMsg))
	goto again
}

func runPrxy() {
	itr := &interact{bufio.NewScanner(os.Stdin)}
	prt := port
	itr.tryScan(&prt,
		prt != "",
		color.Blue("Port number")+color.Grey("(e.g. 8080)"),
		"Invalid port number. Try again.",
		func(p string) bool {
			num, err := strconv.Atoi(p)
			if err != nil {
				return false
			}
			return num > -1 && num < 65536
		},
	)
	hst := target
	itr.tryScan(&hst,
		hst != "",
		color.Blue("Target host")+color.Grey("(e.g. example.com:81)"),
		"Invalid host. Try again.",
		regexp.MustCompile(`^([a-z0-9\-._~%]+|\[[a-z0-9\-._~%!$&'()*+,;=:]+\])(:[0-9]{0,5})?$`).MatchString,
	)

	go runProxyServerPrc(prt, hst)
	color.Println(color.BlueBg(" START ", color.Blk), "request will proxy.", color.Cyan("localhost:"+prt), "⇒", color.Blue(hst))
}

func runProxyServerPrc(port string, forwardHost string) {
	dirct := func(r *http.Request) {
		log.Println(color.Yellow("(proxy)"), color.Green(r.Method), color.Cyan(r.URL.String()),
			"from", r.RemoteAddr)
		r.URL.Scheme = "http"
		r.URL.Host = forwardHost
	}
	rp := &httputil.ReverseProxy{Director: dirct}
	server := http.Server{
		Addr:    ":" + port,
		Handler: rp,
	}

	if err := server.ListenAndServe(); err != nil {
		log.Fatal(err.Error())
	}
}
