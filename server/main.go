package main

import (
	"context"
	"fmt"
	"log"
	g "minesweeper/game"
	pb "minesweeper/proto"
	"net"
	"os"

	"github.com/hajimehoshi/ebiten/v2"
	"google.golang.org/grpc"
)

type gameServer struct {
	pb.UnimplementedMinesweeperServer
	game *g.Game
}

func startGRPCServer(game *g.Game) {
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		fmt.Errorf("failed to listen: %v", err)
		os.Exit(1)
	}

	s := grpc.NewServer()
	pb.RegisterMinesweeperServer(s, &gameServer{game: game})

	log.Printf("Starting gRPC server on :50051")
	if err := s.Serve(lis); err != nil {
		fmt.Errorf("failed to serve: %v", err)
		os.Exit(1)
	}
}

func startEbitenWindow(game *g.Game) {
	ebiten.SetWindowSize(g.DefaultWindowWidth, g.DefaultWindowHeight)
	ebiten.SetWindowTitle("MineSweeper")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)

	game.AudioManager.LoadSound("totalmenchi", "assets/sounds/totalmenchi.mp3")

	if err := ebiten.RunGame(game); err != nil {
		fmt.Errorf("ebiten error: %v", err)
		os.Exit(1)
	}
}

func main() {
	game, err := g.NewGame(g.Medium)
	if err != nil {
		log.Fatal(err)
	}

	go startGRPCServer(game)

	startEbitenWindow(game)
}

func (s *gameServer) MakeMove(ctx context.Context, move *pb.Move) (*pb.Empty, error) {
	var action g.MouseAction
	switch move.Action {
	case 0:
		action = g.LeftClick
	case 1:
		action = g.RightClick
	default:
		return nil, fmt.Errorf("invalid action: %d", move.Action)
	}

	coord := g.Coordinates{X: int(move.X), Y: int(move.Y)}

	posx, posy := s.game.BoardToScreen(coord)

	pos := g.Coordinates{X: int(posx), Y: int(posy)}

	err := s.game.HandleInput(pos, action)
	if err != nil {
		return nil, err
	}

	return &pb.Empty{}, nil
}
