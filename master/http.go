package master

import (
	"errors"
	"fmt"
	"strconv"
	"net/http"
	"net/url"
	"encoding/json"

	"github.com/desperado-bvb/dortmund/util"
	"github.com/golang/glog"
)

type httpServer struct {
	Service *Server
}

func (s *httpServer) ServeHTTP(w http.ResponseWriter, req *http.Request) {

	err := s.Router(w, req)
	if err != nil {
		glog.Infof("ERROR: %s", err)
		util.ApiResponse(w, 404, "NOT_FOUND", nil)
	}
}

func (s *httpServer) Router(w http.ResponseWriter, req *http.Request) error {
	switch req.URL.Path {
	case "/stop":
		s.doStop(w, req)
	case "/query":
		s.doQuery(w, req)
	case "/start":
		s.doStart(w, req)
	case "/servers":
		s.doServer(w, req)

	default:
		return errors.New(fmt.Sprintf("404 %s", req.URL.Path))
	}

	return nil
}

func (s *httpServer) doStart(w http.ResponseWriter, req *http.Request) {
	cmd, err := s.getParamsFromQuery(req, "cmd")
	if err != nil {
		fmt.Println("1", err)
        	util.ApiResponse(w, err.(util.HTTPError).Code, err.(util.HTTPError).Text, nil)
                return
        }

	serverName, err := s.getParamsFromQuery(req, "name")
        if err != nil {
		fmt.Println("2", err)
               util.ApiResponse(w, err.(util.HTTPError).Code, err.(util.HTTPError).Text, nil)
               return
        }

	client, ok := s.Service.clients[serverName]
	if !ok {
                util.ApiResponse(w, 404, "NOT FOUND RESOURCE", nil)
		return
	}

	err =  client.StartProcess(cmd)
	if err != nil {
		util.ApiResponse(w, 200, "FAIL TO STOP PROCESS", nil)
                return
	}

	util.ApiResponse(w, 200, "ok", nil)
	
}

func (s *httpServer) doStop(w http.ResponseWriter, req *http.Request) {
        pid, err := s.getParamsFromQuery(req, "pid")
        if err != nil {
               util.ApiResponse(w, err.(util.HTTPError).Code, err.(util.HTTPError).Text, nil)
                return
        }       
        
        serverName, err := s.getParamsFromQuery(req, "name")
        if err != nil { 
               util.ApiResponse(w, err.(util.HTTPError).Code, err.(util.HTTPError).Text, nil)
                return
        }       
        
        client, ok := s.Service.clients[serverName]
        if !ok {
                util.ApiResponse(w, 404, "NOT FOUND RESOURCE", nil)
		return
        }

	i, err := strconv.Atoi(pid)
	if  err != nil {
		util.ApiResponse(w, err.(util.HTTPError).Code, err.(util.HTTPError).Text, nil)
                return
	}
        
        err =  client.StopProcess(i)
        if err != nil {
                util.ApiResponse(w, 200, "FAIL TO STOP PROCESS", nil)
                return
        }       

        util.ApiResponse(w, 200, "ok", nil)
        
}

func (s *httpServer) doQuery(w http.ResponseWriter, req *http.Request) {
	serverName, err := s.getParamsFromQuery(req, "name")
        if err != nil {
		fmt.Println(err) 
               util.ApiResponse(w, err.(util.HTTPError).Code, err.(util.HTTPError).Text, nil)
                return
        }      

        
        client, ok := s.Service.clients[serverName]

        if !ok {
                util.ApiResponse(w, 404, "NOT FOUND RESOURCE", nil)
                return
        }       
        
        detail, err :=  client.QueryProcess()
        if err != nil {
                util.ApiResponse(w, 403, "NOT FOUND RESOURCE", nil)
                return
        }       
        
        util.ApiResponse(w, 200, string(detail), nil)
}

func (s *httpServer) doServer(w http.ResponseWriter, req *http.Request) {

	var names []string
        for name, _ := range  s.Service.clients {
		names = append(names, name)
	}

	detail, err := json.Marshal(names)
        if err != nil {
        	util.ApiResponse(w, err.(util.HTTPError).Code, err.(util.HTTPError).Text, nil)
                return
	}	

        util.ApiResponse(w, 200, string(detail), nil)
}       

func (s *httpServer) getParamsFromQuery(req *http.Request, param string) (string, error) {
	reqParams, err := url.ParseQuery(req.URL.RawQuery)
	if err != nil {
		glog.Errorf("server/http: failed to parse request params - %s", err)
		return "", util.HTTPError{400, "INVALID_REQUEST"}
	}

	topicNames, ok := reqParams[param]
	if !ok {
		return "", util.HTTPError{400, "MISSING_ARG_TOPIC"}
	}
	topicName := topicNames[0]

	return topicName, nil
}
