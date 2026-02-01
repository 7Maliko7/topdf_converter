package bot

import "sync"

type photoBuffer struct {
	sync.Mutex
	data map[int64][]string //user ID, list of file IDs
}

func newPhotoBuffer() *photoBuffer {
	return &photoBuffer{
		data: make(map[int64][]string),
	}
}

func (pb *photoBuffer) add(userID int64, fileID string) {
	pb.Lock()
	defer pb.Unlock()
	pb.data[userID] = append(pb.data[userID], fileID)
}

func (pb *photoBuffer) get(userID int64) []string {
	pb.Lock()
	defer pb.Unlock()
	// return a copy to avoid race
	arr := make([]string, len(pb.data[userID]))
	copy(arr, pb.data[userID])
	return arr
}

func (pb *photoBuffer) clear(userID int64) {
	pb.Lock()
	defer pb.Unlock()
	delete(pb.data, userID)
}
