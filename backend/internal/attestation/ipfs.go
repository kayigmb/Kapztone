package attestation

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"time"
)

// PinataURL is Pinata's endpoint for pinning one file to public IPFS.
const PinataURL = "https://api.pinata.cloud/pinning/pinFileToIPFS"

const pinTimeout = 30 * time.Second

// Pinata pins files to IPFS through Pinata. A pinned file is public: anyone
// with its content id can fetch it.
type Pinata struct {
	JWT string
	URL string // PinataURL, or a test server
}

// Pin uploads the content and returns its IPFS content id.
func (p Pinata) Pin(ctx context.Context, name string, content io.Reader) (string, error) {
	var body bytes.Buffer
	mw := multipart.NewWriter(&body)
	part, err := mw.CreateFormFile("file", name)
	if err != nil {
		return "", err
	}
	if _, err := io.Copy(part, content); err != nil {
		return "", err
	}
	if err := mw.Close(); err != nil {
		return "", err
	}

	ctx, cancel := context.WithTimeout(ctx, pinTimeout)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, p.URL, &body)
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", mw.FormDataContentType())
	req.Header.Set("Authorization", "Bearer "+p.JWT)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return "", fmt.Errorf("pinata answered %s", res.Status)
	}
	var out struct {
		IpfsHash string `json:"IpfsHash"`
	}
	if err := json.NewDecoder(io.LimitReader(res.Body, 1<<20)).Decode(&out); err != nil || out.IpfsHash == "" {
		return "", fmt.Errorf("pinata answer has no content id")
	}
	return out.IpfsHash, nil
}
