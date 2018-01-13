package common

import (
	"encoding/hex"
	"fmt"
)

func (msg *NewMessage) ClientNewMessageString() *string {
	str := fmt.Sprintf("CLIENT MESSAGE contents %s", msg.Text)
	return &str
}

func (msg *NewPrivateMessage) ClientNewPrivateMessageString() *string {
	str := fmt.Sprintf("CLIENT PRIVATE MESSAGE destination %s contents %s", msg.Dest, msg.Text)
	return &str
}

func (file *NewFile) ClientNewFileString() *string {
	str := fmt.Sprintf("CLIENT FILE path %s", file.Path)
	return &str
}

func (fm *FileRequest) ClientNewFileRequestString() *string {
	metahashStr := hex.EncodeToString(fm.MetaHash)
	str := fmt.Sprintf("CLIENT FILE REQUEST filename %s metahash %s destination %s", fm.FileName, metahashStr, fm.Destination)
	return &str
}

func (fm *FileRequest) GossiperAlreadyHasFileString() *string {
	str := fmt.Sprintf("GOSSIPER ALREADY HAS FILE %s ", fm.FileName)
	return &str

}

/***** Client Notification *****/

func DownloadingMetafileNotification(filename, peername string) *string {
	str := fmt.Sprintf("DOWNLOADING metafile of %s from %s", filename, peername)
	return &str
}

func DownloadingChunkNotification(filename, peername string, chunkNb int) *string {
	str := fmt.Sprintf("DOWNLOADING %s chunk %d from %s", filename, chunkNb, peername)
	return &str
}

func ReconstructedNotification(filename string) *string {
	str := fmt.Sprintf("RECONSTRUCTED file %s", filename)
	return &str
}

func AlreadyHaveFileNotification(filename string) *string {
	str := fmt.Sprintf("FILE ALREADY PRESENT %s", filename)
	return &str
}
