/*
Copyright 2019 The arhat.dev Authors.

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

package libpod

import (
	"context"
	"fmt"
	"os"

	"arhat.dev/aranya-proto/aranyagopb"
	imagetypes "github.com/containers/image/v5/types"
	libpodimage "github.com/containers/libpod/v2/libpod/image"
	libpodutil "github.com/containers/libpod/v2/pkg/util"
	imagespec "github.com/opencontainers/image-spec/specs-go/v1"
)

func (r *libpodRuntime) getImageConfig(ctx context.Context, image *libpodimage.Image) (*imagespec.ImageConfig, error) {
	imageData, err := image.Inspect(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to inspect image: %v", err)
	}

	if imageData == nil {
		imageData = &image.ImageData
	}

	if imageData.Config == nil {
		imageData.Config = &imagespec.ImageConfig{}
	}

	return imageData.Config, nil
}

var pullTypeMapping = map[aranyagopb.ImagePullPolicy]libpodutil.PullType{
	aranyagopb.IMAGE_PULL_ALWAYS:         libpodutil.PullImageAlways,
	aranyagopb.IMAGE_PULL_IF_NOT_PRESENT: libpodutil.PullImageMissing,
	aranyagopb.IMAGE_PULL_NEVER:          libpodutil.PullImageNever,
}

func (r *libpodRuntime) ensureImages(images map[string]*aranyagopb.ImagePull) (map[string]*libpodimage.Image, error) {
	var (
		pulledImages = make(map[string]*libpodimage.Image)
	)

	ctx, cancel := r.ImageActionContext()
	defer cancel()

	for imageName, spec := range images {
		var dockerRegistryOptions *libpodimage.DockerRegistryOptions
		if spec.AuthConfig != nil {
			dockerRegistryOptions = &libpodimage.DockerRegistryOptions{
				DockerRegistryCreds: &imagetypes.DockerAuthConfig{
					Username: spec.AuthConfig.Username,
					Password: spec.AuthConfig.Password,
				},
				DockerCertPath:              "",
				DockerInsecureSkipTLSVerify: imagetypes.NewOptionalBool(false),
			}
		}

		image, err := r.imageClient.New(
			ctx, imageName, "", "",
			os.Stderr, dockerRegistryOptions,
			libpodimage.SigningOptions{},
			nil,
			pullTypeMapping[spec.PullPolicy],
		)
		if err != nil {
			return nil, err
		}
		pulledImages[imageName] = image
	}
	return pulledImages, nil
}
