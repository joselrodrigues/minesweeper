import os
import sys
import time
import grpc
import logging
from typing import Tuple, List, Optional
from contextlib import contextmanager

root_dir = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
sys.path.append(root_dir)
sys.path.append(os.path.join(root_dir, "proto"))

from proto.minesweeper_pb2 import Move, Empty
from proto.minesweeper_pb2_grpc import MinesweeperStub

logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)


class GRPCError(Exception):
    """Custom exception for gRPC errors"""

    pass


class MinesweeperAPI:
    def __init__(self):
        self.channel = None
        self.stub = None
        self._max_retries = 3
        self._retry_delay = 1
        self._connect()

    def _connect(self) -> None:
        """Establece la conexión gRPC con reintentos"""
        for attempt in range(self._max_retries):
            try:
                if self.channel is not None:
                    self.channel.close()

                options = [
                    ("grpc.enable_retries", 1),
                    ("grpc.max_send_message_length", 4 * 1024 * 1024),
                    ("grpc.max_receive_message_length", 4 * 1024 * 1024),
                    ("grpc.max_message_length", 4 * 1024 * 1024),
                    ("grpc.keepalive_time_ms", 10000),
                    ("grpc.keepalive_timeout_ms", 5000),
                    ("grpc.keepalive_permit_without_calls", 1),
                    ("grpc.http2.min_time_between_pings_ms", 10000),
                    ("grpc.http2.max_pings_without_data", 0),
                ]

                self.channel = grpc.insecure_channel("localhost:50051", options=options)
                self.stub = MinesweeperStub(self.channel)

                # Verificar conexión
                grpc.channel_ready_future(self.channel).result(timeout=5)
                logger.info("Connected to gRPC server")
                return

            except grpc.FutureTimeoutError as e:
                logger.warning(f"Connection attempt {attempt + 1} failed: {e}")
                if attempt < self._max_retries - 1:
                    time.sleep(self._retry_delay)
                    self._retry_delay *= 2
                else:
                    raise GRPCError("Failed to connect to gRPC server") from e

    @contextmanager
    def _handle_grpc_errors(self):
        """Context manager para manejar errores gRPC"""
        try:
            yield
        except grpc.RpcError as e:
            status_code = e.code()
            if status_code == grpc.StatusCode.UNAVAILABLE:
                logger.warning("Server unavailable, attempting to reconnect...")
                self._connect()
                raise GRPCError("Server temporarily unavailable, please retry") from e
            elif status_code == grpc.StatusCode.DEADLINE_EXCEEDED:
                raise GRPCError("Request timeout") from e
            else:
                raise GRPCError(f"gRPC error: {e.details()}") from e
        except Exception as e:
            raise GRPCError(f"Unexpected error: {str(e)}") from e

    def convert_board_to_array(self, proto_board) -> List[List[int]]:
        """Convierte el tablero del formato protobuf a una lista de listas"""
        return [list(board.Cell) for board in proto_board]

    def make_move(
        self, x: int, y: int, action: int
    ) -> Tuple[List[List[int]], int, int]:
        """
        Realiza un movimiento en el juego.

        Args:
            x: Coordenada X
            y: Coordenada Y
            action: Tipo de acción (0: RevealCell, 1: ToggleFlag)

        Returns:
            Tuple[List[List[int]], int, int]: (tablero, recompensa, estado)
        """
        with self._handle_grpc_errors():
            move = Move(x=x, y=y, action=action)
            response = self.stub.MakeMove(move, timeout=5, wait_for_ready=True)
            return (
                self.convert_board_to_array(response.Board),
                response.Reward,
                response.State,
            )

    def reset(self) -> List[List[int]]:
        """Reinicia el juego y retorna el estado inicial del tablero"""
        with self._handle_grpc_errors():
            response = self.stub.Reset(Empty(), timeout=5, wait_for_ready=True)
            return self.convert_board_to_array(response.Board)

    def __del__(self):
        """Cleanup al destruir la instancia"""
        if self.channel is not None:
            self.channel.close()
