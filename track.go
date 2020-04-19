package main

type Track struct {
	artist      string
	albumArtist string
	album       string
	title       string
	trackNumber int
}

func (t *Track) isDifferent(o *Track) bool {
	return o == nil ||
		t.artist != o.artist ||
		t.albumArtist != o.albumArtist ||
		t.album != o.album ||
		t.title != o.title ||
		t.trackNumber != o.trackNumber
}
