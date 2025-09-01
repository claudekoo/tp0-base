import socket
import logging
import signal
import sys
import struct
from .utils import Bet, store_bets

RESPONSE_OK = 0
RESPONSE_ERROR = 1


class Server:
    def __init__(self, port, listen_backlog):
        # Initialize server socket
        self._server_socket = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        self._server_socket.setsockopt(socket.SOL_SOCKET, socket.SO_REUSEADDR, 1)
        self._server_socket.bind(('', port))
        self._server_socket.listen(listen_backlog)
        self._running = True
        
        signal.signal(signal.SIGTERM, self._signal_handler)

    def _signal_handler(self, sig, frame):
        logging.info("action: sigterm_received | result: success")
        self._running = False
        if self._server_socket:
            logging.info("action: close_server_socket | result: success")
            self._server_socket.close()
        sys.exit(0)

    def run(self):
        """
        Dummy Server loop

        Server that accept a new connections and establishes a
        communication with a client. After client with communucation
        finishes, servers starts to accept new connections again
        """

        try:
            while self._running:
                try:
                    client_sock = self.__accept_new_connection()
                    self.__handle_client_connection(client_sock)
                except OSError as e:
                    if self._running:
                        logging.error(f"action: accept_connections | result: fail | error: {e}")
                    break
        except KeyboardInterrupt:
            logging.info("action: sigterm_received | result: success")
        finally:
            if self._server_socket:
                logging.info("action: close_server_socket | result: success")
                self._server_socket.close()

    def __handle_client_connection(self, client_sock):
        """
        Read bet data from client and store it
        
        Protocol: client_id(4), nombre_len(4), nombre, apellido_len(4), apellido, documento(4), nacimiento(4), numero(4)
        """
        try:            
            # Client ID (4 bytes)
            client_id_bytes = self._recv_all(client_sock, 4)
            if not client_id_bytes:
                return
            client_id = struct.unpack('!I', client_id_bytes)[0]
            
            # Nombre length (4 bytes)
            nombre_len_bytes = self._recv_all(client_sock, 4)
            if not nombre_len_bytes:
                return
            nombre_len = struct.unpack('!I', nombre_len_bytes)[0]

            # Nombre
            nombre_bytes = self._recv_all(client_sock, nombre_len)
            if not nombre_bytes:
                return
            nombre = nombre_bytes.decode('utf-8')

            # Apellido length (4 bytes)
            apellido_len_bytes = self._recv_all(client_sock, 4)
            if not apellido_len_bytes:
                return
            apellido_len = struct.unpack('!I', apellido_len_bytes)[0]

            # Apellido
            apellido_bytes = self._recv_all(client_sock, apellido_len)
            if not apellido_bytes:
                return
            apellido = apellido_bytes.decode('utf-8')

            # Documento (4 bytes)
            documento_bytes = self._recv_all(client_sock, 4)
            if not documento_bytes:
                return
            documento = struct.unpack('!I', documento_bytes)[0]

            # Nacimiento (4 bytes)
            nacimiento_bytes = self._recv_all(client_sock, 4)
            if not nacimiento_bytes:
                return
            nacimiento_int = struct.unpack('!I', nacimiento_bytes)[0]

            year = nacimiento_int // 10000
            month = (nacimiento_int % 10000) // 100
            day = nacimiento_int % 100
            nacimiento = f"{year:04d}-{month:02d}-{day:02d}"
            
            # Numero (4 bytes)
            numero_bytes = self._recv_all(client_sock, 4)
            if not numero_bytes:
                return
            numero = struct.unpack('!I', numero_bytes)[0]

            bet = Bet(str(client_id), nombre, apellido, str(documento), nacimiento, str(numero))
            store_bets([bet])

            logging.info(f'action: apuesta_almacenada | result: success | dni: {documento} | numero: {numero}')
            client_sock.send(bytes([RESPONSE_OK]))

        except Exception as e:
            logging.error(f"action: receive_message | result: fail | error: {e}")
            try:
                client_sock.send(bytes([RESPONSE_ERROR]))
            except:
                pass
        finally:
            logging.info("action: close_client_socket | result: success")
            client_sock.close()

    def _recv_all(self, sock, n):
        data = b''
        while len(data) < n:
            packet = sock.recv(n - len(data))
            if not packet:
                return None
            data += packet
        return data

    def __accept_new_connection(self):
        """
        Accept new connections

        Function blocks until a connection to a client is made.
        Then connection created is printed and returned
        """

        # Connection arrived
        logging.info('action: accept_connections | result: in_progress')
        c, addr = self._server_socket.accept()
        logging.info(f'action: accept_connections | result: success | ip: {addr[0]}')
        return c
