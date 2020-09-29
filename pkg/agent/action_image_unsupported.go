// +build rt_none

package agent

func (b *Agent) handleImageList(sid uint64, data []byte) {
	b.handleUnknownCmd(sid, "image.list", nil)
}

func (b *Agent) handleImageEnsure(sid uint64, data []byte) {
	b.handleUnknownCmd(sid, "image.ensure", nil)
}

func (b *Agent) handleImageDelete(sid uint64, data []byte) {
	b.handleUnknownCmd(sid, "image.delete", nil)
}
