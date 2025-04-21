package server

import (
	"context"
	"errors"

	"github.com/llravell/go-pass/internal/entity"
)

type CardsUseCase struct {
	repo CardsRepository
}

func NewCardsUseCase(repo CardsRepository) *CardsUseCase {
	return &CardsUseCase{
		repo: repo,
	}
}

func (uc *CardsUseCase) AddNewCard(
	ctx context.Context,
	userID int,
	card *entity.Card,
) error {
	return uc.repo.AddNewCard(ctx, userID, card)
}

func (uc *CardsUseCase) DeleteCardByName(
	ctx context.Context,
	userID int,
	name string,
) error {
	return uc.repo.DeleteCardByName(ctx, userID, name)
}

func (uc *CardsUseCase) GetCards(
	ctx context.Context,
	userID int,
) ([]*entity.Card, error) {
	return uc.repo.GetCards(ctx, userID)
}

func (uc *CardsUseCase) SyncCard(
	ctx context.Context,
	userID int,
	card *entity.Card,
) error {
	err := uc.repo.UpdateByName(
		ctx,
		userID,
		card.Name,
		func(actualCard *entity.Card) (*entity.Card, error) {
			return entity.ChooseMostActualEntity(actualCard, card)
		},
	)
	if err != nil {
		if !errors.Is(err, entity.ErrCardDoesNotExist) {
			return err
		}

		err = uc.AddNewCard(ctx, userID, card)
		if err != nil {
			return err
		}
	}

	return nil
}
