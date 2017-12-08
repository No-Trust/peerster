// File Metadata data structure
package main

import (
	"bytes"
	"fmt"
	"sync"
)

type FileMetadata struct {
	Name     string
	Size     uint
	Metafile []byte
	Metahash []byte
}

func GetNumberOfChunks(metafile []byte, hashlen uint) uint {
	length := uint(len(metafile))
	hashbytelength := hashlen / 8
	return (length / hashbytelength)
}

// get the hash of the i-th chunk in fm
func (fm FileMetadata) GetChunkHash(i int, hashlen uint) []byte {
	chunkHashes := fm.ChunkHashes(hashlen)
	return chunkHashes[i]
}

// return the slice of the hashes of the chunks, in order
func (fm FileMetadata) ChunkHashes(hashlen uint) [][]byte {
	offset := int(hashlen) / 8
	r := make([][]byte, 0)
	for i := 0; i < len(fm.Metafile); i += offset {
		current := fm.Metafile[i : i+offset]
		new := make([]byte, len(current))
		copy(new, current)
		r = append(r, new)
	}

	return r
}

// Return the chunk number associated with the hash
func (fm FileMetadata) GetPositionOfChunk(hash []byte, hashlen uint) *int {
	hashes := fm.ChunkHashes(hashlen)
	for i := 0; i < len(hashes); i++ {
		if bytes.Equal(hash, hashes[i]) {
			return &i
		}
	}
	return nil
}

/***** Metadata Set *****/

type MetadataSet struct {
	metadatas []FileMetadata
	mutex     *sync.Mutex
}

func (ms *MetadataSet) Add(meta FileMetadata) {
	if ms.Contains(meta) {
		// stays a set
		return
	}
	ms.mutex.Lock()
	ms.metadatas = append(ms.metadatas, meta)
	ms.mutex.Unlock()
}

// get the FileMetadata that has the given hash, nil if not present
func (ms *MetadataSet) Get(hash []byte) *FileMetadata {
	ms.mutex.Lock()
	for _, fm := range ms.metadatas {
		if bytes.Equal(fm.Metahash, hash) {
			fmt.Println("WTF : ", fm.Name, "\nmetahash : ", fm.Metahash, "\nhash : ", hash)
			ms.mutex.Unlock()
			return &fm
		}
	}
	ms.mutex.Unlock()
	return nil
}

// return the FileMetadatas with given filename
func (ms *MetadataSet) GetByName(filename string) []FileMetadata {
	r := make([]FileMetadata, 0)
	ms.mutex.Lock()
	for _, fm := range ms.metadatas {
		if fm.Name == filename {
			r = append(r, fm)
		}
	}
	ms.mutex.Unlock()
	return r
}

//
func (ms *MetadataSet) Contains(meta FileMetadata) bool {
	// Assumption : metadata1 == metadata2 if and only if matadata1.Metahash = matadata2.Metahash
	ms.mutex.Lock()
	for _, m := range ms.metadatas {
		if bytes.Equal(meta.Metahash, m.Metahash) {
			ms.mutex.Unlock()
			return true
		}
	}
	ms.mutex.Unlock()
	return false
}

func NewMetadataSet() MetadataSet {
	return MetadataSet{
		metadatas: make([]FileMetadata, 0),
		mutex:     &sync.Mutex{},
	}
}
