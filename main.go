package main

import (
	"encoding/csv"
	"fmt"
	"os"
	"sort"
	"strconv"
	"time"
)

type Prices struct {
	elements []Price
}

type Price struct {
	Time  time.Time
	Value float64
}

func (p Prices) Range() (earliest, latest time.Time) {
	return p.elements[0].Time, p.elements[len(p.elements)-1].Time
}

func (p Prices) AsOf(t time.Time) Price {
	i := sort.Search(len(p.elements), func(i int) bool {
		t1 := p.elements[i].Time
		return t == t1 || t1.After(t)
	})
	e := p.elements[i]
	return e
}

func LoadBTC() (*Prices, error) {
	return LoadCSV("Coinbase_BTCUSD_d.csv", 1, 3)
}
func LoadSP500() (*Prices, error) {
	return LoadCSV("sp500.csv", 0, 4)
}

func LoadCSV(file string, day, price int) (*Prices, error) {
	var p Prices
	f, err := os.Open(file)
	check(err)
	defer f.Close()
	rows, err := csv.NewReader(f).ReadAll()
	check(err)
	for i, r := range rows {
		if i == 0 {
			continue
		}
		day := r[day]
		t, err := time.Parse("2006-01-02", day)
		if err != nil {
			return nil, err
		}
		price, err := strconv.ParseFloat(r[price], 64)
		check(err)
		p.elements = append(p.elements, Price{Time: t, Value: price})
	}
	sort.Slice(p.elements, func(i, j int) bool {
		return p.elements[i].Time.Before(p.elements[j].Time)
	})
	return &p, nil
}

func main() {
	const (
		day  = 24 * time.Hour
		year = 365 * day
	)
	type gen func() (*Prices, error)
	run := func(name string, g gen) {
		prices, err := g()
		check(err)
		t0, latest := prices.Range()
		fmt.Printf("%s data from %v to %v\n", name, t0, latest)
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
	}
	run("sp500", LoadSP500)
	run("btc", LoadBTC)
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}
