package models

import "goblog/pck/types"

type BaseModel struct {
	ID uint64
}

func (a BaseModel) GetStringID() string {
	return types.Unit64ToString(a.ID)
}
