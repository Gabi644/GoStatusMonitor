package main

import (
	"bufio"
	"context"
	"encoding/csv"
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"
)

type Result struct {
	URL      string
	Estado   string
	Latencia string
}

func checkURL(ctx context.Context, url string, wg *sync.WaitGroup, results chan<- Result) {
	defer wg.Done()

	start := time.Now()

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		results <- Result{URL: url, Estado: "Inactivo", Latencia: "-1 ms"}
		return
	}

	client := http.DefaultClient
	resp, err := client.Do(req)

	latency := time.Since(start).Milliseconds()
	estado := "Activo"

	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			estado = "Inactivo (Timeout)"
		} else {
			estado = "Inactivo"
		}
		latency = -1
	} else {
		defer resp.Body.Close()
	}

	results <- Result{URL: url, Estado: estado, Latencia: fmt.Sprintf("%d ms", latency)}
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	file, err := os.Open("sites.txt")
	if err != nil {
		fmt.Printf("Error al abrir el archivo: %v\n", err)
		return
	}
	defer file.Close()

	var wg sync.WaitGroup
	results := make(chan Result)
	var urls []string

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		url := scanner.Text()
		urls = append(urls, url)
	}

	for _, url := range urls {
		wg.Add(1)
		go checkURL(ctx, url, &wg, results)
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	outputFile, _ := os.Create("results.csv")
	defer outputFile.Close()
	writer := csv.NewWriter(outputFile)
	defer writer.Flush()

	writer.Write([]string{"URL", "Estado", "Latencia"})
	for result := range results {
		writer.Write([]string{result.URL, result.Estado, result.Latencia})
	}

	fmt.Println("Escaneando sitios...")
	for res := range results {
		fmt.Printf("✓ %s: %s (%s)\n", res.URL, res.Estado, res.Latencia)
		writer.Write([]string{res.URL, res.Estado, res.Latencia})
	}

	fmt.Println("¡Proceso finalizado! Resultados guardados en results.csv")
}
