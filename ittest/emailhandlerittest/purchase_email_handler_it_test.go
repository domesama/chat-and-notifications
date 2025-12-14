package ittest

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/domesama/chat-and-notifications/emailhandler/service"
	"github.com/domesama/chat-and-notifications/ittest/ittesthelper"
	"github.com/domesama/chat-and-notifications/ittest/stub"
	"github.com/domesama/chat-and-notifications/model"
	"github.com/stretchr/testify/suite"
)

type PurchaseEmailHandlerITTestSuite struct {
	BaseEmailHandlerITTestSuite
}

func (t *PurchaseEmailHandlerITTestSuite) SetupSuite() {
	t.BaseEmailHandlerITTestSuite.SetupSuite()
}

func TestPurchaseEmailHandlerITTestSuite(t *testing.T) {
	suite.Run(t, new(PurchaseEmailHandlerITTestSuite))
}

func (t *PurchaseEmailHandlerITTestSuite) callPurchaseMailingServer(
	ctx context.Context,
	expectedStatusCode int,
	purchase ...model.PurchaseUpdate,
) {
	for _, p := range purchase {
		t.callEmailRoute(ctx, "/email/purchased", expectedStatusCode, p)
	}
}

func (t *PurchaseEmailHandlerITTestSuite) TestPurchaseEmailSendingSuccess() {
	ctx := context.Background()

	t.insertEmailInfo(
		ctx, service.EmailInfo{
			UserID: "shop-owner-1",
			Email:  "shopowner1@example.com",
			Name:   "Shop Owner One",
		},
	)
	t.insertEmailInfo(
		ctx, service.EmailInfo{
			UserID: "shop-owner-2",
			Email:  "shopowner2@example.com",
			Name:   "Shop Owner Two",
		},
	)

	// Add stub for successful email sending with verification
	t.cnt.SimpleEmailSender.AddStub(
		ittesthelper.SimpleEmailSenderStub{
			Predicates: stub.NewPredicates(
				stub.WithReceiverEmail("shopowner2@example.com"),
				stub.WithEmailSubject("New purchase order: Premium Coffee"),
			),
			ExpectedError: nil,
		},
	)

	purchase := stub.CreatePurchaseUpdate(
		"ORD-001",
		"buyer-1",
		"shop-owner-2",
		"Premium Coffee",
		15.99,
	)

	t.callPurchaseMailingServer(ctx, http.StatusCreated, purchase)
}

func (t *PurchaseEmailHandlerITTestSuite) TestPurchaseEmailSendingFailure() {
	ctx := context.Background()

	// Setup EmailInfo data
	t.insertEmailInfo(
		ctx, service.EmailInfo{
			UserID: "shop-owner-3",
			Email:  "shopowner3@example.com",
			Name:   "Shop Owner Three",
		},
	)
	t.insertEmailInfo(
		ctx, service.EmailInfo{
			UserID: "shop-owner-4",
			Email:  "shopowner4@example.com",
			Name:   "Shop Owner Four",
		},
	)

	t.cnt.SimpleEmailSender.AddStub(
		ittesthelper.SimpleEmailSenderStub{
			Predicates: stub.NewPredicates(
				stub.WithReceiverEmail("shopowner4@example.com"),
			),
			ExpectedError: errors.New("failed to send email"),
		},
	)

	purchase := stub.CreatePurchaseUpdate(
		"ORD-002",
		"buyer-2",
		"shop-owner-4",
		"Wireless Mouse",
		29.99,
	)

	t.callPurchaseMailingServer(ctx, http.StatusInternalServerError, purchase)
}

func (t *PurchaseEmailHandlerITTestSuite) TestPurchaseMissingEmailInfo() {
	ctx := context.Background()

	// Do not insert EmailInfo for shop owner - simulating missing user email data
	purchase := stub.CreatePurchaseUpdate(
		"ORD-003",
		"buyer-5",
		"shop-owner-nonexistent",
		"Notebook Set",
		12.50,
	)

	t.callPurchaseMailingServer(ctx, http.StatusInternalServerError, purchase)
}
