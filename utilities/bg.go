package utilities

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
)

func RemoveImageBackground(img io.ReadCloser, ext string) io.ReadCloser {
	// Send to the API for transformation
	url := fmt.Sprintf("%s/transforms/%s/background_removal", os.Getenv("SECONDARY_API_URL"), "pokemons")
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("upload", "image."+ext)
	if err != nil {
		return nil
	}
	io.Copy(part, img)
	writer.Close()
	img.Close()

	r, err := http.NewRequest("POST", url, body)
	if err != nil {
		return nil
	}
	r.Header.Set("Content-Type", writer.FormDataContentType())
	client := &http.Client{}
	rep, err := client.Do(r)
	if err != nil {
		return nil
	}
	return rep.Body
}
