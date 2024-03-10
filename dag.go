package merkledag  
  
import (  
	"crypto/sha256"  
	"encoding/binary"  
	"encoding/json"  
	"hash"  
)  
  
type Link struct {  
	Name string  
	Hash []byte  
	Size int  
}  
  
type Object struct {  
	Links []Link  
	Data  []byte  
}  
  
func Add(store KVStore, node Node, h hash.Hash) ([]byte, error) {  
	// 将分片写到KVstore中  
	// 判断数据类型  
	switch node.Type() {  
	case TYPE_FILE:  
		return StoreFile(store, node.(File), h)  
	case TYPE_DIR:  
		return StoreDir(store, node.(Dir), h)  
	default:  
		return nil, ErrUnknownNodeType  
	}  
}  
  
// 存文件的方法  
func StoreFile(store KVStore, file File, h hash.Hash) ([]byte, error) {  
	// 1. 获取文件数据  
	data := file.Bytes() // 获取数据  
	h.Write(data)        // 计算哈希值  
	hashValue := h.Sum(nil)  
  
	// 2. 将数据写入KVstore，检查错误  
	err := store.Put(hashValue, data)  
	if err != nil {  
		return nil, err  
	}  
	return hashValue, nil  
}  
  
func StoreDir(store KVStore, dir Dir, h hash.Hash) ([]byte, error) {  
	var tree Object  
	iter := dir.It() // 调用了dir目录的It方法，获取目录迭代器  
  
	// 迭代器遍历每个子节点  
	for iter.Next() {  
		childNode := iter.Node() // 调用迭代器Node方法，获取每个节点信息  
		childHash, err := Add(store, childNode, sha256.New()) // 递归调用 Add 函数存储子节点数据  
		if err != nil {  
			return nil, err  
		}  
  
		newLink := Link{  
			Name: childNode.Name(),  
			Hash: childHash,  
			Size: int(childNode.Size()),  
		}  
  
		tree.Links = append(tree.Links, newLink)  
	}  
  
	// 对整个目录节点进行序列化，计算哈希值  
	data, err := json.Marshal(tree)  
	if err != nil {  
		return nil, err  
	}  
	h.Write(data)  
	hashValue := h.Sum(nil) // 计算最终的哈希  
  
	// 将数据写入KVstore  
	err = store.Put(hashValue, data)  
	if err != nil { // 错误处理  
		return nil, err  
	}  
  
	return hashValue, nil // 返回计算得到的哈希值  
}