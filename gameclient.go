package main

import (
	"fmt"
	"net/rpc"
	"os"

	"github.com/nsf/termbox-go"
)

// Defina a estrutura GameState
type GameState struct {
	mapa                        [][]Elemento
	jogadores                   map[string]Posicao
	ultimoElementoSobPersonagem Elemento
	statusMsg                   string
	efeitoNeblina               bool
	revelado                    [][]bool
	raioVisao                   int
}

// Defina a estrutura Elemento
type Elemento struct {
	simbolo  rune
	cor      termbox.Attribute
	corFundo termbox.Attribute
	tangivel bool
}

// Defina a estrutura Posicao
type Posicao struct {
	x int
	y int
}

func main() {

	if len(os.Args) != 3 {
		fmt.Println("Uso:", os.Args[0], " <maquina> <nome_do_jogador>")
		return
	}

	porta := 8973
	maquina := os.Args[1]
	jogador := os.Args[2]

	client, err := rpc.Dial("tcp", fmt.Sprintf("%s:%d", maquina, porta))
	if err != nil {
		fmt.Println("Erro ao conectar ao servidor:", err)
		return
	}

	var game GameState
	err = client.Call("Servidor.GetGameState", jogador, &game)
	if err != nil {
		fmt.Println("Erro ao obter estado do jogo:", err)
	} else {
		fmt.Printf("Sucesso em obter o estado do jogo")
	}
}
