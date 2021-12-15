package utils

import (
	"bytes"
	"encoding/base64"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"time"
)

func HttpClient() *http.Client {
	client := &http.Client{Timeout: 10 * time.Second}
	return client
}

func SendRequest(client *http.Client, method string, url string, headers map[string]string, payload []byte) []byte {
	log.Printf("[%s] Sending request to %s", method, url)

	req, err := http.NewRequest(method, url, bytes.NewBuffer(payload))

	if err != nil {
		log.Fatalf("Error Occurred. %+v", err)
	}

	for k, v := range headers {
		req.Header.Set(k, v)
	}

	response, err := client.Do(req)
	if err != nil {
		log.Fatalf("Error sending request to API endpoint. %+v", err)
	}

	// Close the connection to reuse it
	defer response.Body.Close()

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.Fatalf("Couldn't parse response body. %+v", err)
	}

	if response.StatusCode >= 300 {
		log.Fatalf("Server returns status code %d and message: \n %s", response.StatusCode, string(body))
	}

	return body
}

func ExposeEnv(key string, value string) {
	cmdLog, err := exec.Command("bitrise", "envman", "add", "--key", key, "--value", value).CombinedOutput()
	if err != nil {
		log.Printf("Failed to expose output with envman, error: %#v | output: %s", err, cmdLog)
		os.Exit(1)
	}
}

func Base64Encode(data string) string {
	encodedStr := base64.StdEncoding.EncodeToString([]byte(data))

	return string(encodedStr)
}

func Base64Decode(encodedData string) string {
	decodedData, err := base64.StdEncoding.DecodeString(encodedData)
	if err != nil {
		log.Printf("Error decoding Base64 encoded data %v", err)
	}

	return string(decodedData)
}
