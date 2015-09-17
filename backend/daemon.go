package backend

import (
	"fmt"
	"log"
	"net/http"
	"encoding/json"
)

type backendType struct {
	Name string
	Uris []string
}

type addDataType struct {
	HostName string
	Rule string
	Backend backendType
}

func AddHostRule(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var reqData map[string]string
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&reqData)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"message" : "%s" }`, err.Error()), 500)
		return
	}
	host, okHost := reqData["hostname"]
	backend, okBackend :=reqData["backend"]
	rule, okRule := reqData["rule"]
	if !okHost || !okBackend {
		http.Error(w, `{"message" : "Unknown request" }`, 400)
		return
	}
	if !okRule {
		rule = ""
	}
	err = addHostRule(host, backend, rule)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"message" : "%s" }`, err.Error()), 500)
		return
	}
	fmt.Fprintf(w, `{"message" : "Success"}`)
}

func AddBackendSystem(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var reqData map[string]string
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&reqData)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"message" : "%s" }`, err.Error()), 500)
		return
	}
	backend, okBackend := reqData["backend"]
	hostUri, okHostUri := reqData["hosturi"]
	if !okHostUri || !okBackend {
		http.Error(w, `{"message" : "Unknown request" }`, 400)
		return
	}
	addBackendSystem(backend, hostUri)
	fmt.Fprintf(w, `{"message" : "Success"}`)
}

func AddNewProxy(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var reqData addDataType
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&reqData)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"message" : "%s" }`, err.Error()), 500)
		return
	}
	if reqData.HostName == "" ||  len(reqData.Backend.Uris) < 1 {
		http.Error(w, `{"message" : "Unknown request" }`, 400)
		return
	}
	if reqData.Backend.Name == "" {
		reqData.Backend.Name = RandStringBytes(10)
	}
	err = addHostRule(reqData.HostName, reqData.Backend.Name, reqData.Rule)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"message" : "%s" }`, err.Error()), 500)
		return
	}
	for i := range reqData.Backend.Uris {
		addBackendSystem(reqData.Backend.Name, reqData.Backend.Uris[i])
	}
	fmt.Fprintf(w, `{"message" : "Success"}`)
}

func RemoveHostRule(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var reqData map[string]string
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&reqData)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"message" : "%s" }`, err.Error()), 500)
		return
	}
	host, ok := reqData["hostname"]
	if !ok {
		http.Error(w, `{"message" : "Unknown request" }`, 400)
		return
	}
	deleteHostRule(host)
	fmt.Fprintf(w, `{"message" : "Success"}`)
}


func listen() {
	http.HandleFunc("/addnewproxy", AddNewProxy)
	http.HandleFunc("/addhostrule", AddHostRule)
	http.HandleFunc("/addbackendsystem", AddBackendSystem)
	http.HandleFunc("/removehostrule", RemoveHostRule)
	err := http.ListenAndServe(":9090", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
