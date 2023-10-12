package main

import (
	"flag"
	"fmt"
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

var (
	port   int
	target string
)

func init() {
	signal.Ignore()
	flag.IntVar(&port, "p", -1, "port to listen on")
	flag.StringVar(&target, "t", "", "target host to forward")
	cOut := colorable.NewColorableStdout()
	log.SetOutput(cOut)
	color.SetOutput(cOut)
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

func tryInteractScan[T any](a *T, inid bool, msg string, errMsg string, validators ...func(T) bool) {
	msg = msg + " " + color.Cyan(">>") + " "
	if inid {
		goto validate
	}
again:
	color.Print(msg)
	if _, err := fmt.Scanln(a); err != nil {
		goto invalidInput
	}
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
	prt := port
	tryInteractScan(&prt,
		prt != -1,
		color.Blue("Port number")+" "+color.Black("(e.g. 80)"),
		"Invalid port number. Try again.",
		func(p int) bool { return -1 < p && p < 65536 },
	)
	hst := target
	tryInteractScan(&hst,
		hst != "",
		color.Blue("Target host")+" "+color.Black("(e.g. example.com:8080)"),
		"Invalid host. Try again.",
		regexp.MustCompile(`^([a-z0-9\-._~%]+|\[[a-z0-9\-._~%!$&'()*+,;=:]+\])(:[0-9]{0,5})?$`).MatchString,
	)

	go runProxyServerPrc(prt, hst)
	color.Println(color.BlueBg(" START ", color.Blk), "request will proxy.", color.Cyan("localhost:"+strconv.Itoa(prt)), "⇒", color.Blue(hst))
}

func runProxyServerPrc(port int, forwardHost string) {
	dirct := func(r *http.Request) {
		log.Println(color.Yellow("(proxy)"), color.Green(r.Method), color.Cyan(r.URL.String()),
			"from", r.RemoteAddr)
		r.URL.Scheme = "http"
		r.URL.Host = forwardHost
	}
	rp := &httputil.ReverseProxy{Director: dirct}
	server := http.Server{
		Addr:    ":" + strconv.Itoa(port),
		Handler: rp,
	}

	if err := server.ListenAndServe(); err != nil {
		log.Fatal(err.Error())
	}
}
