package files

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

func DownloadFile(filepath, url string) (err error) {
	// Create the file
	out, err := os.Create(filepath)
	if err != nil {
		return fmt.Errorf("DownloadFile Create filepath-%s err: %v", filepath, err)
	}
	defer out.Close()
	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("DownloadFile Get url-%s err: %v", url, err)
	}
	defer resp.Body.Close()
	// Check server response
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("DownloadFile Get url-%s err: bad status: %s", url, resp.Status)
	}
	// Writer the body to file
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return fmt.Errorf("DownloadFile Copy err: %v", err)
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
				return "", nil, fmt.Errorf("CreateForm Open err: %v", err)
			}
			defer file.Close()
			part, err := mp.CreateFormFile(key, val)
			if err != nil {
				return "", nil, fmt.Errorf("CreateForm CreateFormFile err: %v", err)
			}
			io.Copy(part, file)
		} else {
			mp.WriteField(key, val)
		}
	}
	return mp.FormDataContentType(), body, nil
}

func CreateFormV2(formFiles map[string]string, formFields map[string]string) (string, io.Reader, error) {
	fmt.Println("CreateFormV2::", formFields)
	body := new(bytes.Buffer)
	mp := multipart.NewWriter(body)
	defer mp.Close()
	for key, val := range formFiles {
		if strings.HasPrefix(val, "@") {
			val = val[1:]
			file, err := os.Open(val)
			if err != nil {
				return "", nil, fmt.Errorf("CreateFormV2 Open err: %v", err)
			}
			defer file.Close()
			part, err := mp.CreateFormFile(key, val)
			if err != nil {
				return "", nil, fmt.Errorf("CreateFormV2 CreateFormFile err: %v", err)
			}
			io.Copy(part, file)
		}
	}
	for key, val := range formFields {
		mp.WriteField(key, val)
	}
	return mp.FormDataContentType(), body, nil
}

// delete all files from dir
func RemoveContentsFromDir(dir string) error {
	d, err := os.Open(dir)
	if err != nil {
		return err
	}
	defer d.Close()
	names, err := d.Readdirnames(-1)
	if err != nil {
		return err
	}
	for _, name := range names {
		err = os.RemoveAll(filepath.Join(dir, name))
		if err != nil {
			return err
		}
	}
	return nil
}
