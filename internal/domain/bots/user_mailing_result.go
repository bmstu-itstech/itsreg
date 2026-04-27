package bots

import "errors"

type UserMailingResult struct {
	userID   UserID
	success  bool
	errorMsg *string
}

func NewSuccessMailingResult(userID UserID) UserMailingResult {
	return UserMailingResult{
		success: true,
		userID:  userID,
	}
}

func NewErrorMailingResult(userID UserID, msg string) UserMailingResult {
	return UserMailingResult{
		success:  false,
		userID:   userID,
		errorMsg: &msg,
	}
}

func RestoreUserMailingResult(userID UserID, success bool, errorMsg *string) (UserMailingResult, error) {
	if userID.IsZero() {
		return UserMailingResult{}, errors.New("user id is zero")
	}

	return UserMailingResult{
		userID:   userID,
		success:  success,
		errorMsg: errorMsg,
	}, nil
}

func (r UserMailingResult) UserID() UserID {
	return r.userID
}

func (r UserMailingResult) Success() bool {
	return r.success
}

func (r UserMailingResult) ErrorMessage() *string {
	return r.errorMsg
}
