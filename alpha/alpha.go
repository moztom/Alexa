package main

import (
  "encoding/json"
  "errors"
  "net/http"
  "github.com/gorilla/mux"
  "strings"
  "io"
)

const (
  URI = "http://api.wolframalpha.com/v1/result"
  APPID = ""
)

// Type to handle text JSON
type JSONText struct {
  Text    string  `json:"text"`
}

// Handler function
func Alpha(w http.ResponseWriter, r *http.Request) {
  var question JSONText
  err := json.NewDecoder(r.Body).Decode(&question)
  if err != nil {
    http.Error(w, "error decoding JSON input", http.StatusBadRequest)
  }

  // Query the Wolfram Alpha API
  answer, err := Service(question.Text)
  if err != nil {
    http.Error(w, err.Error(), http.StatusInternalServerError)
  } else {
    w.WriteHeader(http.StatusOK)
    json.NewEncoder( w ).Encode(JSONText{answer})
  }
}

// A function to query Wolfram Alpha's computational knowledge engine API
func Service(question string) (string, error) {
  client := &http.Client{}

  q := strings.Replace(question, " ", "+", 100)
  uri := URI + "?appid=" + APPID + "&i=" + q

  req, err := http.NewRequest("GET", uri, nil)
  if err != nil {
    return "", errors.New("error creating Wolfram request")
  }

  rsp, err := client.Do(req)
  if err != nil {
    return "", errors.New("error while making Wolfram request")
  }

  //defer rsp.Body.Close()

  if rsp.StatusCode == http.StatusOK {
    answer, err := io.ReadAll(rsp.Body)
    if err != nil {
      return "", errors.New("error reading response")
    }
    return string(answer), nil
  }
  return "", errors.New("something bad happened")
}

func main() {
  r := mux.NewRouter()
  r.HandleFunc("/alpha", Alpha).Methods("POST")
  http.ListenAndServe(":3001", r)
}
