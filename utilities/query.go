package utilities

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"pokebloat/data"

	"github.com/yyewolf/arikawa/v3/discord"
)

var c = &http.Client{}

func DownloadImage(url string) (io.Reader, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}

	return resp.Body, nil
}

func PredictImage(bdy io.Reader, ext string) ([]byte, error) {
	// Send to the API
	url := fmt.Sprintf("%s/models/%s/predict?k=3", os.Getenv("API_URL"), "pokemons")
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("upload", "image."+ext)
	if err != nil {
		return nil, err
	}
	io.Copy(part, bdy)
	writer.Close()

	r, err := http.NewRequest("POST", url, body)
	if err != nil {
		return nil, err
	}
	r.Header.Add("Content-Type", writer.FormDataContentType())
	rep, err := c.Do(r)
	if err != nil {
		return nil, err
	}
	resp, err := io.ReadAll(rep.Body)
	if err != nil {
		return nil, err
	}
	rep.Body.Close()

	return resp, nil
}

func RemoveBgFromImage(bdy io.Reader, ext string) (io.ReadCloser, error) {
	// Send to the API for transformation
	url := fmt.Sprintf("%s/transforms/%s/background_removal", os.Getenv("SECONDARY_API_URL"), "pokemons")
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("upload", "image."+ext)
	if err != nil {
		return nil, err
	}
	io.Copy(part, bdy)
	writer.Close()

	r, err := http.NewRequest("POST", url, body)
	if err != nil {
		return nil, err
	}
	r.Header.Set("Content-Type", writer.FormDataContentType())
	client := &http.Client{}
	rep, err := client.Do(r)
	if err != nil {
		return nil, err
	}
	return rep.Body, nil
}

func PerfectCropImage(bdy io.Reader, ext string) (io.ReadCloser, error) {
	// Send to the API for transformation
	url := fmt.Sprintf("%s/transforms/%s/perfect_crop", os.Getenv("SECONDARY_API_URL"), "pokemons")
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("upload", "image."+ext)
	if err != nil {
		return nil, err
	}
	io.Copy(part, bdy)
	writer.Close()

	r, err := http.NewRequest("POST", url, body)
	if err != nil {
		return nil, err
	}
	r.Header.Set("Content-Type", writer.FormDataContentType())
	client := &http.Client{}
	rep, err := client.Do(r)
	if err != nil {
		return nil, err
	}
	return rep.Body, nil
}

func ApplyTransforms(img io.Reader, id discord.UserID, ext string) (io.Reader, error) {
	if t, ok := data.Transforms[id]; ok {
		for _, transform := range t {
			switch transform {
			case "pc":
				a, err := PerfectCropImage(img, ext)
				if err != nil {
					return nil, err
				}
				img = a
			case "bg":
				a, err := RemoveBgFromImage(img, ext)
				if err != nil {
					return nil, err
				}
				img = a
			}
		}
	} else {
		a, err := RemoveBgFromImage(img, ext)
		if err != nil {
			return nil, err
		}
		img = a
	}
	return img, nil
}

func ApplyTransformsManual(img io.Reader, transform string, ext string) (io.Reader, error) {
	switch transform {
	case "pc":
		a, err := PerfectCropImage(img, ext)
		if err != nil {
			return nil, err
		}
		img = a
	case "bg":
		a, err := RemoveBgFromImage(img, ext)
		if err != nil {
			return nil, err
		}
		img = a
	}
	return img, nil
}
