package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"image"
	"image/png"
	"mime/multipart"
	"net/http"

	"github.com/pkg/errors"
)

type VRaptor struct {
	url   string
	token string
}

func newVRaptor(url string, user, pass string) (*VRaptor, error) {
	token, err := getToken(url, user, pass)
	if err != nil {
		return nil, err
	}
	return &VRaptor{
		url:   url,
		token: token,
	}, nil
}

func (v *VRaptor) ImageMode(enable bool) error {
	type Payload struct {
		Status bool `json:"status"`
	}

	data := Payload{
		Status: enable,
	}
	payloadBytes, err := json.Marshal(data)
	if err != nil {
		return errors.Wrap(err, "failed to set oled image mode")
	}
	body := bytes.NewReader(payloadBytes)

	req, err := http.NewRequest("PUT", v.url+"/display/image", body)
	if err != nil {
		return errors.Wrap(err, "failed to set oled image mode")
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", v.token))
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return errors.Wrap(err, "failed to set oled image mode")
	}
	defer resp.Body.Close()

	return nil
}

func (v *VRaptor) SetImage(img image.Image) error {
	bounds := img.Bounds()
	if bounds.Dx() != 256 || bounds.Dy() != 64 {
		return fmt.Errorf("image size should be 256x64, but got %dx%d", bounds.Dx(), bounds.Dy())
	}

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", "image") // image.png
	if err != nil {
		return errors.Wrap(err, "fail to set image")
	}

	if err := png.Encode(part, img); err != nil {
		return errors.Wrap(err, "fail to set image")
	}

	// io.Copy(part, ditherImg)
	writer.Close()

	req, err := http.NewRequest("PUT", v.url+"/display/image/file", body)
	if err != nil {
		return errors.Wrap(err, "fail to set image")
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", v.token))
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return errors.Wrap(err, "fail to set image")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusAccepted {
		buf := &bytes.Buffer{}
		_, err := buf.ReadFrom(resp.Body)
		if err != nil {
			return errors.Wrap(err, "fail to set image")
		}
		return fmt.Errorf("fail to set image: %s", buf.String())
	}

	return nil
}

// ---

func getToken(url string, user, pass string) (string, error) {
	data := map[string]interface{}{
		"username": user,
		"password": pass,
	}
	payloadBytes, err := json.Marshal(data)
	if err != nil {
		return "", errors.Wrap(err, "failed to get token")
	}
	body := bytes.NewReader(payloadBytes)

	req, err := http.NewRequest("POST", url+"/login", body)
	if err != nil {
		return "", errors.Wrap(err, "failed to get token")
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", errors.Wrap(err, "failed to get token")
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", errors.Wrap(err, "failed to get token")
	}

	if token, ok := result["access_token"]; ok {
		return token.(string), nil
	}

	return "", fmt.Errorf("failed to get token: %v", result)
}
