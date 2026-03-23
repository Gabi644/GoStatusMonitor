# Go Status Monitor

Un monitor de estado de sitios web de alto rendimiento diseñado para realizar chequeos de disponibilidad y latencia de forma concurrente. Este proyecto fue desarrollado como práctica profesional para profundizar en el modelo de concurrencia de Go y el manejo de contextos.

## 🚀 Características

- **Concurrencia con Goroutines:** Procesa múltiples sitios en paralelo mediante el uso de `sync.WaitGroup`.
- **Comunicación mediante Channels:** Sincronización segura de resultados entre hilos de ejecución.
- **Gestión de Timeouts:** Implementación de `context.WithTimeout` para garantizar que el programa no se bloquee ante sitios lentos.
- **Reporte CSV:** Generación automática de un informe con el estado (Activo/Caído/Timeout) y la latencia en milisegundos.

## 📋 Requisitos

- Go 1.20 o superior.
