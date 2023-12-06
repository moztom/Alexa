package main

import (
  "github.com/gorilla/mux"
  "net/http"
  "io"
  "encoding/json"
  "errors"
  "bytes"
)

const (
  uri_stt = "http://localhost:3002/stt"
  uri_alpha = "http://localhost:3001/alpha"
  uri_tts = "http://localhost:3003/tts"
)

// Type to handle text JSON
type JSONText struct {
  Text    string  `json:"text"`
}

// Type to handle speech JSON
type JSONSpeech struct {
  Speech  string  `json:"speech"`
}

// Handler function
func Alexa(w http.ResponseWriter, r *http.Request) {
  // Send the input speech to the STT API, and obtain text
  text, err := SpeechToText(r.Body)
  if err != nil {
    http.Error(w, err.Error(), http.StatusInternalServerError)
  }

  // Encode question string in JSON
  question, err := json.Marshal(JSONText{text})
  if err != nil {
    http.Error(w, "error encoding JSON (question text)", http.StatusBadRequest)
  }

  // Send encoded question to the Alpha API, and obtain an answer
  answer, err := Alpha(question)
  if err != nil {
    http.Error(w, err.Error(), http.StatusInternalServerError)
  }

  // Encode answer string in JSON
  ans, err := json.Marshal(JSONText{answer})
  if err != nil {
    http.Error(w, "error encoding JSON (answer text)", http.StatusBadRequest)
  }

  // Send encoded answer to the TTS API, and obtain speech (in base64)
  speech, err := TextToSpeech(ans)
  if err != nil {
    http.Error(w, err.Error(), http.StatusInternalServerError)
  }

  json.NewEncoder( w ).Encode( JSONSpeech{speech} )
}

func SpeechToText( speech io.ReadCloser ) ( string, error ) {
  client := &http.Client{}

  req, err := http.NewRequest( "POST", uri_stt, speech )
  if err != nil {
    return "", errors.New("error while creating speech-to-text request")
  }

  rsp, err := client.Do(req)
  if err != nil {
    return "", errors.New("error while making speech-to-text request")
  }

  defer rsp.Body.Close()

  var text JSONText

  if rsp.StatusCode == http.StatusOK {
    err := json.NewDecoder(rsp.Body).Decode(&text)
    if err != nil {
      return "", errors.New("error while decoding speech-to-text JSON response")
    }
  }

  return text.Text , nil
}

func Alpha( question []byte ) ( string, error ) {
  client := &http.Client{}

  req, err := http.NewRequest( "POST", uri_alpha, bytes.NewReader(question) )
  if err != nil {
    return "", errors.New("error while creating alpha request")
  }

  rsp, err := client.Do(req)
  if err != nil {
    return "", errors.New("error while making alpha request")
  }

  defer rsp.Body.Close()

  var answer JSONText

  if rsp.StatusCode == http.StatusOK {
    err := json.NewDecoder(rsp.Body).Decode(&answer)
    if err != nil {
      return "", errors.New("error while decoding alpha JSON response")
    }
  }

  return answer.Text, nil
}

func TextToSpeech( answer []byte ) ( string, error ) {
  client := &http.Client{}

  req, err := http.NewRequest( "POST", uri_tts, bytes.NewReader(answer) )
  if err != nil {
    return "", errors.New("error while creating alpha request")
  }

  rsp, err := client.Do(req)
  if err != nil {
    return "", errors.New("error while making alpha request")
  }

  defer rsp.Body.Close()

  var speech JSONSpeech

  if rsp.StatusCode == http.StatusOK {
    err := json.NewDecoder(rsp.Body).Decode(&speech)
    if err != nil {
      return "", errors.New("error while decoding alpha JSON response")
    }
  }

  return speech.Speech, nil
}

func main() {
  r := mux.NewRouter()
  r.HandleFunc( "/alexa", Alexa ).Methods( "POST" )
  http.ListenAndServe( ":3000", r )
}
