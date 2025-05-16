package server

import (
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"regexp"
)

type Handler = func(req *Request) Response

type Endpoint struct {
	method  string
	path    string
	handler Handler
}

type route struct {
	method  string
	pattern *regexp.Regexp
	handler Handler
}

type Router struct {
	routes   []route
	notFound Handler
}

type Server struct {
	TmpDir            string
	Address           string
	EncodingManager   *EncodeManager
	Router            *Router
	SupportedEncoding string
}

func NewRouter() *Router {
	return &Router{
		routes: make([]route, 0),
		notFound: func(req *Request) Response {
			return Response{
				StatusCode: 404,
				StatusMsg:  statusText(404),
			}
		},
	}
}

func NewServer(address string, tmpDir string, supportedEncoding string) Server {
	return Server{
		TmpDir:            tmpDir,
		Address:           address,
		EncodingManager:   NewEncodeManager(),
		Router:            NewRouter(),
		SupportedEncoding: supportedEncoding,
	}
}

func (r *Router) Match(req *Request) Handler {
	for _, route := range r.routes {
		if req.Method != route.method {
			continue
		}

		matches := route.pattern.FindStringSubmatch(req.URI)
		if matches == nil {
			continue
		}

		// uri matched!, so we extract args for uri
		urlVars := route.pattern.SubexpNames()
		for i, varName := range urlVars {
			req.Params[varName] = matches[i]
		}

		return route.handler
	}
	return r.notFound
}

func (s Server) serveRequest(conn net.Conn) {
	defer conn.Close()

	keepAlive := true
	for keepAlive {
		req, err := parseRequest(conn)

		// client closed connection without warning
		if err == io.EOF {
			return
		}

		if err != nil {
			fmt.Printf("Error parsing request: %v", err)
			writeResponse(conn, Response{
				StatusCode: 400,
				StatusMsg:  statusText(400),
			})
			return
		}

		handler := s.Router.Match(req)
		response := handler(req)

		if err := s.EncodingManager.ApplyEncoding(req, &response); err != nil {
			log.Printf("Error applying encoding: %v", err)
		}

		if req.Headers["connection"] == "close" {
			response.Headers["Connection"] = "close"
			keepAlive = !keepAlive
		}

		writeResponse(conn, response)
	}
}

func (s Server) ListenAndServe() error {
	var err error
	l, err := net.Listen("tcp", s.Address)

	if err != nil {
		return fmt.Errorf("failed to listen on %s: %v", s.Address, err)
	}
	defer l.Close()

	fmt.Printf("Server started in address: %s and its waiting for new connections\n", s.Address)

	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
			os.Exit(1)
		}
		go func() {
			s.serveRequest(conn)
		}()
	}
}

func (s *Server) RegisterEndpoint(method string, path string, handler Handler) {
	// match anything in curly braces like {param}
	re := regexp.MustCompile(`\{(\w+)\}`)
	// replace with named capture group: (?P<param>\w+)
	regexPath := re.ReplaceAllString(path, `(?P<$1>\w+)`)
	regexPath = "^" + regexPath + "$"

	regexCompiled, err := regexp.Compile(regexPath)
	if err != nil {
		log.Fatalf("Invalid route pattern %q: %v", path, err)
	}

	newRoute := route{
		method:  method,
		pattern: regexCompiled,
		handler: handler,
	}

	s.Router.routes = append(s.Router.routes, newRoute)
}

func writeResponse(conn io.Writer, resp Response) error {
	fmt.Printf("response: %q\n", resp.String())
	_, err := conn.Write([]byte(resp.String()))
	return err
}
