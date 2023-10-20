package main

import (
  "encoding/xml"
  "fmt"
  "io"
  "net/http"
  "net/url"
  "time"
  "os"
)

type MediaContainer struct {
  MachineIdentifier string   `xml:"machineIdentifier,attr"`
}

func getHeaders(clientId string) (headers map[string]string) {
  name := os.Getenv("APP_NAME")
  description := os.Getenv("APP_DESCRIPTION")

  return map[string]string{
    "X-Plex-Client-Identifier": clientId,
    "X-Plex-Product": "Plex OAuth",
    "X-Plex-Version": description,
    "X-Plex-Platform": "Plex",
    "X-Plex-Platform-Version": "1.0",
    "X-Plex-Device": "github.com/SushyDev",
    "X-Plex-Device-Name": name,
    "X-Plex-Model": "X-Plex-Model",
    "X-Plex-Device-Screen-Resolution": "640x480",
    "X-Plex-Layout": "desktop",
    "X-Plex-Language": "en",
  }
}

func getQuery(clientId string, code string) (query map[string]string) {
  name := os.Getenv("APP_NAME")
  description := os.Getenv("APP_DESCRIPTION")

  return map[string]string{
    "clientID": clientId,
    "context[device][product]": "Plex OAuth",
    "context[device][version]": description,
    "context[device][platform]": "Plex",
    "context[device][platformVersion]": "1.0",
    "context[device][device]": "github.com/SushyDev",
    "context[device][deviceName]": name,
    "context[device][model]": "X-Plex-Model",
    "context[device][screenResolution]": "640x480",
    "code": code,
  }
}

func getIdentifier() (identifier string, err error) {
  u := "http://localhost:32400/identity"

  // Create an HTTP client
  client := &http.Client{}

  // Send a GET request
  req, err := http.NewRequest("GET", u, nil)
  if err != nil {
    fmt.Println("Error creating GET request:", err)
    return "", err
  }

  // Perform the request
  resp, err := client.Do(req)
  if err != nil {
    fmt.Println("Error performing GET request:", err)
    return "", err
  }
  defer resp.Body.Close()

  // Read the response body
  body, err := io.ReadAll(resp.Body)
  if err != nil {
    fmt.Println("Error reading response body:", err)
    return "", err
  }

  // Parse the XML response
  var mediaContainer MediaContainer
  err = xml.Unmarshal(body, &mediaContainer)
  if err != nil {
    fmt.Println("Error parsing XML response:", err)
    return "", err
  }

  return mediaContainer.MachineIdentifier, nil
}

type Pin struct {
  Id string `xml:"id,attr"`
  Code string `xml:"code,attr"`
  AuthToken string `xml:"authToken,attr"`
}

func getPins(clientId string) (id string, code string, err error) {
  u := "https://plex.tv/api/v2/pins?strong=true"

  // Create an HTTP client
  client := &http.Client{}

  // Send a POST request
  req, err := http.NewRequest("POST", u, nil)
  if err != nil {
    fmt.Println("Error creating POST request:", err)
    return "", "", err
  }

  // Set the request headers
  headers := getHeaders(clientId)

  // for each data set header
  for key, value := range headers {
    req.Header.Set(key, fmt.Sprint(value))
  }


  // Perform the request
  resp, err := client.Do(req)
  if err != nil {
    fmt.Println("Error performing POST request:", err)
    return "", "", err
  }

  // Read the response body
  body, err := io.ReadAll(resp.Body)
  if err != nil {
    fmt.Println("Error reading response body:", err)
    return "", "", err
  }

  // Parse the XML response
  var pin Pin
  err = xml.Unmarshal(body, &pin)
  if err != nil {
    fmt.Println("Error parsing XML response:", err)
    return "", "", err
  }

  return pin.Id, pin.Code, nil
}

func buildOAuthUrl(clientId string, code string) (u string, err error) {
  query := getQuery(clientId, code)

  // Convert the data to a URL-encoded query string
  params := url.Values{}
  for key, value := range query {
    params.Add(key, fmt.Sprint(value))
  }

  u = "https://app.plex.tv/auth/#!?" + params.Encode()

  return u, nil
}

func pollForToken(clientId string, id string, code string) (token string, err error) {
  for {
    u := "https://plex.tv/api/v2/pins/" + id;

    // Create an HTTP client
    client := &http.Client{}

    // Send a GET request
    req, err := http.NewRequest("GET", u, nil)

    headers := getHeaders(clientId)

    // for each data set header
    for key, value := range headers {
      req.Header.Set(key, fmt.Sprint(value))
    }

    // Perform the request
    resp, err := client.Do(req)
    if err != nil {
      fmt.Println("Error performing GET request:", err)
      return "", err
    }

    // Read the response body
    body, err := io.ReadAll(resp.Body)
    if err != nil {
      fmt.Println("Error reading response body:", err)
      return "", err
    }

    // Parse the XML response
    var pin Pin
    err = xml.Unmarshal(body, &pin)
    if err != nil {
      fmt.Println("Error parsing XML response:", err)
      return "", err
    }

    if pin.AuthToken != "" {
      return pin.AuthToken, nil
    }

    time.Sleep(1 * time.Second)
    pollForToken(clientId, id, code)
  }
}

func main() {
  // get machine identifier
  clientId, err := getIdentifier()
  if err != nil {
    fmt.Println("Error getting machine identifier:", err)
    return
  }

  // get pins
  id, code, err := getPins(clientId)
  if err != nil {
    fmt.Println("Error getting pins:", err)
    return
  }

  // build oauth url
  u, err := buildOAuthUrl(clientId, code)
  if err != nil {
    fmt.Println("Error building oauth url:", err)
    return
  }

  fmt.Println(u)

  // poll for token
  token, err := pollForToken(clientId, id, code)
  if err != nil {
    fmt.Println("Error polling for token:", err)
    return
  }

  fmt.Println("Token:")
  fmt.Println(token)
}
