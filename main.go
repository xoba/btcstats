package main

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"sort"
	"strconv"
	"time"
)

type Prices []Price

type Price struct {
	Time  time.Time
	Value float64
}

func (p Prices) Range() (earliest, latest time.Time) {
	return p[0].Time, p[len(p)-1].Time
}

func (p Prices) AsOf(t time.Time) Price {
	return p[sort.Search(len(p), func(i int) bool {
		t1 := p[i].Time
		return t == t1 || t1.After(t)
	})]
}

func LoadBTC() (Prices, error) {
	return LoadCSV("Coinbase_BTCUSD_d.csv", 1, 3)
}
func LoadSP500() (Prices, error) {
	return LoadCSV("sp500.csv", 0, 4)
}

// LoadCSV assumes and loads a csv with header row
func LoadCSV(file string, dayIndex, priceIndex int) (Prices, error) {
	var p Prices
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	rows, err := csv.NewReader(f).ReadAll()
	if err != nil {
		return nil, err
	}
	for i, r := range rows {
		if i == 0 {
			continue
		}
		t, err := time.Parse(iso, r[dayIndex])
		if err != nil {
			return nil, err
		}
		price, err := strconv.ParseFloat(r[priceIndex], 64)
		if err != nil {
			return nil, err
		}
		p = append(p, Price{Time: t, Value: price})
	}
	sort.Slice(p, func(i, j int) bool {
		return p[i].Time.Before(p[j].Time)
	})
	return p, nil
}

const (
	day  = 24 * time.Hour
	year = 365 * day
	iso  = "2006-01-02"
)

func main() {
	if err := Run(); err != nil {
		log.Fatal(err)
	}
}

func Run() error {
	type gen func() (Prices, error)
	run := func(name string, g gen) error {
		prices, err := g()
		if err != nil {
			return err
		}
		t0, latest := prices.Range()
		fmt.Printf("%s data from %s to %s:\n", name, t0.Format(iso), latest.Format(iso))
		max := latest.Add(-year)
		var returns []float64
		for {
			if t0.After(max) {
				break
			}
			p0 := prices.AsOf(t0).Value
			p1 := prices.AsOf(t0.Add(year)).Value
			returns = append(returns, 100*(p1-p0)/p0)
			t0 = t0.Add(day)
		}
		sort.Float64s(returns)
		for i := 0; i <= 100; i += 5 {
			index := i * len(returns) / 100
			switch {
			case index < 0:
				index = 0
			case index >= len(returns):
				index = len(returns) - 1
			}
			fmt.Printf("  %3d'th percentile return: %+5.0f%%\n",
				i,
				returns[index],
			)
		}
		fmt.Println()
		return nil
	}
	if err := run("sp500", LoadSP500); err != nil {
		return err
	}
	if err := run("btc", LoadBTC); err != nil {
		return err
	}
	return nil
}
