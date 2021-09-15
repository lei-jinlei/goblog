package article

import (
	"goblog/pck/route"
	"strconv"
)

// Article 文章模型
type Article struct {
	ID    int64
	Title string
	Body  string
}

// Link 方法用来生成文章链接
func (article Article) Link() string {
	return route.Name2URL("articles.show", "id", strconv.FormatInt(article.ID, 10))
}