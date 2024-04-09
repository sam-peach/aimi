package ai

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
)

// Structs to model the OpenAI API request and response
type OpenAIRequest struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type OpenAIResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

type Classification struct {
	PrimaryCategory   Category
	SecondaryCategory Category
}

type Category struct {
	Key       string
	Relevance float64
}

func IsArticleRelevant(inputText string) bool {
	apiKey := os.Getenv("OPENAI_API_KEY")

	userPrompt := "Returning only 'true' of 'false' is the following article specifically about the application, or integration, of AI in the military:\n" + inputText

	requestData := OpenAIRequest{
		Model: "gpt-4-turbo-preview", // Specify the model you're using
		Messages: []Message{
			{Role: "system", Content: ""},
			{Role: "user", Content: userPrompt},
		},
	}

	// Marshal the requestData to JSON
	requestBody, err := json.Marshal(requestData)
	if err != nil {
		panic(err)
	}

	// Create a new HTTP request
	req, err := http.NewRequest("POST", "https://api.openai.com/v1/chat/completions", bytes.NewBuffer(requestBody))
	if err != nil {
		panic(err)
	}

	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	// Unmarshal the JSON response into the OpenAIResponse struct
	var response OpenAIResponse
	if err := json.Unmarshal(responseBody, &response); err != nil {
		panic(err)
	}

	// Check if there is at least one choice and print the content
	if len(response.Choices) == 0 {
		panic(errors.New("no response from GPT"))
	}

	return strings.ToLower(response.Choices[0].Message.Content) == "true"
}

func Categorize(inputText string) (Classification, error) {
	// Your OpenAI API key
	apiKey := os.Getenv("OPENAI_API_KEY")
	classification := Classification{}

	systemPrompt := `Given the following categories:

	E1: government AI defense funding, military AI budget, AI research allocation military, budget, dollars
	E2: AI defense investment forecast, future military AI spending, defense AI economic outlook, earmark
	E3: AI defense industry growth, military AI startups, national AI military sector, private, commerce, enterprise
	G1: AI defense partnerships, military AI cooperation, international AI defense agreements, allies, assistance, ties
	G2: AI military rivalry, adversarial AI capabilities, AI defense competition, fear, escalation, aggression, sanctions
	G3: global AI defense policy, AI military treaties, international AI defense ethics, participation, forum, gathering, meeting
	M1: military AI adoption, AI defense leadership, AI military initiatives., need, imperative, desire, modernization
	M2: AI incorporation, AI use in military, cross-service AI applications, usage
	M3: AI defense R&D, military AI technology development, AI military patents, innovation, cutting-edge
	P1: political support, defense policy, government AI initiatives, president, prime minister, cabinet, ministry
	P2: AI defense legislation, AI military laws, AI defense regulatory framework, courts, ruling, restriction,
	P3: public opinion AI defense, AI military debate, AI defense academic research, activism, NGO, civil society, population
	00: A null category for irrelevant data
	
	Extract the primary, most relevant category the input data fits into and a secondary category. 
	Only output the two category keys and the category relevance scores (a number between 0.0000 and 0.9999).
	Each category and score should be on a new line with in the format of 'key:score' with the null category always being '00:0.0000'`

	// Create an instance of OpenAIRequest with your prompt
	requestData := OpenAIRequest{
		Model: "gpt-4-turbo-preview", // Specify the model you're using
		Messages: []Message{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: inputText},
		},
	}

	// Marshal the requestData to JSON
	requestBody, err := json.Marshal(requestData)
	if err != nil {
		return classification, err
	}

	// Create a new HTTP request
	req, err := http.NewRequest("POST", "https://api.openai.com/v1/chat/completions", bytes.NewBuffer(requestBody))
	if err != nil {
		return classification, err
	}

	// Set the required HTTP headers
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")

	// Send the request using the http.DefaultClient
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return classification, err
	}
	defer resp.Body.Close()

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return classification, err
	}

	// Unmarshal the JSON response into the OpenAIResponse struct
	var response OpenAIResponse
	if err := json.Unmarshal(responseBody, &response); err != nil {
		return classification, err
	}

	// Check if there is at least one choice and print the content
	if len(response.Choices) == 0 {
		return classification, errors.New("no response from GPT")
	}

	c, err := parseCategories(response.Choices[0].Message.Content)
	if err != nil {
		return classification, errors.New("no response from GPT")
	}

	classification.PrimaryCategory = c[0]
	classification.SecondaryCategory = c[1]

	return classification, nil
}

func parseCategories(input string) ([]Category, error) {
	var categories []Category

	// Split the input string by newline to get individual category representations
	lines := strings.Split(input, "\n")

	for _, line := range lines {
		// Split each line by ':' to separate the key and the relevance score
		parts := strings.Split(line, ":")

		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid line format: %s", line)
		}

		// Convert the relevance part from string to float64
		relevance, err := strconv.ParseFloat(parts[1], 64)
		if err != nil {
			panic(fmt.Sprintf("invalid relevance value for %s: %v", parts[0], err))
		}

		// Create a Category struct and append it to the categories slice
		category := Category{
			Key:       parts[0],
			Relevance: relevance,
		}

		categories = append(categories, category)
	}
	return categories, nil
}
