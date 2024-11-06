package main

import (
	"context"
	"fmt"
	"log"
	g "minesweeper/game"
	pb "minesweeper/proto"
	"net"

	"github.com/hajimehoshi/ebiten/v2"
	"google.golang.org/grpc"
)

type gameServer struct {
	pb.UnimplementedMinesweeperServer
	game *g.Game
}

//	func startGRPCServer(game *g.Game) error {
//		lis, err := net.Listen("tcp", ":50051")
//		if err != nil {
//			return fmt.Errorf("failed to listen: %v", err)
//		}
//
//		s := grpc.NewServer()
//		pb.RegisterMinesweeperServer(s, &gameServer{game: game})
//
//		log.Printf("Starting gRPC server on :50051")
//		if err := s.Serve(lis); err != nil {
//			return fmt.Errorf("failed to serve: %v", err)
//		}
//
//		return nil
//	}
func startGRPCServer(game *g.Game, errChan chan error) {
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		errChan <- fmt.Errorf("failed to listen: %v", err)
		return
	}

	s := grpc.NewServer()
	pb.RegisterMinesweeperServer(s, &gameServer{game: game})

	log.Printf("Starting gRPC server on :50051")
	if err := s.Serve(lis); err != nil {
		errChan <- fmt.Errorf("failed to serve: %v", err)
	}
}

func startEbitenWindow(game *g.Game, errChan chan error) {
	ebiten.SetWindowSize(g.DefaultWindowWidth, g.DefaultWindowHeight)
	ebiten.SetWindowTitle("MineSweeper")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	game.AudioManager.LoadSound("totalmenchi", "assets/sounds/totalmenchi.mp3")
	if err := ebiten.RunGame(game); err != nil {
		errChan <- fmt.Errorf("ebiten error: %v", err)
	}
}

func main() {
	game, err := g.NewGame(g.Medium)
	if err != nil {
		log.Fatal(err)
	}

	// // Iniciar el servidor gRPC en una goroutine
	// go func() {
	// 	if err := startGRPCServer(game); err != nil {
	// 		log.Printf("gRPC server error: %v", err)
	// 	}
	// }()
	//
	// // Configurar Ebiten
	// ebiten.SetWindowSize(g.DefaultWindowWidth, g.DefaultWindowHeight)
	// ebiten.SetWindowTitle("MineSweeper")
	// ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	//
	// // Iniciar Ebiten en el thread principal
	// if err := ebiten.RunGame(game); err != nil {
	// 	log.Fatal(err)
	// }
	//
	errChan := make(chan error, 2)

	// Iniciar servidor gRPC en una goroutine
	go startGRPCServer(game, errChan)

	// Iniciar Ebiten en otra goroutine
	startEbitenWindow(game, errChan)

	// Esperar por cualquier error
	if err := <-errChan; err != nil {
		log.Fatal(err)
	}
}

func (s *gameServer) MakeMove(ctx context.Context, move *pb.Move) (*pb.Empty, error) {
	log.Printf("Received move: x=%d, y=%d, action=%d", move.X, move.Y, move.Action)
	// Convertir el int32 a MouseAction
	var action g.MouseAction
	switch move.Action {
	case 0:
		action = g.LeftClick // Asegúrate de que estas constantes estén definidas en game
	case 1:
		action = g.RightClick
	default:
		return nil, fmt.Errorf("invalid action: %d", move.Action)
	}

	// Crear Coordinates usando el tipo del paquete game
	coord := g.Coordinates{X: int(move.X), Y: int(move.Y)}

	posx, posy := s.game.BoardToScreen(coord)

	pos := g.Coordinates{X: int(posx), Y: int(posy)}
	fmt.Printf("pos: %v\n", pos)
	err := s.game.HandleInput(pos, action)
	if err != nil {
		return nil, err
	}

	return &pb.Empty{}, nil
}

// func (s *gameServer) MakeMove(ctx context.Context, move *pb.Move) (*pb.Empty, error) {
// 	action := g.MouseAction(move.Action)
// 	pos := g.Coordinates{X: int(move.X), Y: int(move.Y)}
// 	s.game.HandleInput(pos, action)
// 	return &pb.Empty{}, nil
// }
//
// func main() {
// 	ebiten.SetWindowSize(g.DefaultWindowWidth, g.DefaultWindowHeight)
// 	ebiten.SetWindowTitle("MineSweeper")
// 	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
//
// 	game, err := g.NewGame(g.Medium)
// 	if err != nil {
// 		log.Fatal(err)
// 	}
//
// 	game.AudioManager.LoadSound("totalmenchi", "../game/assets/sounds/totalmenchi.mp3")
//
// 	if err := ebiten.RunGame(game); err != nil {
// 		log.Fatal(err)
// 	}
//
// 	lis, err := net.Listen("tcp", ":50051")
// 	if err != nil {
// 		log.Fatalf("failed to listen: %v", err)
// 	}
//
// 	s := grpc.NewServer()
// 	pb.RegisterMinesweeperServer(s, &gameServer{})
//
// 	log.Printf("Starting gRPC server on :50051")
// 	if err := s.Serve(lis); err != nil {
// 		log.Fatalf("failed to serve: %v", err)
// 	}
// }
