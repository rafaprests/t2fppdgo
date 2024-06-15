package main

import (
	"fmt"
	"net/rpc"
	"os"

	"github.com/nsf/termbox-go"
)

// Defina a estrutura GameState
// Estrutura para representar um aluno
type GameState struct {
	mapa                        [][]Elemento
	jogador1                    Player
	jogador2                    Player
	ultimoElementoSobPersonagem Elemento
	statusMsg                   string
	efeitoNeblina               bool
	revelado                    [][]bool
	raioVisao                   int
}

// estrutura para o jogador
type Player struct {
	posicao Posicao
	id      int
	nome    string
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

	// Chamar método remoto para registrar o cliente
	var reply int
	err = client.Call("Servidor.RegisterClient", jogador, &reply)
	if err != nil {
		fmt.Println("Erro ao registrar cliente:", err)
		return
	}
	fmt.Printf("Jogador registrado com sucesso. ID: %d\n", reply)

	var game GameState
	err = client.Call("Servidor.GetGameState", jogador, &game)
	if err != nil {
		fmt.Println("Erro ao obter estado do jogo:", err)
	} else {
		fmt.Printf("Sucesso em obter o estado do jogo")
	}
}
