package bot

import "sync"

type PhotoBuffer struct {
	sync.Mutex
	Data map[int64][]string //user ID, list of file IDs
}

func NewPhotoBuffer() *PhotoBuffer {
	return &PhotoBuffer{
		Data: make(map[int64][]string),
	}
}

func (pb *PhotoBuffer) Add(userID int64, fileID string) {
	pb.Lock()
	defer pb.Unlock()
	pb.Data[userID] = append(pb.Data[userID], fileID)
}

func (pb *PhotoBuffer) Get(userID int64) []string {
	pb.Lock()
	defer pb.Unlock()
	// return a copy to avoid race
	arr := make([]string, len(pb.Data[userID]))
	copy(arr, pb.Data[userID])
	return arr
}

func (pb *PhotoBuffer) Clear(userID int64) {
	pb.Lock()
	defer pb.Unlock()
	delete(pb.Data, userID)
}
