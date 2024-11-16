package main

import (
	"context"
	"fmt"
	"log"
	g "minesweeper/game"
	pb "minesweeper/proto"
	"net"
	"os"
	"os/signal"
	"runtime"
	"sync"
	"syscall"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
)

type gameServer struct {
	pb.UnimplementedMinesweeperServer
	game *g.Game
	mu   sync.Mutex
}

var (
	kaProps = keepalive.ServerParameters{
		MaxConnectionIdle:     15 * time.Second,
		MaxConnectionAge:      30 * time.Second,
		MaxConnectionAgeGrace: 5 * time.Second,
		Time:                  5 * time.Second,
		Timeout:               1 * time.Second,
	}

	kaPolicy = keepalive.EnforcementPolicy{
		MinTime:             5 * time.Second,
		PermitWithoutStream: true,
	}
)

func newGameServer(game *g.Game) *gameServer {
	return &gameServer{
		game: game,
	}
}

func (s *gameServer) MakeMove(ctx context.Context, move *pb.Move) (*pb.GameState, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	defer func() {
		if r := recover(); r != nil {
			log.Printf("Recovered from panic in MakeMove: %v", r)
			runtime.GC()
		}
	}()

	var action g.ActionEvent
	switch move.Action {
	case 0:
		action = g.RevealCell
	case 1:
		action = g.ToggleFlag
	default:
		return nil, fmt.Errorf("invalid action: %d", move.Action)
	}

	coord := g.Coordinates{X: int(move.X), Y: int(move.Y)}
	oldCellState := s.game.GetCellState(coord)

	if err := s.game.HandleInput(coord, action); err != nil {
		return nil, err
	}

	modelState := s.game.ModelState()
	reward := s.game.CalculateModelReward(oldCellState, action)

	protoRows := make([]*pb.Row, len(modelState))
	for i, row := range modelState {
		protoRows[i] = &pb.Row{
			Cell: row,
		}
	}

	return &pb.GameState{
		Board:  protoRows,
		Reward: int32(reward),
		State:  int32(s.game.State),
	}, nil
}

func (s *gameServer) Reset(ctx context.Context, _ *pb.Empty) (*pb.GameState, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.game.Restart()

	modelState := s.game.ModelState()
	protoRows := make([]*pb.Row, len(modelState))
	for i, row := range modelState {
		protoRows[i] = &pb.Row{
			Cell: row,
		}
	}

	return &pb.GameState{
		Board:  protoRows,
		Reward: 0,
		State:  int32(s.game.State),
	}, nil
}

func startGRPCServer(game *g.Game) {
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	opts := []grpc.ServerOption{
		grpc.KeepaliveParams(kaProps),
		grpc.KeepaliveEnforcementPolicy(kaPolicy),
		grpc.MaxConcurrentStreams(100),
		grpc.WriteBufferSize(1024 * 1024),
		grpc.ReadBufferSize(1024 * 1024),
		grpc.MaxRecvMsgSize(4 * 1024 * 1024),
		grpc.MaxSendMsgSize(4 * 1024 * 1024),
		grpc.ConnectionTimeout(5 * time.Second),
	}

	s := grpc.NewServer(opts...)
	srv := newGameServer(game)
	pb.RegisterMinesweeperServer(s, srv)

	// Manejo graceful shutdown
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan
		log.Println("Shutting down gRPC server...")
		s.GracefulStop()
	}()

	log.Printf("Starting gRPC server on :50051")
	if err := s.Serve(lis); err != nil {
		log.Printf("failed to serve: %v", err)
	}
}

func main() {
	game, err := g.NewGame(g.Medium)
	if err != nil {
		log.Fatal(err)
	}

	// Canal para coordinar el cierre
	done := make(chan struct{})

	// Iniciar servidor gRPC
	go func() {
		startGRPCServer(game)
		close(done)
	}()

	// Iniciar ventana Ebiten
	ebiten.SetWindowSize(g.DefaultWindowWidth, g.DefaultWindowHeight)
	ebiten.SetWindowTitle("MineSweeper")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)

	if err := ebiten.RunGame(game); err != nil {
		log.Printf("Ebiten error: %v", err)
	}
	close(done)

	// Esperar a que el servidor gRPC termine
	<-done
}
