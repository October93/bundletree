package bundletree

import (
	"fmt"

	"github.com/google/btree"
	"github.com/october93/core/chanlog"
	"github.com/oleiade/lane"
)

const HARDCODED_BTREE_DEGREE = 4

type BundleTree struct {
	scores              *btree.BTree     //stores ItemBundles
	Deque               *lane.Deque      //stores Items
	item_to_score       map[Item]float64 //strictly keeps score for exactly the items in the tree
	deque_usage_counter map[Item]int     //how many times is item in deque?

	Logger         chanlog.Logger
	size           int
	is_capped      bool   //when false, deque is never used.
	Display_string string //Used as a prefix when displaying items.

	//special behavior options:
	//increment_deque_usage_counter_when_updating_existing_items		bool
}

type Item interface{}

type ItemBundle struct {
	score           float64
	items_in_bundle map[Item]bool
	Display_string  string
}

func (bundle *ItemBundle) Less(than btree.Item) bool {
	//IMPLEMENTING LESS BACKWARDS
	//We want to ascend from the max item down rather than the min item up.
	return bundle.score > than.(*ItemBundle).score
}

func (bundle ItemBundle) String() string {
	out_str := ""
	except_first := true
	for item := range bundle.items_in_bundle {
		if except_first {
			out_str = fmt.Sprintf("%v", bundle.displayify(item))
			except_first = false
		} else {
			out_str = fmt.Sprintf("%v, %v", out_str, bundle.displayify(item))
		}
	}
	return fmt.Sprintf("[%.3v: %v]", bundle.score, out_str)
}

func (bundle *ItemBundle) displayify(item Item) string {
	//modifies to-print items according to the tree's display string.
	disp := bundle.Display_string
	if disp == "" {
		return fmt.Sprintf("%v", item)
	} else {
		return fmt.Sprintf("<%v %v>", disp, item)
	}
}

//Instantiation
func NewBundleTree(n_memory int, logger chanlog.Logger, disp string) *BundleTree {
	//n_memory is the number of distinct items which can be stored.

	bt := BundleTree{}

	bt.Logger = logger
	bt.scores = btree.New(HARDCODED_BTREE_DEGREE)
	bt.Deque = lane.NewCappedDeque(n_memory)
	bt.item_to_score = make(map[Item]float64)
	bt.deque_usage_counter = make(map[Item]int)
	bt.is_capped = true
	bt.Display_string = disp
	//bt.increment_deque_usage_counter_when_updating_existing_items = iducwuei

	return &bt
}

func NewUncappedBundleTree(logger chanlog.Logger, disp string) *BundleTree {
	bt := NewBundleTree(-1, logger, disp)
	//lane.NewCappedDeque(-1) is equivalent to NewDeque(): there is no capacity limit.
	bt.is_capped = false

	return bt
}

func NewItemBundle(item_score float64, disp string) *ItemBundle {
	ip := ItemBundle{score: item_score}
	ip.items_in_bundle = make(map[Item]bool)
	ip.Display_string = disp

	return &ip
}

//Core functionality
func (bundletree *BundleTree) Insert_item(item Item, item_score float64) (removed_item Item) {
	L := bundletree.Logger

	bundle := bundletree.check_or_make_bundle(item_score)
	old_score, has_score := bundletree.item_to_score[item]
	is_identical := old_score == item_score

	L.Logf("tree", "Adding %v to Bundle %v\n", bundle.displayify(item), bundle)

	if has_score && !is_identical { //purge old version of same item, unless scores identical. if identical, just leave it there.
		bundletree.Remove_item(item, old_score)
	}

	if !has_score || !is_identical { //on other hand, if identical, don't increase size!
		bundletree.size++
	}

	bundle.items_in_bundle[item] = true         //bundle now tracks this item
	bundletree.item_to_score[item] = item_score //tree now tracks this item
	if bundletree.is_capped {                   //deque only works / is needed when the tree is capped.
		removed_item = bundletree.append_to_deque(item)
	}

	bundletree.Show()

	return
}

func (bundletree *BundleTree) Remove_item(item Item, item_score float64) {
	//Remove_item deletes an item from the tree, *but it may still be in the deque*
	//At time of writing this is considered a feature, not a bug
	L := bundletree.Logger
	test_bundle := NewItemBundle(item_score, bundletree.Display_string)
	check_bundle := bundletree.scores.Get(test_bundle)

	if check_bundle == nil { //no bundle at that score!
		return
		//panic(fmt.Sprintf("Tried to remove an item, but no bundle at that score: %v %v", item, item_score))
	}

	bundle_with_item := check_bundle.(*ItemBundle)

	_, check_item := bundle_with_item.items_in_bundle[item]
	if !check_item {
		return
		//panic(fmt.Sprintf("Tried to remove a nonexistent item: %v %v", item, item_score))
	}

	L.Logf("tree_remove", "Removed %v with Score %.3v\n", bundle_with_item.displayify(item), item_score)

	bundletree.size--
	delete(bundletree.item_to_score, item)          //tree stops tracking item
	delete(bundle_with_item.items_in_bundle, item)  //bundle stops tracking item
	if len(bundle_with_item.items_in_bundle) == 0 { //if no more items with this score ...
		bundletree.scores.Delete(bundle_with_item) //tree stops tracking bundle
	}
}

func (bundletree *BundleTree) Max_item() Item {
	//NOTE: Max/Min switched since Less is backwards
	check_bundle := bundletree.scores.Min()
	return pull_item(check_bundle)
}

func (bundletree *BundleTree) Min_item() Item {
	//NOTE: Max/Min switched since Less is backwards
	check_bundle := bundletree.scores.Max()
	return pull_item(check_bundle)
}

// internal use
func (bundletree *BundleTree) append_to_deque(deque_entry Item) Item {
	var oldest_item Item
	L := bundletree.Logger

	//first, add one to counter
	bundletree.deque_usage_counter[deque_entry] += 1
	L.Logf("tree_deque", "APPEND_TO_DEQUE %v\n", deque_entry)

	//next, shift the deque if it's full
	if bundletree.Deque.Full() { //need to remove an entry first
		oldest_item = bundletree.Deque.Shift() //removes oldest entry from deque
		bundletree.deque_usage_counter[oldest_item] -= 1

		if bundletree.deque_usage_counter[oldest_item] <= 0 && bundletree.Has_item(oldest_item) {
			oldest_score := bundletree.Get_score(oldest_item)
			L.Log("tree_deque", "About to remove oldest item ...")
			bundletree.Remove_item(oldest_item, oldest_score)
		}
	}

	//finally, append to deque
	bundletree.Deque.Append(deque_entry)
	return oldest_item
}

func (bundletree *BundleTree) check_or_make_bundle(item_score float64) *ItemBundle {
	bundle := NewItemBundle(item_score, bundletree.Display_string)
	check_bundle := bundletree.scores.Get(bundle)

	if check_bundle == nil { //need to make new bundle
		bundletree.scores.ReplaceOrInsert(bundle)
		return bundle
	} else { //just use the bundle we found
		return check_bundle.(*ItemBundle)
	}
}

func pull_item(bundle_container interface{}) Item {
	if bundle_container == nil { //There is no item.
		var nil_item Item
		return nil_item //just return an empty Item
	}

	bundle := bundle_container.(*ItemBundle)
	for item, _ := range bundle.items_in_bundle {
		return item //in undefined/pseudorandom fashion, grab an item from the bundle.
	}
	//should never reach here!
	return nil
	//panic("Unexpectedly encountered an empty bundle!")
}

func (bt *BundleTree) copy_deque() *lane.Deque {
	the_copy := lane.NewCappedDeque(bt.Deque.Capacity())
	the_orig := lane.NewCappedDeque(bt.Deque.Capacity())
	sz := bt.Deque.Size()
	var q Item

	for ii := 0; ii < sz; ii++ {
		q = bt.Deque.Shift()
		the_copy.Append(q)
		the_orig.Append(q)
	}
	bt.Deque = the_orig

	return the_copy
}

//Grab properties
// Needs to be edited to have an error return value for when the lookup fails. Maybe just make the map public???
func (bundletree *BundleTree) Get_score(item Item) (item_score float64) {
	item_score, ok := bundletree.item_to_score[item]
	if !ok {
		return
		//panic("Tried to get score of a nonexistent item!")
	}
	return
}

func (bundletree *BundleTree) Get_size() int {
	return bundletree.size
}

func (bundletree *BundleTree) Is_capped() bool {
	return bundletree.is_capped
}

func (bundletree *BundleTree) Has_item(item Item) bool {
	_, ok := bundletree.item_to_score[item]
	return ok
}

func (bundletree *BundleTree) In_top_items(item Item, n int) bool {
	//Returns true if:
	//a) item is among the top n items
	//b) item is in the same bundle as one of the top n items
	//meaning: we finish searching the last bundle before we return.
	var counter int
	var bundle *ItemBundle
	var found bool

	bundletree.scores.Ascend(func(i btree.Item) bool {
		bundle = i.(*ItemBundle)
		if counter >= n {
			return false
		}
		for check_item := range bundle.items_in_bundle {
			counter++
			if item == check_item {
				found = true
				return false
			}
		}
		return true
	})

	return found
}

//for displaying
func (bundletree *BundleTree) Show() {
	L := bundletree.Logger
	if L.Log_Global && L.Log_Channels["TREE_SHOW"] { //avoid all this computation if not asked for
		L.Log("tree_SHOW", "Displaying the tree:")
		//NOTE: Ascend goes from top to bottom, since Less is backwards.
		bundletree.scores.Ascend(func(i btree.Item) bool {
			L.Logf("tree_SHOW", "%v\n", i.(*ItemBundle))
			return true
		})
	}
}

func (bundletree *BundleTree) Show_deque() {
	size := bundletree.Deque.Size()
	tmp := bundletree.copy_deque()
	dsp := "Displaying the deque: "
	for x := 0; x < size; x++ {
		dsp = dsp + fmt.Sprintf("%v ", tmp.Pop())
	}
	bundletree.Logger.Log("tree_show", dsp)
}

func (bundletree *BundleTree) Items() map[Item]float64 {
	return bundletree.item_to_score
}
