package fixedexpense

import (
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestPaymentMethodValidation(t *testing.T) {
	for _, method := range []PaymentMethod{PaymentMethodCreditCard, PaymentMethodDebitCard, PaymentMethodPIX, PaymentMethodAutomaticDebit, PaymentMethodVoucher} {
		t.Run(string(method), func(t *testing.T) {
			expense := newValidFixedExpense(t)
			expense.PaymentMethod = method
			require.NoError(t, expense.Validate())
		})
	}
	for _, method := range []PaymentMethod{"", "CASH", "credit_card"} {
		t.Run("invalid_"+string(method), func(t *testing.T) {
			expense := newValidFixedExpense(t)
			expense.PaymentMethod = method
			require.ErrorIs(t, expense.Validate(), ErrInvalidPaymentMethod)
		})
	}
}

func TestServiceRejectsInvalidPaymentWithoutMutation(t *testing.T) {
	service, storage := newService(t)
	expense := stubFixedExpense()
	before := *expense
	storage.On("GetByIDAndGuild", mock.Anything, int64(1), int64(1)).Return(expense, nil)
	invalid := PaymentMethod("CASH")
	amount := decimal.NewFromInt(999)
	_, err := service.Update(t.Context(), 1, 1, UpdateInput{PaymentMethod: &invalid, Amount: &amount}, validActor())
	require.ErrorIs(t, err, ErrInvalidInput)
	require.Equal(t, before, *expense)
	storage.AssertNotCalled(t, "Update")
	storage.AssertExpectations(t)
}
