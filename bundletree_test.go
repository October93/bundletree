package bundletree

import (
	"fmt"
	"github.com/october93/core/chanlog"
	"testing"
)

func LoggerForTest() chanlog.Logger {
	L := chanlog.NewLogger(true)
	chans := make(map[string]bool)
	chans["tree_deque"] = true
	chans["tree"] = true
	chans["tree_remove"] = true
	chans["tree_show"] = true
	L.Update_channels(chans)

	return L
}

func BundleTreeForTest(capacity int, display string) BundleTree {
	_ = fmt.Println

	L := LoggerForTest()
	T := *NewBundleTree(capacity, L, display)
	return T
}

func UncappedBundleTreeForTest(display string) BundleTree {
	_ = fmt.Println

	L := LoggerForTest()
	T := *NewUncappedBundleTree(L, display)
	return T
}

func TestBundletree(t *testing.T) {
	a := BundleTreeForTest(3, "thing")

	for x := 0; x < 4; x++ {
		a.Insert_item(x, 3)
	}
	a.Show_deque()
}

func TestUncappedBundleTree(t *testing.T) {
	a := UncappedBundleTreeForTest("thing")

	for x := 0; x < 4; x++ {
		a.Insert_item(x, 3)
	}
	a.Show_deque()
}

// func TestExpirationOrder(t *testing.T) {
//     a := NewBundleTree(3, true)
//     for x := 0; x < 300; x++ {
//         a.Insert_item(x, 3)
//     }
// }

func TestScoreUpdate(t *testing.T) {
	a := BundleTreeForTest(4, "thing")
	a.Insert_item(0, 2.5)
	a.Insert_item(1, 0.3)
	if a.Get_score(0) != 2.5 || a.Get_score(1) != 0.3 {
		t.Fail()
		return
	}
	a.Insert_item(2, 0.5)
	a.Insert_item(0, 0.7)
	a.Insert_item(3, 0.7)
	if a.Get_score(0) != 0.7 || a.Get_score(1) != 0.3 || a.Get_score(2) != 0.5 || a.Get_score(3) != 0.7 {
		t.Fail()
		return
	}
	a.Show_deque()
}

func TestExpProblem(t *testing.T) {
	a := BundleTreeForTest(1, "thing")
	a.Insert_item(0, 2.5)
	a.Insert_item(0, 0.7)
	a.Show_deque()
}

func TestExpiration(t *testing.T) {
	a := BundleTreeForTest(3, "thing")
	a.Insert_item(0, 2.5)
	a.Insert_item(1, 0.3)
	if a.Get_score(0) != 2.5 || a.Get_score(1) != 0.3 {
		t.Fail()
		return
	}
	a.Insert_item(2, 0.5)
	a.Insert_item(0, 0.7)
	a.Insert_item(3, 0.1)
	if a.Get_score(2) != 0.5 || a.Get_score(3) != 0.1 {
		t.Fail()
		return
	}
	a.Show_deque()
}

func TestUpdate(t *testing.T) {
	a := BundleTreeForTest(3, "thing")
	a.Insert_item(0, 2.5)
	a.Insert_item(0, 0.7)
	a.Insert_item(0, 0.2)
	a.Insert_item(0, 0.9)
	a.Show_deque()
}

func TestMaxItem(t *testing.T) {
	n_max := 10
	a := BundleTreeForTest(3, "thing")
	for x := 0; x < n_max; x++ {
		a.Insert_item(x, float64(x)+3)
	}
	b := a.Max_item()
	if b != n_max-1 {
		t.Fail()
	}
	if a.Get_score(b) != float64(n_max+2) {
		t.Fail()
	}
	a.Show_deque()
}

func TestMinItem(t *testing.T) {
	n_max := 10
	a := BundleTreeForTest(3, "thing")
	for x := 0; x < n_max; x++ {
		a.Insert_item(x, float64(x)+3)
	}
	b := a.Min_item()
	if b != n_max-3 {
		t.Fail()
	}
	if a.Get_score(b) != float64(n_max) {
		t.Fail()
	}
	a.Show_deque()
}

func TestEmptyTree(t *testing.T) {
	a := BundleTreeForTest(3, "thing")
	_ = a.Max_item()
	a.Show_deque()
}

func TestItemGone_a(t *testing.T) {
	a := BundleTreeForTest(3, "thing")
	a.Insert_item(0, 1.0)
	a.Insert_item(1, 2.0)
	a.Insert_item(2, 4.0)
	a.Remove_item(1, 2.0)
	a.Insert_item(3, 6.0)
	a.Insert_item(4, 7.0)
	a.Insert_item(12, 9.0)
	a.Insert_item(4, 6.0)
	fmt.Printf("Has %v? %v\n", 0, a.Has_item(0))
	fmt.Printf("Has %v? %v\n", 3, a.Has_item(3))
	fmt.Printf("Has %v? %v\n", 12, a.Has_item(12))
	a.Show_deque()
	fmt.Println(a.Get_size())

	// the commented-out code below worked because Get_score() used to panic.
	// defer func() {
	// 	if r := recover(); r != nil {
	// 	}
	// }()
	// // a.Show()
	// a.Get_score(0)
	// t.Fail()

	if a.Get_score(0) != 0 {
		t.Fail()
	}
}

func TestItemGone_b(t *testing.T) {
	a := BundleTreeForTest(3, "thing")
	a.Insert_item(0, 1.0)
	a.Insert_item(1, 2.0)
	a.Insert_item(2, 4.0)
	a.Remove_item(1, 2.0)
	a.Insert_item(3, 6.0)
	a.Show_deque()
	fmt.Println(a.Get_size())

	// the commented-out code below worked because Get_score() used to panic.
	// defer func() {
	// 	if r := recover(); r != nil {
	// 	}
	// }()
	// a.Get_score(1)
	// t.Fail()

	if a.Get_score(1) != 0 {
		t.Fail()
	}
}

func TestIdenticalItems(t *testing.T) {
	a := BundleTreeForTest(3, "thing")
	a.Insert_item("x", 1.0)
	a.Insert_item("y", 2.0)
	a.Insert_item("x", 1.0)
	a.Show_deque()
}

func TestInTopItems(t *testing.T) {
	a := BundleTreeForTest(3, "thing")
	a.Insert_item("x", 1.0)
	a.Insert_item("y", 2.0)
	a.Insert_item("z", 2.0)
	if a.In_top_items("x", 2) {
		t.Fail()
	}
	if !a.In_top_items("x", 3) {
		t.Fail()
	}
	if !a.In_top_items("y", 1) {
		t.Fail()
	}
	if !a.In_top_items("z", 1) {
		t.Fail()
	}
	a.Insert_item("y", 1.5)
	if a.In_top_items("y", 1) {
		t.Fail()
	}
	if !a.In_top_items("y", 2) {
		t.Fail()
	}
	if !a.In_top_items("z", 1) {
		t.Fail()
	}
}

func TestWrite(t *testing.T) {
	a := BundleTreeForTest(3, "thing")
	a.Insert_item("x", 1.0)
	a.Insert_item("y", 2.0)
	a.Insert_item("z", 3.0)
	a.Insert_item("x", 2.0)
	err := a.Write_to_file("temp.txt")
	if err != nil {
		fmt.Println(err)
		t.Fail()
	}

	c := UncappedBundleTreeForTest("bob")
	err = c.Read_from_file("temp.txt")
	if err != nil {
		fmt.Println(err)
		t.Fail()
	}

	if c.Get_size() == 0 {
		t.Fail()
	}

	if !c.Is_capped() {
		t.Fail()
	}

	c.Show()
	c.Show_deque()
	fmt.Println(c.Get_score("x"), c.Get_score("y"), c.Get_score("z"), c.Get_size())
}
