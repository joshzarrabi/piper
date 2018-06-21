package piper

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"

	"github.com/otiai10/copy"
)

const VolumeMountPoint = "/tmp/build"

type VolumeMountBuilder struct{}

func (b VolumeMountBuilder) Build(resources []VolumeMount, inputs, outputs []string) ([]DockerVolumeMount, string, error) {
	localMountPoint, err := ioutil.TempDir("", "piperMount")
	if err != nil {
		return nil, "", err // not tested
	}

	pairsMap := make(map[string]string)

	for _, input := range inputs {
		parts := strings.Split(input, "=")
		if len(parts) != 2 {
			return nil, "", fmt.Errorf("could not parse input %q. must be of form <input-name>=<input-location>", input)
		}

		pairsMap[parts[0]] = parts[1]
	}
	pairsMap, err = copyInputVolumes(pairsMap, localMountPoint)

	for _, output := range outputs {
		parts := strings.Split(output, "=")
		if len(parts) != 2 {
			return nil, "", fmt.Errorf("could not parse output %q. must be of form <output-name>=<output-location>", output)
		}

		pairsMap[parts[0]] = parts[1]
	}
	if err != nil {
		return nil, "", err
	}

	var mounts []DockerVolumeMount
	var missingResources []string
	for _, resource := range resources {
		resourceLocation, ok := pairsMap[resource.Name]
		if !ok {
			missingResources = append(missingResources, resource.Name)
			continue
		}
		var mountPoint string
		if resource.Path == "" {
			mountPoint = filepath.Join(VolumeMountPoint, resource.Name)
		} else {
			mountPoint = filepath.Join(VolumeMountPoint, resource.Path)
		}

		mounts = append(mounts, DockerVolumeMount{
			LocalPath:  resourceLocation,
			RemotePath: filepath.Clean(mountPoint),
		})
	}
	if len(missingResources) != 0 {
		return nil, "", fmt.Errorf("The following required inputs/outputs are not satisfied: %s.", strings.Join(missingResources, ", "))
	}

	return mounts, localMountPoint, nil
}

func copyInputVolumes(pairsMap map[string]string, localMountPoint string) (map[string]string, error) {
	ephemeralLocations := make(map[string]string)
	for resourceName, resourceLocation := range pairsMap {
		ephemeralLocation := filepath.Join(localMountPoint, filepath.Base(resourceLocation))
		err := copy.Copy(resourceLocation, ephemeralLocation)
		if err != nil {
			return nil, err
		}
		ephemeralLocations[resourceName] = ephemeralLocation
	}
	return ephemeralLocations, nil
}
