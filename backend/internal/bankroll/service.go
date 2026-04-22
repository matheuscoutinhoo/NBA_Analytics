package bankroll

import (
	"fmt"

	"github.com/matheuscoutinhoo/better/internal/models"
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Deposit(userID int64, amount float64) (*models.Bankroll, error) {
	bankroll, err := s.repo.GetBankroll(userID)
	if err != nil {
		return nil, err
	}

	if bankroll == nil {
		return s.repo.CreateBankroll(userID, amount)
	}

	newBalance := bankroll.CurrentAmount + amount
	if err := s.repo.UpdateBalance(userID, newBalance); err != nil {
		return nil, err
	}

	s.repo.CreateTransaction(&models.BankrollTransaction{
		UserID:       userID,
		Type:         "deposit",
		Amount:       amount,
		BalanceAfter: newBalance,
	})

	return s.repo.GetBankroll(userID)
}

func (s *Service) Withdraw(userID int64, amount float64) (*models.Bankroll, error) {
	bankroll, err := s.repo.GetBankroll(userID)
	if err != nil {
		return nil, err
	}
	if bankroll == nil {
		return nil, fmt.Errorf("no bankroll found")
	}
	if bankroll.CurrentAmount < amount {
		return nil, fmt.Errorf("insufficient balance")
	}

	newBalance := bankroll.CurrentAmount - amount
	if err := s.repo.UpdateBalance(userID, newBalance); err != nil {
		return nil, err
	}

	s.repo.CreateTransaction(&models.BankrollTransaction{
		UserID:       userID,
		Type:         "withdraw",
		Amount:       amount,
		BalanceAfter: newBalance,
	})

	return s.repo.GetBankroll(userID)
}

func (s *Service) DeductBet(userID int64, betID int64, amount float64) error {
	bankroll, err := s.repo.GetBankroll(userID)
	if err != nil || bankroll == nil {
		return fmt.Errorf("no bankroll found")
	}

	newBalance := bankroll.CurrentAmount - amount
	if err := s.repo.UpdateBalance(userID, newBalance); err != nil {
		return err
	}

	return s.repo.CreateTransaction(&models.BankrollTransaction{
		UserID:       userID,
		BetID:        &betID,
		Type:         "bet_placed",
		Amount:       -amount,
		BalanceAfter: newBalance,
	})
}

func (s *Service) CreditWin(userID int64, betID int64, amount float64) error {
	bankroll, err := s.repo.GetBankroll(userID)
	if err != nil || bankroll == nil {
		return fmt.Errorf("no bankroll found")
	}

	newBalance := bankroll.CurrentAmount + amount
	if err := s.repo.UpdateBalance(userID, newBalance); err != nil {
		return err
	}

	return s.repo.CreateTransaction(&models.BankrollTransaction{
		UserID:       userID,
		BetID:        &betID,
		Type:         "bet_won",
		Amount:       amount,
		BalanceAfter: newBalance,
	})
}

func (s *Service) RefundBet(userID int64, betID int64, amount float64) error {
	bankroll, err := s.repo.GetBankroll(userID)
	if err != nil || bankroll == nil {
		return fmt.Errorf("no bankroll found")
	}

	newBalance := bankroll.CurrentAmount + amount
	if err := s.repo.UpdateBalance(userID, newBalance); err != nil {
		return err
	}

	return s.repo.CreateTransaction(&models.BankrollTransaction{
		UserID:       userID,
		BetID:        &betID,
		Type:         "bet_void",
		Amount:       amount,
		BalanceAfter: newBalance,
	})
}
