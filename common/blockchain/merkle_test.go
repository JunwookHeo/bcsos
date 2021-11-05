package blockchain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewMerkleNode(t *testing.T) {
	data := [][]byte{
		[]byte("node0"),
		[]byte("node1"),
		[]byte("node2"),
		[]byte("node3"),
		[]byte("node4"),
		[]byte("node5"),
		[]byte("node6"),
		[]byte("node7"),
		[]byte("node8"),
	}

	// level 1
	mn1 := CalHashSha256(data[0])
	mn2 := CalHashSha256(data[1])
	mn3 := CalHashSha256(data[2])
	mn4 := CalHashSha256(data[3])
	mn5 := CalHashSha256(data[4])
	mn6 := CalHashSha256(data[5])
	mn7 := CalHashSha256(data[6])
	mn8 := CalHashSha256(data[7])
	mn9 := CalHashSha256(data[8])
	mn10 := CalHashSha256(data[8])

	// level 2
	mn11 := CalMerkleNodeHash(mn1, mn2)
	mn12 := CalMerkleNodeHash(mn3, mn4)
	mn13 := CalMerkleNodeHash(mn5, mn6)
	mn14 := CalMerkleNodeHash(mn7, mn8)
	mn15 := CalMerkleNodeHash(mn9, mn10)
	mn16 := CalMerkleNodeHash(mn9, mn10)

	//level 3

	mn17 := CalMerkleNodeHash(mn11, mn12)
	mn18 := CalMerkleNodeHash(mn13, mn14)
	mn19 := CalMerkleNodeHash(mn15, mn16)
	mn20 := CalMerkleNodeHash(mn15, mn16)

	//level 4

	mn21 := CalMerkleNodeHash(mn17, mn18)
	mn22 := CalMerkleNodeHash(mn19, mn20)

	//level 5

	mn23 := CalMerkleNodeHash(mn21, mn22)

	root := mn23
	root2 := CalMerkleRootHash(data)

	assert.Equal(t, root, root2, "Merkle node root has is equal")

}

func TestNewMerkleNode2(t *testing.T) {
	data := [][]byte{
		[]byte("node0"),
	}

	// level 1
	mn1 := CalHashSha256(data[0])
	mn2 := CalHashSha256(data[0])

	// level 2
	mn3 := CalMerkleNodeHash(mn1, mn2)

	root := mn3

	root2 := CalMerkleRootHash(data)

	assert.Equal(t, root, root2, "Merkle node root has is equal")
}

func TestNewMerkleNode3(t *testing.T) {
	data := [][]byte{
		[]byte("hash1"),
		[]byte("hash2"),
	}

	// level 1
	mn1 := CalHashSha256(data[0])
	mn2 := CalHashSha256(data[1])

	mn3 := CalMerkleNodeHash(mn1, mn2)

	root := mn3

	root2 := CalMerkleRootHash(data)

	assert.Equal(t, root, root2, "Merkle node root has is equal")
}
