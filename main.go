package main

import (
	"flag"
	"fmt"
	"strings"
	"sync"

	"github.com/google/uuid"
	"github.com/knusbaum/go9p"
	"github.com/knusbaum/go9p/fs"
	"github.com/knusbaum/go9p/proto"
)

type TokenFile struct {
	fs.BaseFile
	ownerId string
	mu      *sync.Mutex
}

var (
	FsUser  = flag.String("u", "loup", "file system root user")
	FsGroup = flag.String("g", "loup", "file system root group")
	FsPerms = flag.Uint("p", 0777, "file system root perms")
)

func newTokenFile(stat *proto.Stat) *fs.WrappedFile {
	token := TokenFile{
		BaseFile: *fs.NewBaseFile(stat),
		mu:       new(sync.Mutex),
	}
	return &fs.WrappedFile{
		File: &token,
		WriteF: func(fid uint64, offset uint64, data []byte) (uint32, error) {
			id := strings.TrimSpace(string(data))
			if id != token.ownerId {
				return uint32(len(data)), fmt.Errorf("Not the token owner: %s", id)
			}
			token.mu.Unlock()
			return uint32(len(data)), nil
		},
		ReadF: func(fid uint64, offset uint64, count uint64) ([]byte, error) {
			id := uuid.NewString()
			if offset >= uint64(len(id)) {
				return []byte{}, nil
			}
			token.mu.Lock()
			token.ownerId = id
			return []byte(id + "\n"), nil
		},
	}
}

func main() {
	flag.Parse()

	yates, root := fs.NewFS(*FsUser, *FsGroup, uint32(*FsPerms))

	// yates/new
	root.AddChild(&fs.WrappedFile{
		File: fs.NewBaseFile(yates.NewStat("new", *FsUser, *FsGroup, 0666)),
		WriteF: func(fid uint64, offset uint64, data []byte) (uint32, error) {
			name := strings.TrimSpace(string(data))
			if name == "new" || name == "del" {
				return 1, fmt.Errorf("Cannot create new token: %s", name)
			}
			if _, ok := root.Children()[name]; ok {
				return 1, fmt.Errorf("Token already exists: %s", name)
			}
			root.AddChild(newTokenFile(yates.NewStat(name, *FsUser, *FsGroup, 0666)))
			return uint32(len(data)), nil
		},
	})

	// yates/del
	root.AddChild(&fs.WrappedFile{
		File: fs.NewBaseFile(yates.NewStat("del", *FsUser, *FsGroup, 0666)),
		WriteF: func(fid uint64, offset uint64, data []byte) (uint32, error) {
			name := strings.TrimSpace(string(data))
			if name == "new" || name == "del" {
				return 1, fmt.Errorf("Cannot delete token: %s", name)
			}
			if _, ok := root.Children()[name]; !ok {
				return 1, fmt.Errorf("No such token: %s", name)
			}
			if err := root.DeleteChild(name); err != nil {
				return 1, fmt.Errorf("Cannot delete token: %s", name)
			}
			return uint32(len(data)), nil
		},
	})

	go9p.PostSrv("yates", yates.Server())
}
