package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"sort"
	"strconv"
	"time"
)

type requestResponse struct {
	Base  string             `json:"base"`
	Date  string             `json:"date"`
	Rates map[string]float64 `json:"rates"`
}
type currencyCollection struct {
	rates []*requestResponse
}

const (
	timeform = "2006-01-02"
	url      = "https://api.fixer.io/"
)

var (
	from string
	to   string
)

func main() {
	from = os.Getenv("from")
	to = os.Getenv("to")
	fromTime, err := time.Parse(timeform, from)
	if err != nil {
		panic("couldnt parse the 'from':" + err.Error())
	}
	toTime, err := time.Parse(timeform, to)
	if err != nil {
		panic("couldnt parse the 'to':" + err.Error())
	}

	collection := &currencyCollection{}
	date := fromTime
	fmt.Println("starting...")
	for toTime.After(date) {
		dateString := date.Format(timeform)
		fmt.Printf("fetching rates for: %v \r", dateString)
		resp := query(dateString)
		resp.Date = dateString
		collection.rates = append(collection.rates, resp)
		date = date.AddDate(0, 0, 1)
	}

	f, err := os.Create("rates.csv")
	if err != nil {
		panic(err)
	}
	defer f.Close()

	writer := csv.NewWriter(f)
	header := []string{"date"}
	header = append(header, getSortedKeys(collection.rates[0].Rates)...)
	writer.Write(header)
	writer.Flush()
	writer.WriteAll(collection.toCSVLines())

}

func query(date string) *requestResponse {
	resp, err := http.Get(queryURL(date))
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}
	response := &requestResponse{}
	json.Unmarshal(body, response)
	return response
}

func queryURL(date string) string {
	return url + date
}

func (c *currencyCollection) toCSVLines() [][]string {
	result := [][]string{}
	for _, rate := range c.rates {
		keys := getSortedKeys(rate.Rates)
		row := []string{rate.Date}
		for _, k := range keys {
			s := strconv.FormatFloat(rate.Rates[k], 'f', 8, 64)
			row = append(row, s)
		}
		result = append(result, row)
	}
	return result
}

func getSortedKeys(m map[string]float64) []string {
	keys := []string{}
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
