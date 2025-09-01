package common

import (
	"encoding/binary"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/op/go-logging"
)

var log = logging.MustGetLogger("log")

const (
	ResponseOK    = 0
	ResponseError = 1
)

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

	err = c.sendBetData()
	if err != nil {
		log.Errorf("action: send_bet | result: fail | client_id: %v | error: %v", c.config.ID, err)
		c.closeConnection()
		return
	}

	response := make([]byte, 1)
	n, err := c.conn.Read(response)
	c.closeConnection()

	if err != nil {
		log.Errorf("action: receive_response | result: fail | client_id: %v | error: %v", c.config.ID, err)
		return
	}

	if n == 1 && response[0] == ResponseOK {
		documento, _ := strconv.ParseUint(c.config.Documento, 10, 64)
		numero, _ := strconv.ParseUint(c.config.Numero, 10, 32)
		
		log.Infof("action: apuesta_enviada | result: success | dni: %v | numero: %v", 
			documento, numero)
	} else {
		var responseCode byte
		if n > 0 {
			responseCode = response[0]
		} else {
			responseCode = 255 // Invalid response
		}
		log.Errorf("action: apuesta_enviada | result: fail | client_id: %v | response_code: %v", 
			c.config.ID, responseCode)
	}
}

func (c *Client) sendBetData() error {
	// Protocol: client_id(4), nombre_len(4), nombre, apellido_len(4), apellido, documento(4), nacimiento(4), numero(4)

	// Client ID (4 bytes)
	clientID, err := strconv.ParseUint(c.config.ID, 10, 32)
	if err != nil {
		return err
	}
	err = binary.Write(c.conn, binary.BigEndian, uint32(clientID))
	if err != nil {
		return err
	}
	
	// Nombre length (4 bytes) and Nombre
	nameBytes := []byte(c.config.Nombre)
	nameLen := uint32(len(nameBytes))
	err = binary.Write(c.conn, binary.BigEndian, nameLen)
	if err != nil {
		return err
	}
	_, err = c.conn.Write(nameBytes)
	if err != nil {
		return err
	}

	// Apellido length (4 bytes) and Apellido  
	surnameBytes := []byte(c.config.Apellido)
	surnameLen := uint32(len(surnameBytes))
	err = binary.Write(c.conn, binary.BigEndian, surnameLen)
	if err != nil {
		return err
	}
	_, err = c.conn.Write(surnameBytes)
	if err != nil {
		return err
	}

	// Documento (4 bytes)
	documento, err := strconv.ParseUint(c.config.Documento, 10, 32)
	if err != nil {
		return err
	}
	err = binary.Write(c.conn, binary.BigEndian, uint32(documento))
	if err != nil {
		return err
	}

	// Nacimiento (4 bytes: YYYYMMDD)
	// Parse date string YYYY-MM-DD to integer YYYYMMDD
	dateParts := strings.Split(c.config.Nacimiento, "-")
	if len(dateParts) != 3 {
		return fmt.Errorf("invalid date format: %s", c.config.Nacimiento)
	}
	year, _ := strconv.Atoi(dateParts[0])
	month, _ := strconv.Atoi(dateParts[1])
	day, _ := strconv.Atoi(dateParts[2])
	dateInt := uint32(year*10000 + month*100 + day)
	err = binary.Write(c.conn, binary.BigEndian, dateInt)
	if err != nil {
		return err
	}

	// Numero (4 bytes)
	numero, err := strconv.ParseUint(c.config.Numero, 10, 32)
	if err != nil {
		return err
	}
	err = binary.Write(c.conn, binary.BigEndian, uint32(numero))
	if err != nil {
		return err
	}

	return nil
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
