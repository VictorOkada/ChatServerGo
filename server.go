package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"strings"
	"unicode"
)

type client chan<- string // canal de mensagem

var (
	entering = make(chan client)
	leaving  = make(chan client)
	messages = make(chan string)
)

var combining = &unicode.RangeTable{
	R32: []unicode.Range32{
		{0x0300, 0x036f, 1},   // combining diacritical marks
		{0x1ab0, 0x1aff, 1},   // combining diacritical marks extended
		{0x1dc0, 0x1dff, 1},   // combining diacritical marks supplement
		{0x20d0, 0x20ff, 1},   // combining diacritical marks for symbols
		{0xFE00, 0xFE0F, 1},   // Variation Selectors
		{0xfe20, 0xfe2f, 1},   // combining half marks
		{0x1F3FB, 0x1F3FF, 1}, // combining emoji modifier sequences, including Skin Tone Modifier
	},
}

func broadcaster() {
	clients := make(map[client]bool) // todos os clientes conectados
	for {
		select {
		case msg := <-messages:
			// broadcast de mensagens. Envio para todos
			for cli := range clients {
				cli <- msg
			}
		case cli := <-entering:
			clients[cli] = true
		case cli := <-leaving:
			delete(clients, cli)
			close(cli)
		}
	}
}

func clientWriter(conn net.Conn, ch <-chan string) {
	for msg := range ch {
		fmt.Fprintln(conn, msg)
	}
}

func handleConn(conn net.Conn) {
	ch := make(chan string)
	go clientWriter(conn, ch)
	apelido := conn.RemoteAddr().String()
	mensagem := conn.RemoteAddr().String()
	ch <- "Bem vindo ao Servidor GO!\nComandos disponíveis:\nCriar Nick - docmd nickcreate\nAtualizar Nick - docmd nickupdate\nSair - docmd exit\nPara ação do BOT digite qualquer coisa:"
	fmt.Printf("\nNovo Usuário Conectado: %s", conn.RemoteAddr().String())
	entering <- ch

	input := bufio.NewScanner(conn)
	for input.Scan() {
		if input.Text() == "docmd nickcreate" {

			ch <- "Digite seu novo apelido: "
			input := bufio.NewScanner(conn)
			input.Scan()
			apelido = input.Text()
			messages <- "\nApelido criado: " + apelido
			//fmt.Printf("\nApelido criado: %s", apelido)

		} else if input.Text() == "docmd nickupdate" {
			fmt.Printf(apelido)
			res := strings.HasPrefix(apelido, "127.0.0")
			if res == false {
				ch <- "Digite um novo apelido para atualizar: "
				input := bufio.NewScanner(conn)
				input.Scan()
				apelido = input.Text()
				messages <- "\nApelido atualizado: " + apelido
				//fmt.Printf("\nApelido atualizado: %s", apelido)
			} else {
				messages <- "VocÊ deve criar um apelido primeiro!"
			}

		} else if input.Text() == "docmd gochat" {
			for {
				input := bufio.NewScanner(conn)
				input.Scan()
				mensagem = input.Text()

				if mensagem != "docmd exit" {

					messages <- apelido + ": " + mensagem

				} else {
					messages <- apelido + ": Se foi!"
					//fmt.Printf("\n%s Se foi", apelido)
					leaving <- ch
					conn.Close()
					break
				}

			}

		} else if input.Text() == "docmd exit" {
			messages <- apelido + " se foi "
			fmt.Printf("\n%s Se foi", apelido)
			leaving <- ch
			conn.Close()

		} else {

			//input := bufio.NewScanner(conn)
			//input.Scan()
			mensagem = input.Text()
			reverseAll(false, mensagem)
			reverseAll(true, mensagem)

		}
	}

}

func reverse(s string, verbose bool) string {
	sv := []rune(s)
	if verbose {
		//fmt.Printf("Reversing '%s' ('%d' runes)\n", s, len(sv))
	}
	rv := make([]rune, 0, len(sv)) // final reverse rune sequence
	cv := make([]rune, 0)          // current rune sequence
	pv := make([]rune, 0)          // previous rune sequence
	for ix := len(sv) - 1; ix >= 0; ix-- {
		r := sv[ix]
		if r == '\u200d' {
			cv = append([]rune{r}, pv...)
			if verbose {
				//fmt.Printf("Detect zero width joiner, so combine previous sequence ' %s '\n", string(pv))
			}
			pv = make([]rune, 0)
			continue
		}
		if unicode.In(r, combining) {
			cv = append([]rune{r}, cv...)
			if verbose {
				//fmt.Printf("Detect combining diacritical mark ' %c ' (%x)", r, r)
				//fmt.Printf("\n")
			}
		} else {
			cv = append([]rune{r}, cv...)
			if verbose {
				//combining := "no combining diacritical mark found"
				//plural := ""
				if len(cv) > 1 {
					if len(cv) > 2 {
						// Yes, I know, risking my life here (https://meta.stackexchange.com/q/236746/6309)
						//plural = "s"
					}
					//combining = fmt.Sprintf("with '%d' combining diacritical mark%s '%s'", len(cv)-1, plural, string(cv))
				}
				//fmt.Printf("regular mark '%c' (%x '%s') => '%s'\n", r, r, combining, string(cv))
			}
			if len(pv) > 0 {
				rv = append(rv, pv...)
			}
			pv = cv
			cv = make([]rune, 0)
		}
	}
	rv = append(rv, pv...)
	return string(rv)
}

func reverseAll(verbose bool, strings ...string) {
	for _, s := range strings {
		if verbose {
			reverse(s, true)
			fmt.Println()
		} else {
			messages <- reverse(s, false)
		}
	}
}

func contains(s []string, str string) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}

	return false
}

func main() {
	fmt.Println("Iniciando servidor...")
	listener, err := net.Listen("tcp", "localhost:3000")
	if err != nil {
		log.Fatal(err)
	}
	go broadcaster()
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Print(err)
			continue
		}
		go handleConn(conn)

	}
}

/*
REFERÊNCIAS:

https://www.geeksforgeeks.org/check-if-the-string-starts-with-specified-prefix-in-golang/
https://www.digitalocean.com/community/tutorials/how-to-construct-for-loops-in-go-pt
https://gobyexample.com/if-else
https://www.educative.io/answers/how-to-add-an-element-to-a-slice-in-golang
https://www.digitalocean.com/community/tutorials/how-to-construct-for-loops-in-go-pt
https://stackoverflow.com/questions/34481065/break-out-of-input-scan
https://go.dev/play/p/Wy4SZ_6U3vy

*/
