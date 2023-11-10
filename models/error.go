package models

const (
	InvalidCredentialsErr = "user not registered or bad credentials"
	EmailAlreadyInUseErr  = "this email has been already registered"
	DbErr                 = "generic db err"
	ObjectNotFoundErr     = "the object with the requested id is not found"
)

type CoworkingErr struct {
	StatusCode int    `json:"-"`
	Code       string `json:"code"`
	Message    string `json:"message"`
}

func (c CoworkingErr) Error() string {
	return c.Message
}
