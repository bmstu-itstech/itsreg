package dto

type UserMailingResult struct {
	UserID   int64
	Success  bool
	ErrorMsg *string
}
