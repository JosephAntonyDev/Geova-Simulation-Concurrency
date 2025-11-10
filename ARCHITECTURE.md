# Arquitectura del Proyecto Geova Simulation

## 📁 Estructura del Proyecto

```
Geova-Simulation-Concurrency/
├── main.go              # Punto de entrada de la aplicación
├── assets/              # Gestión de recursos gráficos
│   └── assets.go        # Carga de sprites e imágenes
├── game/                # Lógica de juego y renderizado
│   └── game.go          # Motor de juego principal
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

## 🏗️ Componentes Principales

### 1. **Main (`main.go`)**
- Inicializa el generador de números aleatorios
- Carga todos los assets gráficos
- Crea el estado compartido
- Configura la ventana de Ebitengine
- Lanza el game loop

### 2. **Assets (`assets/assets.go`)**
- **Responsabilidad**: Gestión centralizada de recursos gráficos
- **Funciones principales**:
  - `LoadAssets()`: Carga todos los sprites al iniciar
  - `loadSprite()`: Carga sprites requeridos (falla si no existe)
  - `loadSpriteOptional()`: Carga sprites opcionales (retorna nil si no existe)

#### Assets Disponibles:
- **Fondo**: `Background` (opcional)
- **Hardware**: Trípode Geova con animación de inclinación
- **Backend**: Iconos de Python API, RabbitMQ, WebSocket API (activos e inactivos)
- **Frontend**: Monitor, gauges, barras de progreso
- **UI**: Botones y paquetes de datos animados

### 3. **Game (`game/game.go`)**
- **Responsabilidad**: Renderizado y lógica de juego
- **Estructura**:
  ```go
  type Game struct {
      Assets *assets.Assets      // Referencias a sprites
      State  *state.VisualState  // Estado compartido thread-safe
      BotonRect image.Rectangle   // Área del botón CREAR
      isBotonPressed bool         // Estado del botón
      animPacketCounter int       // Contador de animación
      animIconCounter int         // Contador de animación
  }
  ```

#### Métodos Principales:
- **`Update()`**: Actualiza estado por frame (60 FPS)
  - Maneja input del usuario
  - Actualiza FSM de paquetes
  - Incrementa contadores de animación

- **`Draw()`**: Renderiza la escena
  - Dibuja fondo (si existe)
  - Dibuja elementos en orden de profundidad
  - Muestra instrucciones

#### Métodos de Renderizado:
- `drawBackground()`: Dibuja fondo escalado o color sólido
- `drawTripode()`: Dibuja trípode animado según inclinación
- `drawTiltMeter()`: Muestra medidor de inclinación superior
- `drawIcons()`: Dibuja iconos de backend (activos/inactivos)
- `drawPackets()`: Renderiza paquetes en movimiento
- `drawButton()`: Dibuja botón CREAR
- `drawDashboard()`: Muestra resultados de sensores

#### Helpers:
- `getTripodeFrame()`: Calcula frame según inclinación (-15° a +15°)

### 4. **Simulation (`simulation/`)**
- **`datatypes.go`**: Define estructuras de datos de sensores
  - `TFLunaData`: Distancia del sensor láser
  - `MPUData`: Datos de inclinación (Roll, Pitch, Yaw)
  - `IMXData`: Datos de cámara (Nitidez, Brillo)

- **`workers.go`**: Goroutines para envío de datos
  - `SendPOSTRequest()`: Envía datos de sensores a la API
  - Genera paquetes visuales con colores distintivos
  - Maneja errores de red

### 5. **State (`state/state.go`)**
- **Responsabilidad**: Estado compartido thread-safe
- **Sincronización**: Usa `sync.Mutex` para acceso concurrente
- **Estructuras**:
  ```go
  type VisualState struct {
      Mutex sync.Mutex
      Packets map[string]*PacketState  // Paquetes en tránsito
      CurrentTilt float64               // Inclinación actual
      DisplayDistancia float64          // Último valor de distancia
      DisplayNitidez float64            // Último valor de nitidez
      DisplayRoll float64               // Último valor de roll
      SimulacionIniciada bool           // Estado de simulación
      // Timers para animaciones
      PythonAPITimer int
      RabbitMQTimer int
      WebsocketAPITimer int
  }
  ```

#### Máquina de Estados (FSM) de Paquetes:
```
SendingToAPI → ArrivedAtAPI → ProcessingAtAPI → 
SendingToRabbit → ProcessingAtRabbit → 
SendingToWebsocket → ProcessingAtWebsocket → 
SendingToFrontend → Done
```

## 🎨 Sistema de Animación

### Trípode Geova
- **Sprite**: `geova_tilt_anim.png` (896×128 px)
- **Frames**: 7 frames horizontales de 128×128 px
- **Mapeo de inclinación**:
  - Frame 0: ≤ -12.5° (muy inclinado izquierda)
  - Frame 1: -12.5° a -7.5°
  - Frame 2: -7.5° a -2.5°
  - Frame 3: -2.5° a +2.5° (nivelado)
  - Frame 4: +2.5° a +7.5°
  - Frame 5: +7.5° a +12.5°
  - Frame 6: ≥ +12.5° (muy inclinado derecha)

### Paquetes de Datos
- **Sprite**: `data_packet_anim.png`
- **Animación**: 6 frames ciclando cada 6 frames del juego
- **Colores distintivos**:
  - 🔴 Rojo: TFLuna (distancia)
  - 🔵 Azul: MPU (inclinación)
  - 🟢 Verde: IMX477 (nitidez)

### Iconos Backend
- **Idle**: Sprites estáticos (64×64)
- **Activos**: 6 frames de animación (384×64)
- **Trigger**: Timer > 0 cuando procesan datos

## 🔧 Configuración

### Constantes Principales (`game/game.go`)
```go
const (
    // Posiciones
    tripodeX, tripodeY = 80.0, 200.0
    iconPythonX = 250.0
    iconRabbitX = 400.0
    iconWebsocketX = 550.0
    monitorX = 620.0
    
    // Animación
    packetSpeed = 3.0      // px/frame
    processingDelay = 30   // frames (0.5s a 60 FPS)
    
    // Sprites
    tripodeFrameWidth = 128
    tripodeFrameHeight = 128
    tripodeFrameCount = 7
)
```

### Tamaño de Ventana (`main.go`)
```go
const (
    windowWidth = 900
    windowHeight = 650
)
```

## 🎮 Controles

- **← →**: Inclinar trípode antes de crear simulación (-15° a +15°)
- **Click en CREAR**: Iniciar nueva simulación
- **F11**: Alternar pantalla completa

## 🔄 Flujo de Ejecución

1. **Inicialización**:
   - Cargar assets
   - Crear estado compartido
   - Configurar ventana

2. **Game Loop (60 FPS)**:
   - `Update()`: Procesar input y actualizar estado
   - `Draw()`: Renderizar escena

3. **Simulación**:
   - Usuario ajusta inclinación con teclado
   - Click en CREAR lanza 3 goroutines
   - Cada goroutine envía datos a Python API
   - Paquetes viajan visualmente por el pipeline
   - Resultados se muestran en dashboard

4. **Concurrencia**:
   - 3 goroutines simultáneas (TFLuna, MPU, IMX477)
   - Estado compartido protegido con mutex
   - FSM de paquetes actualizada thread-safe

## 📝 Agregar un Fondo

1. Coloca tu imagen en `images/background.png`
2. El fondo se cargará automáticamente (opcional)
3. Se escalará para llenar la ventana (900×650)
4. Si no existe, usa fondo gris oscuro por defecto

## 🚀 Mejoras Futuras

- [ ] Agregar más sensores
- [ ] Dashboard interactivo
- [ ] Gráficas en tiempo real
- [ ] Configuración de velocidad de simulación
- [ ] Exportar logs de sensores
- [ ] Modo oscuro/claro
