package backend

import (
	"fmt"
	"log"
	"sync"
	"time"
	"errors"
	"strings"
	"net/http"
	"io/ioutil"
	"math/rand"
	"encoding/json"
)

type HostRule struct {
	Rule string
	Backend string
}

type ConfigData struct {
	HostRules map[string]HostRule
	BackendStruct map[string][]string
}

type Backends struct {
	sync.RWMutex
	HostRules map[string]HostRule
	BackendStruct map[string][]string
} 

var (
	x *Backends
)

func random(min, max int) int {
	rand.Seed(time.Now().Unix())
	return rand.Intn(max - min) + min
}

func RandStringBytes(n int) string {
	const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}

func GetMostMatchString(list []string, keyword string) string {
	var tempList []string
	for i := range list {
		if strings.HasPrefix(keyword, list[i]) {
			tempList = append(tempList, list[i])
		}
	}
	if len(tempList) <1 {
		return ""
	}
	mostMatch := ""
	currentSize  := 0
	for i := range tempList {
		if currentSize < len(tempList[i]) {
			mostMatch = tempList[i]
		}
	}
	return mostMatch
}

func Initialize() {
	x = &Backends{HostRules : make(map[string]HostRule), BackendStruct : make(map[string][]string)}
	configData, err := ioutil.ReadFile("./myconfig.json")
	if err != nil {
		log.Printf("File error: %v\n", err)
	} else {
		err = json.Unmarshal(configData, x)
		if err != nil {
			log.Printf("Unable to load config: %v\n", err)
		}
	}
	go listen()
	go configSaver("./myconfig.json")
}

func configSaver(path string) {
	for {
		time.Sleep(time.Second * 10)
		func() {
			x.RLock()
			defer x.RUnlock()
			config := ConfigData{ HostRules : x.HostRules, BackendStruct : x.BackendStruct}
			data, err := json.Marshal(config)
			if err != nil {
				log.Println(err.Error())
				return
			}
			ioutil.WriteFile(path, data, 0644)
		}()
	}
	
}

func addHostRule(host, backend, rule string) error {
	x.Lock()
	defer x.Unlock() 
	if _, ok := x.HostRules[host]; ok {
		log.Println(fmt.Sprintf("HostAdd Error: Host[%s] entry already exists, skipping", host))
		return errors.New(fmt.Sprintf("HostAdd Error: Host[%s] entry already exists, skipping", host))
	}
	x.HostRules[host] = HostRule{Rule : rule, Backend : backend} 
	return nil
}

func updateHostRule(host, newBackend, rule string) error {
	x.Lock()
	defer x.Unlock() 
	if _, ok := x.HostRules[host]; !ok {
		log.Println(fmt.Sprintf("HostUpdate Error: Host[%s] entry does not exist, skipping", host))
		return errors.New(fmt.Sprintf("HostUpdate Error: Host[%s] entry does not exist, skipping", host))
	}
	x.HostRules[host] = HostRule{Rule : rule, Backend : newBackend}
	return nil
}

func deleteHostRule(host string) {
	x.Lock()
	defer x.Unlock()
	delete(x.HostRules, host)
}

func cleanUpRule(host string) {
	x.Lock()
	defer x.Unlock()
	_, ok := x.HostRules[host]
	if !ok {
		return
	}
	removeBackend(x.HostRules[host].Backend)
	deleteHostRule(host)
}

func getHostBackend(host string) string {
	x.RLock()
	defer x.RUnlock()
	keys := make([]string, 0, len(x.HostRules))
	for k := range x.HostRules {
		keys = append(keys, k)
	}
	mostMatch := GetMostMatchString(keys, host)
	if mostMatch == "" {
		return "default"
	}
	if x.HostRules[mostMatch].Rule == "pathbeg" {
		return x.HostRules[mostMatch].Backend
	} else if mostMatch == host {
		return x.HostRules[mostMatch].Backend
	}
	return "default"
}

func addBackendSystem(backend, hostUri string) {
	x.Lock()
	defer x.Unlock()
	x.BackendStruct[backend] = append(x.BackendStruct[backend], hostUri)
}

func removeBackendSystem(backend, hostUri string) {
	x.Lock()
	defer x.Unlock()
	var tempBackends []string
	if _, ok := x.BackendStruct[backend]; !ok {
		log.Println(fmt.Sprintf("removeBackendSystem Error: Host[%s] entry does not exist, skipping", backend))
		return
	}
	for i := range x.BackendStruct[backend] {
		if hostUri != x.BackendStruct[backend][i] {
			tempBackends = append(tempBackends, x.BackendStruct[backend][i])
		}
	}
	x.BackendStruct[backend] = tempBackends
}

func removeBackend(backend string) {
	x.Lock()
	defer x.Unlock()
	delete(x.BackendStruct, backend)
}

func getBackendSystems(backend string) []string {
	x.RLock()
	defer x.RUnlock()
	return x.BackendStruct[backend]
}

func GetTarget(r *http.Request) string {
	backend := getHostBackend(r.Host)
	allBackends := getBackendSystems(backend)
	if len(allBackends) == 0 {
		return ""
	}
	ranNo := random(0, len(allBackends))
	return allBackends[ranNo]
}

