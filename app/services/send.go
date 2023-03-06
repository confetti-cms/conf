package services

import (
    "bytes"
    "encoding/json"
    "io/ioutil"
    "net/http"
    "time"
)

func Send(url string, body any, method string) error {
    payloadB, err := json.Marshal(body)
    if err != nil {
        return err
    }
    payload := bytes.NewBuffer(payloadB)
    // Create request
    client := &http.Client{Timeout: 20 * time.Second}
    req, err := http.NewRequest(method, url, payload)
    if err != nil {
        return err
    }
    req.Header.Add("Accept-Language", "application/json")
    req.Header.Add("Content-Type", "application/json")
    // Do request
    res, err := client.Do(req)
    if err != nil {
        return err
    }
    defer res.Body.Close()
    // Create response
    responseBody, err := ioutil.ReadAll(res.Body)
    if err != nil {
        println("Error response: " + string(responseBody))
        return err
    }
    return nil
}