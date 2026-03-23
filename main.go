package main

import (
	"bufio"
	"context"
	"encoding/csv"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"sync"
	"time"
)

// Result define el esquema de datos para el reporte y la interfaz
type Result struct {
	URL      string
	Estado   string
	Latencia string
}

// Store centraliza el estado de la aplicación.
// Se utiliza RWMutex para permitir lecturas concurrentes (HTTP GET) mientras se bloquea el acceso durante la escritura del worker (Update).
type Store struct {
	sync.RWMutex
	Data []Result
}

var (
	globalStore Store
	// Pre-parsing del template para optimizar el tiempo de respuesta del servidor
	tmpl = template.Must(template.ParseFiles("index.html"))
)

// checkURL realiza la petición HTTP respetando el ciclo de vida del contexto.
func checkURL(ctx context.Context, url string, wg *sync.WaitGroup, results chan<- Result) {
	defer wg.Done()

	start := time.Now()
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		results <- Result{URL: url, Estado: "Error de Request", Latencia: "-1ms"}
		return
	}

	resp, err := http.DefaultClient.Do(req)
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
		// Cierre preventivo del body para evitar fugas de memoria (memory leaks)
		defer resp.Body.Close()
	}

	results <- Result{URL: url, Estado: estado, Latencia: fmt.Sprintf("%dms", latency)}
}

// performScan gestiona el ciclo de vida del escaneo: lectura, ejecución y persistencia.
func performScan() {
	fmt.Printf("[%s] Iniciando ciclo de monitoreo...\n", time.Now().Format("15:04:05"))

	// Apertura dinámica para permitir modificaciones en caliente del listado de sitios
	file, err := os.Open("sites.txt")
	if err != nil {
		fmt.Printf("Fallo crítico: No se pudo abrir sites.txt: %v\n", err)
		return
	}
	defer file.Close()

	var urls []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		if line := scanner.Text(); line != "" {
			urls = append(urls, line)
		}
	}

	// Timeout global para evitar que el worker quede zombi por hilos bloqueados
	ctx, cancel := context.WithTimeout(context.Background(), 25*time.Second)
	defer cancel()

	var wg sync.WaitGroup
	resultsChan := make(chan Result, len(urls))

	for _, url := range urls {
		wg.Add(1)
		go checkURL(ctx, url, &wg, resultsChan)
	}

	// Orquestador de cierre del canal de resultados
	go func() {
		wg.Wait()
		close(resultsChan)
	}()

	var currentBatch []Result
	// Persistencia en CSV para auditoría y backup local
	outputFile, _ := os.Create("results.csv")
	writer := csv.NewWriter(outputFile)
	writer.Write([]string{"URL", "Estado", "Latencia"})

	for r := range resultsChan {
		currentBatch = append(currentBatch, r)
		writer.Write([]string{r.URL, r.Estado, r.Latencia})
	}
	writer.Flush()
	outputFile.Close()

	// Actualización atómica del estado global (Write Lock)
	globalStore.Lock()
	globalStore.Data = currentBatch
	globalStore.Unlock()

	fmt.Println("Escaneo finalizado. Backup actualizado en results.csv")
}

func main() {
	// Worker asíncrono: Ejecuta el primer escaneo y luego opera bajo el ticker
	go func() {
		performScan()
		ticker := time.NewTicker(60 * time.Second)
		for range ticker.C {
			performScan()
		}
	}()

	// Endpoint raíz: Renderizado completo del dashboard
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		globalStore.RLock()
		defer globalStore.RUnlock()
		tmpl.ExecuteTemplate(w, "index.html", globalStore.Data)
	})

	// Endpoint HTMX: Retorno de fragmento parcial para actualización dinámica
	http.HandleFunc("/results", func(w http.ResponseWriter, r *http.Request) {
		globalStore.RLock()
		defer globalStore.RUnlock()
		tmpl.ExecuteTemplate(w, "table-body", globalStore.Data)
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8045"
	}

	fmt.Printf("Servidor iniciado en el puerto %s\n", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		fmt.Printf("Error al levantar el servidor: %v\n", err)
	}
}
