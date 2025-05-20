package torrentfile

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/jackpal/bencode-go"
	"github.com/veggiedefender/torrent-client/peers"
)

type bencodeTrackerResp struct {
	Interval int    `bencode:"interval"`
	Peers    string `bencode:"peers"`
}

func (t *TorrentFile) buildTrackerUrl(peerId [20]byte, port uint16) (string, error) {
	base, err := url.Parse(t.Announce)
	if err != nil {
		return "", err
	}

	params := url.Values{
		"info_hash":  []string{hex.EncodeToString(t.InfoHash[:])}, // Identifies the file we’re trying to download
		"peer_id":    []string{hex.EncodeToString(peerId[:])},     // Identifies us to the tracker
		"port":       []string{fmt.Sprintf("%d", port)},
		"uploaded":   []string{"0"},
		"downloaded": []string{"0"},
		"compact":    []string{"1"},
		"left":       []string{fmt.Sprintf("%d", t.Length)},
	}

	base.RawQuery = params.Encode()
	return base.String(), nil
}

func (t *TorrentFile) requestPeers(peerId [20]byte, port uint16) ([]peers.Peer, error) {
	url, err := t.buildTrackerUrl(peerId, port)
	if err != nil {
		return nil, err
	}

	c := http.Client{}
	resp, err := c.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var trackerResponse bencodeTrackerResp
	err = bencode.Unmarshal(bytes.NewReader(body), &trackerResponse)
	if err != nil {
		return nil, err
	}

	return peers.Unmarshal([]byte(trackerResponse.Peers))
}
