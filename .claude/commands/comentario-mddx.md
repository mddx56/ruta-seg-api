Define el estilo de comentarios de código para este proyecto: una sola línea, texto simple, directo y lo más breve posible, siempre en español.

## Reglas

1. **Una sola línea.** Nunca bloques de varias líneas ni docstrings multilínea. Si no entra en una línea, es que sobra contenido: recórtalo.
2. **Texto simple y directo.** Sin jerga innecesaria, sin rodeos, sin explicar lo obvio que ya dice el nombre de la variable/función.
3. **Lo más breve posible.** Prioriza la frase más corta que siga siendo clara.
4. **Siempre en español**, incluso si el código, los identificadores o el resto del repo están en inglés.
5. Sigue aplicando la regla general: comentar solo cuando aporta algo que el código no dice por sí solo (el porqué, no el qué). Esta skill define el *formato* del comentario cuando decides que hace falta uno, no obliga a comentar todo.

## Ejemplos

Mal (multilínea, en inglés, explica el qué):
```go
// This function calculates the total price of the order
// by summing all item prices and applying the discount
// if the user has a valid coupon code.
func CalcularTotal(...) {}
```

Bien (una línea, español, explica el porqué):
```go
// Redondeo hacia abajo: la pasarela de pago rechaza centavos sueltos.
total := math.Floor(subtotal)
```
