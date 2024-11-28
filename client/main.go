package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
)

type Price struct {
	Bid float64 `json:"bid"`
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 400*time.Millisecond)
	defer cancel()

	select {
	case <-ctx.Done():
		log.Println("requisição não foi processada")
	}

	req, err := http.NewRequestWithContext(ctx, "GET", "http://localhost:8080/cotacao", nil)
	if err != nil {
		panic(err)
	}
	response, err := http.DefaultClient.Do(req)
	if err != nil {
		panic(err)
	}
	defer response.Body.Close()

	var price Price
	err = json.NewDecoder(response.Body).Decode(&price)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Erro ao fazer o parse da resposta: %v\n", err)

	}

	f, err := os.Create("cotacao.txt")
	if err != nil {
		panic(err)
	}
	_, err = f.WriteString(fmt.Sprintf("Dólar: %f", price.Bid))
	defer f.Close()
	fmt.Println("Price:", price.Bid)

}
