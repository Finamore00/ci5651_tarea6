/*
Solución en Golang de cálculo de cuantificadores universal y existencial para
valores booleanos sobre las aristas de un árbol. Pregunta número 2 de la 6ta
tarea de Diseño de Algoritmos I (CI5651). Universidad Simón Bolívar, trimestre
Ene-Mar 2024

Autor: Santiago Finamore
Carnet: 18-10125
*/

package main

import (
	"fmt"
	"math/rand"
)

const N = 13

/*
Implementación de estructura de datos árbol y del algoritmo
Heavy Light Decomposition
*/
type TreeNode struct {
	value       int
	dfsParent   *TreeNode
	chainChild  *TreeNode
	depth       int
	subTreeSize int
	chain       *Chain
}

type Chain struct {
	topNode        *TreeNode
	accumMapAll    map[int]bool
	accumMapExists map[int]bool
}

type Edge struct {
	node   *TreeNode
	weight bool
}

type Tree struct {
	adj map[*TreeNode][]Edge
}

func newTree(n int) Tree {
	return Tree{
		adj: make(map[*TreeNode][]Edge, n),
	}
}

func newNode(n int) TreeNode {
	return TreeNode{
		value:       n,
		dfsParent:   nil,
		chainChild:  nil,
		depth:       -1,
		subTreeSize: -1,
		chain:       nil,
	}
}

/*Se asume que los valores a y b son nodos válidos del árbol*/
func (t Tree) addEdge(a *TreeNode, b *TreeNode, w bool) {
	t.adj[a] = append(t.adj[a], Edge{b, w})
	t.adj[b] = append(t.adj[b], Edge{a, w})
}

/*
Retorna el objeto Edge que conecta a los nodos a y b dentro del arbol.
Si el arista no existe retorna nil
*/
func (t Tree) getEdge(a *TreeNode, b *TreeNode) *Edge {
	edgeLs := t.adj[a]

	for _, edge := range edgeLs {
		if edge.node.value == b.value {
			return &edge
		}
	}

	return nil
}

/*
Cuando se hace DFS determinamos los padres de cada nodo, y sus
profundidades para facilitar hallar el LCA.
*/
func (t Tree) dfs(rootNode *TreeNode) {
	//Funcion auxiliar recursiva
	var dfsRec func(prev *TreeNode, current *TreeNode, depth int)
	dfsRec = func(prev *TreeNode, current *TreeNode, depth int) {
		current.dfsParent = prev
		current.depth = depth + 1
		current.subTreeSize = 0
		for _, adj := range t.adj[current] {
			if adj.node == current.dfsParent {
				continue
			}
			dfsRec(current, adj.node, depth+1)
			current.subTreeSize += adj.node.subTreeSize
		}
		current.subTreeSize += 1
	}
	rootNode.depth = 0
	rootNode.dfsParent = nil
	dfsRec(nil, rootNode, 0)
}

/*
Encuentra el ancestro común más bajo entre los nodos a y b. Se asume
que los nodos ingresados son correctos y existen en el árbol. Se asume
también que ya se ejecutó DFS sobre el árbol que contiene a los nodos.
*/
func (t Tree) LCA(a *TreeNode, b *TreeNode) *TreeNode {
	var lowerNode, higherNode *TreeNode

	if a.depth > b.depth {
		higherNode = b
		lowerNode = a
	} else {
		higherNode = a
		lowerNode = b
	}

	//Compensar por diferencia de alturas
	hDiff := lowerNode.depth - higherNode.depth
	for i := 0; i < hDiff; i += 1 {
		lowerNode = lowerNode.dfsParent
	}

	//Subir por el arbol hasta que tengan el mismo padre
	for lowerNode.dfsParent != higherNode.dfsParent {
		lowerNode = lowerNode.dfsParent
		higherNode = higherNode.dfsParent
	}

	return lowerNode.dfsParent //Regresar el padre común (LCA)
}

/*
Construye las cadenas pesadas del árbol. Se asume que antes de
ejecutar este método sobre el árbol ya se efectuó el precondicio-
namiento pertinente ejecutando el método dfs.
*/
func (t Tree) buildChains(startNode *TreeNode, chain *Chain) {
	//Agregar el nodo a la cadena actual
	startNode.chain = chain

	//Encontrar el hijo más pesado del nodo
	var heaviestChild *TreeNode
	for _, e := range t.adj[startNode] {
		child := e.node
		if child == startNode.dfsParent {
			continue
		}
		if heaviestChild == nil || child.subTreeSize > heaviestChild.subTreeSize {
			heaviestChild = child
		}
	}

	//Si hay un hijo más pesado, seguir construyendo la misma cadena
	if heaviestChild != nil {
		startNode.chainChild = heaviestChild
		t.buildChains(heaviestChild, chain)
	}

	//Construir una nueva cadena para los hijos livianos del nodo
	for _, e := range t.adj[startNode] {
		child := e.node
		if child == startNode.dfsParent {
			continue
		}
		if child != heaviestChild {
			newChain := &Chain{
				topNode:        child,
				accumMapAll:    map[int]bool{},
				accumMapExists: map[int]bool{},
			}
			t.buildChains(child, newChain)
		}
	}
}

/*
Construye los arreglos acumulativos de cada cadena del arbol.
Utilizar arreglos acumulativos en lugar de arboles de segmentos
permite realizar consultas en tiempo O(1) en lugar de O(log(n))
*/
func (t Tree) buildAccumArrays() {
	knownChains := map[*Chain]bool{}

	for k := range t.adj {
		ch := k.chain
		_, ok := knownChains[ch]
		if !ok {
			//No se ha explorado esta cadena
			curr := ch.topNode
			ch.accumMapAll[curr.value] = true
			ch.accumMapExists[curr.value] = false
			for {
				next := curr.chainChild
				if next == nil {
					break
				}
				edge := t.getEdge(curr, next)
				ch.accumMapAll[next.value] = ch.accumMapAll[curr.value] && edge.weight
				ch.accumMapExists[next.value] = ch.accumMapExists[curr.value] || edge.weight
				curr = next
			}
		}

	}
}

/*
Función que (finalmente dios mío) indica si todas las aristas que comprenden
el camino entre los nodos x y y dentro del arbol son true.
*/
func (t Tree) forAll(x *TreeNode, y *TreeNode) bool {
	chX := x.chain
	chY := y.chain
	//Si estan en la misma cadena, retornar la conjuncion de sus valores acumulados
	if chX == chY {
		return chX.accumMapAll[x.value] && chY.accumMapAll[y.value]
	}

	/*
		Si están en cadenas distintas, pero para alguno de los dos su valor acumulativo dentro
		de su cadena ya es false. Retornar false
	*/
	if !chX.accumMapAll[x.value] || !chY.accumMapAll[y.value] {
		return false
	}

	//De lo contrario, se empieza a subir por las cadenas
	lca := t.LCA(x, y)
	//Subida de x
	curr := x
	for curr.chain != lca.chain {
		top := curr.chain.topNode
		edge := t.getEdge(top, top.dfsParent)
		//Si la conexion debil es false, retornar false
		if !edge.weight {
			return false
		}
		curr = top.dfsParent
		if !curr.chain.accumMapAll[curr.value] {
			return false
		}
	}

	//Subida de y
	curr = y
	for curr.chain != lca.chain {
		top := curr.chain.topNode
		edge := t.getEdge(top, top.dfsParent)
		//Si la conexion debil es false, retornar false
		if !edge.weight {
			return false
		}
		curr = top.dfsParent
		//Si para el punto de entrada en la nueva rama ya existe un valor false. Retornar false
		if !curr.chain.accumMapAll[curr.value] {
			return false
		}
	}

	//Si no se encontró ningún false en todo el camino, retornamos true
	return true
}

/*
Función que (por fin, dios mío, otra vez) indica si existe algún arista true
entre los nodos x y y dentro del árbol.
*/
func (t Tree) exists(x *TreeNode, y *TreeNode) bool {
	chX, chY := x.chain, y.chain
	//Si están en la misma cadena, retornar la disyunción de los acumulados de ambos nodos
	if chX == chY {
		return chX.accumMapExists[x.value] || chY.accumMapExists[y.value]
	}

	/*
		Si están en cadenas distintas, pero para su punto dentro de su cadena ya existe algún valor
		true, retornar true.
	*/
	if chX.accumMapExists[x.value] || chY.accumMapExists[y.value] {
		return true
	}

	//Si no se da ninguno de los casos, se empieza a subir por las cadenas
	lca := t.LCA(x, y)
	//Subida de x
	curr := x
	for curr.chain != lca.chain {
		top := curr.chain.topNode
		edge := t.getEdge(top, top.dfsParent)
		//Si la conexión débil es true, retornar true
		if edge.weight {
			return true
		}
		curr = top.dfsParent
		//Si para el punto de entrada en la nueva rama ya existe un valor true. Retornar true
		if curr.chain.accumMapExists[curr.value] {
			return true
		}

	}

	//Subida de y
	curr = y
	for curr.chain != lca.chain {
		top := curr.chain.topNode
		edge := t.getEdge(top, top.dfsParent)
		//Si la conexión débil es true, retornar true
		if edge.weight {
			return true
		}
		curr = top.dfsParent
		//Si para el punto de entrada en la nueva rama ya existe un valor true. Retornar true
		if curr.chain.accumMapExists[curr.value] {
			return true
		}

	}

	//Si para este punto no se ha encontrado ningún true. Se retorna false
	return false

}

func main() {
	nodes := []TreeNode{}

	//Agregar nodos
	for i := 0; i < 13; i += 1 {
		nodes = append(nodes, newNode(i))
	}

	//Agregar aristas. Es el mismo grafo de ejemplo de la clase :)
	t := newTree(13)
	t.addEdge(&(nodes[0]), &(nodes[1]), rand.Int()%2 == 0)
	t.addEdge(&(nodes[0]), &(nodes[2]), rand.Int()%2 == 0)
	t.addEdge(&(nodes[0]), &(nodes[3]), rand.Int()%2 == 0)
	t.addEdge(&(nodes[1]), &(nodes[4]), rand.Int()%2 == 0)
	t.addEdge(&(nodes[1]), &(nodes[5]), rand.Int()%2 == 0)
	t.addEdge(&(nodes[2]), &(nodes[6]), rand.Int()%2 == 0)
	t.addEdge(&(nodes[2]), &(nodes[7]), rand.Int()%2 == 0)
	t.addEdge(&(nodes[4]), &(nodes[8]), rand.Int()%2 == 0)
	t.addEdge(&(nodes[5]), &(nodes[9]), rand.Int()%2 == 0)
	t.addEdge(&(nodes[6]), &(nodes[10]), rand.Int()%2 == 0)
	t.addEdge(&(nodes[9]), &(nodes[11]), rand.Int()%2 == 0)
	t.addEdge(&(nodes[9]), &(nodes[12]), rand.Int()%2 == 0)

	//Precondicionamiento
	t.dfs(&(nodes[0]))

	t.buildChains(&(nodes[0]), &Chain{
		topNode:        &(nodes[0]),
		accumMapAll:    map[int]bool{},
		accumMapExists: map[int]bool{},
	})

	t.buildAccumArrays()

	//Check de que las cadenas se forman bien, si quiere verlas descomentar
	// knownChains := map[*Chain]bool{}
	// counter := 0
	// for k := range t.adj {
	// 	_, ok := knownChains[k.chain]
	// 	if !ok {
	// 		knownChains[k.chain] = true
	// 		top := k.chain.topNode
	// 		fmt.Printf("Cadena %d: ", counter)
	// 		for top != nil {
	// 			fmt.Printf("%d, ", top.value)
	// 			top = top.chainChild
	// 		}
	// 		fmt.Printf("\n")
	// 		counter += 1
	// 	}
	// }

	//Realmente no supe cómo probarlo ya que todos los pesos son aleatorios. Pero ahí ta por si quiere correrlo :)
	fmt.Println("forAll:", t.forAll(&(nodes[0]), &(nodes[5])))
	fmt.Println("exists:", t.exists(&(nodes[9]), &(nodes[4])))

}
