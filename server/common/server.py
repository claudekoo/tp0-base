import socket
import logging
import signal
import sys
import threading
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
        
        self._client_threads = []
        self._storage_lock = threading.Lock()
        self._finished_agencies_lock = threading.Lock()
        self._lottery_lock = threading.Lock()
        
        signal.signal(signal.SIGTERM, self._signal_handler)

    def _signal_handler(self, sig, frame):
        logging.info("action: sigterm_received | result: success")
        self._running = False
        if self._server_socket:
            logging.info("action: close_server_socket | result: success")
            self._server_socket.close()
        
        for thread in self._client_threads:
            if thread.is_alive():
                thread.join(timeout=5.0)
        
        logging.info("action: server_shutdown | result: success")
        sys.exit(0)

    def run(self):
        """
        Server loop that accepts connections and handles them in parallel using threads

        Server that accepts new connections and creates a new thread to handle
        each client connection in parallel.
        """

        try:
            while self._running:
                try:
                    client_sock = self.__accept_new_connection()
                    client_thread = threading.Thread(
                        target=self.__handle_client_connection,
                        args=(client_sock,),
                        daemon=True
                    )
                    
                    self._client_threads.append(client_thread)
                    client_thread.start()
                    self._client_threads = [t for t in self._client_threads if t.is_alive()]
                    
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
            
            for thread in self._client_threads:
                if thread.is_alive():
                    thread.join(timeout=5.0)

    def __handle_client_connection(self, client_sock):
        """
        Handle multiple messages from a single client connection
        Keep the connection open until client disconnects or an error occurs
        """
        client_addr = client_sock.getpeername()
        try:
            while True:
                msg_type = receive_message_type(client_sock)
                if msg_type is None:
                    logging.debug(f"action: client_disconnected | result: success | ip: {client_addr[0]}")
                    break
                    
                if msg_type == MESSAGE_TYPE_BATCH:
                    self.__handle_bet_batch(client_sock)
                elif msg_type == MESSAGE_TYPE_FINISHED_SENDING:
                    self.__handle_finished_notification(client_sock)
                elif msg_type == MESSAGE_TYPE_QUERY_WINNERS:
                    self.__handle_query_winners(client_sock)
                else:
                    logging.error(f"action: handle_client_connection | result: fail | error: unknown message type: {msg_type}")
                    send_response(client_sock, False)
                    break

        except Exception as e:
            logging.error(f"action: handle_client_connection | result: fail | error: {e}")
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
                    
                    with self._storage_lock:
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
                with self._finished_agencies_lock:
                    self._finished_agencies.add(client_id)
                    logging.debug(f"action: finished_notification_received | result: success | client_id: {client_id} | finished_agencies: {len(self._finished_agencies)}")
                    
                    with self._lottery_lock:
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
                with self._lottery_lock:
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
            with self._storage_lock:
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
