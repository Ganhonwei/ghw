package handler

import (
	"goserver/gen/pb"
	"goserver/pkg/data"
)

func BuildPKBankInfo(p *data.User) map[string]*pb.PAKBankInfoArray {
	pkData := make(map[string]*pb.PAKBankInfoArray, 0)
	for i := len(p.PKBank) - 1; i >= 0; i-- {
		pk := p.PKBank[i]
		infos, ok := pkData[pk.BankName]
		if !ok {
			infos = &pb.PAKBankInfoArray{}
			pkData[pk.BankName] = infos
		}
		var bankType int32 = 2
		if _, ok := pkWallet[pk.BankName]; ok {
			bankType = 1 //钱包
		}
		infos.BankType = int32(bankType)
		infos.Accounts = append(infos.Accounts, &pb.PAKBankInfo{
			BankName:   pk.BankName,
			BankNumber: pk.BankAccount,
			BankHolder: pk.BankHolder,
		})
	}
	return pkData
}
