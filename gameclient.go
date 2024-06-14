package main

import (
    "fmt"
    "net/rpc"
    "os"
)

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
    err = client.Call("Servidor.getGameState", jogador, &game)
    if err != nil {
        fmt.Println("Erro ao obter estado do jogo:", err)
    } else {
        fmt.Printf("Nome: %s\n", jogador)
        fmt.Printf("Nota: %.2f\n", nota)
    }
}
