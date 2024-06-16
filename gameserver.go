package main

import (
	"bufio"
	"fmt"
	"net"
	"net/rpc"
	"os"

	"github.com/nsf/termbox-go"
	//"golang.org/x/text/cases"
)

// estrutura do gamestate
type GameState struct {
	Mapa                        [][]Elemento
	Jogador1                    Player
	Jogador2                    Player
	UltimoElementoSobPersonagem Elemento
	StatusMsg                   string
	EfeitoNeblina               bool
	Revelado                    [][]bool
	RaioVisao                   int
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

// Estrutura para o servidor
type Servidor struct {
	State GameState
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

func (s *Servidor) Inicializar() {
	s.State.Jogador1 = Player{Posicao{0, 0}, 1, ""}
	s.State.Jogador2 = Player{Posicao{0, 0}, 2, ""}
	s.CarregarMapa("mapa.txt")
	s.State.UltimoElementoSobPersonagem = vazio
	s.State.StatusMsg = "jogo inicializado"
	s.State.EfeitoNeblina = false
	s.State.RaioVisao = 3
}

// metodo remoto que registra cliente
func (s *Servidor) RegisterClient(nome string, reply *int) error {
	if s.State.Jogador1.Nome == "" {
		s.State.Jogador1.Nome = nome
		*reply = 1
	} else if s.State.Jogador2.Nome == "" {
		s.State.Jogador2.Nome = nome
		*reply = 2
	} else {
		return fmt.Errorf("Limite de jogadores atingido.")
	}
	return nil
}

// metodo remoto que recebe comando do cliente
func (s *Servidor) SendCommand(cmd Command, reply *string) error {
	var player *Player
	if cmd.PlayerID == 1 {
		player = &s.State.Jogador1
	} else if cmd.PlayerID == 2 {
		player = &s.State.Jogador2
	} else {
		return fmt.Errorf("ID de jogador invalido")
	}

	switch cmd.Action {
	case "move_up":
		return s.Mover(player.Nome, 'w')
	case "move_down":
		return s.Mover(player.Nome, 's')
	case "move_left":
		return s.Mover(player.Nome, 'a')
	case "move_right":
		return s.Mover(player.Nome, 'd')
	case "interact":
		return s.Interagir(player.Nome)
	default:
		return fmt.Errorf("acao invalida")
	}
}

// metodo remoto que retorna o estado do jogo
func (s *Servidor) GetGameState(player string, game *GameState) error {
	*game = s.State
	return nil
}

func main() {
	porta := 8973
	servidor := new(Servidor)
	servidor.Inicializar()

	rpc.Register(servidor)
	l, err := net.Listen("tcp", fmt.Sprintf(":%d", porta))
	if err != nil {
		fmt.Println("Erro ao iniciar o servidor:", err)
		return
	}

	fmt.Println("Servidor aguardando conexões na porta", porta)
	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Erro ao aceitar conexão:", err)
			continue
		}
		go rpc.ServeConn(conn)
	}
}

func (s *Servidor) CarregarMapa(nomeArquivo string) error {
	arquivo, err := os.Open(nomeArquivo)
	if err != nil {
		return err
	}
	defer arquivo.Close()

	scanner := bufio.NewScanner(arquivo)
	y := 0
	for scanner.Scan() {
		linhaTexto := scanner.Text()
		var linhaElementos []Elemento
		var linhaRevelada []bool
		for x, char := range linhaTexto {
			elementoAtual := vazio
			switch char {
			case parede.Simbolo:
				elementoAtual = parede
			case barreira.Simbolo:
				elementoAtual = barreira
			case vegetacao.Simbolo:
				elementoAtual = vegetacao
			case personagem.Simbolo:
				// Atualiza a posição inicial do personagem
				if s.State.Jogador1.Posicao.X == 0 && s.State.Jogador1.Posicao.Y == 0 {
					s.State.Jogador1.Posicao.X = x
					s.State.Jogador1.Posicao.Y = y
					fmt.Printf("Jogador1 encontrado na posição (%d, %d)\n", x, y)
				} else {
					s.State.Jogador2.Posicao.X = x
					s.State.Jogador2.Posicao.Y = y
					fmt.Printf("Jogador2 encontrado na posição (%d, %d)\n", x, y)
				}
				elementoAtual = vazio
			}
			linhaElementos = append(linhaElementos, elementoAtual)
			linhaRevelada = append(linhaRevelada, false)
		}
		s.State.Mapa = append(s.State.Mapa, linhaElementos)
		s.State.Revelado = append(s.State.Revelado, linhaRevelada)
		y++
	}
	if err := scanner.Err(); err != nil {
		return err
	}
	return nil
}

// func (s *Servidor) DesenhaTudo() {
// 	termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
// 	for y, linha := range s.State.Mapa {
// 		for x, elem := range linha {
// 			if s.State.EfeitoNeblina == false || s.State.Revelado[y][x] {
// 				termbox.SetCell(x, y, elem.Simbolo, elem.Cor, elem.CorFundo)
// 			} else {
// 				termbox.SetCell(x, y, neblina.Simbolo, neblina.Cor, neblina.CorFundo)
// 			}
// 		}
// 	}

// 	s.DesenhaBarraDeStatus()
// 	termbox.Flush()
// }

// func (s *Servidor) DesenhaBarraDeStatus() {
// 	for i, c := range s.State.StatusMsg {
// 		termbox.SetCell(i, len(s.State.Mapa)+1, c, termbox.ColorBlack, termbox.ColorDefault)
// 	}
// 	msg := "Use WASD para mover e E para interagir. ESC para sair."
// 	for i, c := range msg {
// 		termbox.SetCell(i, len(s.State.Mapa)+3, c, termbox.ColorBlack, termbox.ColorDefault)
// 	}
// }

func (s *Servidor) RevelarArea(username string) {
	var posicao Posicao
	if s.State.Jogador1.Nome == username {
		posicao = s.State.Jogador1.Posicao
	} else if s.State.Jogador2.Nome == username {
		posicao = s.State.Jogador2.Posicao
	}

	minX := max(0, posicao.X-s.State.RaioVisao)
	maxX := min(len(s.State.Mapa[0])-1, posicao.X+s.State.RaioVisao)
	minY := max(0, posicao.Y-s.State.RaioVisao/2)
	maxY := min(len(s.State.Mapa)-1, posicao.Y+s.State.RaioVisao/2)

	for y := minY; y <= maxY; y++ {
		for x := minX; x <= maxX; x++ {
			s.State.Revelado[y][x] = true
		}
	}
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func (s *Servidor) Mover(username string, comando rune) error {
	var player *Player
	if s.State.Jogador1.Nome == username {
		player = &s.State.Jogador1
	} else if s.State.Jogador2.Nome == username {
		player = &s.State.Jogador2
	} else {
		return fmt.Errorf("Jogador não encontrado: %s", username)
	}

	dx, dy := 0, 0
	switch comando {
	case 'w':
		dy = -1
	case 'a':
		dx = -1
	case 's':
		dy = 1
	case 'd':
		dx = 1
	}

	novaPosX, novaPosY := player.Posicao.X+dx, player.Posicao.Y+dy
	if novaPosY >= 0 && novaPosY < len(s.State.Mapa) && novaPosX >= 0 && novaPosX < len(s.State.Mapa[novaPosY]) &&
		s.State.Mapa[novaPosY][novaPosX].Tangivel == false {
		player.Posicao = Posicao{novaPosX, novaPosY}
		return nil
	}

	return nil //fmt.Errorf("Movimento inválido para o jogador %s", username)
}

func (s *Servidor) Interagir(username string) error {
	var posicao Posicao
	if s.State.Jogador1.Nome == username {
		posicao = s.State.Jogador1.Posicao
	} else if s.State.Jogador2.Nome == username {
		posicao = s.State.Jogador2.Posicao
	} else {
		return fmt.Errorf("Jogador não encontrado: %s", username)
	}

	s.State.StatusMsg = fmt.Sprintf("Interagindo em (%d, %d) pelo jogador %s", posicao.X, posicao.Y, username)
	return nil
}
