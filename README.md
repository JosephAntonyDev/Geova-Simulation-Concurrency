# Geova Simulation - Concurrencia y Visualización

Simulación visual de un sistema de sensores IoT con concurrencia en Go, utilizando Ebitengine para renderizado en tiempo real.

## Descripción

Este proyecto simula el flujo de datos de 3 sensores (TFLuna, MPU6050, IMX477) enviando peticiones HTTP concurrentes a una API REST, con visualización en tiempo real del pipeline de procesamiento.

---

## Estructura del Proyecto

```
Geova-Simulation-Concurrency/
├── main.go              # Punto de entrada de la aplicación
├── assets/              # Gestión de recursos gráficos
│   └── assets.go        # Carga de sprites e imágenes
├── game/                # Lógica de juego y renderizado (modular)
│   ├── game.go          # Estructura principal y game loop
│   ├── config.go        # Constantes de posición y configuración
│   ├── input.go         # Manejo de entrada y lanzamiento de simulaciones
│   ├── fsm.go           # Máquina de estados de paquetes (FSM)
│   └── render.go        # Métodos de renderizado
├── simulation/          # Lógica de simulación y workers
│   ├── datatypes.go     # Estructuras de datos de sensores
│   └── workers.go       # Goroutines para peticiones HTTP
├── state/               # Estado compartido y sincronización
│   └── state.go         # Estado visual y de paquetes
└── images/              # Assets gráficos
    ├── background.png   # Fondo de la simulación (opcional)
    ├── geova_tilt_anim.png  # Animación del trípode (7 frames)
    └── ...              # Otros sprites
```

---

## Controles

| Control | Acción |
|---------|--------|
| ← → | Inclinar trípode (-15° a +15°) |
| Click en CREAR | Iniciar simulación continua |
| Click en DETENER | Detener simulación |
| F11 | Pantalla completa |

---

## Instalación y Ejecución

```bash
# Clonar repositorio
git clone https://github.com/JosephAntony37900/Geova-Simulation-Concurrency.git
cd Geova-Simulation-Concurrency

# Ejecutar
go run .

# O compilar
go build -o geova.exe
./geova.exe
```

**Requisitos**:
- Go 1.21+
- API REST corriendo en `localhost:8000`

---

## Componentes Principales

### 1. Main (`main.go`)
- Inicializa el generador de números aleatorios
- Carga todos los assets gráficos
- Crea el estado compartido
- Configura la ventana de Ebitengine (900×650)
- Lanza el game loop

### 2. Assets (`assets/assets.go`)
- **Responsabilidad**: Gestión centralizada de recursos gráficos
- **Funciones principales**:
  - `LoadAssets()`: Carga todos los sprites al iniciar
  - `loadSprite()`: Carga sprites requeridos
  - `loadSpriteOptional()`: Carga sprites opcionales (background)

### 3. Game (`game/`)

#### `game.go` - Estructura Principal
```go
type Game struct {
    Assets *assets.Assets
    State  *state.VisualState
    BotonRect image.Rectangle
    isBotonPressed bool
    animPacketCounter int
    animIconCounter int
}
```

#### `config.go` - Constantes
Centraliza posiciones de hardware, iconos, frontend y dimensiones de sprites.

#### `input.go` - Manejo de Entrada
- `handleInput()`: Detecta teclas y clicks
- `toggleSimulation()`: Inicia/detiene simulación continua
- `runContinuousSimulation()`: Loop de peticiones cada 2 segundos
- `sendBatchRequests()`: Lanza 3 goroutines por batch

#### `fsm.go` - Máquina de Estados
- `updatePacketFSM()`: Actualiza el ciclo de vida de paquetes
- `handlePacketArrival()`: Procesa llegadas a destinos
- `updateDashboard()`: Actualiza valores en pantalla

#### `render.go` - Renderizado
- `drawBackground()`: Fondo escalado o color sólido
- `drawTripode()`: Trípode animado según inclinación
- `drawTiltMeter()`: Medidor de inclinación
- `drawIcons()`: Iconos de backend (activos/inactivos)
- `drawPackets()`: Paquetes en movimiento
- `drawButton()`: Botón CREAR/DETENER
- `drawDashboard()`: Resultados de sensores

### 4. Simulation (`simulation/`)

#### `datatypes.go` - Estructuras de Datos
```go
type TFLunaData struct { /* Distancia láser */ }
type MPUData struct { /* Inclinación Roll/Pitch */ }
type IMXData struct { /* Nitidez de cámara */ }
```

#### `workers.go` - Goroutines HTTP
- `SendPOSTRequest()`: Envía datos a la API
- `GenerateRandom*Data()`: Genera datos aleatorios de sensores

### 5. State (`state/state.go`)
```go
type VisualState struct {
    Mutex   sync.Mutex
    Packets map[string]*PacketState
    
    PythonAPITimer    int
    RabbitMQTimer     int
    WebsocketAPITimer int
    
    DisplayDistancia   float64
    DisplayRoll        float64
    DisplayNitidez     float64
    CurrentTilt        float64
    SimulacionIniciada bool
    
    StopChan   chan struct{}  // Canal para detener simulación
    PacketID   int            // Contador de paquetes únicos
}
```

---

## Patrones de Concurrencia

### Resumen Ejecutivo

Este proyecto utiliza **3 goroutines concurrentes** por batch de simulación, con sincronización mediante **Mutex** y visualización en tiempo real. La simulación es continua hasta que el usuario la detiene.

### Goroutines por Batch: 3

| # | Nombre | Sensor | Color | Endpoint |
|---|--------|--------|-------|----------|
| 1 | `tfluna_N` | TF-Luna | Rojo | `/tfluna/sensor` |
| 2 | `mpu_N` | MPU6050 | Azul | `/mpu/sensor` |
| 3 | `imx_N` | IMX477 | Verde | `/imx477/sensor` |

### 1. Patrón Worker Pool (Fan-Out)

**Ubicación**: `game/input.go` - `sendBatchRequests()`

```go
func (g *Game) sendBatchRequests() {
    g.State.Mutex.Lock()
    tilt := g.State.CurrentTilt
    g.State.PacketID++
    id := g.State.PacketID
    g.State.Mutex.Unlock()

    go simulation.SendPOSTRequest(
        "http://localhost:8000/tfluna/sensor",
        simulation.GenerateRandomTFLunaData(),
        fmt.Sprintf("tfluna_%d", id), g.State, 180.0,
        color.RGBA{R: 255, G: 50, B: 50, A: 255},
    )
    go simulation.SendPOSTRequest(
        "http://localhost:8000/mpu/sensor",
        simulation.GenerateRandomMPUData(tilt),
        fmt.Sprintf("mpu_%d", id), g.State, 200.0,
        color.RGBA{R: 50, G: 150, B: 255, A: 255},
    )
    go simulation.SendPOSTRequest(
        "http://localhost:8000/imx477/sensor",
        simulation.GenerateRandomIMXData(),
        fmt.Sprintf("imx_%d", id), g.State, 220.0,
        color.RGBA{R: 50, G: 255, B: 50, A: 255},
    )
}
```

**Características**:
- 3 workers independientes por batch
- Ejecutan en paralelo sin bloquearse
- Batches cada 2 segundos mientras simulación activa

### 2. Patrón Toggle con Channel

**Ubicación**: `game/input.go` - `toggleSimulation()`

```go
func (g *Game) toggleSimulation() {
    g.State.Mutex.Lock()
    if g.State.SimulacionIniciada {
        close(g.State.StopChan)  // Señal de parada
        g.State.StopChan = nil
        g.State.SimulacionIniciada = false
        g.State.Mutex.Unlock()
        return
    }
    
    g.State.StopChan = make(chan struct{})
    stopChan := g.State.StopChan
    g.State.SimulacionIniciada = true
    g.State.Mutex.Unlock()
    
    go g.runContinuousSimulation(stopChan)
}

func (g *Game) runContinuousSimulation(stopChan chan struct{}) {
    ticker := time.NewTicker(2 * time.Second)
    defer ticker.Stop()
    
    g.sendBatchRequests()  // Primer batch inmediato
    
    for {
        select {
        case <-stopChan:
            return
        case <-ticker.C:
            g.sendBatchRequests()
        }
    }
}
```

### 3. Shared State con Mutex

**Ubicación**: `state/state.go` y `simulation/workers.go`

```go
// En Worker Goroutine
visState.Mutex.Lock()
packet := &state.PacketState{...}
visState.Packets[packetID] = packet
visState.Mutex.Unlock()

// ... HTTP request ...

visState.Mutex.Lock()
defer visState.Mutex.Unlock()
visState.Packets[packetID].Status = state.ArrivedAtAPI
```

**Zonas Críticas Protegidas**:
1. Creación de paquetes
2. Actualización de estado HTTP
3. Actualización de FSM en game loop
4. Toggle de simulación

### 4. FSM Concurrente

**Estados del Paquete**:
```
SendingToAPI → ArrivedAtAPI → ProcessingAtAPI →
SendingToRabbit → ProcessingAtRabbit →
SendingToWebsocket → ProcessingAtWebsocket →
SendingToFrontend → Done
```

---

## Diagrama de Flujo

```
┌─────────────────────────────────────────────────────────────────┐
│                         USUARIO                                  │
│                   Click en "CREAR"                              │
└─────────────────────────────────────────────────────────────────┘
                              ↓
┌─────────────────────────────────────────────────────────────────┐
│                    toggleSimulation()                            │
│                                                                  │
│  1. Crear canal StopChan                                        │
│  2. Lanzar goroutine runContinuousSimulation()                  │
└─────────────────────────────────────────────────────────────────┘
                              ↓
┌─────────────────────────────────────────────────────────────────┐
│               runContinuousSimulation()                          │
│                                                                  │
│  loop {                                                          │
│    select {                                                      │
│      case <-stopChan: return                                    │
│      case <-ticker.C: sendBatchRequests()                       │
│    }                                                             │
│  }                                                               │
└─────────────────────────────────────────────────────────────────┘
                              ↓
           ┌──────────┬──────────┬──────────┐
           ↓          ↓          ↓          
    ┌──────────┐ ┌──────────┐ ┌──────────┐
    │ Goroutine│ │ Goroutine│ │ Goroutine│
    │ TFLuna   │ │   MPU    │ │   IMX    │
    │   🔴     │ │   🔵     │ │   🟢     │
    └──────────┘ └──────────┘ └──────────┘
           ↓          ↓          ↓
           └──────────┴──────────┘
                      ↓
┌─────────────────────────────────────────────────────────────────┐
│              SendPOSTRequest() [EN PARALELO]                     │
│                                                                  │
│  1. Lock Mutex                                                  │
│  2. Crear PacketState inicial                                   │
│  3. Unlock Mutex                                                │
│  4. Sleep (500-1000ms) - Simular latencia                      │
│  5. HTTP POST a localhost:8000/[sensor]/sensor                 │
│  6. Lock Mutex                                                  │
│  7. Actualizar estado (ArrivedAtAPI o Error)                   │
│  8. Unlock Mutex                                                │
└─────────────────────────────────────────────────────────────────┘
                              ↓
┌─────────────────────────────────────────────────────────────────┐
│                      GAME LOOP (60 FPS)                          │
│                                                                  │
│  Update():                                                      │
│    - handleInput()                                              │
│    - updatePacketFSM()                                          │
│                                                                  │
│  Draw():                                                        │
│    - Fondo, Trípode, Iconos, Paquetes, Botón, Dashboard        │
└─────────────────────────────────────────────────────────────────┘
```

---

## Sincronización con Mutex

```
┌─────────────────────────────────────────────────────────────────┐
│                    MUTEX (sync.Mutex)                            │
│                   Protege: VisualState                           │
└─────────────────────────────────────────────────────────────────┘
                              ↓
        ┌─────────────────────┴─────────────────────┐
        ↓                                           ↓
┌──────────────────┐                      ┌──────────────────┐
│  WRITERS         │                      │  READERS         │
│  (Goroutines)    │                      │  (Game Loop)     │
├──────────────────┤                      ├──────────────────┤
│ 1. Lock()        │                      │ 1. Lock()        │
│ 2. Write state   │                      │ 2. Read state    │
│ 3. Unlock()      │                      │ 3. Unlock()      │
└──────────────────┘                      └──────────────────┘
```

---

## Timeline de Ejecución

```
Tiempo (ms)    Evento
──────────────────────────────────────────────────────────────────
0              Usuario presiona CREAR
1              toggleSimulation() ejecuta
2              runContinuousSimulation() inicia
3              Primer batch: 3 goroutines lanzan
4-5            Goroutines crean PacketState (con Mutex)
6-506          Goroutines duermen (simulan latencia de red)
507-1007       HTTP POST ejecuta en paralelo
1008           Primera respuesta llega → ArrivedAtAPI
2000           Segundo batch se dispara (ticker)
...            Continúa cada 2 segundos
N              Usuario presiona DETENER
N+1            close(StopChan) → goroutine controller termina
```

---

## Sistema de Animación

### Trípode Geova
- **Sprite**: `geova_tilt_anim.png` (896×128 px)
- **Frames**: 7 frames horizontales de 128×128 px
- **Mapeo de inclinación**:
  - Frame 0: ≤ -12.5° (muy inclinado izquierda)
  - Frame 1-2: Inclinación izquierda
  - Frame 3: Nivelado (-2.5° a +2.5°)
  - Frame 4-5: Inclinación derecha
  - Frame 6: ≥ +12.5° (muy inclinado derecha)

### Paquetes de Datos
- **Sprite**: `data_packet_anim.png`
- **Animación**: 6 frames ciclando
- **Colores distintivos**: Rojo (TFLuna), Azul (MPU), Verde (IMX)

### Iconos Backend
- **Idle**: Sprites estáticos (64×64)
- **Activos**: 6 frames de animación (384×64)
- **Trigger**: Timer > 0 cuando procesan datos

---

## Análisis de Rendimiento

### Concurrencia vs Secuencial

| Enfoque | Tiempo por Batch | Mejora |
|---------|------------------|--------|
| Secuencial | ~2250ms (750×3) | - |
| **Concurrente** | **~750ms** | **3x** |

### Verificar Race Conditions

```bash
go run -race .
```

Si hay problemas, Go mostrará warnings detallados.

---

## Fondo Personalizado

### Agregar Fondo

1. Preparar imagen (recomendado: 900×650 px, PNG)
2. Copiar a `images/background.png`
3. El fondo se carga automáticamente al iniciar

Si no existe el archivo, se usa fondo gris oscuro por defecto.

### Sugerencias de Diseño

- Colores oscuros (para que elementos resalten)
- Evitar patrones recargados
- Gradientes suaves funcionan bien

### Crear Fondo Simple con Python

```python
from PIL import Image, ImageDraw

img = Image.new('RGB', (900, 650), color='#1a1a1a')
draw = ImageDraw.Draw(img)

for y in range(650):
    gray = int(26 + (y / 650) * 20)
    draw.line([(0, y), (900, y)], fill=(gray, gray, gray))

img.save('images/background.png')
```

---

## Patrones de Concurrencia - Resumen

| Patrón | Usado | Ubicación |
|--------|-------|-----------|
| Worker Pool (Fan-Out) | ✅ | `input.go:sendBatchRequests()` |
| Shared State + Mutex | ✅ | `state.go` + `workers.go` |
| FSM Concurrente | ✅ | `fsm.go` |
| Channel para Toggle | ✅ | `input.go:toggleSimulation()` |
| Select Statement | ✅ | `input.go:runContinuousSimulation()` |
| Ticker (time.Ticker) | ✅ | `input.go:runContinuousSimulation()` |
| Fire-and-Forget | ✅ | `workers.go:SendPOSTRequest()` |
| Producer-Consumer | ✅ | Workers → Game Loop |

---

## Ventajas de Esta Arquitectura

1. **Escalabilidad**: Fácil agregar más sensores
2. **Rendimiento**: I/O concurrente (3x más rápido)
3. **Realismo**: Simula hardware real con latencia
4. **Visual**: Usuario ve concurrencia en tiempo real
5. **Responsive**: UI nunca se bloquea
6. **Modular**: Código separado por responsabilidades
7. **Control**: Toggle para iniciar/detener en cualquier momento

---

## Mejoras Futuras Potenciales

1. **Context para Cancelación**: Timeout automático de requests
2. **Worker Pool con Límite**: Control de goroutines máximas
3. **Métricas**: Contador de requests exitosos/fallidos
4. **Configuración**: Intervalo de peticiones configurable
5. **RWMutex**: Para mejor rendimiento de lecturas

---

## Licencia

MIT License

---

**Goroutines por Batch**: 3 (una por sensor)  
**Total con Control Loop**: 4+ (3 workers × batches + 1 main loop + 1 controller)
