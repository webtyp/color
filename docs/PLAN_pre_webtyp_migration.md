---
PLAN: "Recoger el color del ecosistema en esta pieza"
TAG: v0.1.0
STATUS: running
SESSION: 1171274572088145272
---

# PLAN — Recoger el color del ecosistema en esta pieza

## Antes de escribir código: lee [CONSTRUCTION_HARNESS.md](CONSTRUCTION_HARNESS.md)

**Es vinculante, no orientativo.** Los principios que gobiernan este trabajo:

| # | Principio | Cómo se aplica aquí |
|---|---|---|
| 4 | One way to do each thing | Hay dos parseadores de hex en el ecosistema. Queda uno. |
| 5 | Minimal surface | Esta pieza es **color**. Si aparece un tamaño en milímetros o una regla CSS, se salió de su responsabilidad. |
| 9 | Lego pieces, never forks | Un consumidor que necesita luminancia no la reescribe: falta aquí y se añade aquí. |

---

## 1. Por qué existe

El color está escrito dos veces, con dos implementaciones distintas del mismo cálculo:

| Dónde | Qué hace | Estado |
|---|---|---|
| `pdf/color.go:11` | `Color.parse()` → hex a RGB, con `webtyp/fmt` | producción |
| `css/contrast_test.go:20-22` | `strconv.ParseUint(hex[0:2], 16, 32)` × 3 | **dentro de un test** |

Y en ese mismo archivo de test viven `relativeLuminance` y `contrastRatio`: ciencia del
color de verdad, atrapada donde nadie más puede usarla. `css/AGENTS.md` §6 documenta
además que su auditoría de contraste **no sabe evaluar `color-mix()`** — una carencia que
es de una librería de color, no de una de CSS.

Esta pieza no tiene build tag, así que cruza a WASM. Es deliberado: un color es un valor
que el frontend maneja de verdad, y aquí no hay nada más.

---

## 2. Qué ya está hecho

```go
type Color string
func (c Color) RGB() (r, g, b int, err error)
```

Trasladado de `pdf/color.go`, con tres cambios:

- **`parse()` → `RGB()`, exportado.** Era privado; el propósito de la pieza es que otros
  lo usen.
- **Las tres ramas repetidas colapsan en `channels()`.** El original repetía el mismo
  bloque de tres `Convert().Int(16)` con su error, dos veces.
- **Tests con el porqué de cada caso**, incluida la comprobación de que `#abc` y
  `#aabbcc` se descomponen igual.

`gotest` en verde, 87,5% de cobertura.

---

## 3. Lo que falta

### 3.1 Migrar `webtyp/pdf` — **no es trabajo de este repo**

`pdf` mantiene su propio `Color` y debe pasar a importar éste, pero ese cambio ocurre
**en el repositorio de `pdf`**, no aquí. Está especificado en `pdf/docs/PLAN.md` §5
(Defecto 4): son 20 referencias en cuatro archivos, se sustituyen por `color.Color`
**sin alias**, y allí queda anotado qué ficheros de `fpdf/` son falsos positivos al
buscar `Color`.

Se deja anotado aquí sólo para que se vea la cadena completa. **Un `codejob` sobre este
módulo no puede tocar `pdf`**, y confundirlo lleva a un plan que nadie puede ejecutar.

### 3.2 La superficie que falta

Inventariado contra lo que los consumidores necesitan de verdad — `pdf` sólo usa `RGB()`
(comprobado: todos sus usos son `parse()` → `SetFillColor(r,g,b)`), y todo lo demás lo
pide la auditoría de contraste de `css`:

```go
// Luminancia relativa WCAG. Aplica la curva sRGB, no es un promedio de canales.
func (c Color) Luminance() float64

// Ratio de contraste WCAG entre dos colores: de 1 (idénticos) a 21 (negro/blanco).
func Contrast(a, b Color) float64

// Mezcla en el espacio Oklab, que es lo que significa `color-mix(in oklab, …)`.
func Mix(a, b Color, pct float64) Color

// Compone un color translúcido sobre un fondo opaco.
func (c Color) Over(backdrop Color) Color
```

Las dos primeras salen de `css/contrast_test.go`, donde hoy viven atrapadas. Las dos
últimas no existen en ningún sitio y son las que cierran §3.3.

#### `Mix` es en Oklab, y eso no es un detalle

`color-mix(in oklab, …)` **no** interpola en sRGB. El camino es
sRGB → lineal → LMS → Oklab → interpolar → volver. Una interpolación ingenua de canales
RGB da un color distinto, y para una auditoría de accesibilidad eso significa validar un
color que el navegador nunca pinta.

Si se implementa mal, los tests pasan y la garantía es falsa. **Verificar contra valores
que produzca un navegador real**, no contra los que devuelva la propia implementación.

#### `Over` existe por un caso concreto

`ColorSelection` es `color-mix(in oklab, var(--color-primary), transparent 85%)`
(`css/catalog.go:25`). Mezclar con `transparent` produce un color **con alfa**, y el
contraste de un color translúcido no está definido sin saber sobre qué se pinta.

Así que `Contrast` sobre ese valor no tiene respuesta: primero hay que componerlo sobre
su fondo real con `Over`, y auditar el resultado. Sin esto, `SurfaceSelected` no puede
entrar en `AllPairs()` aunque `Mix` exista.

Implica que `Color` debe admitir alfa —`#RRGGBBAA` y `transparent`—, hoy no soportado.

#### Dónde está la frontera con `css`

| Concern | Dueño |
|---|---|
| Matemática del color: parsear, luminancia, contraste, mezclar, componer | **`color`** |
| Sintaxis CSS: `var()`, `light-dark()`, resolver un token | `css` |

`css` parsea la expresión `color-mix(in oklab, A, B 8%)` —eso es sintaxis suya— y llama
a `Mix` para el cálculo. Esta pieza nunca ve un `var()`.

**Cuidado con el sentido de la dependencia:** `css` es `//go:build !wasm` y esta pieza no
tiene tag. `css → color` es correcto. Al revés sería imposible y además absurdo.

### 3.3 Cerrar la carencia documentada de `css`

`css/AGENTS.md` §6 dice que `resolveColor` no puede evaluar `color-mix()`, y por eso
`AllPairs()` excluye los pares cuyo fondo es compuesto — un hueco declarado en la
garantía de accesibilidad del ecosistema.

Con `Mix` y `Over` aquí, esos dos pares entran en la auditoría:

| Par | Expresión | Qué hace falta |
|---|---|---|
| `SurfaceSunken` | `color-mix(in oklab, var(--color-surface), var(--color-on-surface) 8%)` | `Mix` — dos colores opacos |
| `SurfaceSelected` | `color-mix(in oklab, var(--color-primary), transparent 85%)` | `Mix` **y** `Over` — el resultado lleva alfa |

**Es la razón por la que la pieza merece existir** más allá de deduplicar un parser: hay
una carencia real, declarada y con dueño ausente.

---

## 4. Lo que esta pieza nunca hace

- **No emite CSS.** `var(--x)`, `light-dark()` y los tokens son de `webtyp/css`.
- **No dibuja.** Convertir un color a la instrucción de relleno de un PDF es de `pdf`.
- **No trae paleta.** No hay colores de marca aquí; los declara cada producto.
- **Sin mapas.** Vetados en el ecosistema por lo que añaden al binario TinyGo.

---

## 5. Verificación

Lo que este repo puede verificar por sí mismo:

1. **`Contrast` contra valores conocidos**, no contra lo que devuelva la implementación:
   negro/blanco = 21, un color consigo mismo = 1. Y `#e8a33d` sobre `#1C1C1E` ≈ 7,89,
   que es un par real del catálogo de `css`.
2. **`Mix` contra un navegador, no contra sí misma.** Es el punto donde una
   implementación en sRGB pasa sus propios tests y produce el color equivocado. Al menos
   un caso comprobado con lo que pinta un navegador real para
   `color-mix(in oklab, …)`.
3. **`Over` sobre un fondo conocido**: un color con alfa 0 deja el fondo intacto; con
   alfa 1 lo tapa por completo.
4. `Color` acepta `#RRGGBBAA` y `transparent`, y `RGB()` sigue funcionando igual para
   los colores opacos de siempre.
5. Ningún archivo de este paquete lleva build tag, y compila bajo `GOOS=js GOARCH=wasm`.
6. `grep -rn "map\[" *.go` vacío.
7. No aparece ningún tamaño, margen ni unidad de página en la superficie pública: si
   entra un milímetro, la pieza se salió de su responsabilidad. Tampoco un `var(` ni un
   `light-dark(`: eso es sintaxis CSS, no color.
8. `gotest` en verde.

Lo que sólo se puede verificar cuando los consumidores migren, y que por tanto **no
bloquea este plan**:

9. `grep -rn "Int(16)\|ParseUint" pdf/ css/` no encuentra ningún parser de hex.
10. `css/contrast_test.go` usa `Contrast` de aquí en vez de su propio cálculo.
11. `AllPairs()` de `css` incluye `SurfaceSunken` y `SurfaceSelected`: la carencia de
    §3.3 quedó cerrada de verdad, no sólo con la función disponible.

`docs/ARCHITECTURE.md` y `docs/SPECS.md` se escriben en el mismo commit; `README.md` los
indexa.
