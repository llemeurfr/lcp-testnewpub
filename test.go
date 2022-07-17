// Test the insertion of multiple records in an LCP License Server.
//
// Source issue from Internet Archive:
// I ran  LCP encrypting tool and it errored while notifying the LCP server.
// The error was TLS Handshake Timeout.
// The TLS handshake occurs at the start of establishing the connection with LCP License Server.
// Notify the LCP Server : Put "https://<LCP_License_server>:8989/contents/<content_id>": net/http: TLS handshake timeout
//
// There are several lcpencrypt instances running in parallel, which all send notifications to the unique LCP server.
// lcpencrypt is moving the file and just notifying the server.

package main

import (
	"bytes"
	"crypto/rand"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/google/uuid"
)

type EncryptionNotification struct {
	ContentID   string `json:"content-id"`
	ContentKey  []byte `json:"content-encryption-key"`
	StorageMode int    `json:"storage-mode"`
	Output      string `json:"protected-content-location"`
	FileName    string `json:"protected-content-disposition"`
	Size        int64  `json:"protected-content-length"`
	Checksum    string `json:"protected-content-sha256"`
	ContentType string `json:"protected-content-type,omitempty"`
}

type Problem struct {
	Type     string `json:"type,omitempty"`
	Title    string `json:"title,omitempty"`
	Status   int    `json:"status,omitempty"` //if present = http response code
	Detail   string `json:"detail,omitempty"`
	Instance string `json:"instance,omitempty"`
}

var Counter int

func main() {
	// get parameters
	var serverURL = flag.String("url", "", "LCP License Server URL")
	var tick = flag.Int("tick", 100, "Tick time in ms, 100 ms default")
	var testtime = flag.Int("testtime", 4100, "Test time in ms, 4 sc default")

	flag.Parse()

	if *serverURL == "" {
		fmt.Println("-url param is required")
		return
	}

	start := time.Now()

	ticker := time.NewTicker(time.Duration(*tick) * time.Millisecond)
	done := make(chan bool)

	go func() {
		for {
			select {
			case <-done:
				return
			case t := <-ticker.C:
				callLicenseServer(t, *serverURL)
			}
		}
	}()

	time.Sleep(time.Duration(*testtime) * time.Millisecond)
	ticker.Stop()
	done <- true

	elapsed := time.Since(start)
	fmt.Println("Process took", elapsed, "counter hit", Counter)
}

func callLicenseServer(t time.Time, serverURL string) {

	// init the file name
	filename := fmt.Sprintf("test-%s.epub", t.Format(time.RFC3339Nano))
	// init the content key
	contentKey, err := generateKey(16)
	if err != nil {
		fmt.Println("unable to generate a key")
	}
	// init the encryption notification structure
	notif := EncryptionNotification{
		ContentKey:  contentKey,
		StorageMode: 2, // already stored in a file system
		Output:      "http://edrlab.org/encrypted/" + filename,
		FileName:    filename,
		Size:        65348042,
		Checksum:    "3d2a8964075bd2064d4234ab03ec9da0f96006d8058142301d58c9c1350b6717",
		ContentType: "application/epub+zip",
	}

	// set access to the license server
	username := "laurent"
	password := "laurent"

	// set content ID
	contentID := uuid.New().String()

	// prepare the call to service/content/<id>,
	url := serverURL + "/contents/" + contentID

	jsonBody, err := json.Marshal(notif)
	if err != nil {
		fmt.Println("json marshal error", err)
		return
	}
	req, err := http.NewRequest("PUT", url, bytes.NewReader(jsonBody))
	if err != nil {
		fmt.Println("call error", err)
		return
	}
	req.SetBasicAuth(username, password)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("call error", err)
		return
	}
	if (resp.StatusCode != 302) && (resp.StatusCode/100) != 2 {
		fmt.Printf("lcp server error %d\n", resp.StatusCode)
		// details
		b, err := io.ReadAll(resp.Body)
		if err != nil {
			fmt.Println("unable to read response body")
		}
		var pb Problem
		err = json.Unmarshal(b, &pb)
		if err != nil {
			fmt.Println("unable to unmarshal response body")
		}
		fmt.Println("detail: ", pb.Detail)
	} else {
		Counter++
	}

}

func generateKey(size int) ([]byte, error) {
	k := make([]byte, size)

	_, err := rand.Read(k)
	if err != nil {
		return nil, err
	}

	return k, nil
}
