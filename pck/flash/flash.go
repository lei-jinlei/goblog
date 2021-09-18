package flash

import (
	"encoding/gob"
	"goblog/pck/session"
)

type Flashes map[string]interface{}

var flashKey = "_flashes"

func init()  {
	// 在 gorilla/sessions 中存储 map 和 struct 数据需
	// 要提前注册 gob，方便后续 gob 序列化编码、解码
	gob.Register(Flashes{})
}

// Info 添加 Info 类型的消息提示
func Info(message string)  {
	addFlash("info", message)
}

func Warning(message string)  {
	addFlash("warning", message)
}

func Success(message string)  {
	addFlash("success", message)
}

func Danger(message string)  {
	addFlash("danger", message)
}

func All() Flashes {
	val := session.Get(flashKey)

	flashMessages, ok := val.(Flashes)
	if !ok {
		return nil
	}

	session.Forget(flashKey)
	return flashMessages
}

func addFlash(key string, message string)  {
	flashes := Flashes{}
	flashes[key] = message
	session.Put(flashKey, flashes)
	session.Save()
}