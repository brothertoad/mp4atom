package mp4atom

import (
  "encoding/binary"
  "io"
  "strconv"
  "strings"
  "github.com/brothertoad/btu"
)

type SeekableReader interface {
  io.Reader
  io.Seeker
}

type atomPart struct {
  atomType string
  count    int
}

// Atom path is of the form atomtype-count/atomtype-count/...
func FindAtomPath(reader SeekableReader, atomPath string) int64 {
  // Break the atomPath into segments.
  words := strings.Split(atomPath, "/")
  parts := make([]atomPart, len(words))
  for j, word := range(words) {
    if strings.Contains(word, "-") {
      typeAndCount := strings.Split(word, "-")
      parts[j].atomType = typeAndCount[0]
      count, err := strconv.Atoi(typeAndCount[1])
      btu.CheckError(err)
      parts[j].count = count
    } else {
      parts[j].atomType = word
      parts[j].count = 1
    }
  }
  var size int64 = 0
  for _, part := range(parts) {
    remaining := part.count
    for remaining > 0 {
      size = FindAtom(reader, part.atomType)
      if remaining == 1 || size == 0 {
        break
      }
      reader.Seek(size - 8, io.SeekCurrent)
      remaining--
    }
    // break if we didn't find this part
    if size == 0 {
      break
    }
  }
  return size
}

func FindAtom(reader SeekableReader, magic string) int64 {
  b := make([]byte, 4)
  for {
    _, err := io.ReadFull(reader, b)
    // If we are the end of the reader, we want to return 0 rather than abort,
    // so we check for EOF before calling CheckError.
    if err == io.EOF {
      return 0
    }
    btu.CheckError(err)
    size := int64(binary.BigEndian.Uint32(b))
    _, err = io.ReadFull(reader, b)
    btu.CheckError(err)
    atomType := string(b)
    if atomType == magic {
      return size
    }
    reader.Seek(size - 8, io.SeekCurrent)
  }
  return 0
}
