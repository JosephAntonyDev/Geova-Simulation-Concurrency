# Patrones de Concurrencia en Geova Simulation

## Resumen Ejecutivo

Este proyecto utiliza **3 goroutines concurrentes** por simulación para enviar datos de sensores a una API REST, con sincronización mediante **Mutex** y visualización en tiempo real.

---

## Goroutines Utilizadas

### **Total por Simulación: 3 Goroutines**

Cada vez que el usuario presiona el botón "CREAR", se lanzan 3 goroutines simultáneas:

| # | Nombre | Sensor | Color | Propósito |
|---|--------|--------|-------|-----------|
| 1 | `tfluna` | TF-Luna (Distancia) | Rojo | Envía datos de distancia láser |
| 2 | `mpu` | MPU6050 (Inclinación) | Azul | Envía datos de inclinación/orientación |
| 3 | `imx` | IMX477 (Cámara) | Verde | Envía datos de nitidez de imagen |

**Ubicación del código**: `game/game.go` - Función `startSimulation()` (líneas ~130-145)

```go
func (g *Game) startSimulation() {
    g.State.Mutex.Lock()
    // ... resetear estado ...
    tilt := g.State.CurrentTilt
    g.State.Mutex.Unlock()

    // Lanzar las 3 goroutines con colores distintivos
    go simulation.SendPOSTRequest(
        "http://localhost:8000/tfluna/sensor",
        simulation.GenerateRandomTFLunaData(),
        "tfluna", g.State, 180.0, color.RGBA{R: 255, G: 50, B: 50, A: 255},
    )
    go simulation.SendPOSTRequest(
        "http://localhost:8000/mpu/sensor",
        simulation.GenerateRandomMPUData(tilt),
        "mpu", g.State, 200.0, color.RGBA{R: 50, G: 150, B: 255, A: 255},
    )
    go simulation.SendPOSTRequest(
        "http://localhost:8000/imx477/sensor",
        simulation.GenerateRandomIMXData(),
        "imx", g.State, 220.0, color.RGBA{R: 50, G: 255, B: 50, A: 255},
    )
}
```

---

## Patrones de Concurrencia Implementados

### **1. Patrón Worker Pool (Fan-Out)**
**Ubicación**: `game/game.go` - `startSimulation()`

**Descripción**: Se lanzan múltiples goroutines (workers) simultáneamente para realizar trabajo en paralelo.

**Características**:
- ✅ 3 workers independientes
- ✅ Cada worker maneja un sensor diferente
- ✅ Ejecutan en paralelo sin bloquearse entre sí
- ✅ No hay dependencias entre workers

**Ventajas**:
- Mejora el rendimiento (3 requests simultáneos vs secuenciales)
- Simula hardware real (sensores enviando datos concurrentemente)
- Reduce el tiempo total de ejecución

---

### **2. Patrón Shared State con Mutex**
**Ubicación**: `state/state.go` y `simulation/workers.go`

**Descripción**: Estado compartido protegido con `sync.Mutex` para evitar race conditions.

#### **Estructura del Estado Compartido**:
```go
type VisualState struct {
    Mutex sync.Mutex                    // ← Mutex para proteger acceso
    Packets map[string]*PacketState     // Estado de paquetes en tránsito
    
    // Timers para animaciones
    PythonAPITimer    int
    RabbitMQTimer     int
    WebsocketAPITimer int
    
    // Datos del dashboard
    DisplayDistancia float64
    DisplayRoll      float64
    DisplayNitidez   float64
    CurrentTilt      float64
    SimulacionIniciada bool
}
```

#### **Uso del Mutex**:

**1. En Worker Goroutines** (`simulation/workers.go`):
```go
func SendPOSTRequest(..., visState *state.VisualState, ...) {
    // LOCK antes de escribir
    visState.Mutex.Lock()
    packet := &state.PacketState{...}
    visState.Packets[packetID] = packet
    visState.Mutex.Unlock()
    
    // ... hacer HTTP request ...
    
    // LOCK antes de actualizar estado
    visState.Mutex.Lock()
    defer visState.Mutex.Unlock()
    
    if err != nil {
        visState.Packets[packetID].Status = state.Error
        return
    }
    visState.Packets[packetID].Status = state.ArrivedAtAPI
}
```

**2. En Game Loop** (`game/game.go`):
```go
func (g *Game) updatePacketFSM() {
    g.State.Mutex.Lock()           // ← LOCK al inicio
    defer g.State.Mutex.Unlock()   // ← UNLOCK automático al salir
    
    for _, packet := range g.State.Packets {
        // ... actualizar posiciones y estados ...
    }
}
```

**Características**:
- ✅ Previene race conditions
- ✅ Uso de `defer` para garantizar unlock
- ✅ Locks de corta duración (minimiza contención)
- ✅ Thread-safe: múltiples goroutines + game loop

**Zonas Críticas Protegidas**:
1. Creación de paquetes (línea ~73-84 en `workers.go`)
2. Actualización de estado HTTP (línea ~101-119 en `workers.go`)
3. Actualización de FSM (línea ~150+ en `game.go`)
4. Inicio de simulación (línea ~113-128 en `game.go`)

---

### **3. Patrón Finite State Machine (FSM) Concurrente**
**Ubicación**: `game/game.go` - `updatePacketFSM()` y `handlePacketArrival()`

**Descripción**: Máquina de estados que controla el ciclo de vida de cada paquete de datos.

#### **Estados del Paquete** (`state/state.go`):
```go
const (
    Idle PacketStatus = iota
    SendingToAPI          // 1. Viajando a Python API
    ArrivedAtAPI          // 2. Llegó a Python API
    ProcessingAtAPI       // 3. Procesando en Python API
    SendingToRabbit       // 4. Viajando a RabbitMQ
    ProcessingAtRabbit    // 5. Procesando en RabbitMQ
    SendingToWebsocket    // 6. Viajando a WebSocket API
    ProcessingAtWebsocket // 7. Procesando en WebSocket API
    SendingToFrontend     // 8. Viajando al Monitor
    Done                  // 9. Completado
    Error                 // X. Error en comunicación
)
```

#### **Transiciones de Estado**:
```
[Goroutine Worker]         [Game Loop FSM]
       ↓                          ↓
  SendingToAPI  ─────HTTP────→ ArrivedAtAPI
                                   ↓
                              ProcessingAtAPI (30 frames)
                                   ↓
                              SendingToRabbit
                                   ↓
                              ProcessingAtRabbit (30 frames)
                                   ↓
                              SendingToWebsocket
                                   ↓
                              ProcessingAtWebsocket (30 frames)
                                   ↓
                              SendingToFrontend
                                   ↓
                                 Done
```

**Características**:
- ✅ FSM actualizada a 60 FPS (game loop)
- ✅ Transiciones visuales suaves
- ✅ Timers de procesamiento (30 frames = 0.5s)
- ✅ Manejo de errores (estado `Error`)

**Código de FSM** (`game/game.go` - `handlePacketArrival()`):
```go
func (g *Game) handlePacketArrival(packet *state.PacketState) {
    switch packet.Status {
    case state.ProcessingAtAPI:
        if packet.ProcessingTimer > 0 {
            packet.ProcessingTimer--
        } else {
            packet.Status = state.SendingToRabbit
            packet.TargetX = iconRabbitX
            packet.TargetY = iconRabbitY
        }
    // ... más estados ...
    }
}
```

---

### **4. Patrón Fire-and-Forget con Callback Visual**
**Ubicación**: `simulation/workers.go` - `SendPOSTRequest()`

**Descripción**: Las goroutines se lanzan sin esperar respuesta inmediata (`fire-and-forget`), pero actualizan el estado visual como "callback".

**Flujo**:
```
Usuario Click
     ↓
startSimulation()
     ↓
go SendPOSTRequest() × 3  ← Fire (no esperamos aquí)
     ↓
return inmediatamente
     ↓
[En paralelo]
Goroutines ejecutan HTTP
     ↓
Actualizan estado visual ← Forget (callback visual)
     ↓
Game loop renderiza
```

**Características**:
- ✅ No bloquea UI
- ✅ Respuesta inmediata al usuario
- ✅ Actualización visual en tiempo real
- ✅ Simula latencia de red realista (500-1000ms)

---

### **5. Patrón Producer-Consumer Implícito**
**Ubicación**: `simulation/workers.go` (Producers) + `game/game.go` (Consumer)

**Descripción**: Las goroutines producen eventos de estado, el game loop los consume y visualiza.

**Roles**:
- **Producers (Goroutines)**: 
  - Generan datos de sensores
  - Envían HTTP requests
  - Actualizan estado de paquetes
  
- **Consumer (Game Loop)**:
  - Lee estado de paquetes
  - Actualiza FSM
  - Renderiza visualización

**Sincronización**:
- Sin canales explícitos
- Usa mutex como mecanismo de coordinación
- Game loop a 60 FPS actúa como consumidor periódico

---

## Mecanismos de Sincronización

### **1. Mutex (`sync.Mutex`)**
**Ubicación**: `state/state.go` - Campo `Mutex` en `VisualState`

**Propósito**: Proteger acceso concurrente al estado compartido

**Uso**:
```go
// Escritura
visState.Mutex.Lock()
visState.Packets[id] = newPacket
visState.Mutex.Unlock()

// Lectura con defer
visState.Mutex.Lock()
defer visState.Mutex.Unlock()
for _, packet := range visState.Packets {
    // ... operaciones seguras ...
}
```

**Buenas Prácticas Aplicadas**:
- ✅ `defer` para garantizar unlock
- ✅ Locks de corta duración
- ✅ Sin locks anidados (evita deadlocks)
- ✅ Consistencia: siempre lock antes de acceder

### **2. Timers de Simulación**
**Ubicación**: `simulation/workers.go` - Línea ~97

```go
// Simular latencia de red (500-1000ms)
time.Sleep(time.Duration(500+rand.Intn(500)) * time.Millisecond)
```

**Propósito**: Simular condiciones realistas de red

---

## Análisis de Rendimiento

### **Concurrencia vs Secuencial**

**Escenario**: Envío de 3 requests con latencia ~750ms cada uno

| Enfoque | Tiempo Total | Aprovechamiento CPU |
|---------|--------------|---------------------|
| Secuencial | ~2250ms (750×3) | Bajo (espera I/O) |
| **Concurrente (actual)** | **~750ms** | Alto (3 requests paralelos) |

**Mejora**: **3x más rápido** 🚀

### **Race Condition Prevention**

Sin mutex, podrían ocurrir:
- ❌ Pérdida de actualizaciones de paquetes
- ❌ Corrupción del mapa `Packets`
- ❌ Lecturas inconsistentes en UI

Con mutex:
- ✅ Todas las operaciones son atómicas
- ✅ Estado siempre consistente
- ✅ Sin race conditions (verificable con `go run -race .`)

---

## Cómo Verificar Concurrencia

### **1. Detectar Race Conditions**
```bash
go run -race .
```
Si hay problemas, Go mostrará warnings detallados.

### **2. Ver Goroutines Activas**
Agrega al código (temporal):
```go
import "runtime"

func (g *Game) Update() error {
    fmt.Printf("Goroutines activas: %d\n", runtime.NumGoroutine())
    // ...
}
```

### **3. Profiling de Concurrencia**
```bash
go build -o geova.exe
go tool trace trace.out
```

---

## Ventajas de Esta Arquitectura

1. **Escalabilidad**: Fácil agregar más sensores (más goroutines)
2. **Rendimiento**: I/O concurrente aprovecha mejor el CPU
3. **Realismo**: Simula hardware real que envía datos en paralelo
4. **Mantenibilidad**: Código limpio y separado por responsabilidades
5. **Visualización**: Usuario ve el paralelismo en tiempo real

---

## Mejoras Futuras Potenciales

### **1. Usar Channels en lugar de Mutex puro**
```go
type PacketUpdate struct {
    PacketID string
    NewStatus state.PacketStatus
}

updatesChan := make(chan PacketUpdate, 10)

// En worker:
updatesChan <- PacketUpdate{packetID, state.ArrivedAtAPI}

// En game loop:
select {
case update := <-updatesChan:
    // procesar sin mutex
default:
    // continuar
}
```

**Ventajas**: 
- Más idiomático en Go
- Mejor para alta concurrencia
- Menos contención de locks

### **2. Worker Pool con Límite**
```go
type WorkerPool struct {
    tasks chan Task
    workers int
}

// Limitar a N goroutines máximo
pool := NewWorkerPool(maxWorkers)
```

**Ventajas**:
- Control de recursos
- Evita crear demasiadas goroutines

### **3. Context para Cancelación**
```go
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

go SendPOSTRequestWithContext(ctx, url, data, state)
```

**Ventajas**:
- Timeout automático
- Cancelación coordinada
- Liberación de recursos

---

## Patrones de Concurrencia - Referencia

| Patrón | Usado | Ubicación |
|--------|-------|-----------|
| Worker Pool (Fan-Out) | ✅ | `game.go:130-145` |
| Shared State + Mutex | ✅ | `state.go` + `workers.go` |
| FSM Concurrente | ✅ | `game.go:150+` |
| Fire-and-Forget | ✅ | `workers.go:SendPOSTRequest()` |
| Producer-Consumer | ✅ | Workers → Game Loop |
| Channels | ❌ | (Mejora futura) |
| Select Statement | ❌ | (Mejora futura) |
| Context | ❌ | (Mejora futura) |
| WaitGroup | ❌ | (No necesario) |
| Once | ❌ | (No necesario) |

---

## Conclusión

Este proyecto es un **excelente ejemplo** de:
- ✅ Concurrencia básica bien implementada
- ✅ Sincronización correcta con Mutex
- ✅ Visualización de concurrencia en tiempo real
- ✅ Separación de responsabilidades (Workers vs UI)
- ✅ Código limpio y mantenible

**Ideal para**:
- Aprender Go concurrency
- Visualizar conceptos abstractos
- Simular sistemas distribuidos
- Proyecto educativo/portafolio

**Total de Goroutines por Simulación**: **3** (una por sensor)
**Total con Game Loop**: **4** (3 workers + 1 main loop)
