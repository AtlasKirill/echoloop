package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"sync"
	"time"
	"bytes"
)

var (
	str     pattern
	port    int
	url     string
	fullUrl string
)

type Data struct {
	Text   string `json:"pattern"`
}

type pattern struct {
	sync.Mutex
	patterns bytes.Buffer
}

func (t *pattern) Append(text string) {
	t.Lock()
	t.patterns.WriteString(text + " ")
	t.Unlock()
}

func userInput() {
	patternPtr := flag.String("str", "test", "repeated pattern")
	portPtr := flag.Int("port", 8080, "port to listen")
	urlPtr := flag.String("url", "http://localhost", "url to send text")
	flag.Parse()

	str.Append(*patternPtr)
	port = *portPtr
	url = *urlPtr

	fullUrl = fmt.Sprintf("%s:%d", url, port)

}

func echoloop() {
	time.Sleep(time.Second) //waiting for starting server --- dirty hack
	ticker := time.Tick(time.Second)
	for range ticker {
		str.Lock()
		val := str.patterns.String()
		str.Unlock()

		fmt.Println(val)
	}

}

func recieveDataByServer(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var data Data
	err := decoder.Decode(&data)
	if err != nil {
		panic(err)
	}
	str.Append(data.Text)
}

func sentDataToServer(str string, url string) {
	jsonMap := make(map[string]string)
	jsonMap["pattern"] = str

	jsonStr, _ := json.Marshal(jsonMap)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonStr))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
}

func main() {
	userInput()
	go echoloop()

	http.HandleFunc("/", recieveDataByServer)

	err := http.ListenAndServe(fmt.Sprintf(":%d", port), nil)

	if err != nil {
		fmt.Println("Server has already existed, its echo's value was modified")
		sentDataToServer(str.patterns.String(), fullUrl)
	}
}
