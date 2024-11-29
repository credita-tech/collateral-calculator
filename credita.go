package main

import (
	"encoding/csv"
	"fmt"
	"log"
	"math"
	"net/http"
	"os"
	"sort"
	"strconv"
	"sync"
)

var (
	closePrices   []float64
	dailyReturns  []float64
	stdDev        float64
	dailyVolatility float64
	percentile5   float64
	liquidityPremium = 0.05
	once          sync.Once
)

func loadData() {
	file, err := os.Open("bitcoin.csv")
	if err != nil {
		log.Fatalf("Error opening file: %v", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	rows, err := reader.ReadAll()
	if err != nil {
		log.Fatalf("Error reading CSV: %v", err)
	}

	closePrices = make([]float64, 0, len(rows)-1) // Assuming first row is header
	for i, row := range rows {
		if i == 0 { // Skip header
			continue
		}
		price, err := strconv.ParseFloat(row[1], 64)
		if err != nil {
			log.Fatalf("Error parsing close price on row %d: %v", i+1, err)
		}
		closePrices = append(closePrices, price)
	}

	dailyReturns = make([]float64, len(closePrices)-1)
	for i := 1; i < len(closePrices); i++ {
		dailyReturns[i-1] = (closePrices[i] - closePrices[i-1]) / closePrices[i-1]
	}

	mean := 0.0
	for _, r := range dailyReturns {
		mean += r
	}
	mean /= float64(len(dailyReturns))

	variance := 0.0
	for _, r := range dailyReturns {
		variance += math.Pow(r-mean, 2)
	}
	variance /= float64(len(dailyReturns))
	stdDev = math.Sqrt(variance)
	dailyVolatility = stdDev * math.Sqrt(252)

	sort.Float64s(dailyReturns)
	percentileIndex := int(math.Ceil(0.05 * float64(len(dailyReturns)))) - 1
	percentile5 = dailyReturns[percentileIndex]
}

func calculateCollateral(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	loanAmountStr := r.URL.Query().Get("amount")
	if loanAmountStr == "" {
		http.Error(w, "Amount query parameter is required", http.StatusBadRequest)
		return
	}

	loanAmount, err := strconv.ParseFloat(loanAmountStr, 64)
	if err != nil {
		http.Error(w, "Invalid amount value", http.StatusBadRequest)
		return
	}

	once.Do(loadData)

	amountWorthCollateral := loanAmount * (1 + percentile5 + dailyVolatility + liquidityPremium)
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"amountWorthCollateral": %.2f}`, amountWorthCollateral)
}

func main() {
	http.HandleFunc("/calculate-collateral", calculateCollateral)
	fmt.Println("Server is running on port 8080...")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
