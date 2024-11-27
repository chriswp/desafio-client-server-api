package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"
)

type Price struct {
	Bid string `json:"bid"`
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*400)
	defer cancel()
	select {
	case <-time.After(time.Millisecond * 400):
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
		fmt.Println("Price:", price.Bid)
	case <-ctx.Done():
		fmt.Println("Requisição realizada com sucesso")
	}

}
