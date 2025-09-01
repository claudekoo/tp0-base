import socket
import logging
import signal
import sys
from .utils import Bet, store_bets
from .protocol import receive_bet, send_response


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
        Read bet data from client and store it using protocol functions
        """
        try:
            bet_data = receive_bet(client_sock)
            if bet_data is None:
                send_response(client_sock, False)
                return

            client_id, nombre, apellido, documento, nacimiento, numero = bet_data
            
            bet = Bet(client_id, nombre, apellido, documento, nacimiento, numero)
            store_bets([bet])

            logging.info(f'action: apuesta_almacenada | result: success | dni: {documento} | numero: {numero}')
            send_response(client_sock, True)

        except Exception as e:
            logging.error(f"action: receive_message | result: fail | error: {e}")
            send_response(client_sock, False)
        finally:
            logging.info("action: close_client_socket | result: success")
            client_sock.close()

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
