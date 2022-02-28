package weather

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"time"
)

type Temperature float64

func (t Temperature) Fahrenheit() float64 {
	return (float64(t)-273.15)*(9.0/5.0) + 32.0
}

type Conditions struct {
	Summary     string
	Temperature Temperature
	Pressure    int
	Humidity    int
	Wind        float64
}

type OWMResponse struct {
	Weather []struct {
		Main string
	}
	Main struct {
		Temp     Temperature
		Pressure int
		Humidity int
		Wind     float64
	}
}

type Client struct {
	APIKey     string
	BaseURL    string
	HTTPClient *http.Client
}

func NewClient(key string) *Client {
	return &Client{
		APIKey:  key,
		BaseURL: "https://api.openweathermap.org",
		HTTPClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (c Client) FormatURL(location string) string {
	location = url.QueryEscape(location)
	return fmt.Sprintf("%s/data/2.5/weather?q=%s&appid=%s", c.BaseURL, location, c.APIKey)
}

func (c *Client) GetWeather(location string) (Conditions, error) {
	URL := c.FormatURL(location)
	resp, err := c.HTTPClient.Get(URL)
	if err != nil {
		return Conditions{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNotFound {
		return Conditions{}, fmt.Errorf("could not find location: %s ", location)
	}
	if resp.StatusCode != http.StatusOK {
		return Conditions{}, fmt.Errorf("unexpected response status %q", resp.Status)
	}
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return Conditions{}, err
	}
	conditions, err := ParseResponse(data)
	if err != nil {
		return Conditions{}, err
	}
	return conditions, nil
}

func ParseResponse(data []byte) (Conditions, error) {
	var resp OWMResponse
	err := json.Unmarshal(data, &resp)
	if err != nil {
		return Conditions{}, fmt.Errorf("invalid API response %s: %w", data, err)
	}
	if len(resp.Weather) < 1 {
		return Conditions{}, fmt.Errorf("invalid API response %s: require at least one weather element", data)
	}
	conditions := Conditions{
		Summary:     resp.Weather[0].Main,
		Temperature: resp.Main.Temp,
		Pressure:    resp.Main.Pressure,
		Humidity:    resp.Main.Humidity,
		Wind:        resp.Main.Wind,
	}
	return conditions, nil
}

func FormatURL(baseURL, location, key string) string {
	return fmt.Sprintf("%s/data/2.5/weather?q=%s&appid=%s", baseURL, location, key)
}

func Get(location, key string) (Conditions, error) {
	c := NewClient(key)
	conditions, err := c.GetWeather(location)
	if err != nil {
		return Conditions{}, err
	}
	return conditions, nil
}

func RunCLI() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s LOCATION\n\nExample: %[1]s London,UK", os.Args[0])
		os.Exit(1)
	}
	location := os.Args[1]
	key := os.Getenv("OPENWEATHERMAP_API_KEY")
	if key == "" {
		fmt.Fprintln(os.Stderr, "Please set the environment variable OPENWEATHERMAP_API_KEY")
		os.Exit(1)
	}
	conditions, err := Get(location, key)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Printf("%s %.1f %d %d %.1f\n", conditions.Summary, conditions.Temperature.Fahrenheit(), conditions.Pressure, conditions.Humidity, conditions.Wind)

}
