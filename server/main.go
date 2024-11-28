package main

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
)

type ExchangeRate struct {
	UsdBrl Exchange `json:"USDBRL"`
}

type BidResponse struct {
	Bid float64 `json:"bid"`
}

type Exchange struct {
	ID         int    `gorm:"primaryKey" json:"-"`
	Code       string `json:"code"`
	Codein     string `json:"codein"`
	Name       string `json:"name"`
	High       string `json:"high"`
	Low        string `json:"low"`
	VarBid     string `json:"varBid"`
	PctChange  string `json:"pctChange"`
	Bid        string `json:"bid"`
	Ask        string `json:"ask"`
	Timestamp  string `json:"timestamp"`
	CreateDate string `json:"create_date"`
}

func main() {
	http.HandleFunc("/cotacao", SearchExchange)
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		return
	}
}

func SearchExchange(w http.ResponseWriter, r *http.Request) {
	exchange, err := GetExchange()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	bidValue, err := strconv.ParseFloat(exchange.Bid, 64)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Erro ao converter Bid para float64: %v", err)
		return
	}

	response := BidResponse{
		Bid: bidValue,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)

}

func GetExchange() (*Exchange, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()

	select {
	case <-ctx.Done():
		log.Println("falha ao buscar a cotação")
	}

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}

	req, err := http.NewRequestWithContext(ctx, "GET", "https://economia.awesomeapi.com.br/json/last/USD-BRL", nil)
	if err != nil {
		panic(err)
	}
	response, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer response.Body.Close()

	var exchange ExchangeRate
	err = json.NewDecoder(response.Body).Decode(&exchange)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Erro ao fazer o parse da resposta: %v\n", err)
		return nil, err
	}

	SaveResult(&exchange)
	return &exchange.UsdBrl, nil
}

func SaveResult(exchange *ExchangeRate) {
	db, err := ConnectSqlite()
	if err != nil {
		log.Fatalf("failed to connect database: %v", err)
	}
	defer db.Session(&gorm.Session{AllowGlobalUpdate: true})

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	result := db.WithContext(ctx).Create(&exchange.UsdBrl)
	if result.Error != nil {
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			fmt.Fprintf(os.Stderr, "Operação excedeu o tempo limite de 10ms\n")
		} else {
			fmt.Fprintf(os.Stderr, "Erro ao salvar no banco: %v\n", result.Error)
		}
	}

}

func ConnectSqlite() (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open("./desafio-server.db"), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	db.AutoMigrate(&Exchange{})
	return db, nil
}
