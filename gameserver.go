package main

import (
	"bufio"
	"fmt"
	"net"
	"net/rpc"
	"os"

	"github.com/nsf/termbox-go"
)

// estrutura do gamestate
type GameState struct {
	Mapa                        [][]Elemento
	Jogador1                    Player
	Jogador2                    Player
	//StatusMsg                   string
	EfeitoNeblina               bool
	ReveladoJ1                  [][]bool 
	ReveladoJ2                  [][]bool 
	RaioVisao                   int
	NroJogadores                int
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

var pombo = Elemento{
	Simbolo: 'P',
	Cor: termbox.ColorLightBlue,
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

// metodo remoto que registra cliente
func (s *Servidor) RegisterClient(nome string, reply *int) error{
	if s.State.Jogador1.Nome == ""{
		s.State.Jogador1.Nome = nome
		s.State.NroJogadores++
		*reply = 1
	} else if s.State.Jogador2.Nome == "" {
		s.State.Jogador2.Nome = nome
		s.State.NroJogadores++
		*reply = 2
	} else{
		return fmt.Errorf("Limite de jogadores atingido.")
	}
	return nil
}

// metodo remoto que retira cliente
func (s *Servidor) UnregisterClient(playerID int, reply *string) error{
	if playerID == 1 && s.State.Jogador1.Nome != ""{
		*reply = s.State.Jogador1.Nome + " desconectado."
		s.State.Jogador1.Nome = ""
		s.State.NroJogadores--
	} else if playerID == 2 && s.State.Jogador2.Nome != "" {
		*reply = s.State.Jogador2.Nome + " desconectado."
		s.State.Jogador2.Nome = ""
		s.State.NroJogadores--
	} else{
		return fmt.Errorf("Jogador ja desconectado.")
	}
	return nil
}

// metodo remoto que recebe comando do cliente
func (s *Servidor) SendCommand(cmd Command, reply *string) error {
	switch cmd.Action {
	case "move_up":
		return s.Mover(cmd.PlayerID, 'w')
	case "move_down":
		return s.Mover(cmd.PlayerID, 's')
	case "move_left":
		return s.Mover(cmd.PlayerID, 'a')
	case "move_right":
		return s.Mover(cmd.PlayerID, 'd')
	case "interact":
		return s.Interagir(cmd.PlayerID)
	default:
		return fmt.Errorf("acao invalida")
	}
}

// metodo remoto que retorna o estado do jogo
func (s *Servidor) GetGameState(player string, game *GameState) error {
	*game = s.State
	return nil
}

func (s *Servidor) Inicializar() {
	s.State.Jogador1 = Player{Posicao{0, 0}, 1, ""}
	s.State.Jogador2 = Player{Posicao{0, 0}, 2, ""}
	s.CarregarMapa("mapa.txt")
	//s.State.StatusMsg = "jogo inicializado"
	s.State.EfeitoNeblina = true
	s.State.RaioVisao = 3
	s.State.NroJogadores = 0

	// Inicializar matrizes de visibilidade
	s.State.ReveladoJ1 = make([][]bool, len(s.State.Mapa))
	s.State.ReveladoJ2 = make([][]bool, len(s.State.Mapa))
	for i := range s.State.Mapa {
		s.State.ReveladoJ1[i] = make([]bool, len(s.State.Mapa[i]))
		s.State.ReveladoJ2[i] = make([]bool, len(s.State.Mapa[i]))
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
			case pombo.Simbolo:
				elementoAtual = pombo
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
		y++
	}
	if err := scanner.Err(); err != nil {
		return err
	}
	return nil
}

func (s *Servidor) Mover(playerID int, comando rune) error {
	var player *Player
	if playerID == 1 {
		player = &s.State.Jogador1
	} else if playerID == 2 {
		player = &s.State.Jogador2
	} else {
		return fmt.Errorf("Jogador não encontrado.")
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
		s.RevelarArea(player.Posicao.X, player.Posicao.Y, player.Id)
		return nil
	}

	return nil 
}

func (s *Servidor) Interagir(playerID int) error {
	var posicao *Posicao
	if playerID == 1 {
		posicao = &s.State.Jogador1.Posicao
	} else if playerID == 2 {
		posicao = &s.State.Jogador2.Posicao
	} else {
		return fmt.Errorf("Jogador não encontrado.")
	}

	direcoes := []Posicao{
		{0, -1},  // cima
		{0, 1},   // baixo
		{-1, 0},  // esquerda
		{1, 0},   // direita
	}

	for _, direcao := range direcoes {
		novoX := posicao.X + direcao.X
		novoY := posicao.Y + direcao.Y
		if novoX >= 0 && novoX < len(s.State.Mapa[0]) && novoY >= 0 && novoY < len(s.State.Mapa) {
			if s.State.Mapa[novoY][novoX].Simbolo == pombo.Simbolo {
				//s.State.EfeitoNeblina = false
				// Atualizar a visibilidade para ambos os jogadores
				for i := range s.State.ReveladoJ1 {
					for j := range s.State.ReveladoJ1[i] {
						s.State.ReveladoJ1[i][j] = true
						s.State.ReveladoJ2[i][j] = true
					}
				}
				return nil
			}
		}
	}

	return nil
}

func (s *Servidor) RevelarArea(x, y, playerID int) {
	minX := max(0, x-s.State.RaioVisao)
	maxX := min(len(s.State.Mapa[0])-1, x+s.State.RaioVisao)
	minY := max(0, y-s.State.RaioVisao)
	maxY := min(len(s.State.Mapa)-1, y+s.State.RaioVisao)

	for i := minY; i <= maxY; i++ {
		for j := minX; j <= maxX; j++ {
			if playerID == 1 {
				s.State.ReveladoJ1[i][j] = true
			} else if playerID == 2 {
				s.State.ReveladoJ2[i][j] = true
			}
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
