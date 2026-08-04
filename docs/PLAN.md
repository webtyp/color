---
PLAN: "Recoger el color del ecosistema en esta pieza"
TAG: v0.1.0
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
| `pdf/color.go:11` | `Color.parse()` → hex a RGB, con `tinywasm/fmt` | producción |
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

### 3.1 Migrar `tinywasm/pdf`

`pdf` mantiene su propio `Color` en `color.go`. Debe importar éste.

- **Borrar el tipo `Color` y `parse()`** de `pdf/color.go`.
- **`pdf.Color` pasa a ser un alias** — `type Color = color.Color` — para no tocar las
  123 referencias del repo de golpe. Es un alias de tipo, no un envoltorio: no añade
  método ni comportamiento, así que no es el *fork with a friendlier name* que la regla
  prohíbe. Si se prefiere el corte limpio, se sustituyen las 123 y se elimina el alias;
  decidir explícitamente.
- **`Theme` NO se mueve.** Lleva `Sizes` (H1/H2/H3 en puntos), `Spacing`, `Page` y
  `Margin` en **milímetros**: es el tema de un documento PDF, no color. Que contenga
  colores no lo convierte en uno, y traérselo aquí metería tamaños de página en una
  librería de color (principio 5).
- Ficheros afectados: `element.go`, `table_layout.go`, `document.go`,
  `examples/contract_layout/main.go`. `fpdf/png.go` y `fpdf/spotcolor.go` usan un
  `color` distinto del stdlib — **revisar antes de tocar**, el nombre colisiona.

### 3.2 Recoger la ciencia del color de `css`

Sacar de `css/contrast_test.go` a producción, aquí:

```go
func (c Color) Luminance() float64
func Contrast(a, b Color) float64
```

`css` pasa a importarlas en lugar de tener su copia en un test. Eso convierte su
auditoría de contraste en algo que cualquier consumidor puede ejecutar, no un detalle
interno de su suite.

**Cuidado con el sentido de la dependencia:** `css` es `//go:build !wasm` y esta pieza no
tiene tag. `css → color` es correcto. Al revés sería imposible y además absurdo.

### 3.3 Cerrar la carencia documentada de `css`

`css/AGENTS.md` §6 dice que `resolveColor` no puede evaluar `color-mix()`, y por eso
`AllPairs()` excluye los pares cuyo fondo es compuesto — un hueco declarado en la
garantía de accesibilidad del ecosistema.

Evaluar `color-mix(in oklab, …)` es trabajo de esta pieza. Con `Mix` aquí, esos pares
entran en la auditoría y el hueco se cierra donde le toca.

No es obligatorio en este plan, pero **es la razón por la que la pieza merece existir**
más allá de deduplicar un parser: hay una carencia real esperando dueño.

---

## 4. Lo que esta pieza nunca hace

- **No emite CSS.** `var(--x)`, `light-dark()` y los tokens son de `tinywasm/css`.
- **No dibuja.** Convertir un color a la instrucción de relleno de un PDF es de `pdf`.
- **No trae paleta.** No hay colores de marca aquí; los declara cada producto.
- **Sin mapas.** Vetados en el ecosistema por lo que añaden al binario TinyGo.

---

## 5. Verificación

1. `grep -rn "Int(16)\|ParseUint" pdf/ css/` no encuentra ningún parser de hex: queda
   uno, y está aquí.
2. `pdf` compila contra este módulo y su suite sigue verde.
3. `css/contrast_test.go` usa `Contrast` de aquí y no declara su propio cálculo.
4. Ningún archivo de este paquete lleva build tag, y compila bajo `GOOS=js GOARCH=wasm`.
5. `grep -rn "map\[" *.go` vacío.
6. No aparece ningún tamaño, margen ni unidad de página en la superficie pública.
7. `gotest` en verde en los tres módulos.

`docs/ARCHITECTURE.md` y `docs/SPECS.md` se escriben en el mismo commit; `README.md` los
indexa.
