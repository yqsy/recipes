package service

import (
	"golang.org/x/net/context"
	pb "github.com/yqsy/recipes/recharge/recharge/recharge_protocol"
	"database/sql"
	"github.com/satori/go.uuid"
	"strings"
	"golang.org/x/crypto/bcrypt"
)

type Handler struct {
	DB *sql.DB
}

func (s *Handler) IsUserExist(tx *sql.Tx, in *pb.RegisterRequest) bool {
	var count uint
	if err := tx.QueryRow(
		"SELECT COUNT(*) FROM s_user WHERE phone = ? OR email = ?",
		in.Phone, in.Email).Scan(&count); err != nil {
		panic(err)
	} else {
		if count > 0 {
			return true
		} else {
			return false
		}
	}
}

func (s *Handler) GetUUID() string {
	ud := strings.Replace(uuid.Must(uuid.NewV4()).String(), "-", "", -1)
	if len(ud) != 32 {
		panic("len(uuid) != 32")
	}
	return ud
}

func (s *Handler) InsertUser(tx *sql.Tx, in *pb.RegisterRequest, ud string) error {
	if stmt, err := tx.Prepare("INSERT s_user SET " +
		"id = ?, phone = ?,email = ?, invite_code = ?, passwd = ?, balance=0.0"); err != nil {
		panic(err)
	} else {
		defer stmt.Close()

		passwd, err := bcrypt.GenerateFromPassword([]byte(in.Passwd), bcrypt.DefaultCost)
		if err != nil || len(passwd) != 60 {
			panic("bcrypt get passwd err")
		}
		_, err = stmt.Exec(ud, in.Phone, in.Email, in.InviteCode, passwd)
		return err
	}
}

func (s *Handler) Register(ctx context.Context, in *pb.RegisterRequest) (*pb.RegisterReply, error) {
	// 手机/邮箱/同时注册
	if (in.Email == "" && in.Phone == "") || in.Passwd == "" {
		return &pb.RegisterReply{Status: Error, Msg: "格式不正确"}, nil
	}

	if tx, err := s.DB.Begin(); err != nil {
		panic(err)
	} else {
		defer tx.Rollback()

		if s.IsUserExist(tx, in) {
			return &pb.RegisterReply{Status: Error, Msg: "账号已存在"}, nil
		}

		ud := s.GetUUID()
		if s.InsertUser(tx, in, ud) != nil {
			return &pb.RegisterReply{Status: Error, Msg: "注册失败"}, nil
		}

		if err := tx.Commit(); err != nil {
			panic(err)
		}
		return &pb.RegisterReply{Status: Ok, UserId: ud, Msg: "注册成功"}, nil
	}
}
