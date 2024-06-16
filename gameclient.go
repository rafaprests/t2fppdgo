package main

import (
	"fmt"
	"net/rpc"
	"os"
	"time"

	"github.com/nsf/termbox-go"
)

// Defina a estrutura GameState
// Estrutura para representar um aluno
// Estrutura para representar um aluno
type GameState struct {
	Mapa                        [][]Elemento
	Jogador1                	Player
	Jogador2					Player
	UltimoElementoSobPersonagem Elemento
	StatusMsg                   string
	EfeitoNeblina               bool
	Revelado                    [][]bool
	RaioVisao                   int
}

// estrutura para o jogador
type Player struct {
	Posicao Posicao
	Id int
	Nome string
}

// Defina a estrutura Elemento
type Elemento struct {
	Simbolo  rune
	Cor      termbox.Attribute
	CorFundo termbox.Attribute
	Tangivel bool
}

// Defina a estrutura Posicao
type Posicao struct {
	X int
	Y int
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


	// Inicializar termbox
	err = termbox.Init()
	if err != nil {
		panic(err)
	}
	defer termbox.Close()


	var game GameState
	for {
		err = client.Call("Servidor.GetGameState", jogador, &game)
		if err != nil {
			fmt.Println("Erro ao obter estado do jogo:", err)
			break
		}

		// Desenhar o estado do jogo na tela
		desenharEstadoDoJogo(&game)
		time.Sleep(2 * time.Second)
	}
}


func desenharEstadoDoJogo(game *GameState) {
	termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
	for y, linha := range game.Mapa {
		for x, elem := range linha {
			if game.EfeitoNeblina == false || game.Revelado[y][x] {
				termbox.SetCell(x, y, elem.Simbolo, elem.Cor, elem.CorFundo)
			} else {
				termbox.SetCell(x, y, '.', termbox.ColorDefault, termbox.ColorYellow)
			}
		}
	}

	// Desenhar barra de status
	for i, c := range game.StatusMsg {
		termbox.SetCell(i, len(game.Mapa)+1, c, termbox.ColorBlack, termbox.ColorDefault)
	}
	msg := "Use WASD para mover e E para interagir. ESC para sair."
	for i, c := range msg {
		termbox.SetCell(i, len(game.Mapa)+3, c, termbox.ColorBlack, termbox.ColorDefault)
	}

	termbox.Flush()
}
