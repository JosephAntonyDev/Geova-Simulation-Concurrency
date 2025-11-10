# Guía para Agregar el Fondo

## Requisitos del Fondo

- **Nombre del archivo**: `background.png`
- **Ubicación**: Carpeta `images/`
- **Formato**: PNG (recomendado para transparencia)
- **Resolución recomendada**: 900×650 px o mayor
  - Si es menor, se estirará
  - Si es mayor, se escalará proporcionalmente

## Pasos para Agregar

1. **Preparar la imagen**:
   - Asegúrate de que tenga buena calidad
   - Preferible 900×650 px para evitar distorsión
   - Formato PNG o JPG

2. **Copiar al proyecto**:
   ```bash
   # Copia tu imagen al directorio images con el nombre correcto
   cp tu_fondo.png images/background.png
   ```

3. **Ejecutar el programa**:
   ```bash
   go run .
   ```

4. **Verificar**:
   - El fondo debería aparecer automáticamente
   - Si no existe el archivo, se usa el fondo gris oscuro por defecto
   - Revisa la consola por mensajes de advertencia

## Sugerencias de Diseño

### Para un fondo profesional:
- Usa colores oscuros (para que los elementos resalten)
- Evita patrones muy recargados
- Gradientes suaves funcionan bien
- Considera un tema tecnológico/industrial

### Herramientas Recomendadas:
- **GIMP**: Editor gratuito para crear/editar imágenes
- **Photoshop**: Editor profesional
- **Canva**: Plantillas online
- **Unsplash**: Fondos gratis de alta calidad

### Ejemplos de Fondos:
1. **Gradiente oscuro**: Azul oscuro a negro
2. **Grid tecnológico**: Líneas sutiles en fondo oscuro
3. **Espacio**: Estrellas y nebulosas
4. **Industrial**: Textura de metal oscuro

## Cambiar el Fondo Dinámicamente

Si quieres cambiar el fondo en tiempo de ejecución (avanzado):

```go
// En game/game.go, agrega un método:
func (g *Game) SetBackground(path string) error {
    img, _, err := ebitenutil.NewImageFromFile(path)
    if err != nil {
        return err
    }
    g.Assets.Background = img
    return nil
}
```

## Troubleshooting

### El fondo no aparece:
1. Verifica que el archivo se llame exactamente `background.png`
2. Verifica que esté en la carpeta `images/`
3. Revisa la consola por errores de carga
4. Verifica que el formato sea PNG o JPG válido

### El fondo se ve distorsionado:
- Usa una imagen de 900×650 px exactos
- O usa una relación de aspecto similar (9:6.5)

### El fondo es muy pesado:
- Optimiza la imagen (reduce tamaño del archivo)
- Usa PNG-8 en lugar de PNG-24 si no necesitas muchos colores
- Comprime con herramientas como TinyPNG

## Crear un Fondo Simple con Python (Script Opcional)

```python
from PIL import Image, ImageDraw

# Crear imagen 900x650
img = Image.new('RGB', (900, 650), color='#1a1a1a')
draw = ImageDraw.Draw(img)

# Agregar gradiente simple
for y in range(650):
    gray = int(26 + (y / 650) * 20)  # De #1a1a1a a #2e2e2e
    draw.line([(0, y), (900, y)], fill=(gray, gray, gray))

# Guardar
img.save('images/background.png')
print("✅ Fondo creado en images/background.png")
```

Para ejecutar:
```bash
pip install Pillow
python create_background.py
```
