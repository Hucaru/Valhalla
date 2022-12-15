package translates

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
)

func GetTranslate(SourceLng string, TargetLng string, Text string) (string, error) {
	client := &http.Client{}

	form := url.Values{}
	form.Add("source", SourceLng)
	form.Add("target", TargetLng)
	form.Add("text", Text)

	req, _ := http.NewRequest("POST", getEnv("PAPAGO_API_URL", ""), strings.NewReader(form.Encode()))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")
	req.Header.Add("X-Naver-Client-Id", getEnv("PAPAGO_CLIENT_ID", ""))
	req.Header.Add("X-Naver-Client-Secret", getEnv("PAPAGO_SECRET_ID", ""))
	res, err := client.Do(req)
	if err != nil {
		fmt.Println("PAPAGO ERROR", err)
		return "", err
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		fmt.Println("Body Read error", err.Error())
		return "", err
	}

	data := &PapagoMessage{} // TopTracks
	err = json.Unmarshal(body, data)
	if err != nil {
		fmt.Println("json.Unmarshal error", err.Error())
		return "", err
	}
	return data.Message.Result.Text, nil
}

func getEnv(key string, defaultVal string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}

	return defaultVal
}
