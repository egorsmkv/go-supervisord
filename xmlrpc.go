package main

import (
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/ochinchina/supervisord/process"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
)

// XMLRPC mange the XML RPC servers
// start XML RPC servers to accept the XML RPC request from client side
type XMLRPC struct {
	// all the listeners to accept the XML RPC request
	listeners map[string]net.Listener
}

type httpBasicAuth struct {
	handler  http.Handler
	user     string
	password string
}

// create a new HttpBasicAuth object with username, password and the http request handler
func newHTTPBasicAuth(user, password string, handler http.Handler) *httpBasicAuth {
	if user != "" && password != "" {
		log.Debug("require authentication")
	}
	return &httpBasicAuth{user: user, password: password, handler: handler}
}

func (h *httpBasicAuth) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if h.user == "" || h.password == "" {
		log.Debug("no auth required")
		h.handler.ServeHTTP(w, r)
		return
	}
	username, password, ok := r.BasicAuth()
	if ok && username == h.user {
		if password == h.password {
			log.Debug("Auth with normal password")
			h.handler.ServeHTTP(w, r)
			return
		}
	}
	w.Header().Set("WWW-Authenticate", "Basic realm=\"supervisor\"")
	w.WriteHeader(401)
}

// NewXMLRPC create a new XML RPC object
func NewXMLRPC() *XMLRPC {
	return &XMLRPC{listeners: make(map[string]net.Listener)}
}

// Stop network listening
func (p *XMLRPC) Stop() {
	log.Info("stop listening")
	for _, listener := range p.listeners {
		listener.Close()
	}
	p.listeners = make(map[string]net.Listener)
}

// StartUnixHTTPServer start http server on unix domain socket with path listenAddr. If both user and password are not empty, the user
// must provide user and password for basic authentication when making an XML RPC request.
func (p *XMLRPC) StartUnixHTTPServer(user, password, listenAddr string, s *Supervisor, startedCb func()) {
	os.Remove(listenAddr)
	p.startHTTPServer(user, password, "unix", listenAddr, s, startedCb)
}

// StartInetHTTPServer start http server on tcp with path listenAddr. If both user and password are not empty, the user
// must provide user and password for basic authentication when making an XML RPC request.
func (p *XMLRPC) StartInetHTTPServer(user, password, listenAddr string, s *Supervisor, startedCb func()) {
	p.startHTTPServer(user, password, "tcp", listenAddr, s, startedCb)
}

func (p *XMLRPC) isHTTPServerStartedOnProtocol(protocol string) bool {
	_, ok := p.listeners[protocol]
	return ok
}

func readFile(path string) ([]byte, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	b, err := io.ReadAll(f)
	if err != nil {
		return nil, err
	}

	return b, nil
}

func getProgramConfigPath(programName string, s *Supervisor) string {
	c := s.config.GetProgram(programName)
	if c == nil {
		return ""
	}

	res := c.GetString("conf_file", "")
	return res
}

func (p *XMLRPC) startHTTPServer(user, password, protocol, listenAddr string, s *Supervisor, startedCb func()) {
	if p.isHTTPServerStartedOnProtocol(protocol) {
		startedCb()
		return
	}
	procCollector := process.NewProcCollector(s.procMgr)
	_ = prometheus.Register(procCollector)

	mux := http.NewServeMux()

	progRestHandler := NewSupervisorRestful(s).CreateProgramHandler()
	mux.Handle("/program/", newHTTPBasicAuth(user, password, progRestHandler))

	supervisorRestHandler := NewSupervisorRestful(s).CreateSupervisorHandler()
	mux.Handle("/supervisor/", newHTTPBasicAuth(user, password, supervisorRestHandler))

	// Config file
	confHandler := NewConfAPI(s).CreateHandler()

	mux.Handle("/conf/", newHTTPBasicAuth(user, password, confHandler))

	mux.Handle("/metrics", promhttp.Handler())

	// Register log routing, you can view the log directory
	entryList := s.config.GetPrograms()
	for _, c := range entryList {
		realName := c.GetProgramName()
		if realName == "" {
			continue
		}

		filePath := c.GetString("stdout_logfile", "")
		if filePath == "" {
			continue
		}

		dir := filepath.Dir(filePath)

		log.Println(dir)

		mux.Handle("/log/"+realName+"/", http.StripPrefix("/log/"+realName+"/", http.FileServer(http.Dir(dir))))
	}

	server := &http.Server{
		Addr:              listenAddr,
		Handler:           mux,
		ReadHeaderTimeout: time.Second * 5,
	}

	ln, err := net.Listen(protocol, server.Addr)
	if err != nil {
		startedCb()

		log.WithFields(log.Fields{"addr": listenAddr, "protocol": protocol}).Error("fail to listen")
	} else {
		p.listeners[protocol] = ln

		startedCb()

		log.WithFields(log.Fields{"addr": listenAddr, "protocol": protocol}).Info("start to serve http")

		if err := server.Serve(ln); err != nil {
			log.WithFields(log.Fields{"addr": listenAddr, "protocol": protocol}).Error("fail to serve http")
		}
	}
}
