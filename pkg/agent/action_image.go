// +build !rt_none

package agent

import (
	"fmt"

	"arhat.dev/aranya-proto/aranyagopb"

	"arhat.dev/arhat/pkg/types"
)

func (b *Agent) handleImageList(sid uint64, data []byte) {
	cmd := new(aranyagopb.ImageListCmd)
	err := cmd.Unmarshal(data)
	if err != nil {
		b.handleRuntimeError(sid, fmt.Errorf("failed to unmarshal ImageListCmd: %w", err))
		return
	}

	ri, ok := b.runtime.(types.RuntimeImage)
	if !ok {
		b.handleUnknownCmd(sid, "image.list", cmd)
		return
	}

	b.processInNewGoroutine(sid, "image.list", func() {
		images, err := ri.ListImages(cmd)
		if err != nil {
			b.handleRuntimeError(sid, err)
			return
		}

		err = b.PostMsg(sid, aranyagopb.MSG_IMAGE_STATUS_LIST, &aranyagopb.ImageStatusListMsg{Images: images})
		if err != nil {
			b.handleConnectivityError(sid, err)
			return
		}
	})
}

func (b *Agent) handleImageEnsure(sid uint64, data []byte) {
	cmd := new(aranyagopb.ImageEnsureCmd)
	err := cmd.Unmarshal(data)
	if err != nil {
		b.handleRuntimeError(sid, fmt.Errorf("failed to unmarshal ImageEnsureCmd: %w", err))
		return
	}

	ri, ok := b.runtime.(types.RuntimeImage)
	if !ok {
		b.handleUnknownCmd(sid, "image.ensure", cmd)
		return
	}

	b.processInNewGoroutine(sid, "image.ensure", func() {
		pulledImages, err := ri.EnsureImages(cmd)
		if err != nil {
			b.handleRuntimeError(sid, err)
			return
		}

		err = b.PostMsg(sid, aranyagopb.MSG_IMAGE_STATUS_LIST, &aranyagopb.ImageStatusListMsg{Images: pulledImages})
		if err != nil {
			b.handleConnectivityError(sid, err)
			return
		}
	})
}

func (b *Agent) handleImageDelete(sid uint64, data []byte) {
	cmd := new(aranyagopb.ImageDeleteCmd)
	err := cmd.Unmarshal(data)
	if err != nil {
		b.handleRuntimeError(sid, fmt.Errorf("failed to unmarshal ImageDeleteCmd: %w", err))
		return
	}

	ri, ok := b.runtime.(types.RuntimeImage)
	if !ok {
		b.handleUnknownCmd(sid, "image.delete", cmd)
		return
	}

	b.processInNewGoroutine(sid, "image.delete", func() {
		deletedImages, err := ri.DeleteImages(cmd)
		if err != nil {
			b.handleRuntimeError(sid, err)
			return
		}

		err = b.PostMsg(sid, aranyagopb.MSG_IMAGE_STATUS_LIST, &aranyagopb.ImageStatusListMsg{Images: deletedImages})
		if err != nil {
			b.handleConnectivityError(sid, err)
			return
		}
	})
}
