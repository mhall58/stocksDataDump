package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"github.com/piquette/finance-go/chart"
	"github.com/piquette/finance-go/datetime"
	"github.com/piquette/finance-go/quote"
	"io/ioutil"
	"net/http"
	"os"
	"time"
)

type TradedCompanies struct {
	Company  string `json:"name"`
	Symbol   string `json:"ticker"`
	Exchange string `json:"exchange"`
}

func main() {

	minPrice := 00.10
	maxPrice := 5.00
	directory := "export/"

	startDate := datetime.Datetime{
		Day:   1,
		Month: 1,
		Year:  2018,
	}

	endDate := datetime.Datetime{
		Day:   22,
		Month: 2,
		Year:  2021,
	}

	for _, stock := range getCompanies() {
		q, err := quote.Get(stock.Symbol)

		if err != nil {
			fmt.Println(err)
			continue
		}

		if q == nil {
			continue
		}

		if q.RegularMarketPrice < minPrice || q.RegularMarketPrice > maxPrice {
			continue
		}

		params := &chart.Params{
			Symbol:   stock.Symbol,
			Interval: datetime.OneDay,
			Start:    &startDate,
			End:      &endDate,
		}

		data := chart.Get(params)

		if data == nil {
			fmt.Println(stock.Symbol, "No Data")
			continue
		}

		if data.Err() != nil {
			fmt.Println(stock.Symbol, data.Err())
			continue
		}

		if data.Count() == 0 {
			continue
		}

		fmt.Println("Creating csv for ", data.Meta().Symbol, q.RegularMarketPrice)

		go writeCsv(directory, data)

	}

}

func getCompanies() []TradedCompanies {
	var companies []TradedCompanies
	resp, _ := http.Get("https://dumbstockapi.com/stock?exchange=AMEX,NASDAQ,NYSE")

	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	json.Unmarshal(body, &companies)

	return companies
}

func writeCsv(directory string, data *chart.Iter) {
	recordFile, err := os.Create(directory + data.Meta().Symbol + ".csv")

	if err != nil {
		fmt.Println("File Create ERROR ", err)
		return
	}
	writer := csv.NewWriter(recordFile)

	defer writer.Flush()
	defer recordFile.Close()

	var lines = [][]string{
		{
			"human_date",
			"open",
			"close",
			"high",
			"low",
			"adj_close",
			"volume",
			"timestamp",
		},
	}

	for data.Next() {
		if data.Err() != nil {
			continue
		}
		var line = []string{
			time.Unix(int64(data.Bar().Timestamp), 0).Format(time.RFC3339),
			fmt.Sprintf("%v", data.Bar().Close),
			fmt.Sprintf("%v", data.Bar().High),
			fmt.Sprintf("%v", data.Bar().Low),
			fmt.Sprintf("%v", data.Bar().AdjClose),
			fmt.Sprintf("%v", data.Bar().Volume),
			fmt.Sprintf("%v", data.Bar().Timestamp),
		}

		lines = append(lines, line)
	}

	writer.WriteAll(lines)
}
