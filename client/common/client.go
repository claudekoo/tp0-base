package common

import (
	"net"
	"strconv"
	"time"

	"github.com/op/go-logging"
)

var log = logging.MustGetLogger("log")

// ClientConfig Configuration used by the client
type ClientConfig struct {
	ID            string
	ServerAddress string
	LoopAmount    int
	LoopPeriod    time.Duration
	Nombre        string
	Apellido      string
	Documento     string
	Nacimiento    string
	Numero        string
}

// Client Entity that encapsulates how
type Client struct {
	config ClientConfig
	conn   net.Conn
	shutdown bool
}

// NewClient Initializes a new client receiving the configuration
// as a parameter
func NewClient(config ClientConfig) *Client {
	client := &Client{
		config:   config,
		shutdown: false,
	}
	return client
}

// CreateClientSocket Initializes client socket. In case of
// failure, error is printed in stdout/stderr and exit 1
// is returned
func (c *Client) createClientSocket() error {
	conn, err := net.Dial("tcp", c.config.ServerAddress)
	if err != nil {
		log.Criticalf(
			"action: connect | result: fail | client_id: %v | error: %v",
			c.config.ID,
			err,
		)
	}
	c.conn = conn
	return nil
}

// StartClientLoop Send bet data to server once
func (c *Client) StartClientLoop() {
	// Send bet data to server once
	err := c.createClientSocket()
	if err != nil || c.shutdown {
		return
	}

	betData := BetData{
		ClientID:   c.config.ID,
		Nombre:     c.config.Nombre,
		Apellido:   c.config.Apellido,
		Documento:  c.config.Documento,
		Nacimiento: c.config.Nacimiento,
		Numero:     c.config.Numero,
	}

	err = SendBet(c.conn, betData)
	if err != nil {
		log.Errorf("action: send_bet | result: fail | client_id: %v | error: %v", c.config.ID, err)
		c.closeConnection()
		return
	}

	success, err := ReceiveResponse(c.conn)
	c.closeConnection()

	if err != nil {
		log.Errorf("action: receive_response | result: fail | client_id: %v | error: %v", c.config.ID, err)
		return
	}

	if success {
		documento, _ := strconv.ParseUint(c.config.Documento, 10, 64)
		numero, _ := strconv.ParseUint(c.config.Numero, 10, 32)
		
		log.Infof("action: apuesta_enviada | result: success | dni: %v | numero: %v", 
			documento, numero)
	} else {
		log.Errorf("action: apuesta_enviada | result: fail | client_id: %v | response_code: ERROR", 
			c.config.ID)
	}
}

func (c *Client) Shutdown() {
	log.Infof("action: shutdown_initiated | result: success | client_id: %v", c.config.ID)
	c.shutdown = true
	c.closeConnection()
}

func (c *Client) closeConnection() {
	if c.conn != nil {
		log.Infof("action: close_connection | result: success | client_id: %v", c.config.ID)
		c.conn.Close()
		c.conn = nil
	}
}
