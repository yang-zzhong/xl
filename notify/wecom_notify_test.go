package notify

import (
	"context"
	"fmt"
	"testing"
)

func TestWecom_Notify(t *testing.T) {
	noti := Wecom(WecomAllGroup, "https://qyapi.weixin.qq.com/cgi-bin/webhook/send?key=9dd10b2b-0cc7-47bd-b566-ad53c15388d7")
	err := noti.Notify(context.Background(), "裁撤看板有任务失败阻塞", fmt.Sprintf("处理事实表第[%d]页出错，重试[%d]次未恢复。错误信息: %s", 3, 20, "dial tcp: lookup bj-cdb-attdt9ai.sql.tencentcdb.com: no such host"))
	if err != nil {
		t.Fatal(err)
	}
}
