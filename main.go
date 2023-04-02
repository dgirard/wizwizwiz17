package main

// wizwizwiz project
import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
)

type Translation struct {
	DetectedSourceLanguage string `json:"detected_source_language"`
	Text                   string `json:"text"`
}
type TranslateResponse struct {
	Translations []Translation `json:"translations"`
}

type CompletionRequest struct {
	Messages         []Message `json:"messages"`
	Temperature      float32   `json:"temperature"`
	MaxTokens        int       `json:"max_tokens"`
	TopP             float32   `json:"top_p"`
	FrequencyPenalty float32   `json:"frequency_penalty"`
	PresencePenalty  float32   `json:"presence_penalty"`
	Model            string    `json:"model"`
	Stream           bool      `json:"stream"`
}
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// build an http server
func main() {
	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/translate", translateHandler)
	fs := http.FileServer(http.Dir("."))
	http.Handle("/static/", http.StripPrefix("/static/", fs))
	fmt.Println("Server listening on port 8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
func indexHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "index.html")
}

func makeDeepLRequest(apiKey string, text string, targetLang string) (string, error) {
	apiUrl := "https://api-free.deepl.com/v2/translate"
	reqBody := fmt.Sprintf("text=%s&target_lang=%s", url.QueryEscape(text), targetLang)
	req, err := http.NewRequest("POST", apiUrl, bytes.NewReader([]byte(reqBody)))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Authorization", "DeepL-Auth-Key "+apiKey)
	req.Header.Set("User-Agent", "YourApp/1.2.3")

	client := http.Client{}
	resp, err := client.Do(req)

	if err != nil {
		return "", err
	}

	defer resp.Body.Close()
	respBody, err := ioutil.ReadAll(resp.Body)
	log.Println("The status code we got is:", resp.StatusCode)

	if err != nil {
		return "", err
	}

	return string(respBody), nil
}

func translateHandler(w http.ResponseWriter, r *http.Request) {

	// Get the API key from the query parameters
	apiKey := r.URL.Query().Get("apiKey")
	openaiApiKey := r.URL.Query().Get("openaiApiKey")

	// Get the text and target language from the query parameters
	text := r.URL.Query().Get("text")

	targetLangInput := r.URL.Query().Get("targetLang")
	log.Println("The target language is  gg:", targetLangInput)

	log.Println("Dans la boucle Here")
	if targetLangInput == "GPT" {
		log.Println("Dans la boucle GPT")
    userText :=  r.URL.Query().Get("userText")

    
		respBody, err := openaiChatCompletionsRequest(openaiApiKey, userText, text)
		if err != nil {
			fmt.Println("Error:", err)
			return
		}
		_, content, err := parseJSONOpenaiResponseMessage(string(respBody))

		if err != nil {

			fmt.Println(err)
			return
		}
          log.Println("coucou")
    fmt.Fprintf(w, "%s", content)
	} else {
		// Check if the API key is empty
		if apiKey == "" {
			http.Error(w, "Missing API key", http.StatusBadRequest)
			return
		}

		_, targetLang, err := getTranslationLanguages(targetLangInput)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// Make the DeepL API request
		translatedText, err := makeRequest(w, apiKey, text, targetLang)
		if err != nil {
			log.Println("Error")

			return
		}
		if err == nil && targetLangInput == "FRENFR" {
			translatedText, err = makeRequest(w, apiKey, translatedText, "FR")
		}
		// Write the translated text to the response
    log.Println(translatedText)
		fmt.Fprintf(w, "%s", translatedText)
	}
}

func makeRequest(w http.ResponseWriter, apiKey string, text string, targetLang string) (string, error) {
	respBody, err := makeDeepLRequest(apiKey, text, targetLang)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return "", err
	}
	// Parse the JSON response body
	var respBodyVar TranslateResponse
	if err := json.Unmarshal([]byte(respBody), &respBodyVar); err != nil {
		return "", err
	}
	// Extract the translated text
	translatedText := ""
	if len(respBodyVar.Translations) > 0 {
		translatedText = respBodyVar.Translations[0].Text
	}
	return translatedText, nil
}

func getTranslationLanguages(targetLangInput string) (string, string, error) {
	log.Println(targetLangInput)

	sourceLang := ""
	targetLang := ""
	if targetLangInput == "ENFR" {
		sourceLang = "auto"
		targetLang = "FR"
	} else if targetLangInput == "FREN" {
		sourceLang = "auto"
		targetLang = "EN"
	} else if targetLangInput == "FRENFR" {
		sourceLang = "auto"
		targetLang = "EN"
	} else {
		// Invalid targetLang value
		return "", "", errors.New("Invalid targetLang value")
	}
	return sourceLang, targetLang, nil
}

func openaiChatCompletionsRequest(openaiKey string, systemMsg string, userMsg string) ([]byte, error) {
	// Set the necessary headers
	contentType := "application/json"
	authToken := "Bearer " + openaiKey // Replace with your actual API token
	// Build the HTTP request body
	msgs := []Message{
		Message{Role: "system", Content: systemMsg},
		Message{Role: "user", Content: userMsg},
	}
	completionRequest := CompletionRequest{
		Messages:    msgs,
		Temperature: 1,
		MaxTokens:   1693,
		Model:       "gpt-4",
	}
	reqBody, err := json.Marshal(completionRequest)
	if err != nil {
		return nil, err
	}
	// Build the HTTP request
	req, err := http.NewRequest("POST", "https://api.openai.com/v1/chat/completions", bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", contentType)
	req.Header.Set("Authorization", authToken)
	// Send the HTTP request
	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	// Read the response body
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return respBody, nil
}

type ChatCompletion struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int    `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Message struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"message"`
		FinishReason string `json:"finish_reason"`
		Index        int    `json:"index"`
	} `json:"choices"`
}

func parseJSONOpenaiResponseMessage(jsonMessage string) (string, string, error) {
	// Parse the JSON message into a ChatCompletion struct
	var chatCompletion ChatCompletion
	err := json.Unmarshal([]byte(jsonMessage), &chatCompletion)
	if err != nil {
		return "", "", err
	}
	// Extract the model and content fields
	model := chatCompletion.Model
	content := chatCompletion.Choices[0].Message.Content
	return model, content, nil
}
