package libs

import (
	"bytes"
	"fmt"
	"math"
	"strings"
)

type Pager struct {
	Page     int
	Totalnum int
	Pagesize int
	urlpath  string
	urlquery string
	nopath   bool
}

func NewPager(page, totalnum, pagesize int, url string, nopath ...bool) *Pager {
	p := new(Pager)
	p.Page = page
	p.Totalnum = totalnum
	p.Pagesize = pagesize

	arr := strings.Split(url, "?")
	p.urlpath = arr[0]
	if len(arr) > 1 {
		p.urlquery = "?" + arr[1]
	} else {
		p.urlquery = ""
	}

	if len(nopath) > 0 {
		p.nopath = nopath[0]
	} else {
		p.nopath = false
	}

	return p
}

func (this *Pager) url(page int) string {
	if this.nopath { //不使用目录形式
		if this.urlquery != "" {
			return fmt.Sprintf("%s%s&page=%d", this.urlpath, this.urlquery, page)
		} else {
			return fmt.Sprintf("%s?page=%d", this.urlpath, page)
		}
	} else {
		return fmt.Sprintf("%s/page/%d%s", this.urlpath, page, this.urlquery)
	}
}

func (this *Pager) ToString() string {
	var buf bytes.Buffer
	buf.WriteString("<ul class=\"pagination-total\">")
	buf.WriteString(fmt.Sprintf("<li>共<span>%d</span>条</li>", int(this.Totalnum)))
	buf.WriteString("</ul>")
	buf.WriteString("<ul class=\"pagination-jump\">")
	buf.WriteString("<li><select id=\"pagesize_select\">")
	buf.WriteString("<option value=\"20\"")
	if this.Pagesize == 20 {
		buf.WriteString(" selected ")
	}
	buf.WriteString(">20条</option>")
	buf.WriteString("<option value=\"40\"")
	if this.Pagesize == 40 {
		buf.WriteString(" selected ")
	}
	buf.WriteString(">40条</option>")
	buf.WriteString("<option value=\"40\"")
	if this.Pagesize == 50 {
		buf.WriteString(" selected ")
	}
	buf.WriteString(">50条</option>")
	buf.WriteString("<option value=\"100\"")
	if this.Pagesize == 100 {
		buf.WriteString(" selected ")
	}
	buf.WriteString(">100条</option>")
	buf.WriteString("<option value=\"200\"")
	if this.Pagesize == 200 {
		buf.WriteString(" selected ")
	}
	buf.WriteString(">200条</option>")
	buf.WriteString("</select></li>")
	buf.WriteString("</ul>")
	if this.Totalnum <= this.Pagesize {
		return buf.String()
	}

	var from, to, linknum, offset, totalpage int

	offset = 5
	linknum = 10

	totalpage = int(math.Ceil(float64(this.Totalnum) / float64(this.Pagesize)))

	if totalpage < linknum {
		from = 1
		to = totalpage
	} else {
		from = this.Page - offset
		to = from + linknum
		if from < 1 {
			from = 1
			to = from + linknum - 1
		} else if to > totalpage {
			to = totalpage
			from = totalpage - linknum + 1
		}
	}

	buf.WriteString("<ul class=\"pagination\">")
	if this.Page > 1 {
		buf.WriteString(fmt.Sprintf("<li><a href=\"%s\">&laquo;</a></li>", this.url(this.Page-1)))
	} else {
		buf.WriteString("<li class=\"disabled\"><span>&laquo;</span></li>")
	}

	if this.Page > linknum {
		buf.WriteString(fmt.Sprintf("<li><a href=\"%s\">1...</a></li>", this.url(1)))
	}

	for i := from; i <= to; i++ {
		if i == this.Page {
			buf.WriteString(fmt.Sprintf("<li class=\"active\"><span>%d</span></li>", i))
		} else {
			buf.WriteString(fmt.Sprintf("<li><a href=\"%s\">%d</a></li>", this.url(i), i))
		}
	}

	if totalpage > to {
		buf.WriteString(fmt.Sprintf("<li><a href=\"%s\">...%d</a></li>", this.url(totalpage), totalpage))
	}

	if this.Page < totalpage {
		buf.WriteString(fmt.Sprintf("<li><a href=\"%s\">&raquo;</a></li>", this.url(this.Page+1)))
	} else {
		buf.WriteString(fmt.Sprintf("<li class=\"disabled\"><span>&raquo;</span></li>"))
	}
	// buf.WriteString(fmt.Sprintf("<li class=\"pagination-jum\">总共:<span>%d</span>条</li>", int(this.Totalnum)))
	// // buf.WriteString(fmt.Sprintf("<li class=\"pagination-jump\">总共：<span>%s</span>条</li>", count))
	// buf.WriteString(fmt.Sprintf("<li class=\"pagination-jump\">前往<input id=\"input_page\" type=\"text\" class=\"page-input\" value=\"%s\" />页&nbsp;<a href=\"%s\">Go</a></li>", ""))

	buf.WriteString("</ul>")

	buf.WriteString("<ul class=\"pagination-jump\">")
	buf.WriteString(fmt.Sprintf("<li><input id=\"total_page\" type=\"hidden\" class=\"page-input\" value=\"%d\" /></li>", totalpage))
	buf.WriteString(fmt.Sprintf("<li>前往<input id=\"input_page\" type=\"text\" class=\"page-input\" value=\"%s\" />页&nbsp;<a onclick=\"toPage(this)\">Go</a></li>", ""))
	buf.WriteString("</ul>")
	return buf.String()
}
