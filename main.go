package main

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/cors"

	"chat/jwt"
	"chat/postgresql"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type LoginReq struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type User struct {
	Login string `json:"login"`
}

type server struct {
	mu        sync.Mutex
	conns     map[string]*websocket.Conn
	broadcast chan Message
	store     *pgxpool.Pool
}

type Message struct {
	From string `json:"from"`
	To   string `json:"to"`
	Msg  string `json:"msg"`
}

func NewServer() *server {
	host := os.Getenv("POSTGRES_HOST")
	db := os.Getenv("POSTGRES_DB")
	user := os.Getenv("POSTGRES_USER")
	password := os.Getenv("POSTGRES_PASSWORD")

	return &server{
		conns:     make(map[string]*websocket.Conn),
		broadcast: make(chan Message),
		store:     postgresql.New(user, password, host, "5432", db),
	}
}

func (s *server) registration(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		return
	}

	var dto LoginReq

	err := json.NewDecoder(r.Body).Decode(&dto)
	if err != nil {
		log.Println("decode req", err)
		return
	}
	defer r.Body.Close()

	query := `
		INSERT INTO users (login, password) 
		VALUES ($1, $2)
	`

	exec, err := s.store.Exec(r.Context(), query, dto.Login, dto.Password)
	if err != nil {
		log.Println("error:", err)
		w.Write([]byte("error while register"))
		return
	}
	log.Println(exec.RowsAffected())

	json.NewEncoder(w).Encode("successfully")
}

func (s *server) login(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		return
	}

	var dto LoginReq

	err := json.NewDecoder(r.Body).Decode(&dto)
	if err != nil {
		log.Println("decode req", err)
		return
	}
	defer r.Body.Close()

	query := `
		SELECT login, password
		FROM users
		WHERE login = $1
	`

	var login string
	var password string

	err = s.store.QueryRow(r.Context(), query, dto.Login).Scan(&login, &password)
	if err != nil {
		log.Println("wrong login", dto.Login)
		w.Write([]byte(err.Error()))
		return
	}

	log.Println(login, password)

	if password != dto.Password {
		log.Println("wrong pass", dto.Password, password)
		w.Write([]byte("wrong password"))
		return
	}

	token, err := jwt.GenerateAccessToken(login)
	var res struct {
		Login string `json:"login"`
		Token string `json:"token"`
	}

	res.Login = dto.Login
	res.Token = token

	json.NewEncoder(w).Encode(res)
}

func (s *server) getUsers(w http.ResponseWriter, r *http.Request) {
	query := `
		SELECT login
		FROM users
	`

	rows, err := s.store.Query(r.Context(), query)
	if err != nil {
		log.Println(err)
		w.Write([]byte(err.Error()))
		return
	}

	var u []User
	for rows.Next() {
		var user User
		err := rows.Scan(&user.Login)
		if err != nil {
			log.Println(err)
			w.Write([]byte(err.Error()))
			return
		}

		u = append(u, user)
	}

	json.NewEncoder(w).Encode(u)
}

func (s *server) handleWS(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	authHeader := r.Header.Get("Authorization")
	headers := strings.Split(authHeader, " ")
	if len(headers) != 2 {
		ws.WriteMessage(websocket.TextMessage, []byte("unauthorized"))
		ws.Close()
		return
	}

	token, err := jwt.ParseAccessToken(headers[1])
	if err != nil {
		ws.WriteMessage(websocket.TextMessage, []byte("unauthorized"))
		ws.Close()
		return
	}

	from := token["login"]
	to := r.URL.Query().Get("username")

	log.Println("new connection", ws.RemoteAddr().String(), from)
	log.Println(from, "=>", to)

	s.mu.Lock()
	s.conns[from.(string)] = ws
	s.mu.Unlock()

	query := `
		SELECT "from", "to", msg 
		FROM messages
		WHERE ("to" = $1 OR "from" = $1) AND ("to" = $2 OR "from" = $2)
	`

	rows, err := s.store.Query(r.Context(), query, from.(string), to)
	if err != nil {
		log.Printf("Error occurred while loading chat history: %v", err)
	}

	for rows.Next() {
		var msg Message
		err := rows.Scan(&msg.From, &msg.To, &msg.Msg)
		if err != nil {
			log.Printf("Error occurred while scanning chat history: %v", err)
		}

		err = ws.WriteJSON(msg)
		if err != nil {
			log.Printf("Error occurred while sending chat history: %v", err)
			ws.Close()
			delete(s.conns, from.(string))
			break
		}
	}

	rows.Close()

	for {
		var msg Message
		err := ws.ReadJSON(&msg)
		if err != nil {
			if err == io.EOF {
				break
			}
			log.Println("error while reading", err)
			ws.Close()
			continue
		}

		log.Println(msg)

		query = `INSERT INTO messages ("from", "to", msg) VALUES ($1, $2, $3)`

		exec, err := s.store.Exec(r.Context(), query, from, to, msg.Msg)
		if err != nil {
			log.Println(err)
		}
		log.Println(exec.RowsAffected())

		s.broadcast <- msg
	}

}

func (s *server) handleMessages() {
	for {
		msg := <-s.broadcast

		if conn, ok := s.conns[msg.To]; ok {
			err := conn.WriteJSON(msg)
			if err != nil {
				log.Printf("Error occurred while sending message: %v", err)
				conn.Close()
				delete(s.conns, msg.To)
			}
		}
	}
}

func main() {
	server := NewServer()
	go server.handleMessages()

	mux := http.NewServeMux()
	mux.HandleFunc("/ws", server.handleWS)
	mux.HandleFunc("/registration", server.registration)
	mux.HandleFunc("/login", server.login)
	mux.HandleFunc("/users", server.getUsers)

	handler := cors.Default().Handler(mux)
	log.Println("server running on port 5000")
	log.Fatal(http.ListenAndServe(":5000", handler))
}
