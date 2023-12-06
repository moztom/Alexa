package main

import (
  "github.com/gorilla/mux"
  "errors"
  "net/http"
  "bytes"
  "encoding/json"
  "encoding/base64"
)

const (
  REGION = "uksouth"
  URI    = "https://" + REGION + ".stt.speech.microsoft.com/" +
          "speech/recognition/conversation/cognitiveservices/v1?" +
          "language=en-US"
  KEY    = ""
)

// Type to handle Azure response JSON
type JSONAzure struct {
  DT  string  `json:"DisplayText"`
}

// Type to handle speech JSON
type JSONSpeech struct {
  Speech  string  `json:"speech"`
}

// Handler function
func SpeechToText(w http.ResponseWriter, r *http.Request) {
  var speechInput JSONSpeech
  err := json.NewDecoder(r.Body).Decode(&speechInput)
  if err != nil {
    http.Error(w, "error decoding JSON input", http.StatusBadRequest)
  }

  // convert base64 into a byte array
  s := make([]byte, base64.StdEncoding.DecodedLen(len(speechInput.Speech)))
  n, err := base64.StdEncoding.Decode(s, []byte(speechInput.Speech))
  if err != nil {
    http.Error(w, "error converting base64", http.StatusBadRequest)
  }
  speech := s[:n]

  text, err := Service(speech)
  if err != nil {
    http.Error(w, err.Error(), http.StatusInternalServerError)
  } else {
    w.WriteHeader(http.StatusOK)
    output := map[string] string {"text":text}
    json.NewEncoder(w).Encode(output)
  }

}

// A function to query Microsoft Azure's speech-to-text API
func Service(speech []byte) (string, error) {
  client := &http.Client{}

  req, err := http.NewRequest("POST", URI, bytes.NewReader(speech))
  if err != nil {
    return "", errors.New("error creating request")
  }

  req.Header.Set("Content-Type", "audio/wav;codecs=audio/pcm;samplerate=16000")
  req.Header.Set("Ocp-Apim-Subscription-Key", KEY)

  rsp, err := client.Do(req)
  if err != nil {
    return "", errors.New("error performing request")
  }

  defer rsp.Body.Close()

  var text JSONAzure
  if rsp.StatusCode == http.StatusOK {
    err := json.NewDecoder(rsp.Body).Decode(&text)
    if err != nil {
      return "", errors.New("error decoding response")
    }
  }
  return text.DT, nil
}

func main() {
  r := mux.NewRouter()
  r.HandleFunc("/stt", SpeechToText).Methods("POST")
  http.ListenAndServe(":3002", r)
}
