import logging
from typing import Optional, Tuple

RESPONSE_OK = 0
RESPONSE_ERROR = 1

def unpack_uint32_be(data: bytes) -> int:
    if len(data) != 4:
        raise ValueError(f"Expected 4 bytes, got {len(data)}")
    return (data[0] << 24) | (data[1] << 16) | (data[2] << 8) | data[3]

def recv_all(sock, n):
    """Helper function to receive exactly n bytes"""
    data = b''
    while len(data) < n:
        packet = sock.recv(n - len(data))
        if not packet:
            logging.error(f"action: recv_all | result: fail | expected_bytes: {n} | received_bytes: {len(data)} | error: connection closed")
            return None
        data += packet
    return data

def receive_bet(client_sock) -> Optional[Tuple[str, str, str, str, str, str]]:
    """
    Receive bet data from client socket
    
    Protocol: client_id(4), nombre_len(4), nombre, apellido_len(4), apellido, documento(4), nacimiento(4), numero(4)
    """
    try:
        # Client ID (4 bytes)
        client_id_bytes = recv_all(client_sock, 4)
        if not client_id_bytes:
            logging.error("action: receive_bet | result: fail | field: client_id | error: failed to receive data")
            return None
        client_id = unpack_uint32_be(client_id_bytes)
        
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

        logging.debug(f"action: receive_bet | result: success | client_id: {client_id} | dni: {documento}")
        return (str(client_id), nombre, apellido, str(documento), nacimiento, str(numero))

    except ValueError as e:
        logging.error(f"action: receive_bet | result: fail | error: data unpacking failed - {e}")
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
