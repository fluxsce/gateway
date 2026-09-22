package dao

import "testing"

func TestLogPageTotal(t *testing.T) {
	limit := int64(20)
	skip := int64(20)

	total, ok := logPageTotal(&limit, nil, 0)
	if !ok || total != 0 {
		t.Fatalf("第一页 0 行应直接得到总数 0，得到 %d ok=%v", total, ok)
	}

	total, ok = logPageTotal(&limit, nil, 7)
	if !ok || total != 7 {
		t.Fatalf("第一页不满应得到 7，得到 %d ok=%v", total, ok)
	}

	total, ok = logPageTotal(&limit, nil, 20)
	if ok {
		t.Fatalf("满页不能省略计数，得到 %d", total)
	}

	total, ok = logPageTotal(&limit, &skip, 5)
	if !ok || total != 25 {
		t.Fatalf("末页不满应为 skip+行数，得到 %d ok=%v", total, ok)
	}

	total, ok = logPageTotal(&limit, &skip, 0)
	if ok {
		t.Fatalf("翻过末页的空白页不能把 skip 当成总数，得到 %d", total)
	}
}
