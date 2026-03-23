# Go Real-time Status Monitor 🚀

Un servicio de monitoreo de infraestructura de alto rendimiento que combina un motor de escaneo concurrente en **Go** con una interfaz dinámica utilizando **HTMX**. El sistema realiza chequeos de salud constantes y actualiza un dashboard web sin necesidad de recargar la página.

## 🏗️ Arquitectura del Sistema

El proyecto opera bajo un modelo de **doble motor asíncrono**:
1. **Background Worker:** Una Goroutine dedicada que utiliza un `time.Ticker` para ejecutar escaneos cada 60 segundos. Utiliza `context.Context` para gestionar timeouts globales de red.
2. **Web Server:** Un servidor HTTP nativo que sirve los datos almacenados en memoria. La integridad de los datos entre el escáner y el servidor se garantiza mediante un `sync.RWMutex` (bloqueo de lectura/escritura).

## 🛠️ Tecnologías Utilizadas

- **Backend:** Go (Golang) con concurrencia nativa.
- **Frontend:** HTML5 + CSS3 (Dark Mode) + **HTMX** (para el polling asíncrono).
- **Persistencia:** Archivo CSV local para auditoría y backup de cada ciclo.
- **Sincronización:** `sync.WaitGroup` y `channels` para la orquestación de hilos.

## 📋 Requisitos

- Go 1.20 o superior instalado.
- Archivo `sites.txt` en la raíz con las URLs a monitorear.
