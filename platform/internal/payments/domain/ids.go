package domain

import (
	"github.com/google/uuid"
)

type PaymentAccountID struct{ u uuid.UUID }
type PaymentTransactionID struct{ u uuid.UUID }
type PaymentImportID struct{ u uuid.UUID }
type PaymentImportItemID struct{ u uuid.UUID }

func NewPaymentAccountID() PaymentAccountID     { return PaymentAccountID{u: uuid.New()} }
func NewPaymentTransactionID() PaymentTransactionID { return PaymentTransactionID{u: uuid.New()} }
func NewPaymentImportID() PaymentImportID       { return PaymentImportID{u: uuid.New()} }
func NewPaymentImportItemID() PaymentImportItemID { return PaymentImportItemID{u: uuid.New()} }

func PaymentAccountIDFromUUID(u uuid.UUID) PaymentAccountID {
	return PaymentAccountID{u: u}
}
func PaymentTransactionIDFromUUID(u uuid.UUID) PaymentTransactionID {
	return PaymentTransactionID{u: u}
}
func PaymentImportIDFromUUID(u uuid.UUID) PaymentImportID {
	return PaymentImportID{u: u}
}
func PaymentImportItemIDFromUUID(u uuid.UUID) PaymentImportItemID {
	return PaymentImportItemID{u: u}
}

func (id PaymentAccountID) UUID() uuid.UUID     { return id.u }
func (id PaymentTransactionID) UUID() uuid.UUID { return id.u }
func (id PaymentImportID) UUID() uuid.UUID      { return id.u }
func (id PaymentImportItemID) UUID() uuid.UUID  { return id.u }

func (id PaymentAccountID) IsZero() bool     { return id.u == uuid.Nil }
func (id PaymentTransactionID) IsZero() bool { return id.u == uuid.Nil }
func (id PaymentImportID) IsZero() bool      { return id.u == uuid.Nil }
func (id PaymentImportItemID) IsZero() bool  { return id.u == uuid.Nil }
