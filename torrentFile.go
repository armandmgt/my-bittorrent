package main

import (
	"crypto/sha1"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
)

// TorrentFile represents the torrent's general info.
type TorrentFile struct {
	Announce    string
	pieces      string
	InfoHash    [20]byte
	PieceHashes [][20]byte
	PieceLength int
	Length      int
	Name        string
}

func (i *TorrentFile) splitPieceHashes(pieces string) error {
	hashLen := sha1.Size
	buf := []byte(pieces)
	if len(buf)%hashLen != 0 {
		err := fmt.Errorf("received malformed pieces of length %d", len(buf))
		return err
	}
	numHashes := len(buf) / hashLen
	hashes := make([][20]byte, numHashes)

	for i := 0; i < numHashes; i++ {
		copy(hashes[i][:], buf[i*hashLen:(i+1)*hashLen])
	}
	i.PieceHashes = hashes
	return nil
}

func fillTorrentFile(tf *TorrentFile, bEncodeData *bEncodeValue, infoHash chan [20]byte) error {
	if bEncodeData.dict == nil {
		return errors.New("invalid torrent file")
	}
	announce, err := getDictValue(bEncodeData.dict, "announce")
	if err != nil {
		return err
	}
	tf.Announce = *announce.str

	info, err := getDictValue(bEncodeData.dict, "info")
	if err != nil {
		return err
	}
	if info.dict == nil {
		return errors.New("invalid torrent file ('info' field is not a dict)")
	}
	pieces, err := getDictValue(info.dict, "pieces")
	if err != nil {
		return err
	}
	if pieces.str == nil {
		return errors.New("invalid torrent file ('pieces' field is not a string)")
	}
	if err = tf.splitPieceHashes(*pieces.str); err != nil {
		return err
	}
	pieceLength, err := getDictValue(info.dict, "piece length")
	if err != nil {
		return err
	}
	if pieceLength.integer == nil {
		return errors.New("invalid torrent file ('piece length' field is not an int)")
	}
	tf.PieceLength = *pieceLength.integer
	length, err := getDictValue(info.dict, "length")
	if err != nil {
		return err
	}
	if length.integer == nil {
		return errors.New("invalid torrent file ('length' field is not an int)")
	}
	tf.Length = *length.integer
	name, err := getDictValue(info.dict, "name")
	if err != nil {
		return err
	}
	if name.str == nil {
		return errors.New("invalid torrent file ('name' field is not a string)")
	}
	tf.Name = *name.str

	tf.InfoHash = <-infoHash

	return nil
}

func hash(infoTokens chan string, infoHash chan [20]byte) {
	h := sha1.New()
	for token := <-infoTokens; token != ""; token = <-infoTokens {
		io.WriteString(h, token)
	}
	var hashResult [20]byte
	copy(hashResult[:], h.Sum(nil)[:20])
	infoHash <- hashResult
}

// ReadTorrent reads a file into a data structure.
func ReadTorrent(path string) (*TorrentFile, error) {
	tf := TorrentFile{}
	contents, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	tokens := make(chan string)
	infoTokens := make(chan string)
	infoHash := make(chan [20]byte)
	go tokenize(string(contents), tokens)
	go hash(infoTokens, infoHash)
	bEncodeData := bEncodeValue{}
	token := <-tokens
	data(&bEncodeData, token, tokens, infoTokens, false)

	if err := fillTorrentFile(&tf, &bEncodeData, infoHash); err != nil {
		return nil, err
	}

	return &tf, nil
}
