package notify

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

const (
	WecomAllGroup = "@all_group"
)

type WecomMessage struct {
	ChatId        string `json:"chatid,omitempty"`
	PostId        string `json:"post_id,omitempty"`
	VisibleToUser string `json:"visible_to_user,omitempty"`
	Msgtype       string `json:"msgtype,omitempty"`
	Markdown      struct {
		Content string `json:"content,omitempty"`
	} `json:"markdown,omitempty"`
}

func Wecom(chatid, boturl string) Notify {
	return func(ctx context.Context, title, msg string) error {
		message := WecomMessage{
			ChatId:        chatid,
			PostId:        "",
			VisibleToUser: "",
			Msgtype:       "markdown",
			Markdown: struct {
				Content string `json:"content,omitempty"`
			}{
				Content: fmt.Sprintf("### %s\n %s", title, msg),
			},
		}
		messageBytes, err := json.Marshal(message)
		if err != nil {
			return err
		}
		var resp *http.Response
		resp, err = http.Post(boturl, "application/json", bytes.NewBuffer(messageBytes))
		body := resp.Body
		defer body.Close()
		r := &struct {
			Errcode int    `json:"errcode"`
			Errmsg  string `json:"errmsg"`
		}{}
		if err := json.NewDecoder(resp.Body).Decode(r); err != nil {
			return err
		}
		if r.Errcode > 0 {
			err = errors.New(r.Errmsg)
		}
		return nil
	}
}
