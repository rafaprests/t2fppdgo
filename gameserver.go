package main

import (
	"bufio"
	"fmt"
	"net"
	"net/rpc"
	"os"

	"github.com/nsf/termbox-go"
)

// Estrutura para representar um aluno
type GameState struct {
	mapa                        [][]Elemento
	jogador1                	Player
	jogador2					Player
	ultimoElementoSobPersonagem Elemento
	statusMsg                   string
	efeitoNeblina               bool
	revelado                    [][]bool
	raioVisao                   int
}

// estrutura para o jogador
type Player struct {
	posicao Posicao
	id  int
	nome string
}

// Estrutura para o servidor
type Servidor struct {
	state GameState
}

// Define os elementos do jogo
type Elemento struct {
	simbolo  rune
	cor      termbox.Attribute
	corFundo termbox.Attribute
	tangivel bool
}

type Posicao struct {
	x int
	y int
}

// Personagem controlado pelo jogador
var personagem = Elemento{
	simbolo:  '☺',
	cor:      termbox.ColorBlack,
	corFundo: termbox.ColorDefault,
	tangivel: true,
}

// Parede
var parede = Elemento{
	simbolo:  '▤',
	cor:      termbox.ColorBlack | termbox.AttrBold | termbox.AttrDim,
	corFundo: termbox.ColorDarkGray,
	tangivel: true,
}

// Barrreira
var barreira = Elemento{
	simbolo:  '#',
	cor:      termbox.ColorRed,
	corFundo: termbox.ColorDefault,
	tangivel: true,
}

// Vegetação
var vegetacao = Elemento{
	simbolo:  '♣',
	cor:      termbox.ColorGreen,
	corFundo: termbox.ColorDefault,
	tangivel: false,
}

// Elemento vazio
var vazio = Elemento{
	simbolo:  ' ',
	cor:      termbox.ColorDefault,
	corFundo: termbox.ColorDefault,
	tangivel: false,
}

// Elemento para representar áreas não reveladas (efeito de neblina)
var neblina = Elemento{
	simbolo:  '.',
	cor:      termbox.ColorDefault,
	corFundo: termbox.ColorYellow,
	tangivel: false,
}

func (s *Servidor) inicializar() {
	s.carregarMapa("mapa.txt")
	s.state.jogador1 = Player{Posicao{0, 0}, 1, ""}
	s.state.jogador2 = Player{Posicao{1, 0}, 2, ""}
	s.state.ultimoElementoSobPersonagem = vazio
	s.state.statusMsg = "jogo inicializado"
	s.state.efeitoNeblina = false
	s.state.raioVisao = 3
}

// metodo remoto que registra cliente
func (s *Servidor) RegisterClient(nome string, reply *int) error{
	if s.state.jogador1.nome == "" {
		s.state.jogador1.nome = nome
		*reply = 1
	} else if s.state.jogador2.nome == "" {
		s.state.jogador2.nome = nome
		*reply = 2
	} else {
		return fmt.Errorf("Limite de jogadores atingido.")
	}
	return nil
}

// metodo remoto que recebe comando do cliente
func (s *Servidor) SendCommand() error{

}

// metodo remoto que retorna o estado do jogo
func (s *Servidor) GetGameState(player string, game *GameState) error {
	*game = s.state
	return nil
}

func main() {
	porta := 8973
	servidor := new(Servidor)
	servidor.inicializar()

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

func (s *Servidor) carregarMapa(nomeArquivo string) error {
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
			case parede.simbolo:
				elementoAtual = parede
			case barreira.simbolo:
				elementoAtual = barreira
			case vegetacao.simbolo:
				elementoAtual = vegetacao
			case personagem.simbolo:
				// Atualiza a posição inicial do personagem
				// s.state.posX, s.state.posY = x, y
				fmt.Printf("Personagem encontrado na posição (%d, %d)\n", x, y)
				elementoAtual = vazio
			}
			linhaElementos = append(linhaElementos, elementoAtual)
			linhaRevelada = append(linhaRevelada, false)
		}
		s.state.mapa = append(s.state.mapa, linhaElementos)
		s.state.revelado = append(s.state.revelado, linhaRevelada)
		y++
	}
	if err := scanner.Err(); err != nil {
		return err
	}
	return nil
}

func (s *Servidor) desenhaTudo() {
	termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
	for y, linha := range s.state.mapa {
		for x, elem := range linha {
			if s.state.efeitoNeblina == false || s.state.revelado[y][x] {
				termbox.SetCell(x, y, elem.simbolo, elem.cor, elem.corFundo)
			} else {
				termbox.SetCell(x, y, neblina.simbolo, neblina.cor, neblina.corFundo)
			}
		}
	}

	s.desenhaBarraDeStatus()
	termbox.Flush()
}

func (s *Servidor) desenhaBarraDeStatus() {
	for i, c := range s.state.statusMsg {
		termbox.SetCell(i, len(s.state.mapa)+1, c, termbox.ColorBlack, termbox.ColorDefault)
	}
	msg := "Use WASD para mover e E para interagir. ESC para sair."
	for i, c := range msg {
		termbox.SetCell(i, len(s.state.mapa)+3, c, termbox.ColorBlack, termbox.ColorDefault)
	}
}

func (s *Servidor) revelarArea(username string) {
	var posicao Posicao
	if s.state.jogador1.nome == username {
		posicao = s.state.jogador1.pos
	} else if s.state.jogador2.nome == username {
		posicao = s.state.jogador2.pos
	}

	minX := max(0, posicao.x-s.state.raioVisao)
	maxX := min(len(s.state.mapa[0])-1, posicao.x+s.state.raioVisao)
	minY := max(0, posicao.y-s.state.raioVisao/2)
	maxY := min(len(s.state.mapa)-1, posicao.y+s.state.raioVisao/2)

	for y := minY; y <= maxY; y++ {
		for x := minX; x <= maxX; x++ {
			s.state.revelado[y][x] = true
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

func (s *Servidor) mover(username string, comando rune) error {
	posicao, ok := s.state.jogadores[username]
	if !ok {
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

	novaPosX, novaPosY := posicao.x+dx, posicao.y+dy
	if novaPosY >= 0 && novaPosY < len(s.state.mapa) && novaPosX >= 0 && novaPosX < len(s.state.mapa[novaPosY]) &&
		s.state.mapa[novaPosY][novaPosX].tangivel == false {
		s.state.jogadores[username] = Posicao{novaPosX, novaPosY}
		return nil
	}

	return fmt.Errorf("Movimento inválido para o jogador %s", username)
}

func (s *Servidor) interagir(username string) error {
	var posicao Posicao
	if s.state.jogador1.nome == username {
		posicao = s.state.jogador1.pos
	} else if s.state.jogador2.nome == username {
		posicao = s.state.jogador2.pos
	} else {
		return fmt.Errorf("Jogador não encontrado: %s", username)
	}

	s.state.statusMsg = fmt.Sprintf("Interagindo em (%d, %d) pelo jogador %s", posicao.x, posicao.y, username)
	return nil
}
