package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/dpakach/middleware/gherkin"
	"github.com/dpakach/middleware/request"
	"github.com/dpakach/middleware/stepdef"
)

type Execute struct {
	suite gherkin.Suite
}

func VerifyMatchParams(pattern string) (error, string, []interface{}) {
	r, _ := regexp.Compile("(\\d+|\"([^\"]*)\")") 

	data := []interface{}{}

	matches :=  r.FindAllStringSubmatch(pattern, -1)

	replacedStr := pattern
	for _, match := range matches {
		replaceStr := ""
		if match[0][0] == '"' {
			replaceStr = "{{s}}"
			data = append(data, match[0][1:len(match[0]) - 1])
		} else {
			replaceStr = "{{d}}"
			i, err := strconv.Atoi(match[0])
			if err != nil {
				return errors.New("invalid args"), "", []interface{}{}
			}
			data = append(data, i)
		}
		replacedStr = strings.Replace(replacedStr, match[0], replaceStr, -1)
	}

	return nil, replacedStr, data
}

func (e *Execute) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	buf := new(strings.Builder)
    _, err := io.Copy(buf, req.Body)
    if err != nil {
        panic(err)
    }

    var reqStep struct {
		Pattern string `json:"pattern"`
		Table []map[string]string `json:"table"`
	}
    err = json.NewDecoder(strings.NewReader(buf.String())).Decode(&reqStep)

	err, pattern, data := VerifyMatchParams(reqStep.Pattern)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	step := &gherkin.Step{
		StepText: pattern,
		Data: data,
		Table: reqStep.Table,
	}
    if err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    stepDef, err := e.suite.GetMatch(*step)
    if err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }
    args := []interface{}{}

    for _, arg := range step.Data {
		i, ok := arg.(int)
		if ok {
			args = append(args, i)
		} else {
			args = append(args, arg)
		}
    }
    if step.Table != nil && len(step.Table) > 0 {
        args = append(args, step.Table)
    }
    err = stepDef.Run(args...)

	fmt.Println(err)
	if err != nil {
		w.WriteHeader(400)
		w.Write([]byte(err.Error()))
	} else {
		w.Header().Add("Content-Type", "application/json")
		json.NewEncoder(w).Encode(step)
	}
}


type startHandler struct {
	Ocs request.Ocs
}

func (h *startHandler)ServeHTTP(w http.ResponseWriter, req *http.Request) {
	fmt.Println("Start test session")
}


type cleanupHandler struct {
	Ocs request.Ocs
}

func (h *cleanupHandler)ServeHTTP(w http.ResponseWriter, req *http.Request) {
	err := h.Ocs.Cleanup()
	if err != nil {
		w.WriteHeader(400)
		w.Write([]byte(err.Error()))
	} else {
		w.WriteHeader(200)
	}
}

func main() {
	suite := gherkin.NewSuite()

	tr := &http.Transport{
		MaxIdleConns:       10,
		IdleConnTimeout:    30 * time.Second,
	}
	client := http.Client{Transport: tr}

	ocsClient := request.Ocs{
		Client: client,
		Base: "http://localhost",
	}

	// Contexts
	stepDefs := []stepdef.StepDefGroup{
		&stepdef.Provisioning{suite, ocsClient},
	}

	// Register Contexts
	for _, c := range stepDefs {
		c.Register()
	}

	// Register handlers
	sm := http.NewServeMux()

	// Execute Step Handler
	eh := &Execute{
		suite: *suite,
	}
	sm.Handle("/execute", eh)

	
	// Start Handler
	sh := &startHandler{
		Ocs: ocsClient,
	}
	sm.Handle("/start", sh)


	// Cleanup Handler
	ch := &cleanupHandler{
		Ocs: ocsClient,
	}
    sm.Handle("/exit", ch)


	// Start the server
	s := http.Server{
		Addr: ":8000",
		Handler: sm,
	}

	fmt.Println("starting server on " + s.Addr)
    err := s.ListenAndServe()
	if err != nil {
		panic(err)
	}
}
