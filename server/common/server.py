import socket
import logging
import signal
import sys
from .utils import Bet, store_bets, load_bets, has_won
from .protocol import receive_bet_batch, send_response, receive_message_type, receive_finished_notification, receive_query_winners, send_winners, MESSAGE_TYPE_BATCH, MESSAGE_TYPE_FINISHED_SENDING, MESSAGE_TYPE_QUERY_WINNERS


class Server:
    def __init__(self, port, listen_backlog, num_agencies):
        # Initialize server socket
        self._server_socket = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        self._server_socket.setsockopt(socket.SOL_SOCKET, socket.SO_REUSEADDR, 1)
        self._server_socket.bind(('', port))
        self._server_socket.listen(listen_backlog)
        self._running = True
        
        self._finished_agencies = set()
        self._lottery_done = False
        self._num_agencies = num_agencies
        
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
        Handle different types of messages from clients
        """
        try:
            msg_type = receive_message_type(client_sock)
            if msg_type is None:
                send_response(client_sock, False)
                return
                
            if msg_type == MESSAGE_TYPE_BATCH:
                self.__handle_bet_batch(client_sock)
            elif msg_type == MESSAGE_TYPE_FINISHED_SENDING:
                self.__handle_finished_notification(client_sock)
            elif msg_type == MESSAGE_TYPE_QUERY_WINNERS:
                self.__handle_query_winners(client_sock)
            else:
                logging.error(f"action: handle_client_connection | result: fail | error: unknown message type: {msg_type}")
                send_response(client_sock, False)

        except Exception as e:
            logging.error(f"action: handle_client_connection | result: fail | error: {e}")
            send_response(client_sock, False)
        finally:
            logging.info("action: close_client_socket | result: success")
            client_sock.close()

    def __handle_bet_batch(self, client_sock):
        """
        Read batch bet data from client and store it
        """
        try:
            batch_data = receive_bet_batch(client_sock)
            if batch_data is not None:
                client_id, bets_data = batch_data
                cantidad = len(bets_data)
                
                try:
                    bets = []
                    for bet_data in bets_data:
                        nombre, apellido, documento, nacimiento, numero = bet_data
                        bet = Bet(client_id, nombre, apellido, documento, nacimiento, numero)
                        bets.append(bet)
                    
                    store_bets(bets)
                    
                    logging.info(f'action: apuesta_recibida | result: success | cantidad: {cantidad}')
                    
                    send_response(client_sock, True)
                    return
                    
                except Exception as batch_error:
                    logging.info(f'action: apuesta_recibida | result: fail | cantidad: {cantidad}')
                    send_response(client_sock, False)
                    return
            else:
                # No batch data
                send_response(client_sock, False)
                return

        except Exception as e:
            logging.error(f"action: receive_message | result: fail | error: {e}")
            send_response(client_sock, False)

    def __handle_finished_notification(self, client_sock):
        try:
            client_id = receive_finished_notification(client_sock)
            if client_id is not None:
                self._finished_agencies.add(client_id)
                logging.debug(f"action: finished_notification_received | result: success | client_id: {client_id} | finished_agencies: {len(self._finished_agencies)}")
                
                if len(self._finished_agencies) == self._num_agencies and not self._lottery_done:
                    self._lottery_done = True
                    logging.info("action: sorteo | result: success")
            
                send_response(client_sock, True)
            else:
                send_response(client_sock, False)
                
        except Exception as e:
            logging.error(f"action: handle_finished_notification | result: fail | error: {e}")
            send_response(client_sock, False)

    def __handle_query_winners(self, client_sock):
        try:
            client_id = receive_query_winners(client_sock)
            if client_id is not None:
                if not self._lottery_done:
                    send_response(client_sock, False)
                    return
            
                winners = self.__get_winners_for_agency(client_id)
                send_winners(client_sock, winners)
                
            else:
                send_response(client_sock, False)
                
        except Exception as e:
            logging.error(f"action: handle_query_winners | result: fail | error: {e}")
            send_response(client_sock, False)

    def __get_winners_for_agency(self, agency_id: str) -> list[str]:
        try:
            winners = []
            for bet in load_bets():
                if str(bet.agency) == agency_id and has_won(bet):
                    winners.append(bet.document)
            
            logging.debug(f"action: get_winners_for_agency | result: success | agency_id: {agency_id} | winners_count: {len(winners)}")
            return winners
            
        except Exception as e:
            logging.error(f"action: get_winners_for_agency | result: fail | agency_id: {agency_id} | error: {e}")
            return []

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
