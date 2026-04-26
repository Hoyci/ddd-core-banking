package usecases

import (
	"ddd-core-banking/internal/account/domain/entity"
	"ddd-core-banking/internal/account/domain/errors"
	"ddd-core-banking/internal/account/domain/repository"
	"fmt"
)

type TransferFundsUseCase struct {
	transactionRepo repository.TransactionRepository
	accountRepo     repository.AccountRepository
}

func NewTransferFundsUseCase(
	transactionRepo repository.TransactionRepository,
	accountRepo repository.AccountRepository,
) *TransferFundsUseCase {
	return &TransferFundsUseCase{
		transactionRepo: transactionRepo,
		accountRepo:     accountRepo,
	}
}

type TransferFundsInput struct {
	SenderAccountID   string
	ReceiverAccountID string
	Amount            int64
}

func (uc *TransferFundsUseCase) Execute(input TransferFundsInput) error {
	sender, err := uc.accountRepo.FindByID(input.SenderAccountID)
	if err != nil {
		return err
	}
	receiver, err := uc.accountRepo.FindByID(input.ReceiverAccountID)
	if err != nil {
		return err
	}

	if sender.Blocked() != nil || receiver.Blocked() != nil {
		return errors.ErrAccountBlocked
	}

	if input.SenderAccountID == input.ReceiverAccountID {
		return errors.ErrSelfTransferNotAllowed
	}

	balance, err := uc.accountRepo.GetBalanceByAccountID(input.SenderAccountID)
	if err != nil {
		return fmt.Errorf("getting balance: %w", err)
	}
	if balance < input.Amount {
		return errors.ErrInsufficientFunds
	}

	transaction, err := entity.CreateTransaction(entity.CreateTransactionInput{
		SenderAccountID:   input.SenderAccountID,
		ReceiverAccountID: input.ReceiverAccountID,
		Amount:            input.Amount,
	})
	if err != nil {
		return err
	}

	if err := uc.transactionRepo.Save(transaction); err != nil {
		return fmt.Errorf("transfering funds: %w", err)
	}

	return nil
}
