package common

import (
	"encoding/binary"
	"fmt"
	"net"
	"strconv"
	"strings"

	"github.com/op/go-logging"
)

var protocolLog = logging.MustGetLogger("protocol")

const (
	ResponseOK    = 0
	ResponseError = 1
)

type BetData struct {
	ClientID   string
	Nombre     string
	Apellido   string
	Documento  string
	Nacimiento string
	Numero     string
}

// Protocol: client_id(4), nombre_len(4), nombre, apellido_len(4), apellido, documento(4), nacimiento(4), numero(4)
func SendBet(conn net.Conn, bet BetData) error {
	// Client ID (4 bytes)
	clientID, err := strconv.ParseUint(bet.ClientID, 10, 32)
	if err != nil {
		protocolLog.Errorf("action: send_bet | result: fail | field: client_id | error: %v", err)
		return fmt.Errorf("invalid client ID: %v", err)
	}
	err = binary.Write(conn, binary.BigEndian, uint32(clientID))
	if err != nil {
		protocolLog.Errorf("action: send_bet | result: fail | field: client_id | error: %v", err)
		return err
	}
	
	// Nombre length (4 bytes) and Nombre
	nameBytes := []byte(bet.Nombre)
	nameLen := uint32(len(nameBytes))
	err = binary.Write(conn, binary.BigEndian, nameLen)
	if err != nil {
		protocolLog.Errorf("action: send_bet | result: fail | field: nombre_length | error: %v", err)
		return err
	}
	_, err = conn.Write(nameBytes)
	if err != nil {
		protocolLog.Errorf("action: send_bet | result: fail | field: nombre | error: %v", err)
		return err
	}

	// Apellido length (4 bytes) and Apellido  
	surnameBytes := []byte(bet.Apellido)
	surnameLen := uint32(len(surnameBytes))
	err = binary.Write(conn, binary.BigEndian, surnameLen)
	if err != nil {
		protocolLog.Errorf("action: send_bet | result: fail | field: apellido_length | error: %v", err)
		return err
	}
	_, err = conn.Write(surnameBytes)
	if err != nil {
		protocolLog.Errorf("action: send_bet | result: fail | field: apellido | error: %v", err)
		return err
	}

	// Documento (4 bytes)
	documento, err := strconv.ParseUint(bet.Documento, 10, 32)
	if err != nil {
		protocolLog.Errorf("action: send_bet | result: fail | field: documento | error: %v", err)
		return fmt.Errorf("invalid documento: %v", err)
	}
	err = binary.Write(conn, binary.BigEndian, uint32(documento))
	if err != nil {
		protocolLog.Errorf("action: send_bet | result: fail | field: documento | error: %v", err)
		return err
	}

	// Nacimiento (4 bytes: YYYYMMDD)
	// Parse date string YYYY-MM-DD to integer YYYYMMDD
	dateParts := strings.Split(bet.Nacimiento, "-")
	if len(dateParts) != 3 {
		err := fmt.Errorf("invalid date format: %s", bet.Nacimiento)
		protocolLog.Errorf("action: send_bet | result: fail | field: nacimiento | error: %v", err)
		return err
	}
	year, err := strconv.Atoi(dateParts[0])
	if err != nil {
		protocolLog.Errorf("action: send_bet | result: fail | field: nacimiento_year | error: %v", err)
		return fmt.Errorf("invalid year: %v", err)
	}
	month, err := strconv.Atoi(dateParts[1])
	if err != nil {
		protocolLog.Errorf("action: send_bet | result: fail | field: nacimiento_month | error: %v", err)
		return fmt.Errorf("invalid month: %v", err)
	}
	day, err := strconv.Atoi(dateParts[2])
	if err != nil {
		protocolLog.Errorf("action: send_bet | result: fail | field: nacimiento_day | error: %v", err)
		return fmt.Errorf("invalid day: %v", err)
	}
	dateInt := uint32(year*10000 + month*100 + day)
	err = binary.Write(conn, binary.BigEndian, dateInt)
	if err != nil {
		protocolLog.Errorf("action: send_bet | result: fail | field: nacimiento | error: %v", err)
		return err
	}

	// Numero (4 bytes)
	numero, err := strconv.ParseUint(bet.Numero, 10, 32)
	if err != nil {
		protocolLog.Errorf("action: send_bet | result: fail | field: numero | error: %v", err)
		return fmt.Errorf("invalid numero: %v", err)
	}
	err = binary.Write(conn, binary.BigEndian, uint32(numero))
	if err != nil {
		protocolLog.Errorf("action: send_bet | result: fail | field: numero | error: %v", err)
		return err
	}

	protocolLog.Debugf("action: send_bet | result: success | client_id: %s | dni: %s", bet.ClientID, bet.Documento)
	return nil
}

func ReceiveResponse(conn net.Conn) (bool, error) {
	response := make([]byte, 1)
	n, err := conn.Read(response)
	if err != nil {
		protocolLog.Errorf("action: receive_response | result: fail | error: %v", err)
		return false, err
	}
	
	if n != 1 {
		protocolLog.Errorf("action: receive_response | result: fail | expected_bytes: 1 | received_bytes: %d", n)
		return false, fmt.Errorf("invalid response length: %d", n)
	}
	
	success := response[0] == ResponseOK
	protocolLog.Debugf("action: receive_response | result: success | response_code: %d | success: %v", response[0], success)
	return success, nil
}
