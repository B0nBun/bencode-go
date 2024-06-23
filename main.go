package main

import (
	"io"
	"fmt"
	"os"
)

type FileDict struct {
	Length int      `bencode:"length"`
	Md5Sum string   `bencode:"?md5sum"`
	Path   []string `bencode:"path"`
}

type InfoDict struct {
	PieceLength int    `bencode:"piece length"`
	Pieces      string `bencode:"pieces"`
	Private     int    `bencode:"?private"`
	Name        string `bencode:"name"`

	// Single-file mode
	Length int    `bencode:"?length"`
	Md5Sum string `bencode:"?md5sum"`

	// Multiple-file mode
	Files []FileDict `bencode:"?files"`
}

type TorrentFile struct {
	Info         InfoDict   `bencode:"info"`
	Announce     string     `bencode:"announce"`
	AnnounceList [][]string `bencode:"?announce-list"`
	CreationDate int64      `bencode:"?creation date"`
	Comment      string     `bencode:"?comment"`
	CreatedBy    string     `bencode:"?created by"`
	Encoding     string     `bencode:"?encoding"`
	UrlList      []string   `bencode:"?url-list"`
}

func main() {
	args := os.Args
	var bytes []byte 
	var err error
	if len(args) < 2 {
		bytes, err = io.ReadAll(os.Stdin)
	} else {
		bytes, err = os.ReadFile(os.Args[1])
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
	tf := TorrentFile{}
	err = Unmarshal(bytes, &tf)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
	fmt.Println("Announce:", tf.Announce)
	fmt.Println("Name:", tf.Info.Name)
	fmt.Println("Created By:", tf.CreatedBy)
}