package main

import (
	"fmt"
	"net/rpc"
	"os"
	"time"

	"github.com/nsf/termbox-go"
)

// estrutura do gamestate
type GameState struct {
	Mapa     [][]Elemento
	Jogador1 Player
	Jogador2 Player
	//StatusMsg                   string
	EfeitoNeblina bool
	ReveladoJ1    [][]bool
	ReveladoJ2    [][]bool
	RaioVisao     int
	NroJogadores  int
}

// estrutura para o jogador
type Player struct {
	Posicao Posicao
	Id      int
	Nome    string
}

// estrutura para o comando
type Command struct {
	PlayerID int
	Action   string
}

// estrutura para o elemento
type Elemento struct {
	Simbolo  rune
	Cor      termbox.Attribute
	CorFundo termbox.Attribute
	Tangivel bool
}

type Posicao struct {
	X int
	Y int
}

var personagem = Elemento{
	Simbolo:  '☺',
	Cor:      termbox.ColorWhite,
	CorFundo: termbox.ColorDefault,
	Tangivel: true,
}

var parede = Elemento{
	Simbolo:  '▤',
	Cor:      termbox.ColorBlack | termbox.AttrBold | termbox.AttrDim,
	CorFundo: termbox.ColorDarkGray,
	Tangivel: true,
}

var barreira = Elemento{
	Simbolo:  '#',
	Cor:      termbox.ColorRed,
	CorFundo: termbox.ColorDefault,
	Tangivel: true,
}

var vegetacao = Elemento{
	Simbolo:  '♣',
	Cor:      termbox.ColorGreen,
	CorFundo: termbox.ColorDefault,
	Tangivel: false,
}

var vazio = Elemento{
	Simbolo:  ' ',
	Cor:      termbox.ColorDefault,
	CorFundo: termbox.ColorDefault,
	Tangivel: false,
}

var neblina = Elemento{
	Simbolo:  '.',
	Cor:      termbox.ColorDefault,
	CorFundo: termbox.ColorYellow,
	Tangivel: false,
}

// estrutura para comando
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
	// Goroutine para atualizar o estado do jogo periodicamente
	go func() {
		for {
			err = client.Call("Servidor.GetGameState", jogador, &game)
			if err != nil {
				fmt.Println("Erro ao obter estado do jogo:", err)
				break
			}

			// Desenhar o estado do jogo na tela
			desenharEstadoDoJogo(&game, reply)
			time.Sleep(200 * time.Millisecond)
		}
	}()

	for {
		var action string
		switch ev := termbox.PollEvent(); ev.Type {
		case termbox.EventKey:
			if ev.Key == termbox.KeyEsc {
				var response string
				err := client.Call("Servidor.UnregisterClient", reply, &response)
				if err != nil {
					fmt.Println("Erro ao desconectar cliente:", err)
				} else {
					fmt.Println(response)
				}
				return
			}
			if ev.Ch == 'w' {
				action = "move_up"
			}
			if ev.Ch == 's' {
				action = "move_down"
			}
			if ev.Ch == 'a' {
				action = "move_left"
			}
			if ev.Ch == 'd' {
				action = "move_right"
			}
			if ev.Ch == 'e' {
				action = "interact"
			}
			if ev.Ch == 'r' {
				action = "restart"
			}
		}
		if action != "" {
			cmd := Command{PlayerID: reply, Action: action}
			var response string
			err := client.Call("Servidor.SendCommand", cmd, &response)
			if err != nil {
				fmt.Println("erro ao enviar comando", err)
			}
		}
	}
}

func desenharEstadoDoJogo(game *GameState, playerID int) {
	termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
	for y, linha := range game.Mapa {
		for x, elem := range linha {
			var revelado bool
			if playerID == 1 {
				revelado = game.ReveladoJ1[y][x]
			} else if playerID == 2 {
				revelado = game.ReveladoJ2[y][x]
			}

			if game.EfeitoNeblina == false || revelado {
				termbox.SetCell(x, y, elem.Simbolo, elem.Cor, elem.CorFundo)
			} else {
				termbox.SetCell(x, y, neblina.Simbolo, neblina.Cor, neblina.CorFundo)
			}
		}
	}

	// Desenhar os personagens
	termbox.SetCell(game.Jogador1.Posicao.X, game.Jogador1.Posicao.Y, '☺', termbox.ColorWhite, termbox.ColorDefault)
	termbox.SetCell(game.Jogador2.Posicao.X, game.Jogador2.Posicao.Y, '☺', termbox.ColorWhite, termbox.ColorDefault)

	// // Desenhar barra de status
	// for i, c := range game.StatusMsg {
	// 	termbox.SetCell(i, len(game.Mapa)+1, c, termbox.ColorBlack, termbox.ColorDefault)
	// }
	// msg := "Use WASD para mover e E para interagir. ESC para sair."
	// for i, c := range msg {
	// 	termbox.SetCell(i, len(game.Mapa)+3, c, termbox.ColorBlack, termbox.ColorDefault)
	// }

	termbox.Flush()
}
