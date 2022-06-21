package timer

import (
	"das_database/dao"
	"encoding/json"
	"github.com/DeAccountSystems/das-lib/common"
)

type RegisterDetail struct {
	One   int `json:"1,omitempty"`
	Two   int `json:"2,omitempty"`
	Three int `json:"3,omitempty"`
	Four  int `json:"4,omitempty"`
	Five  int `json:"5,omitempty"`
	Sub   int `json:"sub,omitempty"`
}

func (p *ParserTimer) dailyRegister(registeredAt string) {
	accountInfos, err := p.DbDao.GetAccountInfoByRegisteredAt(registeredAt)
	if err != nil {
		log.Error("GetAccountInfoByRegisteredAt err:", err.Error())
		return
	}

	registerInfo := dao.TableRegisterInfo{
		RegisterDate:    registeredAt,
		TotalAccount:    0,
		TotalSubAccount: 0,
		TotalOwner:      0,
		One:             0,
		Two:             0,
		Three:           0,
		Four:            0,
		FiveAndMore:     0,
		RegisterDetail:  "",
	}
	ownerMap := map[string]uint64{}
	registerDetail := RegisterDetail{}

	for _, accountInfo := range accountInfos {
		ownerMap[accountInfo.Owner]++
		if accountInfo.ParentAccountId != "" {
			registerInfo.TotalSubAccount++
			registerDetail.Sub = registerInfo.TotalSubAccount
			continue
		}

		registerInfo.TotalAccount++
		accLen := common.GetAccountLength(accountInfo.Account)
		switch accLen {
		case 1:
			registerInfo.One++
			registerDetail.One = registerInfo.One
		case 2:
			registerInfo.Two++
			registerDetail.Two = registerInfo.Two
		case 3:
			registerInfo.Three++
			registerDetail.Three = registerInfo.Three
		case 4:
			registerInfo.Four++
			registerDetail.Four = registerInfo.Four
		default:
			registerInfo.FiveAndMore++
			registerDetail.Five = registerInfo.FiveAndMore
		}
	}
	registerInfo.TotalOwner = len(ownerMap)
	// register detail
	b, _ := json.Marshal(registerDetail)
	registerInfo.RegisterDetail = string(b)

	if err = p.DbDao.CreateRegisterInfo(registerInfo); err != nil {
		log.Error("CreateRegisterInfo err:", err.Error())
	}
}
