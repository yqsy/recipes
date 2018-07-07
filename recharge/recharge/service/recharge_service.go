package service

import (
	"golang.org/x/net/context"
	pb "github.com/yqsy/recipes/recharge/recharge/recharge_protocol"
	"github.com/shopspring/decimal"
	"database/sql"
)

func (s *Handler) ZeroDecimal() decimal.Decimal {
	dec, err := decimal.NewFromString("0.0")
	if err != nil {
		panic(err)
	}
	return dec
}

func (s *Handler) IsUserIdExist(tx *sql.Tx, in *pb.RechargeRequest) bool {
	var count uint
	if err := tx.QueryRow("SELECT COUNT(*) FROM s_user WHERE id = ?",
		in.UserId).Scan(&count); err != nil {
		panic(err)
	} else {
		if count > 0 {
			return true
		} else {
			return false
		}
	}
}

func (s *Handler) InsertGoldInFlow(tx *sql.Tx, in *pb.RechargeRequest) error {
	if stmt, err := tx.Prepare("INSERT f_goldin_flow SET id = ? , user_id = ?, amount = ?"); err != nil {
		panic(err)
	} else {
		defer stmt.Close()

		_, err := stmt.Exec(in.FGoldinFlowId, in.UserId, in.Amount)
		return err
	}
}

func (s *Handler) AppendUserAmount(tx *sql.Tx, in *pb.RechargeRequest) error {
	if stmt, err := tx.Prepare("UPDATE s_user SET balance = balance + ? WHERE id = ?"); err != nil {
		panic(err)
	} else {
		defer stmt.Close()

		_, err := stmt.Exec(in.Amount, in.UserId)
		return err
	}
}

func (s *Handler) Recharge(ctx context.Context, in *pb.RechargeRequest) (*pb.RechargeReply, error) {
	if len(in.FGoldinFlowId) != 64 || len(in.UserId) != 32 {
		return &pb.RechargeReply{Ok: false, Msg: "Id错误"}, nil
	}

	price, err := decimal.NewFromString(in.Amount)
	if err != nil || price.Cmp(s.ZeroDecimal()) < 0 {
		return &pb.RechargeReply{Ok: false, Msg: "数量错误"}, nil
	}

	if tx, err := s.DB.Begin(); err != nil {
		panic(err)
	} else {
		defer tx.Rollback()

		if !s.IsUserIdExist(tx, in) {
			return &pb.RechargeReply{Ok: false, Msg: "账号不存在"}, nil
		}

		if s.InsertGoldInFlow(tx, in) != nil {
			return &pb.RechargeReply{Ok: false, Msg: "产生流水失败"}, nil
		}

		if s.AppendUserAmount(tx, in) != nil {
			return &pb.RechargeReply{Ok: false, Msg: "入金失败"}, nil
		}

		if err := tx.Commit(); err != nil {
			panic(err)
		}

		return &pb.RechargeReply{Ok: true, Msg: "入金成功"}, nil
	}
}
