package service

import (
	"golang.org/x/net/context"
	pb "github.com/yqsy/recipes/recharge/recharge/recharge_protocol"
	"database/sql"
	"github.com/satori/go.uuid"
	"strings"
	"golang.org/x/crypto/bcrypt"
	"errors"
)

type Handler struct {
	db *sql.DB
}

func (s *Handler) Register(ctx context.Context, in *pb.RegisterRequest) (*pb.RegisterReply, error) {
	// 手机/邮箱/同时注册

	if (in.Email == "" && in.Phone == "") || in.Passwd == "" {
		return &pb.RegisterReply{Ok: false, Msg: "格式不正确"}, nil
	}

	tx, err := s.db.Begin()
	if err != nil {
		return nil, errors.New("db err")
	}

	defer tx.Rollback()

	stmt, err := tx.Prepare("INSERT s_user SET " +
		"id = ?, phone = ?,email = ?, invite_code = ?, passwd = ? ")
	if err != nil {
		return nil, errors.New("db err")
	}

	ud := strings.Replace(uuid.Must(uuid.NewV4()).String(), "-", "", -1)

	passwd, err := bcrypt.GenerateFromPassword([]byte(in.Passwd), bcrypt.DefaultCost)
	if err != nil {
		return nil, errors.New("hash err")
	}

	_, err = stmt.Exec(ud, in.Phone, in.Email, in.InviteCode, passwd)
	if err != nil {
		return &pb.RegisterReply{Ok: false, Msg: "邮箱或手机号重复"}, nil
	}

	tx.Commit()
	return &pb.RegisterReply{Ok: true, Id: ud, Msg: "注册成功"}, nil
}
