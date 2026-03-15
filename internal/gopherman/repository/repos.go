package repository

type Repositories struct {
	User       UserRepository
	Order      OrderRepository
	Withdrawal WithdrawalRepository
}
