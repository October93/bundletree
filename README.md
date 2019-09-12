#bundletree
BTrees modified for fixed capacity and multiple items per score, in Go

Data structure expanding on the functionality of BTree (btree.BTree from github.com/google/btree) in two ways:

- Fixed capacity, kicking out the oldest entry added, via double-ended queue (lane.Deque from github.com/oleiade/lane)
- "Bundles", allowing efficient storage of multiple items with identical scores in the BTree.
