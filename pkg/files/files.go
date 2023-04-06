package files

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"strings"
)

func DownloadFile(filepath string, url string) (err error) {
	// Create the file
	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()
	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	// Check server response
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status: %s", resp.Status)
	}
	// Writer the body to file
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return err
	}
	return nil
}

func CreateForm(form map[string]string) (string, io.Reader, error) {
	fmt.Println("CreateForm::", form)
	body := new(bytes.Buffer)
	mp := multipart.NewWriter(body)
	defer mp.Close()
	for key, val := range form {
			if strings.HasPrefix(val, "@") {
				val = val[1:]
				file, err := os.Open(val)
				if err != nil {
					return "", nil, err
				}
				defer file.Close()
				part, err := mp.CreateFormFile(key, val)
				if err != nil {
					return "", nil, err
				}
				io.Copy(part, file)
			} else {
			   	mp.WriteField(key, val)
			}
	}
	return mp.FormDataContentType(), body, nil
}