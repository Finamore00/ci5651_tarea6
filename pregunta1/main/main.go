package main

import (
	"fmt"
	"math/rand"
)

/*
Solución en Golang de intercambio múltiple de elementos en un arreglo en tiempo logarítmico.
Pregunta 1 de la 6ta tarea de Diseño de Algoritmos I (CI5651). Universidad Simón Bolívar,
trimestre Ene-Mar 2024.

Autor: Santiago Finamore
Carnet: 18-10125
*/

/*
Implementación de estructura de datos treap implícito y funciones asociadas
*/
type TreapNode struct {
	value  int32
	weight int32
	size   int32
	left   *TreapNode
	right  *TreapNode
}

func size(n *TreapNode) int32 {
	if n == nil {
		return 0
	} else {
		return n.size
	}
}

func split(t *TreapNode, left **TreapNode, right **TreapNode, val int32) {
	if t == nil {
		*left = nil
		*right = nil
		return
	}

	if size(t.left) < val {
		split(t.right, &(t.right), right, val-size(t.left)-1)
		*left = t
	} else {
		split(t.left, left, &(t.left), val)
		*right = t
	}
	t.size = 1 + size(t.left) + size(t.right)
}

func merge(t **TreapNode, left *TreapNode, right *TreapNode) {
	if left == nil {
		*t = right
		return
	}
	if right == nil {
		*t = left
		return
	}

	if left.weight < right.weight {
		merge(&(left.right), left.right, right)
		*t = left
	} else {
		merge(&(right.left), left, right.left)
		*t = right
	}
	(*t).size = 1 + size((*t).left) + size((*t).right)
}

func inOrderTraversal(root *TreapNode) {

	var inOrderTraversalAux func(r *TreapNode)

	inOrderTraversalAux = func(r *TreapNode) {
		if r == nil {
			return
		}
		inOrderTraversalAux(r.left)
		fmt.Printf("%d, ", r.value)
		inOrderTraversalAux(r.right)
	}

	inOrderTraversalAux(root)
	fmt.Printf("\n")
}

func main() {
	arr := []int32{0, 1, 2, 3, 4, 5, 6, 7}
	//Go no tiene tuplas f
	actions := [][]int32{
		{1, 5},
		{2, 4},
		{4, 6},
		{2, 6},
	}
	var root *TreapNode

	//Poblar el arbol
	for _, e := range arr {
		merge(&root, root, &TreapNode{
			value:  e,
			weight: rand.Int31n(100),
			size:   1,
			left:   nil,
			right:  nil,
		})
	}

	for _, elem := range actions {
		var i, j int32 = elem[0], elem[1]
		subArrSize := min(j-i, int32(len(arr))-j)
		//Se realiza la separación del arreglo
		var a, b, c, d, e, f, g, h *TreapNode
		split(root, &a, &b, i)
		split(b, &c, &d, subArrSize)
		split(d, &e, &f, j-(i+subArrSize))
		split(f, &g, &h, subArrSize)

		//Se reensamblan las piezas en el order deseado
		merge(&root, a, g)
		merge(&root, root, e)
		merge(&root, root, c)
		merge(&root, root, h)

	}

	inOrderTraversal(root)
}
