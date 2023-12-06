package main

import (
  "github.com/gorilla/mux"
  "errors"
  "io/ioutil"
  "net/http"
  "bytes"
  "encoding/json"
  "encoding/base64"
  "encoding/xml"
)

const (
  REGION = "uksouth"
  URI    = "https://" + REGION + ".tts.speech.microsoft.com/cognitiveservices/v1"
  KEY    = ""
)

// Type to handle text JSON
type JSONText struct {
  Text  string  `json:"text"`
}

// Type to handle speech JSON
type JSONSpeech struct {
  Speech  string  `json:"speech"`
}

// XML (for tts request)
type speak struct {
  Version string  `xml:"version,attr"`
  Lang    string  `xml:"xml:lang,attr"`
  Voice   voice   `xml:"voice"`
}

type voice struct {
  Voice   string  `xml:",chardata"`
  Lang    string  `xml:"xml:lang,attr"`
  Name    string  `xml:"name,attr"`
}

// Handler function
func TextToSpeech(w http.ResponseWriter, r *http.Request) {
  var text JSONText
  err := json.NewDecoder(r.Body).Decode(&text)
  if err != nil {
    http.Error(w, "error decoding JSON input", http.StatusBadRequest)
  }

  // Creating the necessary XML file for querying Azure
  xmlUnformatted := &speak{
    Version: "1.0",
    Lang: "en-US",
    Voice: voice{Voice:text.Text, Lang:"en-US", Name:"en-US-JennyNeural"},
  }
  xmlFormatted, err := xml.MarshalIndent(xmlUnformatted, " ", "  ")
  if err != nil {
    http.Error(w, "error formatting xml", http.StatusBadRequest)
  }

  speech, err := Service(xmlFormatted)
  if err != nil {
    http.Error(w, err.Error(), http.StatusInternalServerError)
  } else {
    // converts byte array to base64
    speechB64 := make([]byte, base64.StdEncoding.EncodedLen(len(speech)))
    base64.StdEncoding.Encode(speechB64, speech)

    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(JSONSpeech{string(speechB64)})
  }
}

// A function to query Microsoft Azure's text-to-speech API
func Service(xml []byte) ([]byte, error) {
  client := &http.Client{}

  req, err := http.NewRequest("POST", URI, bytes.NewBuffer(xml))
  if err != nil {
    return nil, errors.New("error creating request")
  }

  req.Header.Set("Content-Type", "application/ssml+xml")
  req.Header.Set("Ocp-Apim-Subscription-Key", KEY)
  req.Header.Set("X-Microsoft-OutputFormat", "riff-16khz-16bit-mono-pcm")

  rsp, err := client.Do(req)
  if err != nil {
    return nil, errors.New("error while making request")
  }

  defer rsp.Body.Close()

  if rsp.StatusCode == http.StatusOK {
    speech, err := ioutil.ReadAll(rsp.Body)
    if err != nil {
      return nil, errors.New("error reading response")
    }
    return speech, nil
  }
  return nil, errors.New("something bad happened")
}



func main() {
  r := mux.NewRouter()
  r.HandleFunc("/tts", TextToSpeech).Methods("POST")
  http.ListenAndServe(":3003", r)
}
