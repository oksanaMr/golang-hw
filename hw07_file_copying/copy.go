package main

import (
	"errors"
	"io"
	"os"

	"github.com/cheggaaa/pb/v3"
)

var (
	ErrUnsupportedFile       = errors.New("unsupported file")
	ErrOffsetExceedsFileSize = errors.New("offset exceeds file size")
)

func Copy(fromPath, toPath string, offset, limit int64) error {
	srcFile, err := os.Open(fromPath)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	fileInfo, err := srcFile.Stat()
	if err != nil {
		return err
	}

	fileSize := fileInfo.Size()
	if fileSize == 0 {
		return ErrUnsupportedFile
	}

	if offset > fileSize {
		return ErrOffsetExceedsFileSize
	}

	_, err = srcFile.Seek(offset, io.SeekStart)
	if err != nil {
		return err
	}

	bytesToCopy := fileSize - offset
	if limit > 0 && limit < bytesToCopy {
		bytesToCopy = limit
	}

	dstFile, err := os.Create(toPath)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	bar := pb.Full.Start64(bytesToCopy)
	defer bar.Finish()

	var reader io.Reader = srcFile
	if limit > 0 {
		reader = io.LimitReader(srcFile, limit)
	}

	barReader := bar.NewProxyReader(reader)

	_, err = io.Copy(dstFile, barReader)
	return err
}
