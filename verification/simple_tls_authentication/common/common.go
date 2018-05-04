package common

type LoginReq struct {
	User string
	Password string
}


type LoginRes struct {
	LoginResult bool
}
