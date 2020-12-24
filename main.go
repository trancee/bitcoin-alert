package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/hajimehoshi/go-mp3"
	"github.com/hajimehoshi/oto"

	"github.com/trancee/bitcoin-alert/asset"
)

var (
	interval = 1 * 60 // 1 minute

	currency = "USD"
	amount   = 0.0
)

type Spot struct {
	Base     string  `json:"base"`
	Currency string  `json:"currency"`
	Amount   float64 `json:"amount,string"`
}
type Data struct {
	Spot *Spot `json:"data"`
}

func price(currency string) (*Spot, error) {
	url := "https://api.coinbase.com/v2/prices/spot?currency=" + currency

	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var data Data
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}

	return data.Spot, nil
}

func alert() error {
	data, err := asset.Asset("asset/suffer.mp3")
	if err != nil {
		return err
	}

	d, err := mp3.NewDecoder(bytes.NewReader(data))
	if err != nil {
		return err
	}

	c, err := oto.NewContext(d.SampleRate(), 2, 2, 8192)
	if err != nil {
		return err
	}
	defer c.Close()

	p := c.NewPlayer()
	defer p.Close()

	if _, err := io.Copy(p, d); err != nil {
		return err
	}

	return nil
}

func percentage(old float64, new float64) string {
	var sign string

	percentage := ((new - old) / old) * 100

	if math.IsInf(percentage, 0) {
		return ""
	}
	if percentage >= 0.0 {
		sign = "+"
	}

	return fmt.Sprintf("%s%.2f%%", sign, percentage)
}

func check() {
	if data, err := price(currency); err != nil {
		log.Fatal(err)
	} else {
		diff := percentage(amount, data.Amount)

		fmt.Printf("%v %s\n", data, diff)

		if amount > data.Amount || diff == "" {
			go alert()
		}

		amount = data.Amount
	}
}

func main() {
	// create a context with cancel() callback function
	ctx, cancel := context.WithCancel(context.Background())

	// create a channel for listening to OS signals and connecting OS interrupts to the channel
	sig := make(chan os.Signal, 1)

	signal.Notify(
		sig,

		os.Interrupt,

		syscall.SIGHUP,
		syscall.SIGTERM,
		syscall.SIGQUIT,
	)

	go func() {
		fmt.Printf("EXITING: %+v\n", <-sig)

		cancel()
	}()

	fmt.Println("RUNNING")

	check()

loop:
	for {
		select {
		case <-ctx.Done():
			break loop
		case <-time.After(time.Duration(interval) * time.Second):
			check()
		}
	}

	fmt.Println("EXIT")
}
