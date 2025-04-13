package entity

import (
	"time"

	"github.com/llravell/go-pass/pkg/encryption"
	pb "github.com/llravell/go-pass/pkg/grpc"
	timestamppb "google.golang.org/protobuf/types/known/timestamppb"
)

type Card struct {
	Name           string
	CardholderName string
	Number         string
	CVV            string
	Meta           string
	ExpirationDate time.Time
	Version        int
	Deleted        bool
}

func (card *Card) GetVersion() int {
	return card.Version
}

func (card *Card) IsDeleted() bool {
	return card.Deleted
}

func (card *Card) BumpVersion() {
	card.Version++
}

func (card *Card) Open(key *encryption.Key) error {
	decryptedNumber, err := key.Decrypt(card.Number)
	if err != nil {
		return err
	}

	decryptedCVV, err := key.Decrypt(card.CVV)
	if err != nil {
		return err
	}

	card.Number = decryptedNumber
	card.CVV = decryptedCVV

	return nil
}

func (card *Card) Close(key *encryption.Key) error {
	encryptedNumber, err := key.Encrypt(card.Number)
	if err != nil {
		return err
	}

	encryptedCVV, err := key.Encrypt(card.CVV)
	if err != nil {
		return err
	}

	card.Number = encryptedNumber
	card.CVV = encryptedCVV

	return nil
}

func (card *Card) Equal(target *Card) bool {
	return (card.Name == target.Name &&
		card.CardholderName == target.CardholderName &&
		card.Number == target.Number &&
		card.CVV == target.CVV &&
		card.ExpirationDate.Year() == target.ExpirationDate.Year() &&
		card.ExpirationDate.Month() == target.ExpirationDate.Month() &&
		card.Meta == target.Meta &&
		card.Version == target.Version)
}

func (card *Card) Clone() *Card {
	return &Card{
		Name:           card.Name,
		CardholderName: card.CardholderName,
		Number:         card.Number,
		CVV:            card.CVV,
		Meta:           card.Meta,
		ExpirationDate: card.ExpirationDate,
		Version:        card.Version,
		Deleted:        card.Deleted,
	}
}

func (card *Card) ToPB() *pb.Card {
	return &pb.Card{
		Name:           card.Name,
		CardholderName: card.CardholderName,
		Number:         card.Number,
		Cvv:            card.CVV,
		Meta:           card.Meta,
		ExpirationDate: timestamppb.New(card.ExpirationDate),
		Version:        int32(card.Version), //nolint:gosec
	}
}

func NewCardFromPB(card *pb.Card) *Card {
	return &Card{
		Name:           card.GetName(),
		CardholderName: card.GetCardholderName(),
		Number:         card.GetNumber(),
		CVV:            card.GetCvv(),
		Meta:           card.GetMeta(),
		ExpirationDate: card.GetExpirationDate().AsTime(),
		Version:        int(card.GetVersion()),
	}
}
