package main

import (
	"encoding/json"
	"io"
	"log"
	"math/rand"
	"net"
	"strconv"
	"strings"
	"time"
)
//решил в пулл собирать коннекты вместо каналов
type (
	Entity struct {
		Line    net.Conn
		Name    string `json:"Name"`
		Message string `json:"Message"`
	}
	lineEssence map[net.Conn]bool
	roomPool    []lineEssence
)

var (
	rooms        roomPool
	join         = make(chan Entity)
	leave        = make(chan Entity)
	messages     = make(chan Entity)
	gameExchange = make(chan Entity)
)

func (rooms *roomPool) purgeClient(cli Entity) {
	for _, room := range *rooms {
		if _, exist := room[cli.Line]; exist {
			delete(room, cli.Line)
		}
	}
}

func main() {
	line, err := net.Listen("tcp", ":8080")
	if err != nil {
		log.Panic(err)
	}
	defer line.Close()
	log.Println("Сервер Запущен")

	go clientpoolControl()
	go GameEngine(gameExchange)

	chatRoom := make(lineEssence)
	rooms = append(rooms, chatRoom)
	for {
		conn, err := line.Accept()
		if err != nil {
			log.Println(err)
			break
		}
		go handleconnection(conn)
	}
}

func handleconnection(conn net.Conn) {
	defer conn.Close()
	thisUser := Entity{}
	thisUser.Line = conn
	for {
		bufer := make([]byte, 64)
		n, err := conn.Read(bufer)
		if err != nil {
			if err == io.EOF {
				break
			} else {
				log.Println(err)
				continue
			}
		}
		bufer = bufer[0:n]
		err = json.Unmarshal(bufer, &thisUser)
		if err != nil {
			log.Println(err)
			continue
		}

		var foundClient bool
		for _, rm := range rooms {
			if _, found := rm[thisUser.Line]; found {
				foundClient = true
				break
			}
		}

		if !foundClient {
			if len(thisUser.Name) < 2 {
				conn.Write([]byte("Ник не может быть короче двух символов\nОтключен от сервером..."))
				break
			}
			thisUser.Message = "Вошел в чат"
			join <- thisUser
		}
		messages <- thisUser
	}
	thisUser.Message = "Вышел"
	messages <- thisUser
	leave <- thisUser
}

func clientpoolControl() {
	for {
		select {
		case cli := <-join:
			rooms[0][cli.Line] = true
		case cli := <-leave:
			rooms.purgeClient(cli)
		case msg := <-messages:
			dataClientRoute(msg)
		}
	}
}

func dataClientRoute(msg Entity) {
	switch msg.Message {
	case "/g":
		if _, exist := rooms[0][msg.Line]; exist {
			rooms.purgeClient(msg)
			if len(rooms) == 1 {
				gameRoom := make(lineEssence)
				rooms = append(rooms, gameRoom)
			}
			rooms[len(rooms)-1][msg.Line] = true
		}
	case "/q":
		if _, exist := rooms[0][msg.Line]; !exist {
			rooms.purgeClient(msg)
			rooms[0][msg.Line] = true
		}
	default:
		//просто жутко приколхозил сюда мотор игры, хочу с заделом на многопользовательскую игру, с множеством комнат; нужно запускать на каждую игровую комнату свою рутину, с игровым движком... но по срокам боюсь не успею, надеюсь данная реализация - хотя бы отчасти подходит под ТЗ)))
		for i, room := range rooms {
			if _, exist := room[msg.Line]; exist {
				for ln := range room {
					if ln != msg.Line {
						ln.Write([]byte(msg.Name + ": " + msg.Message))
					}
					if i != 0 && ln != msg.Line {
						gameExchange <- msg
					} else if i != 0 && len(room) == 1 {
						gameExchange <- msg
					}
				}
			}
		}
	}
}

//------------------------------------------------------------------------------

type MathExpression struct {
	Proection, Result string
	num1, num2        int
	Contraulation     bool
}

//Some Numbers & Expression Method-Generate...
func (gameExpression *MathExpression) gameExpressionGenerate() {
	rand.Seed(time.Now().UnixNano())
	gameExpression.num1 = rand.Intn(100)
	gameExpression.num2 = rand.Intn(100)
	randomizer := rand.Intn(24)

	if gameExpression.num1 > gameExpression.num2 {
		if randomizer*2 < gameExpression.num1 && gameExpression.num2 != 0 {
			gameExpression.Proection = "Чему равно выражение: " + strconv.Itoa(gameExpression.num1) + " / " + strconv.Itoa(gameExpression.num2) + " = ? (округлить до целых)"
			gameExpression.Result = strconv.Itoa(gameExpression.num1 / gameExpression.num2)
		} else {
			gameExpression.Proection = "Чему равно выражение: " + strconv.Itoa(gameExpression.num1) + " - " + strconv.Itoa(gameExpression.num2) + " = ? "
			gameExpression.Result = strconv.Itoa(gameExpression.num1 - gameExpression.num2)
		}
	} else {
		if randomizer*2 < gameExpression.num1 {
			gameExpression.Proection = "Чему равно выражение: " + strconv.Itoa(gameExpression.num1) + " + " + strconv.Itoa(gameExpression.num2) + " = ? "
			gameExpression.Result = strconv.Itoa(gameExpression.num1 + gameExpression.num2)
		} else {
			gameExpression.Proection = "Чему равно выражение: " + strconv.Itoa(gameExpression.num1) + " * " + strconv.Itoa(gameExpression.num2) + " = ? "
			gameExpression.Result = strconv.Itoa(gameExpression.num1 * gameExpression.num2)
		}
	}
}

//Game Engine...
func GameEngine(ch chan Entity) {
	randExpression := MathExpression{}
	for {
		answer := <-ch
		if !randExpression.Contraulation {
			randExpression.gameExpressionGenerate()
			randExpression.Contraulation = true
		}

		if strings.Trim(answer.Message, " \n\r") == randExpression.Result {
			randExpression.Contraulation = false
		}

		for ln := range rooms[len(rooms)-1] {
			if !randExpression.Contraulation {
				ln.Write([]byte("\nSERVER: Победил " + answer.Name + "!"))
			} else {
				ln.Write([]byte("\nSERVER: " + randExpression.Proection))
			}
		}
	}
}
