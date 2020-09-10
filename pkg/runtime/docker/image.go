/*
Copyright 2020 The arhat.dev Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package docker

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"arhat.dev/aranya-proto/aranyagopb"
	dockertype "github.com/docker/docker/api/types"
	dockerfilter "github.com/docker/docker/api/types/filters"
	dockermessage "github.com/docker/docker/pkg/jsonmessage"
)

func (r *dockerRuntime) ensureImages(
	images map[string]*aranyagopb.ImagePull,
) (map[string]*dockertype.ImageSummary, error) {
	var (
		pulledImages = make(map[string]*dockertype.ImageSummary)
		imagesToPull []string
	)

	pullCtx, cancelPull := r.ImageActionContext()
	defer cancelPull()

	for imageName, spec := range images {
		if spec.PullPolicy == aranyagopb.IMAGE_PULL_ALWAYS {
			imagesToPull = append(imagesToPull, imageName)
			continue
		}

		image, err := r.getImage(pullCtx, imageName)
		if err == nil {
			// image exists
			switch spec.PullPolicy {
			case aranyagopb.IMAGE_PULL_NEVER, aranyagopb.IMAGE_PULL_IF_NOT_PRESENT:
				pulledImages[imageName] = image
			}
		} else {
			// image does not exist
			switch spec.PullPolicy {
			case aranyagopb.IMAGE_PULL_NEVER:
				return nil, fmt.Errorf("failed to ensure image [%s]: %w", imageName, err)
			case aranyagopb.IMAGE_PULL_IF_NOT_PRESENT:
				imagesToPull = append(imagesToPull, imageName)
			}
		}
	}

	for _, imageName := range imagesToPull {
		authStr := ""
		if spec, ok := images[imageName]; ok && spec.AuthConfig != nil {
			authCfg := dockertype.AuthConfig{
				Username:      spec.AuthConfig.Username,
				Password:      spec.AuthConfig.Password,
				ServerAddress: spec.AuthConfig.ServerAddress,
				IdentityToken: spec.AuthConfig.IdentityToken,
				RegistryToken: spec.AuthConfig.RegistryToken,
			}
			encodedJSON, err := json.Marshal(authCfg)
			if err != nil {
				return nil, fmt.Errorf("unable to encode auth config json: %w", err)
			}

			authStr = base64.URLEncoding.EncodeToString(encodedJSON)
		}

		out, err := r.imageClient.ImagePull(pullCtx, imageName, dockertype.ImagePullOptions{
			RegistryAuth: authStr,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to pull image [%s]: %w", imageName, err)
		}
		err = func() error {
			defer func() { _ = out.Close() }()
			decoder := json.NewDecoder(out)
			for {
				var msg dockermessage.JSONMessage
				err = decoder.Decode(&msg)
				if err == io.EOF {
					break
				}
				if err != nil {
					return err
				}
				if msg.Error != nil {
					return msg.Error
				}
			}
			return nil
		}()
		if err != nil {
			return nil, fmt.Errorf("failed to decode output: %w", err)
		}

		image, err := r.getImage(pullCtx, imageName)
		if err != nil {
			return nil, fmt.Errorf("failed to get pulled image: %w", err)
		}
		pulledImages[imageName] = image
	}

	return pulledImages, nil
}

func (r *dockerRuntime) getImage(ctx context.Context, imageName string) (*dockertype.ImageSummary, error) {
	imageName = strings.TrimPrefix(imageName, "docker.io/")
	imageList, err := r.imageClient.ImageList(ctx, dockertype.ImageListOptions{
		Filters: dockerfilter.NewArgs(dockerfilter.Arg("reference", imageName)),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list images: %v", err)
	}

	if len(imageList) == 0 {
		return nil, fmt.Errorf("image [%s] not found", imageName)
	}

	return &imageList[0], nil
}
