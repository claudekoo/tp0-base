import logging
from typing import Optional, Tuple

RESPONSE_OK = 0
RESPONSE_ERROR = 1

MESSAGE_TYPE_BATCH = 1
MESSAGE_TYPE_FINISHED_SENDING = 2
MESSAGE_TYPE_QUERY_WINNERS = 3

def unpack_uint32_be(data: bytes) -> int:
    """Helper function to unpack a 4-byte big-endian unsigned integer from bytes"""
    if len(data) != 4:
        raise ValueError(f"Expected 4 bytes, got {len(data)}")
    return (data[0] << 24) | (data[1] << 16) | (data[2] << 8) | data[3]

def recv_all(sock, n):
    """Helper function to receive exactly n bytes"""
    data = b''
    while len(data) < n:
        packet = sock.recv(n - len(data))
        if not packet:
            return None
        data += packet
    return data

def receive_bet_batch(client_sock) -> Optional[Tuple[str, list]]:
    """
    Receive a batch of bets from client socket
    
    Protocol: client_id(4), batch_size(4), then batch_size number of bets
    Where each bet: nombre_len(4), nombre, apellido_len(4), apellido, documento(4), nacimiento(4), numero(4)
    """
    try:
        client_id_bytes = recv_all(client_sock, 4)
        if not client_id_bytes:
            logging.error("action: receive_bet_batch | result: fail | field: client_id | error: failed to receive data")
            return None
        client_id = str(unpack_uint32_be(client_id_bytes))
        
        batch_size_bytes = recv_all(client_sock, 4)
        if not batch_size_bytes:
            logging.error("action: receive_bet_batch | result: fail | field: batch_size | error: failed to receive data")
            return None
        
        batch_size = unpack_uint32_be(batch_size_bytes)
        logging.debug(f"action: receive_bet_batch | result: in_progress | client_id: {client_id} | batch_size: {batch_size}")
        
        if batch_size == 0:
            logging.info(f"action: receive_bet_batch | result: success | client_id: {client_id} | batch_size: 0")
            return (client_id, [])
        
        bets_data = []
        
        for i in range(batch_size):
            bet_result = receive_bet(client_sock)
            if bet_result is None:
                logging.error(f"action: receive_bet_batch | result: fail | bet_number: {i+1} | error: failed to receive bet")
                return None
            
            bets_data.append(bet_result)
        
        logging.info(f"action: receive_bet_batch | result: success | client_id: {client_id} | batch_size: {batch_size}")
        return (client_id, bets_data)
        
    except Exception as e:
        logging.error(f"action: receive_bet_batch | result: fail | error: {e}")
        return None

def receive_bet(client_sock) -> Optional[Tuple[str, str, str, str, str]]:
    """
    Receive a bet data from client socket
    
    Protocol: nombre_len(4), nombre, apellido_len(4), apellido, documento(4), nacimiento(4), numero(4)
    """
    try:
        # Nombre length (4 bytes)
        nombre_len_bytes = recv_all(client_sock, 4)
        if not nombre_len_bytes:
            logging.error("action: receive_bet | result: fail | field: nombre_length | error: failed to receive data")
            return None
        nombre_len = unpack_uint32_be(nombre_len_bytes)

        # Nombre
        nombre_bytes = recv_all(client_sock, nombre_len)
        if not nombre_bytes:
            logging.error(f"action: receive_bet | result: fail | field: nombre | expected_length: {nombre_len} | error: failed to receive data")
            return None
        nombre = nombre_bytes.decode('utf-8')

        # Apellido length (4 bytes)
        apellido_len_bytes = recv_all(client_sock, 4)
        if not apellido_len_bytes:
            logging.error("action: receive_bet | result: fail | field: apellido_length | error: failed to receive data")
            return None
        apellido_len = unpack_uint32_be(apellido_len_bytes)

        # Apellido
        apellido_bytes = recv_all(client_sock, apellido_len)
        if not apellido_bytes:
            logging.error(f"action: receive_bet | result: fail | field: apellido | expected_length: {apellido_len} | error: failed to receive data")
            return None
        apellido = apellido_bytes.decode('utf-8')

        # Documento (4 bytes)
        documento_bytes = recv_all(client_sock, 4)
        if not documento_bytes:
            logging.error("action: receive_bet | result: fail | field: documento | error: failed to receive data")
            return None
        documento = unpack_uint32_be(documento_bytes)

        # Nacimiento (4 bytes)
        nacimiento_bytes = recv_all(client_sock, 4)
        if not nacimiento_bytes:
            logging.error("action: receive_bet | result: fail | field: nacimiento | error: failed to receive data")
            return None
        nacimiento_int = unpack_uint32_be(nacimiento_bytes)

        year = nacimiento_int // 10000
        month = (nacimiento_int % 10000) // 100
        day = nacimiento_int % 100
        nacimiento = f"{year:04d}-{month:02d}-{day:02d}"
        
        # Numero (4 bytes)
        numero_bytes = recv_all(client_sock, 4)
        if not numero_bytes:
            logging.error("action: receive_bet | result: fail | field: numero | error: failed to receive data")
            return None
        numero = unpack_uint32_be(numero_bytes)

        return (nombre, apellido, str(documento), nacimiento, str(numero))

    except ValueError as e:
        logging.error(f"action: receive_bet | result: fail | error: unpacking failed - {e}")
        return None
    except UnicodeDecodeError as e:
        logging.error(f"action: receive_bet | result: fail | error: unicode decode failed - {e}")
        return None
    except Exception as e:
        logging.error(f"action: receive_bet | result: fail | error: {e}")
        return None


def send_response(client_sock, success: bool) -> None:
    response_code = RESPONSE_OK if success else RESPONSE_ERROR
    try:
        client_sock.send(bytes([response_code]))
        logging.debug(f"action: send_response | result: success | response_code: {response_code}")
    except Exception as e:
        logging.error(f"action: send_response | result: fail | response_code: {response_code} | error: {e}")

def receive_message_type(client_sock) -> Optional[int]:
    """Receive message type (4 bytes)"""
    try:
        msg_type_bytes = recv_all(client_sock, 4)
        if not msg_type_bytes:
            return None
        msg_type = unpack_uint32_be(msg_type_bytes)
        logging.debug(f"action: receive_message_type | result: success | message_type: {msg_type}")
        return msg_type
    except Exception as e:
        logging.error(f"action: receive_message_type | result: fail | error: {e}")
        return None

def receive_finished_notification(client_sock) -> Optional[str]:
    """Receive finished notification message"""
    try:
        client_id_bytes = recv_all(client_sock, 4)
        if not client_id_bytes:
            logging.error("action: receive_finished_notification | result: fail | field: client_id | error: failed to receive data")
            return None
        client_id = str(unpack_uint32_be(client_id_bytes))
        
        logging.debug(f"action: receive_finished_notification | result: success | client_id: {client_id}")
        return client_id
    except Exception as e:
        logging.error(f"action: receive_finished_notification | result: fail | error: {e}")
        return None

def receive_query_winners(client_sock) -> Optional[str]:
    """Receive query winners message"""
    try:
        client_id_bytes = recv_all(client_sock, 4)
        if not client_id_bytes:
            logging.error("action: receive_query_winners | result: fail | field: client_id | error: failed to receive data")
            return None
        client_id = str(unpack_uint32_be(client_id_bytes))
        
        logging.debug(f"action: receive_query_winners | result: success | client_id: {client_id}")
        return client_id
    except Exception as e:
        logging.error(f"action: receive_query_winners | result: fail | error: {e}")
        return None

def pack_uint32_be(value: int) -> bytes:
    """Pack a 4-byte big-endian unsigned integer to bytes"""
    return bytes([(value >> 24) & 0xFF, (value >> 16) & 0xFF, (value >> 8) & 0xFF, value & 0xFF])

def send_winners(client_sock, winners: list[str]) -> None:
    try:
        client_sock.send(bytes([RESPONSE_OK]))
        
        winners_count = len(winners)
        client_sock.send(pack_uint32_be(winners_count))
        
        for winner_documento in winners:
            documento_int = int(winner_documento)
            client_sock.send(pack_uint32_be(documento_int))
        
        logging.debug(f"action: send_winners | result: success | winners_count: {winners_count}")
    except Exception as e:
        logging.error(f"action: send_winners | result: fail | error: {e}")
        try:
            client_sock.send(bytes([RESPONSE_ERROR]))
        except:
            pass
