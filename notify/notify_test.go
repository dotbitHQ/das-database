package notify

import (
	"fmt"
	"testing"
	"time"
)

func TestSendLarkTextNotify(t *testing.T) {
	// notify
	msg := "> Transaction hash：%s\n> Action：%s\n> Timestamp：%s\n> Error message：%s"
	msg = fmt.Sprintf(msg, "0x31db70391cb1b541cccbd5146e77870c095b48bc4dc8d763f222c9b0afe19424", "edit_offer", time.Now().Format("2006-01-02 15:04:05"), "ArgsToHex err...")
	err := SendLarkTextNotify(
		"https://open.larksuite.com/open-apis/bot/v2/hook/a5225cf9-7865-486e-917d-2284b0395e98",
		"DasDatabase BlockParser",
		msg,
	)
	if err != nil {
		t.Log("SendLarkTextNotify err:", err.Error())
	}
}
