package handler

import (
	"goserver/gen/pb"
)

func updateConfig(d []byte, node string) (err error) {
	save := false
	if node == "dbms" {
		save = true
	}
	msg := new(pb.UploadConfig)
	err = msg.Unmarshal(d)
	if err != nil {
		return err
	}
	switch msg.Id {
	case "0":
		err = updateHall(d, save, msg.FileName)
	case "1":
		err = updateTP(d, save)
	case "2":
		err = updateLHD(d, save)
	case "3":
		err = updateUP(d, save)
	case "4", "12":
		err = updateRM(d, save)
	case "5":
		err = updateAK47(d, save)
	case "6":
		err = updateJOKER(d, save)
	case "7":
		err = updateCRASH(d, save)
	case "8":
		err = updateAB(d, save)
	case "9":
		err = updateLottery(d, save)
	case "10":
		err = updatePLANE(d, save)
	}
	return
}
