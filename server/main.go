package main

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"io"
	"log"
	"net/http"
	"os"
)

type ExchangeRate struct {
	UsdBrl Exchange `json:"USDBRL"`
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
	http.ListenAndServe(":8080", nil)
}

func SearchExchange(w http.ResponseWriter, r *http.Request) {
	exchange, err := GetExchange()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(exchange)
}

func GetExchange() (*Exchange, error) {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}
	request, err := client.Get("https://economia.awesomeapi.com.br/json/last/USD-BRL")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Erro ao realizar a requisição: %v\n", err)
		return nil, err
	}
	defer request.Body.Close()
	response, err := io.ReadAll(request.Body)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Erro ao ao ler a resposta: %v\n", err)
		return nil, err
	}
	var exchange ExchangeRate
	err = json.Unmarshal(response, &exchange)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Erro ao parse da resposta: %v\n", err)
		return nil, err
	}
	db, err := ConnectSqlite()
	if err != nil {
		log.Fatalf("failed to connect database: %v", err)
		return nil, err
	}
	defer db.Session(&gorm.Session{AllowGlobalUpdate: true})
	db.Create(&exchange.UsdBrl)
	return &exchange.UsdBrl, nil
}

func ConnectSqlite() (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open("./desafio-server.db"), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	db.AutoMigrate(&Exchange{})
	return db, nil
}
