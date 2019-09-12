package bundletree

import (
	"bytes"
	"encoding/gob"
	// "fmt"
	"github.com/google/btree"
	// "github.com/october93/chanlog"
	"github.com/oleiade/lane"
	"os"
	// "strconv"
)

//bundletrees can be mapped to and from memory.
func (bt *BundleTree) Write_to_file(filename string) error {
	fd, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_SYNC, 0777)
	if err != nil {
		return err
	}
	defer fd.Close()
	var fb bytes.Buffer
	enc := gob.NewEncoder(&fb)
	err = enc.Encode(bt)
	if err != nil {
		return err
	}
	_, err = fb.WriteTo(fd)
	if err != nil {
		return err
	}
	return nil
}

func (bt *BundleTree) Read_from_file(filename string) error {
	fd, err := os.OpenFile(filename, os.O_RDONLY|os.O_SYNC, 0777)
	if err != nil {
		return err
	}
	defer fd.Close()
	dec := gob.NewDecoder(fd)
	err = dec.Decode(bt)
	if err != nil {
		return err
	}
	return nil
}

//need all objects to satisfy Gob_Encoder and Gob_Decoder.
func (bt *BundleTree) GobEncode() ([]byte, error) {
	var b bytes.Buffer
	var err error
	enc := gob.NewEncoder(&b)

	err = enc.Encode(bt.size) //int
	if err != nil {
		return b.Bytes(), err
	}
	err = enc.Encode(bt.is_capped) //bool
	if err != nil {
		return b.Bytes(), err
	}
	err = enc.Encode(bt.Display_string) //string
	if err != nil {
		return b.Bytes(), err
	}
	err = enc.Encode(bt.item_to_score) //map[Item]float64 , Items are interface{}
	if err != nil {
		return b.Bytes(), err
	}
	err = enc.Encode(bt.deque_usage_counter) //map[Item]int
	if err != nil {
		return b.Bytes(), err
	}
	err = bt.Serialize_btree(enc) //int and sequence of ItemBundles (#bundles first)
	if err != nil {
		return b.Bytes(), err
	}
	err = bt.Serialize_deque(enc) //int and []Item (capacity and contents)
	if err != nil {
		return b.Bytes(), err
	}
	err = enc.Encode(bt.Logger) //Logger has all public, builtin types! Super easy.
	if err != nil {
		return b.Bytes(), err
	}

	return b.Bytes(), err
}

func (bt *BundleTree) GobDecode(data []byte) error {
	var err error
	b := bytes.NewBuffer(data)
	dec := gob.NewDecoder(b)

	err = dec.Decode(&bt.size) //int
	if err != nil {
		return err
	}
	err = dec.Decode(&bt.is_capped) //bool
	if err != nil {
		return err
	}
	err = dec.Decode(&bt.Display_string) //string
	if err != nil {
		return err
	}
	err = dec.Decode(&bt.item_to_score) //map[Item]float64
	if err != nil {
		return err
	}
	err = dec.Decode(&bt.deque_usage_counter) //map[Item]int
	if err != nil {
		return err
	}
	err = bt.Deserialize_btree(dec) //rebuilds btree.BTree
	if err != nil {
		return err
	}
	err = bt.Deserialize_deque(dec) //rebuilds lane.deque
	if err != nil {
		return err
	}
	err = dec.Decode(&bt.Logger) //rebuilds chanlog.Logger
	if err != nil {
		return err
	}
	return err
}

func (bt *BundleTree) Serialize_btree(enc *gob.Encoder) error {
	var err error

	err = enc.Encode(bt.scores.Len()) //int
	if err != nil {
		return err
	}

	bt.scores.Ascend(func(i btree.Item) bool {
		err = enc.Encode(i.(*ItemBundle))
		if err != nil {
			return false
		}
		return true
	})

	return err
}

func (bt *BundleTree) Deserialize_btree(dec *gob.Decoder) error {
	var err error
	var ib *ItemBundle

	bt.scores = btree.New(HARDCODED_BTREE_DEGREE)

	var n_bundles int
	err = dec.Decode(&n_bundles)
	if err != nil {
		return err
	}

	for ii := 0; ii < n_bundles; ii++ {
		ib = new(ItemBundle)
		err = dec.Decode(&ib)
		if err != nil {
			return err
		}
		bt.scores.ReplaceOrInsert(ib) //directly insert into btree to avoid triggering events
	}

	return err
}

func (bt *BundleTree) Serialize_deque(enc *gob.Encoder) error {
	var err error
	var dlist []Item
	d := bt.copy_deque() //create a copy to pop from

	err = enc.Encode(d.Capacity()) //int
	if err != nil {
		return err
	}

	sz := d.Size()

	for ii := 0; ii < sz; ii++ {
		dlist = append(dlist, d.Shift())
	}
	err = enc.Encode(dlist) //[]Item
	return err
}

func (bt *BundleTree) Deserialize_deque(dec *gob.Decoder) error {
	var err error
	var dcap int
	var dlist []Item

	err = dec.Decode(&dcap) //int
	if err != nil {
		return err
	}
	err = dec.Decode(&dlist) //[]Item
	if err != nil {
		return err
	}

	bt.Deque = lane.NewCappedDeque(dcap)

	for _, item := range dlist {
		bt.Deque.Append(item)
	}

	return err
}

func (ib *ItemBundle) GobEncode() ([]byte, error) {
	var b bytes.Buffer
	var err error
	enc := gob.NewEncoder(&b)

	err = enc.Encode(ib.score) //int
	if err != nil {
		return b.Bytes(), err
	}
	err = enc.Encode(ib.items_in_bundle) //map[Item]bool
	if err != nil {
		return b.Bytes(), err
	}
	err = enc.Encode(ib.Display_string) //string
	if err != nil {
		return b.Bytes(), err
	}

	return b.Bytes(), err
}

func (ib *ItemBundle) GobDecode(data []byte) error {
	b := bytes.NewBuffer(data)
	var err error
	dec := gob.NewDecoder(b)

	err = dec.Decode(&ib.score) //int
	if err != nil {
		return err
	}
	err = dec.Decode(&ib.items_in_bundle) //map[Item]bool
	if err != nil {
		return err
	}
	err = dec.Decode(&ib.Display_string) //string
	if err != nil {
		return err
	}

	return err
}
