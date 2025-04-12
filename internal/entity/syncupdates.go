package entity

type PasswordsUpdates struct {
	ToAdd    []*Password
	ToUpdate []*Password
	ToSync   []*Password
}
