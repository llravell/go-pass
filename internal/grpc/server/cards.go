package server

import (
	"context"
	"errors"

	"github.com/llravell/go-pass/internal/entity"
	usecase "github.com/llravell/go-pass/internal/usecase/server"
	pb "github.com/llravell/go-pass/pkg/grpc"
	"github.com/rs/zerolog"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
)

type CardsServer struct {
	pb.UnimplementedCardsServer

	cardsUC *usecase.CardsUseCase
	log     *zerolog.Logger
}

func NewCardsServer(
	cardsUC *usecase.CardsUseCase,
	log *zerolog.Logger,
) *CardsServer {
	return &CardsServer{
		cardsUC: cardsUC,
		log:     log,
	}
}

func (s *CardsServer) Sync(ctx context.Context, in *pb.Card) (*pb.CardSyncResponse, error) {
	userID, ok := GetUserIDFromContext(ctx)
	if !ok {
		s.log.Error().Msg("getting userID from ctx failed")

		return nil, status.Error(codes.Unauthenticated, "failed to resolve user id")
	}

	err := s.cardsUC.SyncCard(ctx, userID, entity.NewCardFromPB(in))
	if err != nil {
		var conflictErr *entity.ConflictError[*entity.Card]

		if errors.As(err, &conflictErr) {
			s.log.Info().
				Str("conflict_type", string(conflictErr.Type())).
				Msg("sync conflict")

			return &pb.CardSyncResponse{
				Success: false,
				Conflict: &pb.CardConflict{
					Card: conflictErr.Actual().ToPB(),
					Type: conflictErr.TypePB(),
				},
			}, nil
		}

		s.log.Error().Err(err).Msg("sync failed")

		return nil, status.Error(codes.Internal, "sync failed")
	}

	return &pb.CardSyncResponse{Success: true}, nil
}

func (s *CardsServer) Delete(ctx context.Context, in *pb.CardDeleteRequest) (*emptypb.Empty, error) {
	userID, ok := GetUserIDFromContext(ctx)
	if !ok {
		s.log.Error().Msg("getting userID from ctx failed")

		return nil, status.Error(codes.Unauthenticated, "failed to resolve user id")
	}

	err := s.cardsUC.DeleteCardByName(ctx, userID, in.GetName())
	if err != nil {
		s.log.Error().Err(err).Msg("card deleting failed")

		return nil, status.Error(codes.Internal, "deleting failed")
	}

	return &emptypb.Empty{}, nil
}

func (s *CardsServer) GetList(ctx context.Context, _ *emptypb.Empty) (*pb.CardGetListResponse, error) {
	userID, ok := GetUserIDFromContext(ctx)
	if !ok {
		s.log.Error().Msg("getting userID from ctx failed")

		return nil, status.Error(codes.Unauthenticated, "failed to resolve user id")
	}

	Cards, err := s.cardsUC.GetCards(ctx, userID)
	if err != nil {
		s.log.Error().Err(err).Msg("cards fetching failed")

		return nil, status.Error(codes.Internal, "cards fetching failed")
	}

	response := &pb.CardGetListResponse{
		Cards: make([]*pb.Card, 0, len(Cards)),
	}

	for _, Card := range Cards {
		response.Cards = append(response.Cards, Card.ToPB())
	}

	return response, nil
}
