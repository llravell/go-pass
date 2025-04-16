package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/llravell/go-pass/internal/entity"
)

type CardsPostgresRepository struct {
	conn *sql.DB
}

func NewCardsPostgresRepository(conn *sql.DB) *CardsPostgresRepository {
	return &CardsPostgresRepository{
		conn: conn,
	}
}

func (repo *CardsPostgresRepository) GetCards(
	ctx context.Context,
	userID int,
) ([]*entity.Card, error) {
	cards := make([]*entity.Card, 0)

	rows, err := repo.conn.QueryContext(ctx, `
		SELECT name, cardholder_name, number_encrypted, cvv_encrypted, expiration_date, meta, version
		FROM cards
		WHERE user_id=$1 AND NOT is_deleted;
	`, userID)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	for rows.Next() {
		var card entity.Card

		err = rows.Scan(
			&card.Name, &card.CardholderName, &card.Number, &card.CVV,
			&card.ExpirationDate, &card.Meta, &card.Version,
		)
		if err != nil {
			return nil, err
		}

		cards = append(cards, &card)
	}

	if err = rows.Err(); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return cards, nil
		}

		return nil, err
	}

	return cards, nil
}

func (repo *CardsPostgresRepository) AddNewCard(
	ctx context.Context,
	userID int,
	card *entity.Card,
) error {
	_, err := repo.conn.ExecContext(ctx, `
		INSERT INTO cards (name, cardholder_name, number_encrypted, cvv_encrypted, expiration_date, meta, version, user_id)
		VALUES
			($1, $2, $3, $4, $5);
	`, card.Name, card.CardholderName, card.Number, card.CVV,
		card.ExpirationDate, card.Meta, card.Version, userID)
	if err != nil {
		return err
	}

	return nil
}

func (repo *CardsPostgresRepository) DeleteCardByName(
	ctx context.Context,
	userID int,
	name string,
) error {
	_, err := repo.conn.ExecContext(ctx, `
		UPDATE cards
		SET is_deleted=TRUE
		WHERE user_id=$1 AND name=$2;
	`, userID, name)
	if err != nil {
		return err
	}

	return nil
}

func (repo *CardsPostgresRepository) UpdateByName(
	ctx context.Context,
	userID int,
	name string,
	updateFn func(Card *entity.Card) (*entity.Card, error),
) error {
	return runInTx(ctx, repo.conn, func(tx *sql.Tx) error {
		var card entity.Card

		row := tx.QueryRowContext(ctx, `
			SELECT name, cardholder_name, number_encrypted, cvv_encrypted, expiration_date, meta, version, is_deleted
			FROM Cards
			WHERE user_id=$1 AND name=$2
			FOR UPDATE;
		`, userID, name)

		err := row.Scan(
			&card.Name, &card.CardholderName, &card.Number, &card.CVV,
			&card.ExpirationDate, &card.Meta, &card.Version, &card.Deleted,
		)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return entity.ErrCardDoesNotExist
			}

			return err
		}

		updatedCard, err := updateFn(&card)
		if err != nil {
			return err
		}

		if updatedCard == nil {
			return nil
		}

		_, err = tx.ExecContext(ctx, `
			UPDATE Cards
			SET cardholder_name=$1, number_encrypted=$2, cvv_encrypted=$3, expiration_date=$4, meta=$5, version=$6, is_deleted=$7
			WHERE user_id=$7 AND name=$8;
		`, updatedCard.CardholderName, updatedCard.Number, updatedCard.CVV, updatedCard.ExpirationDate,
			updatedCard.Meta, updatedCard.Version, updatedCard.Deleted, userID, card.Name,
		)
		if err != nil {
			return err
		}

		return nil
	})
}
